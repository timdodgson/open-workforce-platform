import { StorageProvider } from './types';
import { LocalStorageProvider } from './local-provider';
import { S3StorageProvider } from './s3-provider';

/**
 * StorageFactory — Returns the configured storage provider.
 *
 * Configuration via environment variable:
 *   STORAGE_PROVIDER=local  (default)
 *   STORAGE_PROVIDER=s3
 *
 * For S3, also set:
 *   PFRS_S3_BUCKET=pfrs-research-lab-data  (optional, has default)
 *   AWS_REGION=eu-west-1                    (optional, has default)
 */

let instance: StorageProvider | null = null;

export function getStorageProvider(): StorageProvider {
  if (instance) return instance;

  const provider = process.env.STORAGE_PROVIDER ?? 'local';

  switch (provider) {
    case 's3':
      instance = new S3StorageProvider();
      break;
    case 'local':
    default:
      instance = new LocalStorageProvider();
      break;
  }

  return instance;
}
