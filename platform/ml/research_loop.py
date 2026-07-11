"""
Step 8 — Human-in-the-loop closed research loop.

Reads validation, registry, and harness signals; proposes experiments that
require explicit human approval before execution.
"""

from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path
from typing import Any

POLICY_DIR = Path(__file__).resolve().parent / "policies"
GO_CWD = "platform/go"
POLICY_DIR_REL = "../ml/policies"
MAX_PROPOSALS = 8
VERIFY_SEED = 99


def _read_json(path: Path) -> dict[str, Any] | None:
    if not path.exists():
        return None
    try:
        with open(path) as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return None


def _harness_paths(repo_root: Path) -> list[Path]:
    return [
        repo_root / "docs" / "reports" / "ml-harness" / "latest.json",
        repo_root / "platform" / "web" / "pfrs-lab" / "data" / "ml-harness-latest.json",
    ]


def _load_harness(repo_root: Path) -> dict[str, Any] | None:
    for path in _harness_paths(repo_root):
        data = _read_json(path)
        if data:
            return data
    return None


def _slug(*parts: str) -> str:
    raw = "-".join(p for p in parts if p)
    return "".join(c if c.isalnum() or c in "-_" else "-" for c in raw.lower())


def _command_nrp(mode: str, policy: str, seed: int, label: str) -> str:
    return (
        f"go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode {mode} "
        f"--pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 "
        f"--seeds {seed} --worker-decision-mode assist --policy-mode {policy} "
        f"--policy-dir {POLICY_DIR_REL} --pfrs-run-label {label} --storage s3"
    )


def _command_cvrp(mode: str, policy: str, seed: int, label: str, instance_file: str, slug: str) -> str:
    return (
        f"go run ./cmd/owp solve cvrp --instance {instance_file} --mode {mode} "
        f"--iterations 500000 --policy-mode {policy} --policy-dir {POLICY_DIR_REL} "
        f"--seed {seed} --run-label {label} --storage s3"
    )


def _command_jss(mode: str, policy: str, seed: int, label: str) -> str:
    return (
        f"go run ./cmd/owp solve jobshop "
        f"--instance internal/infrastructure/jobshop/testdata/la01.txt --mode {mode} "
        f"--iterations 100000 --policy-mode {policy} --policy-dir {POLICY_DIR_REL} "
        f"--seed {seed} --run-label {label} --storage s3"
    )


def _command_vrptw(mode: str, policy: str, seed: int, label: str) -> str:
    return (
        f"go run ./cmd/owp solve vrptw --instance ../../examples/vrptw/C101.txt --mode {mode} "
        f"--iterations 100000 --policy-mode {policy} --policy-dir {POLICY_DIR_REL} "
        f"--seed {seed} --run-label {label} --storage s3"
    )


def build_experiment_command(
    domain: str,
    algorithm: str,
    policy_mode: str,
    seed: int,
    label: str,
    instance: str = "",
) -> str:
    """Build an owp command (run from platform/go)."""
    domain = domain.lower()
    algorithm = algorithm.lower()
    if domain == "nrp":
        return _command_nrp(algorithm, policy_mode, seed, label)
    if domain == "cvrp":
        inst = instance.lower() if instance else "a32k5"
        if inst in ("a80k10", "a-n80-k10"):
            vrp = "../../examples/cvrp/A-n80-k10.vrp"
            slug = "a80k10"
        else:
            vrp = "../../examples/cvrp/A-n32-k5.vrp"
            slug = "a32k5"
        return _command_cvrp(algorithm, policy_mode, seed, label, vrp, slug)
    if domain in ("jss", "jobshop"):
        return _command_jss(algorithm, policy_mode, seed, label)
    if domain == "vrptw":
        return _command_vrptw(algorithm, policy_mode, seed, label)
    return _command_nrp(algorithm, policy_mode, seed, label)


