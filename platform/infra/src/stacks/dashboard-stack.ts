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
 * DashboardStack - ECS Fargate service for PFRS Research Lab dashboard.
 *
 * Architecture:
 *   ECR -> Fargate (Next.js container) -> S3 (reads run data)
 *   ALB provides stable HTTPS endpoint.
 */
export class DashboardStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const bucketName = this.node.tryGetContext('bucketName') || 'pfrs-research-lab-data';

    // --- ECR Repository ---
    const repo = new ecr.Repository(this, 'DashboardRepo', {
      repositoryName: 'pfrs-lab-dashboard',
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      lifecycleRules: [{ maxImageCount: 10 }],
    });

    // --- VPC (default, no cost) ---
    const vpc = new ec2.Vpc(this, 'Vpc', {
      maxAzs: 2,
      natGateways: 0, // No NAT to save cost; tasks use public subnet.
      subnetConfiguration: [
        { name: 'Public', subnetType: ec2.SubnetType.PUBLIC, cidrMask: 24 },
      ],
    });

    // --- ECS Cluster ---
    const cluster = new ecs.Cluster(this, 'Cluster', { vpc });

    // --- Task Definition ---
    const taskDef = new ecs.FargateTaskDefinition(this, 'TaskDef', {
      cpu: 256,    // 0.25 vCPU
      memoryLimitMiB: 512,
    });

    // Grant S3 read/write access. Write needed for manifest management (hide runs).
    const dataBucket = s3.Bucket.fromBucketName(this, 'DataBucket', bucketName);
    dataBucket.grantReadWrite(taskDef.taskRole);

    // Grant Bedrock invoke access for the optimisation assistant.
    taskDef.taskRole.addToPrincipalPolicy(new iam.PolicyStatement({
      actions: ['bedrock:InvokeModel', 'aws-marketplace:ViewSubscriptions', 'aws-marketplace:Subscribe'],
      resources: ['*'],
    }));

    // --- Cognito User Pool (for assistant auth) ---
    const userPool = new cognito.UserPool(this, 'AssistantUserPool', {
      userPoolName: 'pfrs-lab-users',
      selfSignUpEnabled: false, // Admin creates users only.
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

    // Container.
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

    // --- ALB ---
    const alb = new elbv2.ApplicationLoadBalancer(this, 'ALB', {
      vpc,
      internetFacing: true,
    });

    const listener = alb.addListener('HTTP', {
      port: 80,
      protocol: elbv2.ApplicationProtocol.HTTP,
    });

    // --- Fargate Service ---
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

    // --- GitHub Actions OIDC + Deploy Role ---
    const githubOrg = this.node.tryGetContext('githubOrg') || 'timdodgson';
    const githubRepo = this.node.tryGetContext('githubRepo') || 'open-workforce-platform';

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
        'sts:AssumeRoleWithWebIdentity'
      ),
      description: 'GitHub Actions deploy role for PFRS Lab',
    });

    repo.grantPush(deployRole);
    deployRole.addToPolicy(new iam.PolicyStatement({
      actions: ['ecr:GetAuthorizationToken'],
      resources: ['*'],
    }));
    // Allow force new deployment.
    deployRole.addToPolicy(new iam.PolicyStatement({
      actions: ['ecs:UpdateService', 'ecs:DescribeServices'],
      resources: [service.serviceArn],
    }));

    // OpenNext / SST pilot deploy (CloudFront + Lambda + supporting resources).
    // Scoped loosely: SST creates ephemeral stacks named pfrs-lab-production / sst-* .
    deployRole.addToPolicy(new iam.PolicyStatement({
      actions: [
        'cloudformation:CreateStack',
        'cloudformation:UpdateStack',
        'cloudformation:DeleteStack',
        'cloudformation:DescribeStacks',
        'cloudformation:DescribeStackEvents',
        'cloudformation:DescribeStackResources',
        'cloudformation:GetTemplate',
        'cloudformation:CreateChangeSet',
        'cloudformation:DeleteChangeSet',
        'cloudformation:DescribeChangeSet',
        'cloudformation:ExecuteChangeSet',
        'cloudformation:ListStacks',
        'cloudformation:ListStackResources',
      ],
      resources: ['*'],
    }));
    deployRole.addToPolicy(new iam.PolicyStatement({
      actions: [
        'lambda:*',
        'iam:CreateRole',
        'iam:DeleteRole',
        'iam:GetRole',
        'iam:PassRole',
        'iam:AttachRolePolicy',
        'iam:DetachRolePolicy',
        'iam:PutRolePolicy',
        'iam:DeleteRolePolicy',
        'iam:GetRolePolicy',
        'iam:TagRole',
        'iam:UntagRole',
        'iam:CreateServiceLinkedRole',
        'logs:*',
        's3:*',
        'cloudfront:*',
        // OpenNext ISR revalidation (SST creates these automatically).
        'sqs:*',
        'dynamodb:*',
        'ssm:GetParameter',
        'ssm:GetParameters',
        'ssm:PutParameter',
        'ssm:DeleteParameter',
        'ssm:AddTagsToResource',
        'ssm:RemoveTagsFromResource',
        'ssm:DescribeParameters',
        'route53:ListHostedZones',
        'route53:ChangeResourceRecordSets',
        'route53:GetChange',
        'route53:ListResourceRecordSets',
      ],
      resources: ['*'],
    }));

    // --- Outputs ---
    new cdk.CfnOutput(this, 'DashboardUrl', {
      value: `http://${alb.loadBalancerDnsName}`,
      description: 'Dashboard URL',
    });
    new cdk.CfnOutput(this, 'EcrRepoUri', {
      value: repo.repositoryUri,
      description: 'ECR repo for pushing images',
    });
    new cdk.CfnOutput(this, 'DeployRoleArn', {
      value: deployRole.roleArn,
      description: 'Add as AWS_DEPLOY_ROLE_ARN GitHub secret',
    });
    new cdk.CfnOutput(this, 'ServiceArn', {
      value: service.serviceArn,
      description: 'ECS service ARN for force deploy',
    });
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
