# OpenNext + Lambda Migration

Pilot: CloudFront → Lambda (OpenNext via SST)  
Rollback: existing ECS + ALB stack stays live until cutover.

## Why

| | ECS + ALB (today) | OpenNext + Lambda (pilot) |
|--|--|--|
| Cost | ALB + always-on Fargate | Pay per request; no ALB |
| Static assets | Served from container | CloudFront edge cache |
| Freshness | force-dynamic pages | Same — **do not CDN-cache data routes** |
| Timeout | ALB ~60s | CloudFront origin ~60s (we use 50s Lambda) |

## Prerequisites

- AWS credentials configured (`eu-west-1`)
- Node 20+
- Cognito IDs from existing dashboard stack (assistant auth)

```powershell
aws cloudformation describe-stacks --stack-name PfrsDashboardStack --region eu-west-1 `
  --query "Stacks[0].Outputs[?OutputKey=='CognitoUserPoolId' || OutputKey=='CognitoClientId'].[OutputKey,OutputValue]" `
  --output table
```

## Deploy pilot

**Prefer GitHub Actions (Linux)** — OpenNext needs symlinks; Windows often fails with `EPERM` unless Developer Mode is on. Use WSL or CI.

### Option A — GitHub Actions (recommended)

Auth is **OIDC** via existing `AWS_DEPLOY_ROLE_ARN` — no Cognito GitHub secrets.
Cognito IDs are read at deploy time from `PfrsDashboardStack` CloudFormation outputs.

1. Ensure deploy role can also manage SST resources (Lambda, CloudFront, S3, IAM, CloudFormation) — today it may only cover ECR + ECS
2. Actions → **Build & Deploy** → **Run workflow**
3. Enable **Deploy OpenNext Lambda/CloudFront pilot**
4. Read the job log for the CloudFront URL

### Option B — Local (WSL or Windows Developer Mode)

```powershell
cd platform\web\pfrs-lab
npm install
npx sst deploy --stage production
```

If you see `EPERM: symlink`, enable **Windows Settings → System → For developers → Developer Mode**, or use WSL.

## Freshness rules (hard requirements)

**Never cache at CloudFront** (must see new runs immediately):

- `/intelligence*`
- `/benchmarks`
- `/runs*`
- `/predictions`
- `/api/*`

**Cache hard:**

- `/_next/static/*`

SST/OpenNext already puts assets on CloudFront with long TTLs and keeps SSR routes as Lambda origins. Do not add custom cache policies that blanket-cache HTML.

## Validation checklist

Before cutting over from ALB:

1. Upload a new `--storage s3` run from your PC
2. Confirm it appears on the **SST URL** within a minute (hard refresh)
3. `/intelligence`, `/benchmarks`, `/runs` load without 504
4. Cognito login + Assistant (`/api/chat`) work
5. Cost / cold-start acceptable in CloudWatch

## Cutover

1. Point domain / bookmark to SST CloudFront URL
2. Keep ECS running 1–2 weeks as rollback
3. Then scale ECS desiredCount to 0 (or destroy `PfrsDashboardStack` carefully — Cognito is RETAIN)

## Rollback

Use the ECS ALB URL from `PfrsDashboardStack` outputs / existing ELB DNS. No code change required.

## CI (optional)

`.github/workflows/deploy.yml` can gain a parallel `deploy-opennext` job later. Until then, deploy pilot manually with `npx sst deploy --stage production`.

## Files

| Path | Role |
|------|------|
| `platform/web/pfrs-lab/sst.config.ts` | SST app — CloudFront + Lambda + S3/Bedrock IAM |
| `platform/web/pfrs-lab/package.json` | `sst:deploy`, `opennext:build` scripts |
| `platform/infra/.../dashboard-stack.ts` | Legacy ECS — keep until cutover done |
