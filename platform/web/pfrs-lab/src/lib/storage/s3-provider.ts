import { StorageProvider } from './types';
import { S3Client, ListObjectsV2Command, GetObjectCommand, HeadObjectCommand, PutObjectCommand } from '@aws-sdk/client-s3';

/**
 * S3StorageProvider — Reads/writes run data from an S3 bucket.
 * Uses AWS SDK v3 with the default credential chain.
 */
export class S3StorageProvider implements StorageProvider {
  private readonly bucketName: string;
  private s3Client: S3Client;

  constructor(bucketName?: string) {
    this.bucketName = bucketName ?? process.env.PFRS_S3_BUCKET ?? 'pfrs-research-lab-data';
    this.s3Client = new S3Client({
      region: process.env.AWS_REGION ?? 'eu-west-1',
    });
  }

  async listRuns(): Promise<string[]> {
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
      await this.s3Client.send(new HeadObjectCommand({
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
      const response = await this.s3Client.send(new GetObjectCommand({
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
      const response = await this.s3Client.send(new GetObjectCommand({
        Bucket: this.bucketName,
        Key: filename,
      }));
      return await response.Body?.transformToString('utf-8') ?? null;
    } catch {
      return null;
    }
  }

  async writeFile(runId: string, filename: string, content: string): Promise<void> {
    await this.s3Client.send(new PutObjectCommand({
      Bucket: this.bucketName,
      Key: `runs/${runId}/${filename}`,
      Body: content,
      ContentType: filename.endsWith('.json') ? 'application/json' : 'text/csv',
    }));
  }

  async writeRootFile(filename: string, content: string): Promise<void> {
    await this.s3Client.send(new PutObjectCommand({
      Bucket: this.bucketName,
      Key: filename,
      Body: content,
      ContentType: filename.endsWith('.json') ? 'application/json' : 'text/plain',
    }));
  }

  async hideRun(runId: string): Promise<void> {
    const manifestContent = await this.readRootFile('manifest.json');
    if (!manifestContent) {
      throw new Error(`Cannot hide run '${runId}': manifest.json not found`);
    }

    const manifest = JSON.parse(manifestContent);
    if (!manifest.runs || !Array.isArray(manifest.runs)) {
      throw new Error(`Cannot hide run '${runId}': manifest.json has no runs array`);
    }

    const before = manifest.runs.length;
    manifest.runs = manifest.runs.filter((r: any) => r.runId !== runId);

    if (manifest.runs.length === before) {
      // Run wasn't in the manifest — nothing to do, not an error.
      return;
    }

    await this.writeRootFile('manifest.json', JSON.stringify(manifest, null, 2));
  }
}