def _proposal(
    *,
    proposal_type: str,
    priority: int,
    rationale: str,
    domain: str,
    algorithm: str,
    policy_mode: str,
    seed: int,
    instance: str = "",
    extra: dict[str, Any] | None = None,
) -> dict[str, Any]:
    label = f"ml-exp-{_slug(domain, instance or 'all', algorithm, policy_mode, f's{seed}')}"
    pid = _slug(proposal_type, domain, instance, algorithm, policy_mode)
    cmd = build_experiment_command(domain, algorithm, policy_mode, seed, label, instance)
    item: dict[str, Any] = {
        "id": pid,
        "type": proposal_type,
        "priority": priority,
        "rationale": rationale,
        "requires_approval": True,
        "status": "proposed",
        "domain": domain,
        "algorithm": algorithm,
        "instance": instance or None,
        "policy_mode": policy_mode,
        "seed": seed,
        "run_label": label,
        "command": cmd,
        "cwd": GO_CWD,
    }
    if extra:
        item.update(extra)
    return item


def proposals_from_registry(registry: dict[str, Any]) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    entries = registry.get("versions") or registry.get("policies") or []
    for entry in entries:
        if entry.get("promotion_ready", True):
            continue
        domain = str(entry.get("domain", "nrp"))
        decision = str(entry.get("decision_type", "stagnation"))
        algo = str(entry.get("algorithm", "*"))
        algorithm = "sa" if algo == "*" else algo
        if domain == "jss" and algorithm == "*":
            algorithm = "tabu"
        if domain == "cvrp" and algorithm == "*":
            algorithm = "sa"
        regret = float(entry.get("regret_vs_rules", 0))
        accuracy = float(entry.get("offline_accuracy", 0))
        rationale = (
            f"{decision} policy for {domain} not promotion-ready "
            f"(accuracy={accuracy:.3f}, regret_vs_rules={regret:.2f}). "
            "Collect hybrid shadow telemetry on val matrix cell."
        )
        out.append(
            _proposal(
                proposal_type="fill_promotion_gap",
                priority=90 if decision == "stagnation" else 70,
                rationale=rationale,
                domain=domain,
                algorithm=algorithm,
                policy_mode="hybrid",
                seed=VERIFY_SEED,
                extra={
                    "policy_id": entry.get("id"),
                    "decision_type": decision,
                    "signal": "registry",
                },
            ),
        )
    return out


def proposals_from_harness(harness: dict[str, Any]) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    for row in harness.get("comparisons", []):
        if row.get("verdict") != "worse":
            continue
        mode_b = str(row.get("modeB", "hybrid"))
        domain = str(row.get("domain", "nrp"))
        algorithm = str(row.get("algorithm", "sa"))
        instance = str(row.get("instance", ""))
        delta = float(row.get("objectiveDelta", 0))
        roi = float(row.get("roi", 0))
        rationale = (
            f"Harness regression: {mode_b} worse than rules on "
            f"{domain}/{instance or '—'}/{algorithm} "
            f"(Δobj={delta:.2f}, ROI={roi:.2f}). Re-verify on fresh seed."
        )
        out.append(
            _proposal(
                proposal_type="harness_regression",
                priority=95,
                rationale=rationale,
                domain=domain,
                algorithm=algorithm,
                policy_mode=mode_b,
                seed=VERIFY_SEED,
                instance=instance,
                extra={
                    "objective_delta": delta,
                    "roi": roi,
                    "signal": "harness",
                },
            ),
        )
    return out


def proposals_from_neural(stagnation: dict[str, Any] | None) -> list[dict[str, Any]]:
    if not stagnation:
        return []
    neural = stagnation.get("neural") or {}
    if not neural.get("promotion_ready"):
        return []
    out: list[dict[str, Any]] = []
    for clf in neural.get("classifiers", []):
        if float(clf.get("gain_vs_trajectory", 0)) < 0.003:
            continue
        domain = str(clf.get("domain", ""))
        algorithm = str(clf.get("algorithm", "sa"))
        instance = str(clf.get("instance", ""))
        gain = float(clf.get("gain_vs_trajectory", 0))
        rationale = (
            f"Neural promoted on {domain}/{instance}/{algorithm} "
            f"(gain_vs_trajectory={gain:.4f}). Verify on seed {VERIFY_SEED}."
        )
        out.append(
            _proposal(
                proposal_type="neural_verify",
                priority=60,
                rationale=rationale,
                domain=domain,
                algorithm=algorithm,
                policy_mode="learned",
                seed=VERIFY_SEED,
                instance=instance,
                extra={
                    "gain_vs_trajectory": gain,
                    "signal": "neural",
                },
            ),
        )
    return out


