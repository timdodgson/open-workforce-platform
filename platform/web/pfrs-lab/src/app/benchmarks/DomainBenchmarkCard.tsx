'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import type { BenchmarkRun } from './page';
import type { BenchmarkSuite } from '@/lib/benchmark-suites';
import { computeSiTripletCoverage } from '@/lib/benchmark-si-coverage';
import { BATCH_SCRIPTS, commandsForDomain, GO_CWD } from '@/lib/benchmark-commands';

type ModeKey = 'sa' | 'lahc' | 'tabu' | 'portfolio' | 'adaptive' | 'ilp';

interface Row {
  instance: string;
  sa: number | null;
  lahc: number | null;
  tabu: number | null;
  portfolio: number | null;
  adaptive: number | null;
  reference: number | null;
  best: number | null;
  gapPct: number | null;
  winner: string;
}

function normaliseMode(mode: string): ModeKey {
  const lower = mode.toLowerCase();
  if (lower === 'adaptive' || lower === 'hyper-heuristic') return 'adaptive';
  if (lower.includes('portfolio') || lower.includes('sa,lahc') || lower.includes('sa,lahc,tabu')) return 'portfolio';
  if (lower === 'sa' || lower === 'simulated-annealing') return 'sa';
  if (lower === 'lahc' || lower === 'late-acceptance') return 'lahc';
  if (lower === 'tabu' || lower === 'tabu-search') return 'tabu';
  if (lower === 'ilp') return 'ilp';
  return lower as ModeKey;
}

function pickBest(values: Array<number | null>): number | null {
  const v = values.filter((x): x is number => x !== null);
  return v.length ? Math.min(...v) : null;
}

function winnerLabel(row: Row): string {
  if (row.best === null) return '—';
  if (row.adaptive === row.best) return 'Adaptive';
  if (row.portfolio === row.best) return 'Portfolio';
  if (row.lahc === row.best) return 'LAHC';
  if (row.sa === row.best) return 'SA';
  if (row.tabu === row.best) return 'Tabu';
  return '—';
}

function CoverageBadge({ complete, expected, label }: { complete: number; expected: number; label: string }) {
  const pct = expected > 0 ? complete / expected : 0;
  const colour =
    pct >= 1 ? 'bg-emerald-900/60 text-emerald-300 border-emerald-700'
      : pct >= 0.5 ? 'bg-amber-900/40 text-amber-300 border-amber-800'
        : 'bg-gray-800 text-gray-400 border-gray-700';
  return (
    <span className={`text-[10px] px-2 py-0.5 rounded border ${colour}`}>
      {label}: {complete}/{expected} triplets
    </span>
  );
}

function PolicyBadge({ policy, count, expected }: { policy: string; count: number; expected: number }) {
  const ok = count >= expected;
  return (
    <span
      className={`text-[9px] px-1.5 py-0.5 rounded font-mono ${
        ok ? 'bg-emerald-950 text-emerald-400' : count > 0 ? 'bg-amber-950 text-amber-400' : 'bg-gray-900 text-gray-600'
      }`}
    >
      {policy} {count}/{expected}
    </span>
  );
}

