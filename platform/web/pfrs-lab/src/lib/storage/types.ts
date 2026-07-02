/**
 * StorageProvider — Abstraction over run data storage.
 * Implementations: LocalStorageProvider (filesystem) and S3StorageProvider (AWS).
 */
export interface RunMetadataRecord {
  runId: string;
  [key: string]: unknown;
}

export interface StorageProvider {
  /** List all available run IDs. */
  listRuns(): Promise<string[]>;

  /** Check if a specific file exists within a run. */
  exists(runId: string, filename: string): Promise<boolean>;

  /** Read a file from a run as UTF-8 string. Returns null if not found. */
  readFile(runId: string, filename: string): Promise<string | null>;

  /** Read a file from the root storage (e.g. manifest.json). Returns null if not found. */
  readRootFile(filename: string): Promise<string | null>;

  /** Write a file to a run. */
  writeFile(runId: string, filename: string, content: string): Promise<void>;

  /** Write a file to root storage. */
  writeRootFile(filename: string, content: string): Promise<void>;

  /**
   * Hide a run from listings.
   * Local: deletes the directory.
   * S3: removes the entry from manifest.json (data remains versioned in bucket).
   */
  hideRun(runId: string): Promise<void>;
}