def proposals_from_active_regression(
    registry: dict[str, Any],
    harness: dict[str, Any] | None,
) -> list[dict[str, Any]]:
    """Propose watch/re-verify when policies are active but harness still shows regressions."""
    if not harness:
        return []
    active = int(registry.get("active_count", 0))
    if active < 1:
        return []
    worse = [c for c in harness.get("comparisons", []) if c.get("verdict") == "worse"]
    if not worse:
        return []
    row = worse[0]
    domain = str(row.get("domain", "nrp"))
    algorithm = str(row.get("algorithm", "sa"))
    mode_b = str(row.get("modeB", "hybrid"))
    instance = str(row.get("instance", ""))
    rationale = (
        f"{active} lifecycle policies active, but harness still flags {len(worse)} "
        f"regression(s) vs rules (e.g. {domain}/{algorithm}/{mode_b}). "
        "Schedule approved re-verify runs before claiming domain-wide SI2 wins."
    )
    label = _slug("ml-watch", domain, algorithm, mode_b, f"s{VERIFY_SEED}")
    return [
        _proposal(
            proposal_type="active_regression_watch",
            priority=85,
            rationale=rationale,
            domain=domain,
            algorithm=algorithm,
            policy_mode=mode_b,
            seed=VERIFY_SEED,
            instance=instance,
            extra={
                "signal": "registry",
                "regression_count": len(worse),
                "active_policies": active,
            },
        ),
    ]


def proposals_from_retrain(validation: dict[str, Any], registry: dict[str, Any] | None) -> list[dict[str, Any]]:
    global_metrics = validation.get("global", {})
    samples = int(global_metrics.get("total_checkpoints", 0) or global_metrics.get("samples", 0))
    ready = int(registry.get("promotion_ready_count", 0)) if registry else 0
    total = int(registry.get("promotion_total", 0)) if registry else 0
    if samples < 1000 or ready >= total:
        return []
    rationale = (
        f"Registry {ready}/{total} promotion-ready with {samples} validation checkpoints. "
        "After approved shadow runs, retrain policies and re-validate."
    )
    return [{
        "id": "retrain-policies-post-shadow",
        "type": "retrain_policies",
        "priority": 50,
        "rationale": rationale,
        "requires_approval": True,
        "status": "proposed",
        "command": (
            "python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies"
        ),
        "cwd": "platform/ml",
        "signal": "retrain",
    }]


def dedupe_proposals(proposals: list[dict[str, Any]]) -> list[dict[str, Any]]:
    seen: set[str] = set()
    out: list[dict[str, Any]] = []
    for p in sorted(proposals, key=lambda x: (-int(x.get("priority", 0)), x.get("id", ""))):
        key = p.get("id", "")
        if key in seen:
            continue
        seen.add(key)
        out.append(p)
        if len(out) >= MAX_PROPOSALS:
            break
    return out


def evaluate_step8_gate(queue: dict[str, Any]) -> dict[str, Any]:
    proposals = queue.get("proposals", [])
    signals = set(queue.get("summary", {}).get("signals", []))
    all_require_approval = bool(proposals) and all(p.get("requires_approval") for p in proposals)
    all_have_commands = all(p.get("command") for p in proposals)
    loop_ok = (
        queue.get("human_approval_required") is True
        and len(proposals) >= 1
        and all_require_approval
        and all_have_commands
    )
    promote_ok = loop_ok and len(signals) >= 2
    return {
        "status": "ready" if loop_ok else "incomplete",
        "proposal_count": len(proposals),
        "signals": sorted(signals),
        "loop_ok": loop_ok,
        "promotion_ready": promote_ok,
        "human_approval_required": queue.get("human_approval_required", True),
    }


