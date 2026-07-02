/**
 * Jest mock for @aws-sdk/client-s3.
 * Provides minimal stubs so the S3StorageProvider can be instantiated in tests.
 */
export class S3Client {
  constructor(_config: any) {}
  send(_command: any) { return Promise.resolve({}); }
}

export class ListObjectsV2Command {
  constructor(_input: any) {}
}

export class GetObjectCommand {
  constructor(_input: any) {}
}

export class HeadObjectCommand {
  constructor(_input: any) {}
}

export class PutObjectCommand {
  constructor(_input: any) {}
}
