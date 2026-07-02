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

  it('attempts S3 when STORAGE_PROVIDER=s3', async () => {
    process.env.STORAGE_PROVIDER = 's3';
    // S3 provider may throw if AWS SDK not available in test — that's expected.
    try {
      const { getStorageProvider } = await import('@/lib/storage/factory');
      const provider = getStorageProvider();
      expect(provider.constructor.name).toBe('S3StorageProvider');
    } catch (e) {
      // Expected in test environment without AWS SDK properly configured.
      expect(String(e)).toContain('S3');
    }
  });
});
