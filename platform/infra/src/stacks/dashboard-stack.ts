import * as cdk from 'aws-cdk-lib';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import { Construct } from 'constructs';

/**
 * DashboardStack — auth + CI deploy role for PFRS Lab.
 *
 * Production dashboard: OpenNext (SST) — CloudFront + Lambda.
 * Legacy ECS/Fargate + ALB is optional via context `enableEcsDashboard` (default false).
 *
 * Set enableEcsDashboard=true only for rollback during cutover.
 */
export class DashboardStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const bucketName = this.node.tryGetContext('bucketName') || 'pfrs-research-lab-data';
    const enableEcs = this.node.tryGetContext('enableEcsDashboard') === true;
    const githubOrg = this.node.tryGetContext('githubOrg') || 'timdodgson';
    const githubRepo = this.node.tryGetContext('githubRepo') || 'open-workforce-platform';

    const dataBucket = s3.Bucket.fromBucketName(this, 'DataBucket', bucketName);

    // --- Cognito User Pool (assistant auth — shared by OpenNext + legacy ECS) ---
    const userPool = new cognito.UserPool(this, 'AssistantUserPool', {
      userPoolName: 'pfrs-lab-users',
      selfSignUpEnabled: false,
      signInAliases: { email: true },
      passwordPolicy: {
        minLength: 8,
        requireUppercase: false,
        requireDigits: false,
        requireSymbols: false,
      },
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    const userPoolClient = userPool.addClient('DashboardClient', {
      userPoolClientName: 'pfrs-dashboard',
      authFlows: {
        userPassword: true,
        userSrp: true,
      },
      generateSecret: false,
    });

    // --- GitHub Actions OIDC + Deploy Role ---
    const oidcProvider = new iam.OpenIdConnectProvider(this, 'GitHubOIDC', {
      url: 'https://token.actions.githubusercontent.com',
      clientIds: ['sts.amazonaws.com'],
    });

    const deployRole = new iam.Role(this, 'GitHubActionsDeployRole', {
      roleName: 'pfrs-lab-github-deploy',
      assumedBy: new iam.FederatedPrincipal(
        oidcProvider.openIdConnectProviderArn,
        {
          StringEquals: { 'token.actions.githubusercontent.com:aud': 'sts.amazonaws.com' },
          StringLike: { 'token.actions.githubusercontent.com:sub': `repo:${githubOrg}/${githubRepo}:*` },
        },
        'sts:AssumeRoleWithWebIdentity',
      ),
      description: 'GitHub Actions deploy role for PFRS Lab (OpenNext + optional ECS)',
    });

    dataBucket.grantReadWrite(deployRole);

    // OpenNext / SST deploy — SST documents Action:"*" for the Nextjs component.
    // See https://sst.dev/docs/iam-credentials/
    deployRole.addToPolicy(new iam.PolicyStatement({
      sid: 'SstOpenNextDeployments',
      actions: ['*'],
      resources: ['*'],
    }));

    let albDns: string | undefined;
    let ecrUri: string | undefined;
    let serviceArn: string | undefined;

    if (enableEcs) {
      const repo = new ecr.Repository(this, 'DashboardRepo', {
        repositoryName: 'pfrs-lab-dashboard',
        removalPolicy: cdk.RemovalPolicy.RETAIN,
        lifecycleRules: [{ maxImageCount: 10 }],
      });
      ecrUri = repo.repositoryUri;

      const vpc = new ec2.Vpc(this, 'Vpc', {
        maxAzs: 2,
        natGateways: 0,
        subnetConfiguration: [
          { name: 'Public', subnetType: ec2.SubnetType.PUBLIC, cidrMask: 24 },
        ],
      });

      const cluster = new ecs.Cluster(this, 'Cluster', { vpc });

      const taskDef = new ecs.FargateTaskDefinition(this, 'TaskDef', {
        cpu: 256,
        memoryLimitMiB: 512,
      });

      dataBucket.grantReadWrite(taskDef.taskRole);
      taskDef.taskRole.addToPrincipalPolicy(new iam.PolicyStatement({
        actions: ['bedrock:InvokeModel', 'aws-marketplace:ViewSubscriptions', 'aws-marketplace:Subscribe'],
        resources: ['*'],
      }));

      taskDef.addContainer('Dashboard', {
        image: ecs.ContainerImage.fromEcrRepository(repo, 'latest'),
        portMappings: [{ containerPort: 3000 }],
        environment: {
          STORAGE_PROVIDER: 's3',
          PFRS_S3_BUCKET: bucketName,
          AWS_REGION: this.region,
          NODE_ENV: 'production',
          COGNITO_USER_POOL_ID: userPool.userPoolId,
          COGNITO_CLIENT_ID: userPoolClient.userPoolClientId,
        },
        logging: ecs.LogDrivers.awsLogs({ streamPrefix: 'pfrs-dashboard' }),
      });

      const alb = new elbv2.ApplicationLoadBalancer(this, 'ALB', {
        vpc,
        internetFacing: true,
      });

      const listener = alb.addListener('HTTP', {
        port: 80,
        protocol: elbv2.ApplicationProtocol.HTTP,
      });

      const service = new ecs.FargateService(this, 'Service', {
        cluster,
        taskDefinition: taskDef,
        desiredCount: 1,
        assignPublicIp: true,
        capacityProviderStrategies: [
          { capacityProvider: 'FARGATE_SPOT', weight: 1 },
        ],
      });

      listener.addTargets('Dashboard', {
        port: 3000,
        protocol: elbv2.ApplicationProtocol.HTTP,
        targets: [service],
        healthCheck: { path: '/', interval: cdk.Duration.seconds(30) },
      });

      albDns = alb.loadBalancerDnsName;
      serviceArn = service.serviceArn;

      repo.grantPush(deployRole);
      deployRole.addToPolicy(new iam.PolicyStatement({
        actions: ['ecr:GetAuthorizationToken'],
        resources: ['*'],
      }));
      deployRole.addToPolicy(new iam.PolicyStatement({
        actions: ['ecs:UpdateService', 'ecs:DescribeServices'],
        resources: [service.serviceArn],
      }));
    }

    // --- Outputs ---
    new cdk.CfnOutput(this, 'DashboardMode', {
      value: enableEcs ? 'ecs-legacy' : 'opennext',
      description: 'Production dashboard runtime',
    });

    if (albDns) {
      new cdk.CfnOutput(this, 'DashboardUrl', {
        value: `http://${albDns}`,
        description: 'Legacy ECS dashboard URL (rollback only)',
      });
    }

    if (ecrUri) {
      new cdk.CfnOutput(this, 'EcrRepoUri', {
        value: ecrUri,
        description: 'ECR repo for legacy ECS images',
      });
    }

    new cdk.CfnOutput(this, 'DeployRoleArn', {
      value: deployRole.roleArn,
      description: 'Add as AWS_DEPLOY_ROLE_ARN GitHub secret',
    });

    if (serviceArn) {
      new cdk.CfnOutput(this, 'ServiceArn', {
        value: serviceArn,
        description: 'ECS service ARN (legacy rollback)',
      });
    }

    new cdk.CfnOutput(this, 'CognitoUserPoolId', {
      value: userPool.userPoolId,
      description: 'Cognito User Pool ID for assistant auth',
    });

    new cdk.CfnOutput(this, 'CognitoClientId', {
      value: userPoolClient.userPoolClientId,
      description: 'Cognito Client ID for dashboard',
    });
  }
}