def build_research_queue(
    policy_dir: Path,
    repo_root: Path | None = None,
) -> dict[str, Any]:
    repo_root = repo_root or policy_dir.resolve().parents[2]
    registry = _read_json(policy_dir / "policy_registry.json") or {}
    validation = _read_json(policy_dir / "validation_results.json") or {}
    harness = _load_harness(repo_root) or {}
    stagnation = _read_json(policy_dir / "stagnation_policy.json")

    proposals: list[dict[str, Any]] = []
    proposals.extend(proposals_from_registry(registry))
    proposals.extend(proposals_from_active_regression(registry, harness))
    proposals.extend(proposals_from_harness(harness))
    proposals.extend(proposals_from_neural(stagnation))
    proposals.extend(proposals_from_retrain(validation, registry))
    proposals = dedupe_proposals(proposals)

    signals = sorted({p.get("signal", p.get("type", "unknown")) for p in proposals})
    queue = {
        "generated_at": datetime.now().isoformat(),
        "status": "ready",
        "human_approval_required": True,
        "step": 8,
        "proposals": proposals,
        "summary": {
            "proposal_count": len(proposals),
            "signals": signals,
            "by_type": {
                t: sum(1 for p in proposals if p.get("type") == t)
                for t in sorted({p.get("type") for p in proposals})
            },
        },
    }
    gate = evaluate_step8_gate(queue)
    queue["step8_loop_ok"] = gate["loop_ok"]
    queue["step8_promotion_ready"] = gate["promotion_ready"]
    queue["gate"] = gate
    return queue


def write_research_queue(queue: dict[str, Any], policy_dir: Path, repo_root: Path | None = None) -> list[Path]:
    repo_root = repo_root or policy_dir.resolve().parents[2]
    paths = [
        policy_dir / "research_queue.json",
        repo_root / "docs" / "reports" / "ml-research" / "queue.json",
    ]
    written: list[Path] = []
    payload = json.dumps(queue, indent=2)
    for path in paths:
        path.parent.mkdir(parents=True, exist_ok=True)
        with open(path, "w") as f:
            f.write(payload)
        written.append(path)
    return written


def merge_research_into_validation(
    validation: dict[str, Any],
    queue: dict[str, Any],
) -> dict[str, Any]:
    gate = evaluate_step8_gate(queue)
    validation["research"] = {
        "status": gate["status"],
        "proposal_count": gate["proposal_count"],
        "signals": gate["signals"],
        "loop_ok": gate["loop_ok"],
        "promotion_ready": gate["promotion_ready"],
        "human_approval_required": True,
        "evaluated_at": datetime.now().isoformat(),
    }
    validation["step8_promotion_ready"] = gate["promotion_ready"]
    validation["step8_loop_ok"] = gate["loop_ok"]
    return validation


def propose_and_write_queue(policy_dir: Path | None = None, repo_root: Path | None = None) -> dict[str, Any]:
    policy_dir = policy_dir or POLICY_DIR
    queue = build_research_queue(policy_dir, repo_root)
    write_research_queue(queue, policy_dir, repo_root)
    validation_path = policy_dir / "validation_results.json"
    validation = _read_json(validation_path)
    if validation is not None:
        validation = merge_research_into_validation(validation, queue)
        with open(validation_path, "w") as f:
            json.dump(validation, f, indent=2)
    return queue


def main() -> None:
    import argparse

    parser = argparse.ArgumentParser(description="Step 8 — propose ML research experiments")
    parser.add_argument("--policy-dir", type=Path, default=POLICY_DIR)
    parser.add_argument("--repo-root", type=Path, default=None)
    args = parser.parse_args()

    queue = propose_and_write_queue(args.policy_dir, args.repo_root)
    gate = queue.get("gate", {})
    print(f"Step 8 research queue — {queue['summary']['proposal_count']} proposals")
    print(f"  signals: {', '.join(queue['summary']['signals'])}")
    print(f"  step8_loop_ok: {gate.get('loop_ok')}  step8_promote_ok: {gate.get('promotion_ready')}")
    for p in queue["proposals"]:
        print(f"  [{p['priority']}] {p['id']}: {p['rationale'][:80]}...")


if __name__ == "__main__":
    main()
