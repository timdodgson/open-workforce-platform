#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { PfrsResearchLabStack } from './stacks/pfrs-research-lab-stack';

const app = new cdk.App();

new PfrsResearchLabStack(app, 'PfrsResearchLabStack', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION ?? 'eu-west-1',
  },
  description: 'PFRS Research Lab — S3 storage for optimisation run telemetry',
});
