"""
Search Intelligence 2.0 — Retrospective Policy Validation.

Outcome-based promotion gates (primary):
  - outcome_accuracy >= 80%
  - regret_vs_rules <= 0

Rule agreement is reported as diagnostic only.

Usage:
    python validate_policies.py --data-dir ../web/pfrs-lab/data/runs --policy-dir policies
    python validate_policies.py --data-dir ... --policy-dir ... --merge-registry
"""

import argparse
import json
import os
from pathlib import Path

from policy_registry import (
    load_registry,
    merge_validation_into_registry,
    save_registry,
    sync_dashboard_registry,
)
from policy_validation import validate_all


def parse_args():
    parser = argparse.ArgumentParser(description="Validate SI 2.0 policies")
    parser.add_argument("--data-dir", required=True)
    parser.add_argument("--policy-dir", default="policies")
    parser.add_argument("--output", default="docs/reports/search-intelligence-v2-validation.md")
    parser.add_argument(
        "--merge-registry",
        action="store_true",
        default=True,
        help="Merge validation metrics into policy_registry.json (default: true)",
    )
    parser.add_argument("--training-report", default="", help="Optional training_report.json for worker CV")
    return parser.parse_args()


def _gate_label(passed: bool) -> str:
    return "✅ PASS" if passed else "❌ FAIL"


def generate_report(results: dict, output_path: str):
    lines = [
        "# Search Intelligence 2.0 — Validation Report",
        "",
        f"## Status: {'Validated' if results.get('status') == 'validated' else 'Not Yet Evaluated'}",
        "",
        f"Generated: {results.get('validated_at', 'N/A')}",
        "",
        "---",
        "",
        "## Outcome-Based Promotion (Primary)",
        "",
        "Ex-post optimal stop/continue vs learned and rule decisions.",
        "",
        "| Metric | Value |",
        "|--------|-------|",
        f"| Total checkpoints | {results.get('total_checkpoints', 0)} |",
    ]

    g = results.get("global", {})
    lines.extend([
        f"| Learned outcome accuracy | {g.get('outcome_accuracy', 0) * 100:.1f}% |",
        f"| Rule outcome accuracy | {g.get('rule_outcome_accuracy', 0) * 100:.1f}% |",
        f"| Regret vs rules | {g.get('regret_vs_rules', 0):.4f} |",
        f"| Rule agreement (diagnostic) | {results.get('agreement_rate', 0) * 100:.1f}% |",
        "",
        "---",
        "",
        "## Per-Domain Stagnation",
        "",
        "| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |",
        "|--------|---------|-------------|-----------------|-----------|-----------|",
    ])

    for domain, stats in sorted(results.get("domain_stats", {}).items()):
        ready = "✅" if stats.get("promotion_ready") else "❌"
        lines.append(
            f"| {domain.upper()} | {stats['total']} | "
            f"{stats.get('outcome_accuracy', 0) * 100:.1f}% | "
            f"{stats.get('regret_vs_rules', 0):.4f} | "
            f"{stats.get('agreement_rate', 0) * 100:.1f}% | {ready} |"
        )

    gate = results.get("promotion_gate", {})
    min_acc = gate.get("min_outcome_accuracy", 0.80)
    max_regret = gate.get("max_regret_vs_rules", 0.0)
    outcome_acc = g.get("outcome_accuracy", 0)
    regret = g.get("regret_vs_rules", 0)

    lines.extend([
        "",
        "---",
        "",
        "## Acceptance Criteria",
        "",
        "| Criterion | Result |",
        "|-----------|--------|",
        f"| Outcome accuracy >= {min_acc:.0%} | {_gate_label(outcome_acc >= min_acc)} ({outcome_acc * 100:.1f}%) |",
        f"| Regret vs rules <= {max_regret} | {_gate_label(regret <= max_regret)} ({regret:.4f}) |",
        f"| Learned policy loaded | {_gate_label(results.get('stagnation_model_loaded', False))} |",
        "",
        "---",
        "",
        "## Recommendation",
        "",
    ])

    if results.get("promotion_ready"):
        lines.append("**Promote to shadow** — outcome gates passed globally.")
    elif any(s.get("promotion_ready") for s in results.get("domain_stats", {}).values()):
        lines.append("**Promote per-domain** — some domains pass outcome gates; use hybrid for passing domains.")
    else:
        lines.append("**Collect more shadow data** — outcome gates not yet met.")

    lines.extend([
        "",
        "---",
        "",
        "## Methodology",
        "",
        "- Data: `generic_search_assist.csv` shadow checkpoints",
        "- Ex-post label: stop if `best_penalty - final_best_penalty <= 1`",
        "- Learned: stagnation curve P(improve) model (domain + algorithm scoped)",
        "- Rules: 50k plateau stagnation window",
        "- Promotion: outcome accuracy >= 80% AND regret_vs_rules <= 0",
        "",
    ])

    os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")


def main():
    args = parse_args()
    data_dir = Path(args.data_dir)
    policy_dir = Path(args.policy_dir)

    training_results = None
    if args.training_report:
        with open(args.training_report) as f:
            report = json.load(f)
            training_results = report.get("results", report)

    print(f"Validating policies from: {policy_dir}")
    print(f"Against telemetry from: {data_dir}")
    print()

    results = validate_all(data_dir, policy_dir, training_results)

    if results.get("status") == "no_data":
        print("No search assist data found. Cannot validate.")
        return

    g = results.get("global", {})
    print(f"Total checkpoints: {results.get('total_checkpoints', 0)}")
    print(f"Outcome accuracy: {g.get('outcome_accuracy', 0) * 100:.1f}%")
    print(f"Regret vs rules: {g.get('regret_vs_rules', 0):.4f}")
    print(f"Rule agreement (diagnostic): {results.get('agreement_rate', 0) * 100:.1f}%")
    print()
    for domain, stats in sorted(results.get("domain_stats", {}).items()):
        tag = "READY" if stats.get("promotion_ready") else "WAIT"
        print(
            f"  {domain.upper()}: outcome {stats.get('outcome_accuracy', 0) * 100:.0f}% "
            f"regret {stats.get('regret_vs_rules', 0):.4f} [{tag}]"
        )
    print()

    generate_report(results, args.output)
    print(f"Report written: {args.output}")

    results_path = policy_dir / "validation_results.json"
    with open(results_path, "w") as f:
        json.dump(results, f, indent=2)
    print(f"Results written: {results_path}")

    if args.merge_registry:
        registry_path = policy_dir / "policy_registry.json"
        registry = load_registry(registry_path)
        if not registry.get("versions"):
            training_path = policy_dir / "training_report.json"
            if training_path.exists():
                with open(training_path) as f:
                    tr = json.load(f)
                from policy_registry import build_lifecycle_registry

                registry = build_lifecycle_registry(tr.get("results", {}))

        merged = merge_validation_into_registry(registry, results, training_results)
        save_registry(registry_path, merged)
        sync_dashboard_registry(policy_dir, merged)
        ready = merged.get("promotion_ready_count", 0)
        total = merged.get("promotion_total", 0)
        print(f"Registry updated: {registry_path} ({ready}/{total} promotion-ready)")


if __name__ == "__main__":
    main()
