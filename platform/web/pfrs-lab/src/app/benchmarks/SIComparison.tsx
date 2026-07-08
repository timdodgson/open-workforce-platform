'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import type { BenchmarkRun } from './page';

const POLICIES = ['rules', 'hybrid', 'learned'] as const;
type PolicyMode = (typeof POLICIES)[number];

interface Triplet {
  key: string;
  problemType: string;
  instance: string;
  mode: string;
  seed: number;
  rules: number | null;
  hybrid: number | null;
  learned: number | null;
}

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

export default function SIComparison({ runs }: { runs: BenchmarkRun[] }) {
  const { triplets, summary, byDomain } = useMemo(() => {
    const valRuns = runs.filter((r) => r.id.startsWith('val-') && parsePolicyMode(r));
    const groups = new Map<string, Triplet>();

    for (const run of valRuns) {
      const policy = parsePolicyMode(run)!;
      const mode = normaliseMode(run.mode);
      const key = `${run.problemType}:${run.instance}:${mode}:${run.seed}`;
      const row = groups.get(key) ?? {
        key,
        problemType: run.problemType,
        instance: run.instance,
        mode,
        seed: run.seed,
        rules: null,
        hybrid: null,
        learned: null,
      };
      row[policy] = run.penalty;
      groups.set(key, row);
    }

    const all = [...groups.values()].sort((a, b) =>
      a.problemType.localeCompare(b.problemType)
        || a.instance.localeCompare(b.instance)
        || a.mode.localeCompare(b.mode)
        || a.seed - b.seed,
    );

    const complete = all.filter((t) => t.rules !== null && t.hybrid !== null && t.learned !== null);

    let hybridWins = 0;
    let learnedWins = 0;
    let rulesWins = 0;
    let ties = 0;
    let hybridTotalDelta = 0;
    let learnedTotalDelta = 0;

    for (const t of complete) {
      const scores = [
        { p: 'rules' as const, v: t.rules! },
        { p: 'hybrid' as const, v: t.hybrid! },
        { p: 'learned' as const, v: t.learned! },
      ].sort((a, b) => a.v - b.v);
      const best = scores[0].v;
      const winners = scores.filter((s) => s.v === best).map((s) => s.p);
      if (winners.length > 1) ties++;
      else if (winners[0] === 'hybrid') hybridWins++;
      else if (winners[0] === 'learned') learnedWins++;
      else rulesWins++;
      hybridTotalDelta += t.rules! - t.hybrid!;
      learnedTotalDelta += t.rules! - t.learned!;
    }

    const domainMap = new Map<string, { complete: number; hybridWins: number; learnedWins: number; rulesWins: number }>();
    for (const t of complete) {
      const d = domainMap.get(t.problemType) ?? { complete: 0, hybridWins: 0, learnedWins: 0, rulesWins: 0 };
      d.complete++;
      const best = Math.min(t.rules!, t.hybrid!, t.learned!);
      if (t.hybrid === best && t.hybrid! < t.rules!) d.hybridWins++;
      if (t.learned === best && t.learned! < t.rules!) d.learnedWins++;
      if (t.rules === best && t.rules! < t.hybrid! && t.rules! < t.learned!) d.rulesWins++;
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

      {triplets.length > 0 && (
        <Card title="Policy head-to-head (val-* runs)">
          <div className="overflow-x-auto max-h-96 overflow-y-auto">
            <table className="w-full text-[10px]">
              <thead className="text-gray-500 uppercase sticky top-0 bg-gray-900">
                <tr>
                  <th className="text-left p-1">Domain</th>
                  <th className="text-left p-1">Instance</th>
                  <th className="text-left p-1">Mode</th>
                  <th className="text-right p-1">Seed</th>
                  <th className="text-right p-1">Rules</th>
                  <th className="text-right p-1">Hybrid</th>
                  <th className="text-right p-1">Learned</th>
                  <th className="text-center p-1">Winner</th>
                </tr>
              </thead>
              <tbody>
                {triplets.map((t) => {
                  const vals = [t.rules, t.hybrid, t.learned].filter((v): v is number => v !== null);
                  const best = vals.length ? Math.min(...vals) : null;
                  const winner =
                    best === null ? '—'
                      : t.rules === best && t.hybrid === best && t.learned === best ? 'tie'
                        : t.hybrid === best && t.hybrid! < (t.rules ?? Infinity) && t.hybrid! <= (t.learned ?? Infinity) ? 'hybrid'
                          : t.learned === best && t.learned! < (t.rules ?? Infinity) ? 'learned'
                            : t.rules === best ? 'rules' : 'tie';
                  return (
                    <tr key={t.key} className="border-t border-gray-800 text-gray-300">
                      <td className="p-1">{t.problemType}</td>
                      <td className="p-1 font-mono">{t.instance}</td>
                      <td className="p-1">{t.mode}</td>
                      <td className="p-1 text-right">{t.seed}</td>
                      <td className={`p-1 text-right ${t.rules === best ? 'text-emerald-400 font-semibold' : ''}`}>{t.rules ?? '—'}</td>
                      <td className={`p-1 text-right ${t.hybrid === best ? 'text-emerald-400 font-semibold' : ''}`}>{t.hybrid ?? '—'}</td>
                      <td className={`p-1 text-right ${t.learned === best ? 'text-emerald-400 font-semibold' : ''}`}>{t.learned ?? '—'}</td>
                      <td className="p-1 text-center text-amber-400">{winner}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
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
