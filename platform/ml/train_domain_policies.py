"""
Domain-scoped Search Intelligence policy training.

Trains policies for a single domain (nrp, cvrp, jss, vrptw) and applies the
0.80 offline agreement gate before marking models as promotion-ready.

Usage:
    python train_domain_policies.py --domain cvrp --data-dir ../web/pfrs-lab/data/runs --output-dir policies/cvrp
"""

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path

# Reuse core trainers from the global pipeline.
from train_policies import (
    MIN_LEARNED_POLICY_AGREEMENT,
    build_lifecycle_registry,
    detect_domain,
    generate_training_report,
    load_portfolio_assist_data,
    load_run_metadata,
    load_search_assist_data,
    load_worker_assist_data,
    train_portfolio_budget_policy,
    train_restart_policy,
    train_stagnation_policy,
    train_worker_policy,
)

DOMAIN_POLICIES = {
    "nrp": ["worker"],
    "cvrp": ["budget", "stagnation", "restart"],
    "jss": ["budget", "stagnation", "restart"],
    "vrptw": ["budget", "stagnation", "restart"],
}


def filter_by_domain(df, domain: str, run_col: str = "run_id"):
    if df.empty or run_col not in df.columns:
        return df
    mask = df[run_col].apply(lambda rid: detect_domain(str(rid)) == domain)
    return df[mask].copy()


def passes_promotion_gate(result: dict) -> bool:
    if result.get("status") != "trained":
        return False
    agreement = result.get("accuracy", result.get("cv_mean", 0))
    return agreement >= MIN_LEARNED_POLICY_AGREEMENT


def annotate_gate(result: dict) -> dict:
    agreement = result.get("accuracy", result.get("cv_mean", 0))
    result["agreement_rate"] = round(float(agreement), 4)
    result["promotion_ready"] = passes_promotion_gate(result)
    result["promotion_gate"] = MIN_LEARNED_POLICY_AGREEMENT
    if result.get("status") == "trained" and not result["promotion_ready"]:
        result["status"] = "below_promotion_gate"
    return result


def parse_args():
    parser = argparse.ArgumentParser(description="Train SI policies for one domain")
    parser.add_argument("--domain", required=True, choices=sorted(DOMAIN_POLICIES.keys()))
    parser.add_argument("--data-dir", required=True, help="Path to data/runs directory")
    parser.add_argument("--output-dir", required=True, help="Output directory for domain policies")
    parser.add_argument("--min-samples", type=int, default=20)
    return parser.parse_args()


def main():
    args = parse_args()
    domain = args.domain
    data_dir = Path(args.data_dir)
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    if not data_dir.exists():
        print(f"Error: data directory not found: {data_dir}")
        sys.exit(1)

    print(f"Domain: {domain}")
    print(f"Promotion gate: {MIN_LEARNED_POLICY_AGREEMENT:.0%} offline agreement")
    print()

    portfolio_df = filter_by_domain(load_portfolio_assist_data(data_dir), domain)
    search_df = filter_by_domain(load_search_assist_data(data_dir), domain)
    worker_df = filter_by_domain(load_worker_assist_data(data_dir), domain)
    metadata_df = filter_by_domain(load_run_metadata(data_dir), domain)

    results = {}
    policy_set = DOMAIN_POLICIES[domain]

    if "budget" in policy_set:
        results["budget_policy"] = annotate_gate(
            train_portfolio_budget_policy(portfolio_df, args.min_samples)
        )
    if "stagnation" in policy_set:
        results["stagnation_policy"] = annotate_gate(
            train_stagnation_policy(search_df, metadata_df, args.min_samples)
        )
    if "restart" in policy_set:
        results["restart_policy"] = annotate_gate(
            train_restart_policy(search_df, args.min_samples)
        )
    if "worker" in policy_set:
        results["worker_policy"] = annotate_gate(
            train_worker_policy(worker_df, args.min_samples)
        )

    for key, result in results.items():
        if result.get("status") in ("trained", "below_promotion_gate"):
            out_file = output_dir / f"{key.replace('_policy', '')}_policy.json"
            with open(out_file, "w") as f:
                json.dump(result, f, indent=2)

    lifecycle = build_lifecycle_registry(results)
    lifecycle["domain"] = domain
    lifecycle["promotion_gate"] = MIN_LEARNED_POLICY_AGREEMENT

    with open(output_dir / "policy_registry.json", "w") as f:
        json.dump(lifecycle, f, indent=2)

    report = generate_training_report(results)
    report["domain"] = domain
    report["promotion_gate"] = MIN_LEARNED_POLICY_AGREEMENT
    report["promotion_ready"] = sum(1 for r in results.values() if r.get("promotion_ready"))

    with open(output_dir / "training_report.json", "w") as f:
        json.dump(report, f, indent=2)

    with open(output_dir / "domain_training_manifest.json", "w") as f:
        json.dump({
            "domain": domain,
            "generated_at": datetime.now().isoformat(),
            "policies": list(results.keys()),
            "promotion_ready": report["promotion_ready"],
            "promotion_gate": MIN_LEARNED_POLICY_AGREEMENT,
        }, f, indent=2)

    print(f"Done. {report['promotion_ready']} policies promotion-ready for {domain}.")
    print(f"Output: {output_dir}")


if __name__ == "__main__":
    main()
