'use client';

import { Fragment, useMemo, useState } from 'react';
import Card from '@/components/Card';
import type { BenchmarkRun } from './page';
import type { PolicyMode, ProblemType } from '@/lib/benchmark-suites';
import { buildSiGoCommand, GO_CWD } from '@/lib/benchmark-commands';

const POLICIES = ['rules', 'hybrid', 'learned'] as const;

interface Triplet {
  key: string;
  problemType: string;
  instance: string;
  mode: string;
  seed: number;
  tier: 'fast' | 'deep';
  rules: number | null;
  hybrid: number | null;
  learned: number | null;
}

type SortKey = 'problemType' | 'instance' | 'mode' | 'seed' | 'rules' | 'hybrid' | 'learned' | 'winner';
type SortDir = 'asc' | 'desc';

function parsePolicyMode(run: BenchmarkRun): PolicyMode | null {
  const fromMeta = run.policyMode?.toLowerCase();
  if (POLICIES.includes(fromMeta as PolicyMode)) return fromMeta as PolicyMode;
  const m = run.id.match(/-(rules|hybrid|learned)-s\d+$/);
  return m ? (m[1] as PolicyMode) : null;
}

function normaliseMode(mode: string): string {
  const lower = mode.toLowerCase();
  if (lower.includes('portfolio')) return 'portfolio';
  if (lower === 'tabu') return 'tabu';
  if (lower === 'lahc') return 'lahc';
  if (lower === 'sa') return 'sa';
  return lower;
}

function inferTier(runId: string): 'fast' | 'deep' {
  return runId.startsWith('val-deep-') ? 'deep' : 'fast';
}

function tripletWinner(t: Triplet): string {
  const vals = [t.rules, t.hybrid, t.learned].filter((v): v is number => v !== null);
  const best = vals.length ? Math.min(...vals) : null;
  if (best === null) return '—';
  if (t.rules === best && t.hybrid === best && t.learned === best) return 'tie';
  if (t.hybrid === best && t.hybrid! < (t.rules ?? Infinity) && t.hybrid! <= (t.learned ?? Infinity)) return 'hybrid';
  if (t.learned === best && t.learned! < (t.rules ?? Infinity)) return 'learned';
  if (t.rules === best) return 'rules';
  return 'tie';
}

function winnerRank(w: string): number {
  if (w === 'hybrid') return 0;
  if (w === 'learned') return 1;
  if (w === 'rules') return 2;
  if (w === 'tie') return 3;
  return 4;
}