export default function DomainBenchmarkCard({
  suite,
  runs,
  knownOptimal,
}: {
  suite: BenchmarkSuite;
  runs: BenchmarkRun[];
  knownOptimal: Record<string, { value: number; source: string }>;
}) {
  const { rows, coverage, siFast, siDeep } = useMemo(() => {
    const suiteRuns = runs
      .filter((r) => r.problemType === suite.id)
      .filter((r) => suite.instances.includes(r.instance));

    const byInstance = new Map<string, BenchmarkRun[]>();
    for (const r of suiteRuns) {
      const list = byInstance.get(r.instance) ?? [];
      list.push(r);
      byInstance.set(r.instance, list);
    }

    const out: Row[] = [];
    for (const instance of suite.instances) {
      const instRuns = byInstance.get(instance) ?? [];
      const bestByMode = new Map<ModeKey, number>();
      for (const r of instRuns) {
        const m = normaliseMode(r.mode);
        if (!bestByMode.has(m) || (r.penalty > 0 && r.penalty < (bestByMode.get(m) ?? Infinity))) {
          bestByMode.set(m, r.penalty);
        }
      }

      const sa = bestByMode.get('sa') ?? null;
      const lahc = bestByMode.get('lahc') ?? null;
      const tabu = bestByMode.get('tabu') ?? null;
      const portfolio = bestByMode.get('portfolio') ?? null;
      const adaptive = bestByMode.get('adaptive') ?? null;

      const known = knownOptimal[instance];
      const reference = known?.value ?? (bestByMode.get('ilp') ?? null);
      const best = pickBest([sa, lahc, tabu, portfolio, adaptive]);
      const gapPct = reference && best ? ((best - reference) / reference) * 100 : null;

      out.push({
        instance,
        sa,
        lahc,
        tabu,
        portfolio,
        adaptive,
        reference,
        best,
        gapPct,
        winner: '—',
      });
    }

    for (const r of out) r.winner = winnerLabel(r);

    const expectedRuns = suite.instances.length * suite.seeds.length * suite.algorithms.length;
    const actualRuns = suiteRuns.length;
    const instanceCoverage = out.filter((r) => r.best !== null).length;
    const referenceCoverage = out.filter((r) => r.reference !== null).length;

    return {
      rows: out,
      coverage: { expectedRuns, actualRuns, instanceCoverage, referenceCoverage },
      siFast: computeSiTripletCoverage(runs, suite.id, suite.siFast),
      siDeep: computeSiTripletCoverage(runs, suite.id, suite.siDeep),
    };
  }, [runs, suite, knownOptimal]);

  const fastGo = commandsForDomain(suite.id, 'fast');
  const deepGo = commandsForDomain(suite.id, 'deep');

  const fastBatch = `${GO_CWD}\n${BATCH_SCRIPTS.fast}`;
  const deepBatch = `${GO_CWD}\n${BATCH_SCRIPTS.deep}`;
  const retrainBatch = `${GO_CWD}\n${BATCH_SCRIPTS.retrain}`;

  return (
    <Card title={suite.title}>
      <div className="flex flex-wrap items-center gap-2 mb-3">
        <p className="text-xs text-gray-500">{suite.subtitle}</p>
        <span className="text-[10px] px-2 py-0.5 rounded bg-gray-800 text-gray-400">
          Ladder: {suite.instances.length} instance{suite.instances.length === 1 ? '' : 's'} · {suite.seeds.length} seeds · {suite.algorithms.length} algos
        </span>
        <span className="text-[10px] px-2 py-0.5 rounded bg-gray-800 text-gray-400">
          Runs: {coverage.actualRuns}/{coverage.expectedRuns}
        </span>
      </div>

      {/* SI policy triplet completeness */}
      <div className="mb-3 p-3 rounded-lg bg-gray-900/80 border border-gray-800">
        <p className="text-[10px] text-gray-500 uppercase mb-2">Search Intelligence (rules / hybrid / learned)</p>
        <div className="flex flex-wrap gap-2 mb-2">
          <CoverageBadge complete={siFast.complete} expected={siFast.expected} label="Fast val" />
          <CoverageBadge complete={siDeep.complete} expected={siDeep.expected} label="Deep val" />
        </div>
        <div className="flex flex-wrap gap-1.5 mb-1">
          <span className="text-[9px] text-gray-600 w-full">Fast policy runs (of {siFast.expected}):</span>
          <PolicyBadge policy="rules" count={siFast.byPolicy.rules} expected={siFast.expected} />
          <PolicyBadge policy="hybrid" count={siFast.byPolicy.hybrid} expected={siFast.expected} />
          <PolicyBadge policy="learned" count={siFast.byPolicy.learned} expected={siFast.expected} />
        </div>
        <div className="flex flex-wrap gap-1.5">
          <span className="text-[9px] text-gray-600 w-full">Deep policy runs (of {siDeep.expected}):</span>
          <PolicyBadge policy="rules" count={siDeep.byPolicy.rules} expected={siDeep.expected} />
          <PolicyBadge policy="hybrid" count={siDeep.byPolicy.hybrid} expected={siDeep.expected} />
          <PolicyBadge policy="learned" count={siDeep.byPolicy.learned} expected={siDeep.expected} />
        </div>
        <p className="text-[9px] text-gray-600 mt-2">
          Triplet = same instance + mode + seed with all three policy modes. Partial groups: fast {siFast.partial} · deep {siDeep.partial}.
        </p>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-[10px] mb-3">
        {Object.entries(suite.settings).map(([k, v]) => (
          <div key={k} className="bg-gray-800 rounded p-2">
            <div className="text-gray-500 uppercase text-[9px]">{k}</div>
            <div className="text-gray-300">{v}</div>
          </div>
        ))}
        <div className="bg-gray-800 rounded p-2">
          <div className="text-gray-500 uppercase text-[9px]">Seeds</div>
          <div className="text-gray-300 font-mono">{suite.seeds.slice(0, 5).join(', ')}{suite.seeds.length > 5 ? '…' : ''}</div>
        </div>
        <div className="bg-gray-800 rounded p-2">
          <div className="text-gray-500 uppercase text-[9px]">Reference</div>
          <div className="text-gray-300">{suite.referenceLabel}</div>
        </div>
      </div>

      {/* Run commands */}
      <div className="mb-3 p-3 rounded-lg bg-gray-900/80 border border-gray-800">
        <p className="text-[10px] text-gray-500 uppercase mb-2">How to reproduce (from platform/go)</p>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mb-3 text-[10px]">
          <div>
            <p className="text-gray-500 uppercase text-[9px] mb-1">Single run — fast tier (example: hybrid, seed 42)</p>
            {fastGo.map(({ mode, example }) => (
              <div key={mode} className="mb-2">
                <span className="text-[9px] text-gray-600 uppercase">{mode}</span>
                <pre className="text-blue-300/90 whitespace-pre-wrap font-mono text-[9px] mt-0.5">{GO_CWD}{'\n'}{example}</pre>
              </div>
            ))}
          </div>
          <div>
            <p className="text-gray-500 uppercase text-[9px] mb-1">Single run — deep tier (example: hybrid, seed 42)</p>
            {deepGo.map(({ mode, example }) => (
              <div key={mode} className="mb-2">
                <span className="text-[9px] text-gray-600 uppercase">{mode}</span>
                <pre className="text-blue-300/90 whitespace-pre-wrap font-mono text-[9px] mt-0.5">{GO_CWD}{'\n'}{example}</pre>
              </div>
            ))}
          </div>
        </div>

        <p className="text-[9px] text-gray-600 mb-2">
          Swap <code className="text-gray-400">hybrid</code> for <code className="text-gray-400">rules</code> or <code className="text-gray-400">learned</code>; change <code className="text-gray-400">--seed</code> / <code className="text-gray-400">--seeds</code> to match the ladder.
        </p>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-2 text-[10px]">
          <div className="bg-gray-900 rounded p-2 border border-gray-800">
            <div className="text-gray-500 uppercase text-[9px] mb-1">Batch — fast suite (all domains)</div>
            <p className="text-gray-600 mb-1">{suite.siFast.scriptHint}</p>
            <pre className="text-gray-400 whitespace-pre-wrap font-mono text-[9px]">{fastBatch}</pre>
          </div>
          <div className="bg-gray-900 rounded p-2 border border-gray-800">
            <div className="text-gray-500 uppercase text-[9px] mb-1">Batch — deep suite (all domains)</div>
            <p className="text-gray-600 mb-1">{suite.siDeep.scriptHint}</p>
            <pre className="text-gray-400 whitespace-pre-wrap font-mono text-[9px]">{deepBatch}</pre>
          </div>
          <div className="bg-gray-900 rounded p-2 border border-gray-800">
            <div className="text-gray-500 uppercase text-[9px] mb-1">After runs: retrain policies</div>
            <p className="text-gray-600 mb-1">Sync S3 → train → validate → upload registry</p>
            <pre className="text-emerald-300/90 whitespace-pre-wrap font-mono text-[9px]">{retrainBatch}</pre>
          </div>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Instance</th>
              <th className="text-right p-2">SA</th>
              <th className="text-right p-2">LAHC</th>
              <th className="text-right p-2">Tabu</th>
              <th className="text-right p-2">Portfolio</th>
              <th className="text-right p-2">Adaptive</th>
              <th className="text-right p-2 text-blue-400">Reference</th>
              <th className="text-right p-2">Gap%</th>
              <th className="text-center p-2">Winner</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr key={r.instance} className="border-t border-gray-800 hover:bg-gray-800/50">
                <td className="p-2 font-mono text-gray-300">{r.instance}</td>
                <Cell value={r.sa} best={r.best} reference={r.reference} />
                <Cell value={r.lahc} best={r.best} reference={r.reference} />
                <Cell value={r.tabu} best={r.best} reference={r.reference} />
                <Cell value={r.portfolio} best={r.best} reference={r.reference} />
                <Cell value={r.adaptive} best={r.best} reference={r.reference} />
                <td className="text-right p-2">
                  {r.reference !== null ? (
                    <span className="text-blue-400 font-semibold">{r.reference.toLocaleString()}</span>
                  ) : (
                    <span className="text-gray-700">—</span>
                  )}
                </td>
                <td className="text-right p-2">
                  {r.gapPct !== null ? (
                    <span
                      className={`font-semibold ${
                        r.gapPct === 0 ? 'text-emerald-400' : r.gapPct < 15 ? 'text-emerald-400' : r.gapPct < 30 ? 'text-amber-400' : 'text-red-400'
                      }`}
                    >
                      {r.gapPct === 0 ? '✓ optimal' : `+${r.gapPct.toFixed(1)}%`}
                    </span>
                  ) : (
                    <span className="text-gray-700">—</span>
                  )}
                </td>
                <td className="text-center p-2">
                  <span className="bg-emerald-900/50 text-emerald-400 px-2 py-0.5 rounded text-[10px] font-semibold">
                    {r.winner}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <p className="text-[10px] text-gray-600 mt-2">
        Instances with results: {coverage.instanceCoverage}/{suite.instances.length} · With reference: {coverage.referenceCoverage}/{suite.instances.length}
      </p>
    </Card>
  );
}

function Cell({ value, best, reference }: { value: number | null; best: number | null; reference: number | null }) {
  if (value === null) return <td className="text-right p-2 text-gray-700">—</td>;
  const isBest = best !== null && value === best;
  const isOptimal = reference !== null && value === reference;
  return (
    <td className={`text-right p-2 ${isOptimal ? 'text-emerald-400 font-bold' : isBest ? 'text-emerald-400 font-semibold' : 'text-gray-300'}`}>
      {value.toLocaleString()}
    </td>
  );
}
