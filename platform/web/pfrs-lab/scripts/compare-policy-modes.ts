/**
 * Step 1 ML harness — compare policy modes from live storage (local or S3).
 * Writes docs/reports/ml-harness/latest.json
 */
import { mkdirSync, readFileSync, writeFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { getStorageProvider } from '../src/lib/storage';

type RunRow = {
  label: string;
  domain: string;
  instance: string;
  algorithm: string;
  policyMode: string;
  objective: number;
  runtimeMs: number;
  feasible: boolean;
};

function detectDomain(label: string): string {
  const l = label.toLowerCase();
  if (l.includes('cvrp')) return 'cvrp';
  if (l.includes('jss') || l.includes('jobshop')) return 'jss';
  if (l.includes('vrptw')) return 'vrptw';
  return 'nrp';
}

function inferPolicy(label: string, meta: Record<string, unknown>): string {
  const pm = meta.policyMode ?? meta.policy_mode;
  if (typeof pm === 'string' && pm) return pm;
  for (const mode of ['rules', 'hybrid', 'learned']) {
    if (label.includes(`-${mode}-`)) return mode;
  }
  return 'unknown';
}

function inferAlgorithm(label: string, meta: Record<string, unknown>): string {
  if (typeof meta.mode === 'string') return meta.mode;
  for (const alg of ['portfolio', 'tabu', 'lahc', 'sa']) {
    if (label.includes(`-${alg}-`)) return alg;
  }
  return 'unknown';
}

function objective(meta: Record<string, unknown>): number {
  for (const key of ['bestObjective', 'bestDistance', 'bestPenalty', 'totalPenalty', 'objective']) {
    const v = meta[key];
    if (typeof v === 'number') return v;
  }
  return 0;
}

function loadValidationGates(): {
  falseStopRate?: number;
  step4PromoteOk?: boolean;
  counterfactualSamples?: number;
} {
  const candidates = [
    join(process.cwd(), '..', '..', 'ml', 'policies', 'validation_results.json'),
    join(process.cwd(), 'data', 'validation_results.json'),
  ];
  for (const p of candidates) {
    if (!existsSync(p)) continue;
    try {
      const raw = JSON.parse(readFileSync(p, 'utf-8')) as Record<string, unknown>;
      const cf = raw.counterfactual as Record<string, unknown> | undefined;
      return {
        falseStopRate: Number(cf?.false_stop_rate ?? raw.false_stop_rate ?? 1),
        step4PromoteOk: Boolean(cf?.promotion_ready ?? raw.step4_promotion_ready),
        counterfactualSamples: Number(cf?.samples ?? 0),
      };
    } catch {
      /* try next */
    }
  }
  return {};
}

function mean(xs: number[]): number {
  return xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : 0;
}

async function loadRows(prefix: string): Promise<RunRow[]> {
  const storage = getStorageProvider();
  const ids = await storage.listRuns();
  const rows: RunRow[] = [];
  for (const id of ids) {
    if (prefix && !id.startsWith(prefix)) continue;
    const raw = await storage.readFile(id, 'run.json');
    if (!raw) continue;
    try {
      const meta = JSON.parse(raw) as Record<string, unknown>;
      rows.push({
        label: id,
        domain: String(meta.problemType ?? detectDomain(id)),
        instance: String(meta.instance ?? ''),
        algorithm: inferAlgorithm(id, meta),
        policyMode: inferPolicy(id, meta),
        objective: objective(meta),
        runtimeMs: Number(meta.runtimeMs ?? meta.runtime_ms ?? 0),
        feasible: meta.feasible !== false,
      });
    } catch {
      /* skip */
    }
  }
  return rows;
}

function buildReport(rows: RunRow[], prefix: string, mlMaturity: number) {
  const byMode = new Map<string, RunRow[]>();
  for (const r of rows) {
    const g = byMode.get(r.policyMode) ?? [];
    g.push(r);
    byMode.set(r.policyMode, g);
  }

  const modeSummaries = [...byMode.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([policyMode, group]) => ({
      policyMode,
      n: group.length,
      meanObjective: Math.round(mean(group.map((r) => r.objective)) * 100) / 100,
      meanRuntimeMs: Math.round(mean(group.map((r) => r.runtimeMs)) * 10) / 10,
      feasibilityRate: group.filter((r) => r.feasible).length / group.length,
    }));

  const byCfg = new Map<string, Map<string, RunRow[]>>();
  for (const r of rows) {
    const key = `${r.domain}|${r.instance}|${r.algorithm}`;
    if (!byCfg.has(key)) byCfg.set(key, new Map());
    const modes = byCfg.get(key)!;
    const g = modes.get(r.policyMode) ?? [];
    g.push(r);
    modes.set(r.policyMode, g);
  }

  const comparisons: Record<string, unknown>[] = [];
  let qualityWins = 0;
  let runtimeWins = 0;

  for (const modes of byCfg.values()) {
    const rules = modes.get('rules');
    if (!rules?.length) continue;
    for (const modeB of ['hybrid', 'learned'] as const) {
      const groupB = modes.get(modeB);
      if (!groupB?.length) continue;
      const meanA = mean(rules.map((r) => r.objective));
      const meanB = mean(groupB.map((r) => r.objective));
      const rtA = mean(rules.map((r) => r.runtimeMs));
      const rtB = mean(groupB.map((r) => r.runtimeMs));
      const saved = rtA > 0 ? ((rtA - rtB) / rtA) * 100 : 0;
      const delta = meanB - meanA;
      let verdict = 'equivalent';
      if (delta < -0.01) verdict = 'better';
      else if (delta > 0.01) verdict = 'worse';
      const rtVerdict = saved > 2 ? 'faster' : saved < -2 ? 'slower' : 'equivalent';
      let roi = (meanA !== 0 ? ((meanA - meanB) / Math.abs(meanA)) * 100 : 0) + saved;
      if (verdict === 'worse') roi = -Math.abs(roi);

      comparisons.push({
        domain: rules[0].domain,
        instance: rules[0].instance,
        algorithm: rules[0].algorithm,
        modeA: 'rules',
        modeB,
        meanObjectiveA: Math.round(meanA * 100) / 100,
        meanObjectiveB: Math.round(meanB * 100) / 100,
        objectiveDelta: Math.round(delta * 100) / 100,
        meanRuntimeA: Math.round(rtA * 10) / 10,
        meanRuntimeB: Math.round(rtB * 10) / 10,
        runtimeSavedPct: Math.round(saved * 100) / 100,
        verdict,
        runtimeVerdict: rtVerdict,
        roi: Math.round(roi * 100) / 100,
      });

      if (verdict === 'better' || verdict === 'equivalent') qualityWins += 1;
      if (rtVerdict === 'faster' && verdict !== 'worse') runtimeWins += 1;
    }
  }

  const validation = loadValidationGates();

  return {
    generatedAt: new Date().toISOString(),
    step: 4,
    mlMaturity,
    totalRuns: rows.length,
    prefix,
    modeSummaries,
    comparisons,
    gates: {
      step1HarnessOk: rows.length > 0,
      step2QualityWins: qualityWins,
      step2RuntimeWins: runtimeWins,
      step2PromoteOk: qualityWins >= 2 || runtimeWins >= 2,
      step4FalseStopRate: validation.falseStopRate,
      step4CounterfactualSamples: validation.counterfactualSamples ?? 0,
      step4PromoteOk: validation.step4PromoteOk ?? false,
    },
  };
}

async function main() {
  const prefix = process.env.HARNESS_PREFIX ?? 'val-';
  const mlMaturity = Number(process.env.ML_MATURITY ?? '6');
  const rows = await loadRows(prefix);
  const report = buildReport(rows, prefix, mlMaturity);

  const root = process.cwd();
  const relPaths = [
    join(root, '..', '..', '..', 'docs', 'reports', 'ml-harness', 'latest.json'),
    join(root, 'data', 'ml-harness-latest.json'),
  ];

  const json = JSON.stringify(report, null, 2);
  for (const outPath of relPaths) {
    try {
      mkdirSync(join(outPath, '..'), { recursive: true });
      writeFileSync(outPath, json);
      console.log(`Wrote ${outPath}`);
    } catch {
      /* try next */
    }
  }

  console.log(`ML harness — ${report.totalRuns} runs`);
  for (const s of report.modeSummaries) {
    console.log(`  ${s.policyMode.padEnd(8)} n=${String(s.n).padStart(3)}  obj=${s.meanObjective}  runtime=${s.meanRuntimeMs}ms`);
  }
  console.log('  gates:', report.gates);
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exit(1);
});
