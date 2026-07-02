import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { ResearchDataBucket } from '../constructs/research-data-bucket';

/**
 * PfrsResearchLabStack — Phase 1
 *
 * Provisions the core S3 storage backend for the PFRS Research Lab.
 * Designed to be extended in later phases with CloudFront, API Gateway,
 * Cognito, and Lambda without modifying the bucket architecture.
 */
export class PfrsResearchLabStack extends cdk.Stack {
  /** The research data bucket — long-term storage for all optimisation runs. */
  public readonly dataBucket: ResearchDataBucket;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Bucket name from CDK context (overridable via cdk.json or --context).
    const bucketName = this.node.tryGetContext('bucketName') as string | undefined;

    this.dataBucket = new ResearchDataBucket(this, 'ResearchData', {
      bucketName,
    });

    // --- Outputs ---
    new cdk.CfnOutput(this, 'BucketName', {
      value: this.dataBucket.bucket.bucketName,
      description: 'S3 bucket name for research run storage',
      exportName: 'PfrsResearchLab-BucketName',
    });

    new cdk.CfnOutput(this, 'BucketArn', {
      value: this.dataBucket.bucket.bucketArn,
      description: 'S3 bucket ARN for research run storage',
      exportName: 'PfrsResearchLab-BucketArn',
    });
  }
}
