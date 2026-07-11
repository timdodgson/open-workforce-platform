/**
 * Audit val-* experiment matrix coverage against current storage (local or S3).
 * Run from platform/web/pfrs-lab:
 *   npx tsx scripts/audit-val-matrix.ts
 *   npx tsx scripts/audit-val-matrix.ts --json
 */
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { listRunsAsync } from '../src/lib/data-loader';
import {
  VARIATION_CONFIGS,
  allLabelsForConfig,
  countMatchingRuns,
  matrixSummary,
} from '../src/lib/experiment-matrix';

const jsonOut = process.argv.includes('--json');

async function main() {
  const runs = await listRunsAsync();
  const ids = new Set(runs.map((r) => r.id));
  const valIds = runs.filter((r) => r.id.startsWith('val-')).map((r) => r.id);

  const expected = new Set<string>();
  for (const cfg of VARIATION_CONFIGS) {
    for (const label of allLabelsForConfig(cfg)) {
      expected.add(label);
    }
  }

  const missing = [...expected].filter((l) => !ids.has(l)).sort();
  const extraVal = valIds.filter((id) => !expected.has(id)).sort();

  const byConfig = VARIATION_CONFIGS.map((cfg) => {
    const found = countMatchingRuns([...ids], cfg);
    return {
      id: cfg.id,
      tier: cfg.tier,
      domain: cfg.domain,
      labelPrefix: cfg.labelPrefix,
      found,
      expected: cfg.variationsPerConfig,
      gap: cfg.variationsPerConfig - found,
      complete: found === cfg.variationsPerConfig,
    };
  });

  const totalFound = byConfig.reduce((n, c) => n + c.found, 0);
  const totalExpected = byConfig.reduce((n, c) => n + c.expected, 0);

  const report = {
    generatedAt: new Date().toISOString(),
    storageProvider: process.env.STORAGE_PROVIDER ?? 'local',
    bucket: process.env.PFRS_S3_BUCKET ?? null,
    totalRuns: runs.length,
    valRuns: valIds.length,
    matrixFound: totalFound,
    matrixExpected: totalExpected,
    matrixGap: totalExpected - totalFound,
    matrixComplete: totalFound === totalExpected,
    summary: matrixSummary(),
    byConfig,
    missingCount: missing.length,
    missing,
    extraValCount: extraVal.length,
    extraVal,
  };

  if (jsonOut) {
    const outPath = join(process.cwd(), 'data', 'val-matrix-audit.json');
    writeFileSync(outPath, JSON.stringify(report, null, 2));
    console.log(`Wrote ${outPath}`);
    return;
  }

  console.log('val-* matrix audit');
  console.log(`  storage: ${report.storageProvider}${report.bucket ? ` (${report.bucket})` : ''}`);
  console.log(`  total runs: ${report.totalRuns}`);
  console.log(`  val-* runs: ${report.valRuns}`);
  console.log(`  matrix: ${report.matrixFound}/${report.matrixExpected} (${report.matrixGap} gap)`);
  console.log('');
  console.log('  config coverage:');
  for (const c of byConfig) {
    const mark = c.complete ? 'OK' : 'GAP';
    console.log(`    [${mark}] ${c.id}: ${c.found}/${c.expected}`);
  }
  if (missing.length > 0) {
    console.log('');
    console.log(`  missing labels (${missing.length}):`);
    for (const m of missing.slice(0, 20)) {
      console.log(`    - ${m}`);
    }
    if (missing.length > 20) {
      console.log(`    ... and ${missing.length - 20} more`);
    }
  }
  if (extraVal.length > 0) {
    console.log('');
    console.log(`  extra val-* not in matrix (${extraVal.length}):`);
    for (const e of extraVal.slice(0, 10)) {
      console.log(`    - ${e}`);
    }
  }
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exit(1);
});
