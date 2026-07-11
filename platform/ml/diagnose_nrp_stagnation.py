"""Diagnose NRP learned stagnation policy — checkpoint vs trajectory false stops."""

from __future__ import annotations

import json
from pathlib import Path

import pandas as pd

from policy_training_utils import (
    enrich_search_features,
    infer_instance_from_run_id,
    predict_row_stop,
)
from trajectory_training import enrich_trajectory_features
from policy_validation import (
    ex_post_should_stop,
    find_stagnation_classifier,
    load_search_assist_data,
    rule_would_stop,
)


def _traj_classifiers(model: dict) -> dict[tuple[str, str, str], dict]:
    block = model.get("trajectory", {})
    out: dict[tuple[str, str, str], dict] = {}
    for c in block.get("classifiers", []):
        out[(c["domain"], c.get("algorithm", ""), c.get("instance", ""))] = c
    return out


def _lookup(clf_map: dict, domain: str, algo: str, inst: str) -> dict | None:
    return (
        clf_map.get((domain, algo, inst))
        or clf_map.get((domain, algo, ""))
        or clf_map.get((domain, "", ""))
        or clf_map.get((domain, "", inst))
    )


def traj_stop_row(row: pd.Series, traj_map: dict, traj_df: pd.DataFrame, idx: int) -> tuple[bool, float]:
    inst = infer_instance_from_run_id(str(row.get("run_id", "")))
    algo = str(row.get("algorithm", "sa"))
    clf = _lookup(traj_map, "nrp", algo, inst)
    if not clf or not clf.get("tree") or idx >= len(traj_df):
        return False, 0.0
    return predict_row_stop(clf["tree"], traj_df.iloc[idx])


def ckpt_stop_row(row: pd.Series, model: dict, ckpt_df: pd.DataFrame, idx: int) -> tuple[bool, float]:
    inst = infer_instance_from_run_id(str(row.get("run_id", "")))
    algo = str(row.get("algorithm", "sa"))
    clf = find_stagnation_classifier(model, "nrp", algo, inst)
    if not clf or not clf.get("tree") or idx >= len(ckpt_df):
        return False, 0.0
    return predict_row_stop(clf["tree"], ckpt_df.iloc[idx])


def main() -> None:
    policy_dir = Path(__file__).parent / "policies"
    data_dir = Path(__file__).parent.parent / "web" / "pfrs-lab" / "data" / "runs"

    with open(policy_dir / "stagnation_policy.json") as f:
        model = json.load(f)
    traj_map = _traj_classifiers(model)

    df = load_search_assist_data(data_dir)
    sa = df[df["run_id"].astype(str).str.contains("nrp", case=False, na=False)]
    sa = sa[sa["run_id"].astype(str).str.contains("-sa-", case=False, na=False)].copy()
    if "algorithm" not in sa.columns:
        sa["algorithm"] = "sa"

    sa_enriched = enrich_search_features(sa.copy())
    sa_traj = enrich_trajectory_features(sa.copy())
    sa_enriched = sa_enriched.reset_index(drop=True)
    sa_traj = sa_traj.reset_index(drop=True)

    stats = {
        "checkpoint": {"stops": 0, "false": 0},
        "trajectory": {"stops": 0, "false": 0},
        "rules": {"stops": 0, "false": 0},
    }
    disagree_traj_ckpt = 0
    false_examples: list[str] = []

    for idx, (_, row) in enumerate(sa.iterrows()):
        should = ex_post_should_stop(row)
        rs, _, _ = rule_would_stop(row)
        cs, cp = ckpt_stop_row(row, model, sa_enriched, idx)
        ts, tp = traj_stop_row(row, traj_map, sa_traj, idx)

        for name, stop in [("rules", rs), ("checkpoint", cs), ("trajectory", ts)]:
            if stop:
                stats[name]["stops"] += 1
                if not should:
                    stats[name]["false"] += 1

        if ts != cs:
            disagree_traj_ckpt += 1
        if ts and not should and len(false_examples) < 5:
            best = float(row.get("best_penalty", 0) or 0)
            final = float(row.get("final_best_penalty", best) or best)
            false_examples.append(
                f"  {row['run_id']} budget={row.get('candidates', 0)}/"
                f"{row.get('iterations_total', 0)} imp_after={best - final:.0f} "
                f"traj_p={tp:.3f} ckpt_p={cp:.3f}"
            )

    n = len(sa)
    print(f"NRP SA diagnostic ({n} checkpoints, {sa['run_id'].nunique()} runs)")
    print()
    for name in ("rules", "checkpoint", "trajectory"):
        s = stats[name]
        rate = s["false"] / s["stops"] if s["stops"] else 0.0
        gate = "PASS" if rate <= 0.05 else "FAIL"
        print(
            f"  {name:12} stops={s['stops']:5d}  false={s['false']:4d}  "
            f"false_rate={rate:.4f}  [{gate}]"
        )
    print(f"\n  trajectory vs checkpoint disagree: {disagree_traj_ckpt} ({100*disagree_traj_ckpt/n:.1f}%)")
    print("\n  Runtime resolver matches Go: neural→trajectory→checkpoint.")
    print("  Trajectory tier requires promotion_ready on each classifier.")
    if false_examples:
        print("\n  Sample trajectory false stops:")
        print("\n".join(false_examples))

    # n012w8 classifier metadata
    inst = "n012w8"
    ck = find_stagnation_classifier(model, "nrp", "sa", inst)
    tr = _lookup(traj_map, "nrp", "sa", inst)
    print("\n  n012w8/sa classifiers:")
    if ck:
        print(f"    checkpoint  cv={ck.get('cv_mean')} positive_rate={ck.get('positive_rate')}")
    if tr:
        print(f"    trajectory  cv={tr.get('cv_mean')} positive_rate={tr.get('positive_rate')}")


if __name__ == "__main__":
    main()
