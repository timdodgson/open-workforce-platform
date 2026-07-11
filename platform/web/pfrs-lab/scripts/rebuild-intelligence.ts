/**
 * CLI entry point for precomputing intelligence dashboard artifacts.
 * Run from platform/web/pfrs-lab with STORAGE_PROVIDER=local|s3.
 */
import { buildIntelligenceArtifacts } from '../src/lib/intelligence/build-artifacts';

async function main() {
  const result = await buildIntelligenceArtifacts();
  console.log(`Wrote ${result.paths.length} artifacts at ${result.generatedAt}`);
  for (const p of result.paths) console.log(`  ${p}`);
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exit(1);
});
