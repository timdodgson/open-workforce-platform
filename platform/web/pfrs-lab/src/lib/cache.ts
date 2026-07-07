/**
 * Simple in-memory cache with TTL.
 *
 * Used to avoid re-reading the filesystem/S3 on every request.
 * Entries expire after `ttlMs` milliseconds.
 */

interface CacheEntry<T> {
  value: T;
  expiresAt: number;
}

const store = new Map<string, CacheEntry<unknown>>();

const DEFAULT_TTL_MS = 60_000; // 60 seconds

/**
 * Get a value from cache, or compute and store it if missing/expired.
 */
export async function cached<T>(
  key: string,
  compute: () => Promise<T>,
  ttlMs: number = DEFAULT_TTL_MS
): Promise<T> {
  const now = Date.now();
  const existing = store.get(key) as CacheEntry<T> | undefined;

  if (existing && existing.expiresAt > now) {
    return existing.value;
  }

  const value = await compute();
  store.set(key, { value, expiresAt: now + ttlMs });
  return value;
}

/**
 * Invalidate a specific cache key.
 */
export function invalidate(key: string): void {
  store.delete(key);
}

/**
 * Invalidate all cache entries.
 */
export function invalidateAll(): void {
  store.clear();
}
