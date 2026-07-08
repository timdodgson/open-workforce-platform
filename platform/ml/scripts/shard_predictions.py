"""
Shard an existing worker_predictions.json into per-run files + index.
Run once after upgrading dashboard to the paginated API.

Usage:
    python scripts/shard_predictions.py --input worker_predictions.json --data-dir ../web/pfrs-lab/data --storage s3
"""

import argparse
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from worker_model.predict import write_prediction_shards


def main():
    parser = argparse.ArgumentParser(description="Shard monolithic predictions JSON")
    parser.add_argument("--input", type=Path, default=Path("worker_predictions.json"))
    parser.add_argument("--data-dir", type=Path, default=Path("../web/pfrs-lab/data"))
    parser.add_argument("--storage", choices=["local", "s3"], default="local")
    parser.add_argument("--s3-bucket", default="pfrs-research-lab-data")
    parser.add_argument("--s3-region", default="eu-west-1")
    args = parser.parse_args()

    if not args.input.exists():
        print(f"ERROR: {args.input} not found", file=sys.stderr)
        sys.exit(1)

    print(f"Loading {args.input}...")
    data = json.load(open(args.input))
    predictions = data.get("predictions", data if isinstance(data, list) else [])
    print(f"  {len(predictions)} predictions")

    index_path = write_prediction_shards(
        predictions,
        args.data_dir,
        upload_s3=args.storage == "s3",
        s3_bucket=args.s3_bucket,
        s3_region=args.s3_region,
    )
    print(f"Done. Index: {index_path}")


if __name__ == "__main__":
    main()
