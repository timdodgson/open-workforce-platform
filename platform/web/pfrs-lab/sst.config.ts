/**
 * PFRS Lab dashboard — OpenNext (CloudFront + Lambda) production runtime.
 * Excluded from Next.js tsconfig (see exclude in tsconfig.json).
 *
 * Legacy ECS/Fargate is disabled by default (CDK context enableEcsDashboard=false).
 *
 * Deploy:
 *   cd platform/web/pfrs-lab
 *   npx sst deploy --stage production
 *
 * Optional custom domain (Route 53 in same account):
 *   DASHBOARD_DOMAIN=pfrs-lab.com npx sst deploy --stage production
 */
export default $config({
  app(input) {
    return {
      name: "pfrs-lab",
      removal: input?.stage === "production" ? "retain" : "remove",
      home: "aws",
      providers: {
        aws: { region: "eu-west-1" },
      },
    };
  },
  async run() {
    const bucketName = process.env.PFRS_S3_BUCKET || "pfrs-research-lab-data";
    const dataBucket = sst.aws.Bucket.get("ResearchData", bucketName);

    // Cognito from existing DashboardStack (set via env when deploying).
    // aws cloudformation describe-stacks --stack-name PfrsDashboardStack --query ...
    const cognitoUserPoolId = process.env.COGNITO_USER_POOL_ID || "eu-west-1_J3FLcGW6P";
    const cognitoClientId = process.env.COGNITO_CLIENT_ID || "dnjtkgqomiq15if0519nalgp4";

    const dashboardDomain = process.env.DASHBOARD_DOMAIN;
    const anthropicSsmPath =
      process.env.ANTHROPIC_API_KEY_SSM || "/pfrs-lab/production/anthropic-api-key";

    const nextjsArgs: sst.aws.NextjsArgs = {
      link: [dataBucket],
      // Keep server warm — reduces cold-start lag on navigation.
      warm: 3,
      // CloudFront origin timeout hard-caps ~60s — keep under that.
      server: {
        timeout: "50 seconds",
        memory: "2048 MB",
      },
      environment: {
        STORAGE_PROVIDER: "s3",
        PFRS_S3_BUCKET: bucketName,
        // Do not set AWS_REGION — Lambda reserves it and CreateFunction fails.
        NODE_ENV: "production",
        COGNITO_USER_POOL_ID: cognitoUserPoolId,
        COGNITO_CLIENT_ID: cognitoClientId,
        PFRS_ADMIN_MODE: "authenticated",
        NEXT_PUBLIC_ADMIN_MODE: "authenticated",
        // Injected by CI on each OpenNext deploy (Admin "Release" panel).
        NEXT_PUBLIC_APP_VERSION: process.env.NEXT_PUBLIC_APP_VERSION || "unknown",
        NEXT_PUBLIC_GIT_SHA: process.env.NEXT_PUBLIC_GIT_SHA || "unknown",
        NEXT_PUBLIC_DEPLOYED_AT: process.env.NEXT_PUBLIC_DEPLOYED_AT || "unknown",
        LLM_PROVIDER: process.env.LLM_PROVIDER || "anthropic",
        ANTHROPIC_MODEL_ID: process.env.ANTHROPIC_MODEL_ID || "claude-haiku-4-5-20251001",
        // Parameter *name* only — key value lives in SSM SecureString, lazy-loaded at runtime.
        ANTHROPIC_API_KEY_SSM: anthropicSsmPath,
        BEDROCK_MODEL_ID: process.env.BEDROCK_MODEL_ID || "eu.anthropic.claude-3-haiku-20240307-v1:0",
      },
      permissions: [
        {
          actions: [
            "s3:GetObject",
            "s3:PutObject",
            "s3:DeleteObject",
            "s3:ListBucket",
            "s3:HeadObject",
          ],
          resources: [
            `arn:aws:s3:::${bucketName}`,
            `arn:aws:s3:::${bucketName}/*`,
          ],
        },
        {
          // Path begins with / → ARN uses :parameter/pfrs-lab/...
          actions: ["ssm:GetParameter"],
          resources: [
            `arn:aws:ssm:eu-west-1:*:parameter/pfrs-lab/production/anthropic-api-key`,
          ],
        },
        {
          actions: ["kms:Decrypt"],
          resources: ["*"],
        },
        {
          actions: [
            "bedrock:InvokeModel",
            "aws-marketplace:ViewSubscriptions",
            "aws-marketplace:Subscribe",
          ],
          resources: ["*"],
        },
      ],
    };

    if (dashboardDomain) {
      nextjsArgs.domain = dashboardDomain;
    }

    const web = new sst.aws.Nextjs("Dashboard", nextjsArgs);

    return {
      url: web.url,
      bucket: bucketName,
      stage: $app.stage,
    };
  },
});
