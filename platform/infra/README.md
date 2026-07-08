# PFRS Research Lab — Infrastructure

AWS CDK (TypeScript) infrastructure for the PFRS Research Lab platform.

## Architecture

- **S3 Bucket** — Versioned, encrypted, private storage for all optimisation run telemetry
- **App Runner** — Serverless container hosting for the Next.js dashboard (scales to zero)
- **ECR** — Container registry for dashboard images

## Prerequisites

- Node.js 18+
- AWS CLI configured with credentials
- CDK CLI: `npm install -g aws-cdk`
- Docker (for building dashboard image)

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

## Dashboard Deployment (OpenNext — production)

The dashboard runs on **CloudFront + Lambda** via SST/OpenNext. ECS/Fargate is **disabled by default**.

### Deploy (CI)

- **Releases** (semantic-release on `main`) → auto-deploy OpenNext
- **Manual**: Actions → *Build & Deploy* → Run workflow → `deploy_opennext=true`

### Deploy (local)

```bash
cd platform/web/pfrs-lab
npm install
STORAGE_PROVIDER=local npx sst deploy --stage production
```

SST prints the CloudFront URL at the end.

### Custom domain (optional)

If `pfrs-lab.com` is in Route 53 (same AWS account):

```bash
DASHBOARD_DOMAIN=pfrs-lab.com npx sst deploy --stage production
```

Or set GitHub repo variable `DASHBOARD_DOMAIN` for CI deploys.

### Decommission ECS/Fargate + ALB

CDK context `enableEcsDashboard` defaults to `false`. Deploying removes the legacy stack resources:

```bash
cd platform/infra
npm run build
npx cdk deploy PfrsDashboardStack
```

Confirm the diff shows ECS cluster, ALB, and Fargate service **deletion**. Cognito + GitHub deploy role are retained.

To temporarily restore ECS for rollback:

```bash
npx cdk deploy PfrsDashboardStack --context enableEcsDashboard=true
```

Then run the workflow with `deploy_ecs=true`.

---

## Dashboard Deployment (ECS legacy — deprecated)

### Deploy Infrastructure

```bash
cd platform/infra
npm run build
npx cdk deploy PfrsDashboardStack
```

This creates:
- ECR repository (`pfrs-lab-dashboard`)
- App Runner service with IAM roles
- Auto-deploy enabled (push to ECR → auto redeploy)

### Build & Push Dashboard Image

```bash
cd platform/web/pfrs-lab

# Get ECR login
aws ecr get-login-password --region eu-west-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.eu-west-1.amazonaws.com

# Build
docker build -t pfrs-lab-dashboard .

# Tag
docker tag pfrs-lab-dashboard:latest <ACCOUNT_ID>.dkr.ecr.eu-west-1.amazonaws.com/pfrs-lab-dashboard:latest

# Push
docker push <ACCOUNT_ID>.dkr.ecr.eu-west-1.amazonaws.com/pfrs-lab-dashboard:latest
```

App Runner will automatically detect the new image and redeploy (~30s).

### Access

After deployment, the `ServiceUrl` output gives you the HTTPS URL:
```
https://xxxxxxxx.eu-west-1.awsapprunner.com
```

### Custom Domain (later)

1. Register domain in Route 53
2. Add custom domain in App Runner console (or CDK)
3. App Runner handles SSL certificate automatically

### Costs

- **Idle**: ~$0/month (scales to zero)
- **Active**: ~$0.007/vCPU-hour + $0.0008/GB-hour
- **First request after idle**: ~10-15s cold start
