import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';

export interface ResearchDataBucketProps {
  /**
   * Optional bucket name. If not provided, CloudFormation generates one.
   * Configurable via CDK context key 'bucketName'.
   */
  bucketName?: string;
}

/**
 * ResearchDataBucket — Immutable storage for optimisation run telemetry.
 *
 * Design principles:
 * - Versioning enabled: every run is immutable, overwrites are tracked.
 * - Encryption: S3 managed (SSE-S3) by default.
 * - No public access: all access via authenticated AWS principals.
 * - RETAIN on deletion: data survives stack teardown.
 * - Intelligent Tiering: older runs move to cheaper storage automatically.
 * - SSL required: no unencrypted transport.
 *
 * Bucket layout (application convention, not enforced by CDK):
 *   manifest.json
 *   runs/<run-id>/metadata.json
 *   runs/<run-id>/summary.json
 *   runs/<run-id>/discoveries.csv
 *   runs/<run-id>/tree.csv
 *   runs/<run-id>/workers.csv
 *   runs/<run-id>/dashboard/index.html
 *   versions/
 */
export class ResearchDataBucket extends Construct {
  public readonly bucket: s3.Bucket;

  constructor(scope: Construct, id: string, props: ResearchDataBucketProps = {}) {
    super(scope, id);

    this.bucket = new s3.Bucket(this, 'Bucket', {
      bucketName: props.bucketName,
      versioned: true,
      encryption: s3.BucketEncryption.S3_MANAGED,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      objectOwnership: s3.ObjectOwnership.BUCKET_OWNER_ENFORCED,
      enforceSSL: true,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      autoDeleteObjects: false,

      // Intelligent Tiering: move infrequently accessed runs to cheaper storage.
      lifecycleRules: [
        {
          id: 'IntelligentTiering',
          enabled: true,
          transitions: [
            {
              storageClass: s3.StorageClass.INTELLIGENT_TIERING,
              transitionAfter: cdk.Duration.days(30),
            },
          ],
        },
      ],

      // CORS: allow dashboard to fetch CSV/JSON from S3 directly (Phase 2).
      cors: [
        {
          allowedMethods: [s3.HttpMethods.GET, s3.HttpMethods.HEAD],
          allowedOrigins: ['*'], // Tightened in Phase 2 when CloudFront is added.
          allowedHeaders: ['*'],
          maxAge: 3600,
        },
      ],
    });
  }
}
