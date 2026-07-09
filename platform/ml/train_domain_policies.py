"""
Domain-scoped Search Intelligence policy training with outcome-based promotion gates.

Usage:
    python train_domain_policies.py --domain cvrp --data-dir ../web/pfrs-lab/data/runs --output-dir policies/cvrp
"""

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path

from policy_registry import (
    build_lifecycle_registry,
    merge_validation_into_registry,
    save_registry,
)
from policy_validation import validate_all
from train_policies import (
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
    print("Promotion gate: outcome accuracy >= 80%, regret_vs_rules <= 0")
    print()

    portfolio_df = filter_by_domain(load_portfolio_assist_data(data_dir), domain)
    search_df = filter_by_domain(load_search_assist_data(data_dir), domain)
    worker_df = filter_by_domain(load_worker_assist_data(data_dir), domain)
    metadata_df = filter_by_domain(load_run_metadata(data_dir), domain)

    results = {}
    policy_set = DOMAIN_POLICIES[domain]

    if "budget" in policy_set:
        results["budget_policy"] = train_portfolio_budget_policy(portfolio_df, args.min_samples)
    if "stagnation" in policy_set:
        results["stagnation_policy"] = train_stagnation_policy(search_df, metadata_df, args.min_samples)
    if "restart" in policy_set:
        results["restart_policy"] = train_restart_policy(search_df, args.min_samples)
    if "worker" in policy_set:
        results["worker_policy"] = train_worker_policy(worker_df, args.min_samples)

    for key, result in results.items():
        if result.get("status") == "trained":
            out_file = output_dir / f"{key.replace('_policy', '')}_policy.json"
            with open(out_file, "w") as f:
                json.dump(result, f, indent=2)

    lifecycle = build_lifecycle_registry(results)
    validation = validate_all(data_dir, output_dir, results)
    if validation.get("status") != "no_data":
        lifecycle = merge_validation_into_registry(lifecycle, validation, results)
        with open(output_dir / "validation_results.json", "w") as f:
            json.dump(validation, f, indent=2)

    lifecycle["domain"] = domain
    save_registry(output_dir / "policy_registry.json", lifecycle)

    report = generate_training_report(results)
    report["domain"] = domain
    report["promotion_ready"] = lifecycle.get("promotion_ready_count", 0)

    with open(output_dir / "training_report.json", "w") as f:
        json.dump(report, f, indent=2)

    with open(output_dir / "domain_training_manifest.json", "w") as f:
        json.dump({
            "domain": domain,
            "generated_at": datetime.now().isoformat(),
            "policies": list(results.keys()),
            "promotion_ready": lifecycle.get("promotion_ready_count", 0),
            "promotion_total": lifecycle.get("promotion_total", 0),
        }, f, indent=2)

    print(f"Done. {lifecycle.get('promotion_ready_count', 0)}/{lifecycle.get('promotion_total', 0)} promotion-ready for {domain}.")
    print(f"Output: {output_dir}")


if __name__ == "__main__":
    main()
