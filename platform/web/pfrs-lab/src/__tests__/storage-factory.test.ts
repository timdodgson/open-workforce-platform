/**
 * Tests for storage provider factory.
 * Verifies correct provider selection based on environment variable.
 */

describe('Storage Factory', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...originalEnv };
  });

  afterAll(() => {
    process.env = originalEnv;
  });

  it('returns LocalStorageProvider when STORAGE_PROVIDER=local', async () => {
    process.env.STORAGE_PROVIDER = 'local';
    const { getStorageProvider } = await import('@/lib/storage/factory');
    const provider = getStorageProvider();
    expect(provider).toBeDefined();
    expect(provider.constructor.name).toBe('LocalStorageProvider');
  });

  it('defaults to local when env var not set', async () => {
    delete process.env.STORAGE_PROVIDER;
    const { getStorageProvider } = await import('@/lib/storage/factory');
    const provider = getStorageProvider();
    expect(provider.constructor.name).toBe('LocalStorageProvider');
  });

  it('returns S3StorageProvider when STORAGE_PROVIDER=s3', async () => {
    process.env.STORAGE_PROVIDER = 's3';
    const { getStorageProvider } = await import('@/lib/storage/factory');
    const provider = getStorageProvider();
    expect(provider.constructor.name).toBe('S3StorageProvider');
  });
});
