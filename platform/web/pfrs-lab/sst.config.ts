/**
 * PFRS Lab dashboard — OpenNext (CloudFront + Lambda) pilot.
 * Excluded from Next.js tsconfig (see exclude in tsconfig.json).
 *
 * Deploys in parallel with the existing ECS/ALB stack so you can cut over
 * after verifying freshness + timeouts. ECS remains the rollback path.
 *
 * Deploy:
 *   cd platform/web/pfrs-lab
 *   npx sst deploy --stage production
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

    const web = new sst.aws.Nextjs("Dashboard", {
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
          actions: [
            "bedrock:InvokeModel",
            "aws-marketplace:ViewSubscriptions",
            "aws-marketplace:Subscribe",
          ],
          resources: ["*"],
        },
      ],
    });

    return {
      url: web.url,
      bucket: bucketName,
      stage: $app.stage,
    };
  },
});
