import { StorageProvider } from './types';

/**
 * S3StorageProvider — Reads/writes run data from an S3 bucket.
 * Uses AWS SDK v3 with the default credential chain.
 *
 * IMPORTANT: Requires @aws-sdk/client-s3 to be installed.
 * Install with: npm install @aws-sdk/client-s3
 *
 * This file uses dynamic require() to avoid breaking builds when the SDK is not installed.
 */
export class S3StorageProvider implements StorageProvider {
  private readonly bucketName: string;
  private s3Client: any = null;
  private sdk: any = null;

  constructor(bucketName?: string) {
    this.bucketName = bucketName ?? process.env.PFRS_S3_BUCKET ?? 'pfrs-research-lab-data';
    try {
      // eslint-disable-next-line @typescript-eslint/no-require-imports
      this.sdk = require('@aws-sdk/client-s3');
      this.s3Client = new this.sdk.S3Client({
        region: process.env.AWS_REGION ?? 'eu-west-1',
      });
    } catch {
      throw new Error('S3StorageProvider requires @aws-sdk/client-s3. Install with: npm install @aws-sdk/client-s3');
    }
  }

  async listRuns(): Promise<string[]> {
    // Use manifest as source of truth for visible runs.
    const content = await this.readRootFile('manifest.json');
    if (!content) return [];
    try {
      const manifest = JSON.parse(content);
      if (manifest.runs && Array.isArray(manifest.runs)) {
        return manifest.runs.map((r: any) => r.runId).filter((id: string) => id);
      }
    } catch { /* ignore */ }
    return [];
  }

  async exists(runId: string, filename: string): Promise<boolean> {
    try {
      await this.s3Client.send(new this.sdk.HeadObjectCommand({
        Bucket: this.bucketName,
        Key: `runs/${runId}/${filename}`,
      }));
      return true;
    } catch {
      return false;
    }
  }

  async readFile(runId: string, filename: string): Promise<string | null> {
    try {
      const response = await this.s3Client.send(new this.sdk.GetObjectCommand({
        Bucket: this.bucketName,
        Key: `runs/${runId}/${filename}`,
      }));
      return await response.Body?.transformToString('utf-8') ?? null;
    } catch {
      return null;
    }
  }

  async readRootFile(filename: string): Promise<string | null> {
    try {
      const response = await this.s3Client.send(new this.sdk.GetObjectCommand({
        Bucket: this.bucketName,
        Key: filename,
      }));
      return await response.Body?.transformToString('utf-8') ?? null;
    } catch {
      return null;
    }
  }

  async writeFile(runId: string, filename: string, content: string): Promise<void> {
    await this.s3Client.send(new this.sdk.PutObjectCommand({
      Bucket: this.bucketName,
      Key: `runs/${runId}/${filename}`,
      Body: content,
      ContentType: filename.endsWith('.json') ? 'application/json' : 'text/csv',
    }));
  }

  async writeRootFile(filename: string, content: string): Promise<void> {
    await this.s3Client.send(new this.sdk.PutObjectCommand({
      Bucket: this.bucketName,
      Key: filename,
      Body: content,
      ContentType: filename.endsWith('.json') ? 'application/json' : 'text/plain',
    }));
  }

  async hideRun(runId: string): Promise<void> {
    // Soft delete: remove from manifest only. Data remains versioned in S3.
    const manifestContent = await this.readRootFile('manifest.json');
    if (!manifestContent) return;

    try {
      const manifest = JSON.parse(manifestContent);
      if (manifest.runs && Array.isArray(manifest.runs)) {
        manifest.runs = manifest.runs.filter((r: any) => r.runId !== runId);
        await this.writeRootFile('manifest.json', JSON.stringify(manifest, null, 2));
      }
    } catch {
      // If manifest is corrupt, nothing to remove.
    }
  }
}
