#!/usr/bin/env python3
"""Step 1 ML harness — compare rules vs hybrid vs learned from run.json telemetry."""

from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path

from policy_validation import detect_domain


def parse_args():
    p = argparse.ArgumentParser(description="Compare SI policy modes (Step 1 harness)")
    p.add_argument("--data-dir", required=True, help="Path to data/runs")
    p.add_argument("--prefix", default="val-", help="Run label prefix")
    p.add_argument("--json-out", default="", help="Write JSON report path")
    p.add_argument("--ml-maturity", type=float, default=4.0)
    return p.parse_args()


def infer_policy(label: str, meta: dict) -> str:
    for key in ("policyMode", "policy_mode"):
        if meta.get(key):
            return str(meta[key])
    for mode in ("rules", "hybrid", "learned"):
        if f"-{mode}-" in label:
            return mode
    return "unknown"


def infer_algorithm(label: str, meta: dict) -> str:
    if meta.get("mode"):
        return str(meta["mode"])
    for alg in ("portfolio", "tabu", "lahc", "sa"):
        if f"-{alg}-" in label or label.endswith(f"-{alg}"):
            return alg
    return "unknown"


def objective(meta: dict) -> int:
    for key in ("bestObjective", "bestDistance", "bestPenalty", "totalPenalty", "objective"):
        if key in meta and meta[key] is not None:
            return int(meta[key])
    return 0


def load_results(data_dir: Path, prefix: str) -> list[dict]:
    rows = []
    for run_dir in sorted(data_dir.iterdir()):
        if not run_dir.is_dir():
            continue
        label = run_dir.name
        if prefix and not label.startswith(prefix):
            continue
        run_json = run_dir / "run.json"
        if not run_json.exists():
            continue
        try:
            meta = json.loads(run_json.read_text(encoding="utf-8"))
        except Exception:
            continue
        rows.append({
            "label": label,
            "domain": meta.get("problemType") or detect_domain(label),
            "instance": meta.get("instance", ""),
            "algorithm": infer_algorithm(label, meta),
            "policyMode": infer_policy(label, meta),
            "objective": objective(meta),
            "runtimeMs": int(meta.get("runtimeMs") or meta.get("runtime_ms") or 0),
            "feasible": bool(meta.get("feasible", True)),
            "seed": int(meta.get("seed") or 0),
        })
    return rows


def mean(xs: list[float]) -> float:
    return sum(xs) / len(xs) if xs else 0.0


def build_report(rows: list[dict], prefix: str, ml_maturity: float) -> dict:
    by_mode: dict[str, list] = {}
    for r in rows:
        by_mode.setdefault(r["policyMode"], []).append(r)

    mode_summaries = []
    for mode, group in sorted(by_mode.items()):
        objs = [r["objective"] for r in group]
        rts = [r["runtimeMs"] for r in group]
        feas = sum(1 for r in group if r["feasible"])
        mode_summaries.append({
            "policyMode": mode,
            "n": len(group),
            "meanObjective": round(mean(objs), 2),
            "meanRuntimeMs": round(mean(rts), 1),
            "feasibilityRate": round(feas / len(group), 4) if group else 0,
        })

    comparisons = []
    quality_wins = 0
    runtime_wins = 0
    by_cfg: dict[str, dict[str, list]] = {}
    for r in rows:
        key = f"{r['domain']}|{r['instance']}|{r['algorithm']}"
        by_cfg.setdefault(key, {}).setdefault(r["policyMode"], []).append(r)

    for modes in by_cfg.values():
        rules = modes.get("rules", [])
        if not rules:
            continue
        for mode_b in ("hybrid", "learned"):
            group_b = modes.get(mode_b, [])
            if not group_b:
                continue
            mean_a = mean([r["objective"] for r in rules])
            mean_b = mean([r["objective"] for r in group_b])
            rt_a = mean([r["runtimeMs"] for r in rules])
            rt_b = mean([r["runtimeMs"] for r in group_b])
            saved = ((rt_a - rt_b) / rt_a * 100) if rt_a else 0
            delta = mean_b - mean_a
            verdict = "equivalent"
            if delta < -0.01:
                verdict = "better"
            elif delta > 0.01:
                verdict = "worse"
            rt_verdict = "faster" if saved > 2 else ("slower" if saved < -2 else "equivalent")
            roi = ((mean_a - mean_b) / abs(mean_a) * 100 if mean_a else 0) + saved
            if verdict == "worse":
                roi = -abs(roi)
            comparisons.append({
                "domain": rules[0]["domain"],
                "instance": rules[0]["instance"],
                "algorithm": rules[0]["algorithm"],
                "modeA": "rules",
                "modeB": mode_b,
                "meanObjectiveA": round(mean_a, 2),
                "meanObjectiveB": round(mean_b, 2),
                "objectiveDelta": round(delta, 2),
                "meanRuntimeA": round(rt_a, 1),
                "meanRuntimeB": round(rt_b, 1),
                "runtimeSavedPct": round(saved, 2),
                "verdict": verdict,
                "runtimeVerdict": rt_verdict,
                "roi": round(roi, 2),
            })
            if verdict in ("better", "equivalent"):
                quality_wins += 1
            if rt_verdict == "faster" and verdict != "worse":
                runtime_wins += 1

    return {
        "generatedAt": datetime.now(timezone.utc).isoformat(),
        "step": 1,
        "mlMaturity": ml_maturity,
        "totalRuns": len(rows),
        "prefix": prefix,
        "modeSummaries": mode_summaries,
        "comparisons": comparisons,
        "gates": {
            "step1HarnessOk": len(rows) > 0,
            "step2QualityWins": quality_wins,
            "step2RuntimeWins": runtime_wins,
            "step2PromoteOk": quality_wins >= 2 or runtime_wins >= 2,
        },
    }


def main():
    args = parse_args()
    data_dir = Path(args.data_dir)
    if not data_dir.exists():
        raise SystemExit(f"data dir not found: {data_dir}")

    rows = load_results(data_dir, args.prefix)
    report = build_report(rows, args.prefix, args.ml_maturity)

    print(f"ML harness — {report['totalRuns']} runs (maturity {report['mlMaturity']}/10)")
    for s in report["modeSummaries"]:
        print(f"  {s['policyMode']:8s} n={s['n']:3d}  obj={s['meanObjective']:.1f}  runtime={s['meanRuntimeMs']:.0f}ms")
    print(f"  Step 2 gate: {report['gates']}")

    if args.json_out:
        out = Path(args.json_out)
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(json.dumps(report, indent=2), encoding="utf-8")
        print(f"Wrote {out}")


if __name__ == "__main__":
    main()
