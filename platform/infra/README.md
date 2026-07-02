# PFRS Research Lab — Infrastructure

AWS CDK (TypeScript) infrastructure for the PFRS Research Lab platform.

## Architecture

Phase 1 provisions:
- **S3 Bucket** — Versioned, encrypted, private storage for all optimisation run telemetry

Future phases will add:
- CloudFront distribution (static dashboard hosting)
- API Gateway + Lambda (run upload API)
- Cognito (authentication)
- DynamoDB (run metadata index)

## Prerequisites

- Node.js 18+
- AWS CLI configured with credentials
- CDK CLI: `npm install -g aws-cdk`

## Setup

```bash
cd platform/infra
npm install
```

## CDK Bootstrap

Required once per AWS account/region:

```bash
npx cdk bootstrap aws://ACCOUNT_ID/eu-west-1
```

Example:
```bash
npx cdk bootstrap aws://123456789012/eu-west-1
```

## Commands

| Command | Description |
|---------|-------------|
| `npm run build` | Compile TypeScript |
| `npm run synth` | Synthesise CloudFormation template |
| `npm run diff` | Show changes vs deployed stack |
| `npm run deploy` | Deploy stack to AWS |
| `npm run destroy` | Destroy stack (bucket is RETAINED) |

## Configuration

The bucket name is configurable via CDK context in `cdk.json`:

```json
{
  "context": {
    "bucketName": "pfrs-research-lab-data"
  }
}
```

Override at deploy time:
```bash
npx cdk deploy --context bucketName=my-custom-bucket-name
```

## Outputs

After deployment, the stack outputs:
- `BucketName` — The S3 bucket name
- `BucketArn` — The S3 bucket ARN

## Bucket Layout

```
s3://pfrs-research-lab-data/
├── manifest.json
├── runs/
│   ├── sa-baseline/
│   │   ├── metadata.json
│   │   ├── summary.json
│   │   ├── discoveries.csv
│   │   ├── tree.csv
│   │   ├── workers.csv
│   │   └── dashboard/
│   │       ├── index.html
│   │       └── assets/
│   └── lahc-budget-diversity/
│       └── ...
└── versions/
```

## Design Decisions

- **Versioning enabled**: Every run is immutable. Overwriting a run preserves the previous version.
- **RETAIN on deletion**: Destroying the CDK stack does NOT delete the bucket or its data.
- **Intelligent Tiering**: Runs older than 30 days automatically move to cheaper storage.
- **CORS pre-configured**: Ready for Phase 2 when the dashboard fetches directly from S3.
- **No public access**: All access requires authenticated AWS credentials.
- **Bucket owner enforced**: Prevents ACL-based access patterns.
- **SSL required**: No unencrypted transport permitted.

## Phase 2 Compatibility

The bucket is designed to be the long-term storage backend. Phase 2 additions (CloudFront, API Gateway, Lambda) will:
- Add an OAI/OAC to allow CloudFront to read from the bucket
- Add a Lambda function to handle uploads
- Add Cognito for authentication
- Add a DynamoDB table for fast metadata queries

None of these require changing the bucket configuration.
