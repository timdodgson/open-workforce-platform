#!/usr/bin/env python3
"""Step 4 CLI — run counterfactual offline eval and write validation_results.json."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

from counterfactual_eval import evaluate_offline_counterfactual, merge_counterfactual_into_validation
from policy_validation import validate_all


def main() -> None:
    parser = argparse.ArgumentParser(description="Counterfactual offline policy evaluation")
    parser.add_argument(
        "--data-dir",
        type=Path,
        default=Path(__file__).resolve().parent.parent / "web" / "pfrs-lab" / "data" / "runs",
    )
    parser.add_argument(
        "--policy-dir",
        type=Path,
        default=Path(__file__).resolve().parent / "policies",
    )
    parser.add_argument("--output", type=Path, default=None, help="Write validation_results.json here")
    args = parser.parse_args()

    validation = validate_all(args.data_dir, args.policy_dir)
    out_path = args.output or (args.policy_dir / "validation_results.json")

    with open(out_path, "w") as f:
        json.dump(validation, f, indent=2)

    cf = validation.get("counterfactual", {})
    print(f"Wrote {out_path}")
    print(f"  checkpoints: {validation.get('total_checkpoints', 0)}")
    print(f"  false_stop_rate: {cf.get('false_stop_rate', 'n/a')}")
    print(f"  step4_promotion_ready: {validation.get('step4_promotion_ready', False)}")


if __name__ == "__main__":
    main()
