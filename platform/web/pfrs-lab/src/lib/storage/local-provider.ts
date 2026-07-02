import { existsSync, readdirSync, readFileSync, writeFileSync, mkdirSync, rmSync } from 'fs';
import path from 'path';
import { StorageProvider } from './types';

/**
 * LocalStorageProvider — Reads/writes run data from the local filesystem.
 * Uses the existing data/runs/ directory structure.
 * Behaviour is identical to the pre-abstraction implementation.
 */
export class LocalStorageProvider implements StorageProvider {
  private readonly baseDir: string;

  constructor(baseDir?: string) {
    this.baseDir = baseDir ?? path.join(process.cwd(), 'data', 'runs');
  }

  async listRuns(): Promise<string[]> {
    if (!existsSync(this.baseDir)) return [];
    const entries = readdirSync(this.baseDir, { withFileTypes: true });
    return entries.filter(e => e.isDirectory()).map(e => e.name);
  }

  async exists(runId: string, filename: string): Promise<boolean> {
    const filePath = path.join(this.baseDir, runId, filename);
    return existsSync(filePath);
  }

  async readFile(runId: string, filename: string): Promise<string | null> {
    const filePath = path.join(this.baseDir, runId, filename);
    if (!existsSync(filePath)) return null;
    return readFileSync(filePath, 'utf-8');
  }

  async readRootFile(filename: string): Promise<string | null> {
    const filePath = path.join(this.baseDir, '..', filename);
    if (!existsSync(filePath)) return null;
    return readFileSync(filePath, 'utf-8');
  }

  async writeFile(runId: string, filename: string, content: string): Promise<void> {
    const dir = path.join(this.baseDir, runId);
    mkdirSync(dir, { recursive: true });
    writeFileSync(path.join(dir, filename), content, 'utf-8');
  }

  async writeRootFile(filename: string, content: string): Promise<void> {
    const filePath = path.join(this.baseDir, '..', filename);
    writeFileSync(filePath, content, 'utf-8');
  }

  async hideRun(runId: string): Promise<void> {
    // Local: actually delete (no versioning on filesystem).
    const dir = path.join(this.baseDir, runId);
    if (existsSync(dir)) {
      rmSync(dir, { recursive: true, force: true });
    }
  }
}