export default function SIComparison({
  runs,
  domainFilter = 'all',
}: {
  runs: BenchmarkRun[];
  domainFilter?: 'all' | ProblemType;
}) {
  const [sortKey, setSortKey] = useState<SortKey>('problemType');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [expandedKey, setExpandedKey] = useState<string | null>(null);

  const { triplets, summary, byDomain } = useMemo(() => {
    const valRuns = runs.filter((r) => r.id.startsWith('val-') && parsePolicyMode(r));
    const groups = new Map<string, Triplet>();

    for (const run of valRuns) {
      const policy = parsePolicyMode(run)!;
      const mode = normaliseMode(run.mode);
      const tier = inferTier(run.id);
      const key = `${run.problemType}:${run.instance}:${mode}:${run.seed}:${tier}`;
      const row = groups.get(key) ?? {
        key,
        problemType: run.problemType,
        instance: run.instance,
        mode,
        seed: run.seed,
        tier,
        rules: null,
        hybrid: null,
        learned: null,
      };
      row[policy] = run.penalty;
      groups.set(key, row);
    }

    const all = [...groups.values()];

    const complete = all.filter((t) => t.rules !== null && t.hybrid !== null && t.learned !== null);

    let hybridWins = 0;
    let learnedWins = 0;
    let rulesWins = 0;
    let ties = 0;
    let hybridTotalDelta = 0;
    let learnedTotalDelta = 0;

    for (const t of complete) {
      const w = tripletWinner(t);
      if (w === 'tie') ties++;
      else if (w === 'hybrid') hybridWins++;
      else if (w === 'learned') learnedWins++;
      else if (w === 'rules') rulesWins++;
      hybridTotalDelta += t.rules! - t.hybrid!;
      learnedTotalDelta += t.rules! - t.learned!;
    }

    const domainMap = new Map<string, { complete: number; hybridWins: number; learnedWins: number; rulesWins: number }>();
    for (const t of complete) {
      const d = domainMap.get(t.problemType) ?? { complete: 0, hybridWins: 0, learnedWins: 0, rulesWins: 0 };
      d.complete++;
      const w = tripletWinner(t);
      if (w === 'hybrid') d.hybridWins++;
      else if (w === 'learned') d.learnedWins++;
      else if (w === 'rules') d.rulesWins++;
      domainMap.set(t.problemType, d);
    }

    return {
      triplets: all,
      summary: {
        totalValRuns: valRuns.length,
        complete: complete.length,
        hybridWins,
        learnedWins,
        rulesWins,
        ties,
        avgHybridDelta: complete.length ? hybridTotalDelta / complete.length : 0,
        avgLearnedDelta: complete.length ? learnedTotalDelta / complete.length : 0,
      },
      byDomain: [...domainMap.entries()].sort((a, b) => a[0].localeCompare(b[0])),
    };
  }, [runs]);

  const filteredTriplets = useMemo(() => {
    let rows = domainFilter === 'all' ? triplets : triplets.filter((t) => t.problemType === domainFilter);

    rows = [...rows].sort((a, b) => {
      let cmp = 0;
      if (sortKey === 'winner') {
        cmp = winnerRank(tripletWinner(a)) - winnerRank(tripletWinner(b));
      } else if (sortKey === 'seed') {
        cmp = a.seed - b.seed;
      } else if (sortKey === 'rules' || sortKey === 'hybrid' || sortKey === 'learned') {
        cmp = (a[sortKey] ?? Infinity) - (b[sortKey] ?? Infinity);
      } else {
        cmp = String(a[sortKey]).localeCompare(String(b[sortKey]));
      }
      return sortDir === 'asc' ? cmp : -cmp;
    });

    return rows;
  }, [triplets, domainFilter, sortKey, sortDir]);

  function toggleSort(key: SortKey) {
    if (sortKey === key) setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    else {
      setSortKey(key);
      setSortDir(key === 'seed' || key === 'rules' || key === 'hybrid' || key === 'learned' ? 'asc' : 'asc');
    }
  }

  function SortHeader({ label, col }: { label: string; col: SortKey }) {
    const active = sortKey === col;
    return (
      <th className="text-left p-1 cursor-pointer select-none hover:text-gray-300" onClick={() => toggleSort(col)}>
        {label}
        {active && <span className="ml-1 text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
      </th>
    );
  }

  if (summary.totalValRuns === 0) {
    return (
      <Card title="Search Intelligence Comparison">
        <p className="text-xs text-gray-500">
          No <code className="text-blue-400">val-*</code> runs with policy modes yet. Run{' '}
          <code className="text-blue-400">validate-si2.ps1</code> to compare rules vs hybrid vs learned.
        </p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card title="Search Intelligence Comparison">
        <p className="text-xs text-gray-500 mb-3">
          Same instance, algorithm, and seed — rules vs hybrid vs learned. Lower objective wins.
          {domainFilter !== 'all' && (
            <span className="ml-2 text-blue-400">Filtered: {domainFilter.toUpperCase()}</span>
          )}
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2 mb-3">
          <Stat label="Val runs" value={summary.totalValRuns} />
          <Stat label="Matched triplets" value={summary.complete} colour="blue" />
          <Stat label="Hybrid wins" value={summary.hybridWins} colour="emerald" />
          <Stat label="Learned wins" value={summary.learnedWins} colour="emerald" />
          <Stat label="Rules wins" value={summary.rulesWins} colour="amber" />
          <Stat label="Ties" value={summary.ties} />
          <Stat
            label="Avg Δ rules→hybrid"
            value={summary.avgHybridDelta >= 0 ? `+${summary.avgHybridDelta.toFixed(1)}` : summary.avgHybridDelta.toFixed(1)}
            colour={summary.avgHybridDelta > 0 ? 'emerald' : 'amber'}
          />
        </div>
        {byDomain.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            {byDomain.map(([domain, d]) => (
              <div key={domain} className="bg-gray-800 rounded p-2 text-[10px]">
                <p className="text-gray-400 uppercase font-semibold">{domain}</p>
                <p className="text-gray-300">{d.complete} triplets</p>
                <p className="text-emerald-400">hybrid {d.hybridWins} · learned {d.learnedWins}</p>
                <p className="text-amber-400">rules {d.rulesWins}</p>
              </div>
            ))}
          </div>
        )}
      </Card>

      {filteredTriplets.length > 0 && (
        <Card title="Policy head-to-head (val-* runs)">
          <p className="text-[10px] text-gray-500 mb-2">
            Click column headers to sort. Click a row to see the exact <code className="text-blue-400">go run</code> commands to reproduce it.
          </p>
          <div className="overflow-x-auto max-h-[28rem] overflow-y-auto">
            <table className="w-full text-[10px]">
              <thead className="text-gray-500 uppercase sticky top-0 bg-gray-900 z-10">
                <tr>
                  <SortHeader label="Domain" col="problemType" />
                  <SortHeader label="Instance" col="instance" />
                  <SortHeader label="Mode" col="mode" />
                  <SortHeader label="Seed" col="seed" />
                  <th className="text-right p-1">Tier</th>
                  <SortHeader label="Rules" col="rules" />
                  <th className="text-right p-1 cursor-pointer select-none hover:text-gray-300" onClick={() => toggleSort('hybrid')}>
                    Hybrid{sortKey === 'hybrid' && <span className="ml-1 text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
                  </th>
                  <th className="text-right p-1 cursor-pointer select-none hover:text-gray-300" onClick={() => toggleSort('learned')}>
                    Learned{sortKey === 'learned' && <span className="ml-1 text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>}
                  </th>
                  <SortHeader label="Winner" col="winner" />
                </tr>
              </thead>
              <tbody>
                {filteredTriplets.map((t) => {
                  const best = [t.rules, t.hybrid, t.learned].filter((v): v is number => v !== null);
                  const bestVal = best.length ? Math.min(...best) : null;
                  const winner = tripletWinner(t);
                  const isExpanded = expandedKey === t.key;
                  return (
                    <Fragment key={t.key}>
                      <tr
                        key={t.key}
                        className={`border-t border-gray-800 text-gray-300 cursor-pointer hover:bg-gray-800/60 ${isExpanded ? 'bg-gray-800/40' : ''}`}
                        onClick={() => setExpandedKey(isExpanded ? null : t.key)}
                      >
                        <td className="p-1">{t.problemType}</td>
                        <td className="p-1 font-mono">{t.instance}</td>
                        <td className="p-1">{t.mode}</td>
                        <td className="p-1 text-right">{t.seed}</td>
                        <td className="p-1 text-right text-gray-500">{t.tier}</td>
                        <td className={`p-1 text-right ${t.rules === bestVal ? 'text-emerald-400 font-semibold' : ''}`}>{t.rules ?? '—'}</td>
                        <td className={`p-1 text-right ${t.hybrid === bestVal ? 'text-emerald-400 font-semibold' : ''}`}>{t.hybrid ?? '—'}</td>
                        <td className={`p-1 text-right ${t.learned === bestVal ? 'text-emerald-400 font-semibold' : ''}`}>{t.learned ?? '—'}</td>
                        <td className="p-1 text-center text-amber-400">{winner}</td>
                      </tr>
                      {isExpanded && (
                        <tr key={`${t.key}-cmd`} className="bg-gray-900/80">
                          <td colSpan={9} className="p-3">
                            <p className="text-[9px] text-gray-500 mb-2 uppercase">Reproduce from platform/go</p>
                            <pre className="text-[9px] text-gray-500 mb-2 font-mono">{GO_CWD}</pre>
                            {POLICIES.map((policy) => {
                              const cmd = buildSiGoCommand(t.problemType, t.mode, t.seed, policy, t.tier);
                              return (
                                <div key={policy} className="mb-2">
                                  <span className="text-[9px] text-gray-500 uppercase">{policy}</span>
                                  <pre className="text-[9px] text-blue-300/90 whitespace-pre-wrap font-mono mt-0.5">
                                    {cmd ?? `# No template for ${t.problemType}/${t.mode}/${t.tier}`}
                                  </pre>
                                </div>
                              );
                            })}
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
          <p className="text-[9px] text-gray-600 mt-2">
            Showing {filteredTriplets.length} of {triplets.length} rows
          </p>
        </Card>
      )}
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: number | string; colour?: string }) {
  const text = colour === 'emerald' ? 'text-emerald-400' : colour === 'amber' ? 'text-amber-400' : colour === 'blue' ? 'text-blue-400' : 'text-gray-200';
  return (
    <div className="bg-gray-800 rounded p-2 text-center">
      <p className={`text-sm font-semibold ${text}`}>{value}</p>
      <p className="text-[9px] text-gray-500 uppercase">{label}</p>
    </div>
  );
}
