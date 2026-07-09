import type { BenchmarkRun } from '@/app/benchmarks/page';
import type { PolicyMode, SiValidationSpec } from '@/lib/benchmark-suites';
import { BENCHMARK_POLICIES } from '@/lib/benchmark-suites';

export interface SiTripletCoverage {
  expected: number;
  complete: number;
  partial: number;
  byPolicy: Record<PolicyMode, number>;
  /** instance:mode:seed groups with at least one val run */
  groupsWithData: number;
}

function parsePolicyMode(run: BenchmarkRun): PolicyMode | null {
  const fromMeta = run.policyMode?.toLowerCase();
  if (BENCHMARK_POLICIES.includes(fromMeta as PolicyMode)) return fromMeta as PolicyMode;
  const m = run.id.match(/-(rules|hybrid|learned)-s\d+$/);
  return m ? (m[1] as PolicyMode) : null;
}

function normaliseValMode(mode: string): string {
  const lower = mode.toLowerCase();
  if (lower.includes('portfolio')) return 'portfolio';
  if (lower === 'tabu') return 'tabu';
  if (lower === 'sa') return 'sa';
  if (lower === 'lahc') return 'lahc';
  return lower;
}

function instanceMatches(runInstance: string, specInstance: string): boolean {
  const a = runInstance.toLowerCase().replace(/\.\w+$/, '');
  const b = specInstance.toLowerCase().replace(/\.\w+$/, '');
  if (a === b) return true;
  // val labels embed instance slugs: a32k5, n012w8, la01, c101, ft10, a80k10
  const slug = b.replace(/[^a-z0-9]/gi, '');
  return a.includes(slug) || runInstance.toLowerCase().includes(slug);
}

export function computeSiTripletCoverage(
  runs: BenchmarkRun[],
  problemType: string,
  spec: SiValidationSpec,
): SiTripletCoverage {
  const expected =
    spec.instances.length * spec.modes.length * spec.seeds.length;

  const valRuns = runs.filter(
    (r) =>
      r.problemType === problemType &&
      r.id.startsWith(spec.labelPrefix) &&
      parsePolicyMode(r) !== null,
  );

  type Group = { rules: boolean; hybrid: boolean; learned: boolean };
  const groups = new Map<string, Group>();

  const byPolicy: Record<PolicyMode, number> = { rules: 0, hybrid: 0, learned: 0 };

  for (const run of valRuns) {
    const policy = parsePolicyMode(run)!;
    byPolicy[policy]++;

    const mode = normaliseValMode(run.mode);
    if (!spec.modes.includes(mode)) continue;

    const inst = spec.instances.find((i) => instanceMatches(run.instance, i));
    if (!inst) continue;
    if (!spec.seeds.includes(run.seed)) continue;

    const key = `${inst}:${mode}:${run.seed}`;
    const g = groups.get(key) ?? { rules: false, hybrid: false, learned: false };
    g[policy] = true;
    groups.set(key, g);
  }

  let complete = 0;
  let partial = 0;
  for (const g of groups.values()) {
    const n = (g.rules ? 1 : 0) + (g.hybrid ? 1 : 0) + (g.learned ? 1 : 0);
    if (n === 3) complete++;
    else if (n > 0) partial++;
  }

  return {
    expected,
    complete,
    partial,
    byPolicy,
    groupsWithData: groups.size,
  };
}
