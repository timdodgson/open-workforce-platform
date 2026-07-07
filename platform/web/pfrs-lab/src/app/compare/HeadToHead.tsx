'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { RunData } from './page';

interface DNAMetric {
  name: string;
  scoreA: number;
  scoreB: number;
}

export default function HeadToHead({ runs }: { runs: RunData[] }) {
  const [runAId, setRunAId] = useState(runs[0]?.id || '');
  const [runBId, setRunBId] = useState(runs[1]?.id || '');

  const runA = runs.find(r => r.id === runAId);
  const runB = runs.find(r => r.id === runBId);

  if (!runA || !runB) return <Card title="Select Runs"><p className="text-gray-500">Select two runs.</p></Card>;

  return (
    <div className="space-y-4">
      {/* Run selector */}
      <Card title="Head-to-Head Comparison">
        <div className="flex gap-4 items-center">
          <div className="flex-1">
            <label className="text-[10px] text-gray-500 uppercase block mb-1">Run A</label>
            <select value={runAId} onChange={e => setRunAId(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-blue-400">
              {runs.map(r => <option key={r.id} value={r.id}>{r.id}</option>)}
            </select>
          </div>
          <span className="text-gray-600 text-lg">vs</span>
          <div className="flex-1">
            <label className="text-[10px] text-gray-500 uppercase block mb-1">Run B</label>
            <select value={runBId} onChange={e => setRunBId(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-rose-400">
              {runs.map(r => <option key={r.id} value={r.id}>{r.id}</option>)}
            </select>
          </div>
        </div>
      </Card>

      <ComparisonBody runA={runA} runB={runB} />
    </div>
  );
}

function diffStr(a: number, b: number): { text: string; color: string } {
  const diff = a - b;
  if (diff === 0) return { text: '=', color: 'text-gray-500' };
  const pct = b !== 0 ? ((diff / b) * 100).toFixed(1) : '∞';
  if (diff > 0) return { text: `+${diff.toLocaleString()} (+${pct}%)`, color: 'text-red-400' };
  return { text: `${diff.toLocaleString()} (${pct}%)`, color: 'text-emerald-400' };
}

function computeDNA(run: RunData): DNAMetric[] {
  const s = run.summary;
  const exploration = Math.min(100, Math.round(s.acceptWorseRate * 10 + (s.totalBranches / Math.max(s.totalWorkers, 1)) * 20));
  const totalImp = s.weeks.reduce((sum, w) => sum + w.improvement, 0);
  const effPerM = s.totalCandidates > 0 ? totalImp / (s.totalCandidates / 1_000_000) : 0;
  const exploitation = Math.min(100, Math.round(effPerM * 2));
  const globalDisc = run.discoveries.filter(d => d.eventType === 'global_best').length;
  const innovation = Math.min(100, Math.round((globalDisc / Math.max(s.totalWorkers, 1)) * 500));
  const retained = run.nodes.filter(t => t.retained);
  const uniqueParents = new Set(retained.map(t => t.parentID)).size;
  const lineage = Math.min(100, Math.round((uniqueParents / Math.max(retained.length, 1)) * 100));
  const improved = run.workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const utilisation = Math.min(100, Math.round((improved / Math.max(run.workers.length, 1)) * 100));

  return [
    { name: 'Exploration', scoreA: exploration, scoreB: 0 },
    { name: 'Exploitation', scoreA: exploitation, scoreB: 0 },
    { name: 'Innovation', scoreA: innovation, scoreB: 0 },
    { name: 'Lineage', scoreA: lineage, scoreB: 0 },
    { name: 'Utilisation', scoreA: utilisation, scoreB: 0 },
  ];
}

function ComparisonBody({ runA, runB }: { runA: RunData; runB: RunData }) {
  const dnaA = useMemo(() => computeDNA(runA), [runA]);
  const dnaB = useMemo(() => computeDNA(runB), [runB]);

  // Detect problem type.
  const problemTypeA = (runA.summary.metadata as unknown as Record<string, unknown>)?.problemType;
  const problemTypeB = (runB.summary.metadata as unknown as Record<string, unknown>)?.problemType;
  const problemType = problemTypeA || problemTypeB || 'nrp';
  const objectiveLabel = problemType === 'cvrp' || problemType === 'vrptw' ? 'Total Distance' :
                         problemType === 'jss' ? 'Makespan' : 'Total Penalty';

  // Merge DNA for comparison.
  const dnaCompared = dnaA.map((m, i) => ({
    name: m.name,
    scoreA: m.scoreA,
    scoreB: computeDNA(runB)[i]?.scoreA || 0,
  }));

  const sA = runA.summary;
  const sB = runB.summary;
  const penaltyDiff = diffStr(sA.totalPenalty, sB.totalPenalty);

  // Per-week penalty overlay data.
  const maxWeeks = Math.max(sA.weeks.length, sB.weeks.length);

  // Entropy comparison.
  const entropyA = computeEntropy(runA.nodes);
  const entropyB = computeEntropy(runB.nodes);

  // Generate narrative.
  const narrative = generateNarrative(runA, runB, dnaCompared);

  return (
    <>
      {/* Summary comparison */}
      <Card title="Summary">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Metric</th>
              <th className="text-right p-2 text-blue-400">Run A</th>
              <th className="text-right p-2 text-rose-400">Run B</th>
              <th className="text-right p-2">Difference</th>
            </tr>
          </thead>
          <tbody>
            {[
              { label: objectiveLabel, a: sA.totalPenalty, b: sB.totalPenalty, always: true },
              { label: 'Workers', a: sA.totalWorkers, b: sB.totalWorkers, always: false },
              { label: 'Candidates', a: sA.totalCandidates, b: sB.totalCandidates, always: true },
              { label: 'Branches', a: sA.totalBranches, b: sB.totalBranches, always: false },
              { label: 'Runtime (ms)', a: sA.totalDurationMs, b: sB.totalDurationMs, always: true },
              { label: 'Global Bests', a: runA.discoveries.filter(d => d.eventType === 'global_best').length, b: runB.discoveries.filter(d => d.eventType === 'global_best').length, always: true },
              { label: 'Plateaus', a: runA.plateaus.length, b: runB.plateaus.length, always: false },
            ].filter(row => row.always || !isCVRP).map(row => {
              const d = diffStr(row.a, row.b);
              return (
                <tr key={row.label} className="border-t border-gray-800">
                  <td className="p-2">{row.label}</td>
                  <td className="text-right p-2 text-blue-400">{row.a.toLocaleString()}</td>
                  <td className="text-right p-2 text-rose-400">{row.b.toLocaleString()}</td>
                  <td className={`text-right p-2 ${d.color}`}>{d.text}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>

      {/* DNA Overlay */}
      <Card title="Search DNA Comparison">
        <div className="space-y-2">
          {dnaCompared.map(m => (
            <div key={m.name}>
              <div className="flex justify-between text-[10px] text-gray-400 mb-0.5">
                <span>{m.name}</span>
                <span>{m.scoreA} vs {m.scoreB}</span>
              </div>
              <div className="relative h-4 bg-gray-800 rounded overflow-hidden">
                <div className="absolute top-0 left-0 h-2 bg-blue-500 rounded-t" style={{ width: `${m.scoreA}%` }} />
                <div className="absolute bottom-0 left-0 h-2 bg-rose-500 rounded-b" style={{ width: `${m.scoreB}%` }} />
              </div>
            </div>
          ))}
        </div>
        <div className="flex gap-4 mt-2 text-[9px] text-gray-500">
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-blue-500 rounded-sm" />Run A</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-rose-500 rounded-sm" />Run B</span>
        </div>
      </Card>

      {/* Penalty Waterfall Overlay */}
      <Card title="Penalty per Week (Overlay)">
        <div className="flex items-end gap-1 h-32">
          {Array.from({ length: maxWeeks }, (_, i) => {
            const pA = sA.weeks[i]?.finalPenalty || 0;
            const pB = sB.weeks[i]?.finalPenalty || 0;
            const maxP = Math.max(...sA.weeks.map(w => w.finalPenalty), ...sB.weeks.map(w => w.finalPenalty), 1);
            return (
              <div key={i} className="flex-1 flex gap-px items-end h-full">
                <div className="flex-1 bg-blue-500 rounded-t" style={{ height: `${(pA / maxP) * 100}%` }} title={`A W${i+1}: ${pA}`} />
                <div className="flex-1 bg-rose-500 rounded-t" style={{ height: `${(pB / maxP) * 100}%` }} title={`B W${i+1}: ${pB}`} />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          {Array.from({ length: maxWeeks }, (_, i) => <span key={i}>W{i+1}</span>)}
        </div>
      </Card>

      {/* Entropy Overlay */}
      <Card title="Beam Entropy (Overlay)">
        <div className="flex items-end gap-1 h-24">
          {Array.from({ length: maxWeeks }, (_, i) => {
            const eA = entropyA[i] || 0;
            const eB = entropyB[i] || 0;
            const maxE = Math.max(...entropyA, ...entropyB, 0.01);
            return (
              <div key={i} className="flex-1 flex gap-px items-end h-full">
                <div className="flex-1 bg-blue-500 rounded-t" style={{ height: `${(eA / maxE) * 100}%` }} title={`A: ${eA.toFixed(2)}`} />
                <div className="flex-1 bg-rose-500 rounded-t" style={{ height: `${(eB / maxE) * 100}%` }} title={`B: ${eB.toFixed(2)}`} />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Week 1</span><span>Week {maxWeeks}</span>
        </div>
      </Card>

      {/* Workers overlay */}
      <Card title="Worker Utilisation">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-[10px] text-blue-400 uppercase mb-1">Run A</p>
            <p className="text-sm">{runA.workers.length} workers, {runA.workers.filter(w => w.producedGlobalBest).length} global finders</p>
          </div>
          <div>
            <p className="text-[10px] text-rose-400 uppercase mb-1">Run B</p>
            <p className="text-sm">{runB.workers.length} workers, {runB.workers.filter(w => w.producedGlobalBest).length} global finders</p>
          </div>
        </div>
      </Card>

      {/* Narrative */}
      <Card title="Comparison Summary">
        <div className="space-y-2">
          {narrative.map((n, i) => (
            <p key={i} className="text-sm text-gray-300">{n}</p>
          ))}
        </div>
      </Card>
    </>
  );
}

function computeEntropy(nodes: RunData['nodes']): number[] {
  const weeks = [...new Set(nodes.map(n => n.week))].sort((a, b) => a - b);
  const parentMap = new Map<number, number>();
  for (const n of nodes) parentMap.set(n.pathID, n.parentID);

  return weeks.map(week => {
    const retained = nodes.filter(n => n.week === week && n.retained);
    if (retained.length <= 1) return 0;
    const families = new Map<number, number>();
    for (const n of retained) {
      let cur = n.pathID;
      let iter = 0;
      while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) {
        cur = parentMap.get(cur)!;
        iter++;
      }
      families.set(cur, (families.get(cur) || 0) + 1);
    }
    let entropy = 0;
    for (const count of families.values()) {
      const p = count / retained.length;
      if (p > 0) entropy -= p * Math.log2(p);
    }
    return entropy;
  });
}

function generateNarrative(runA: RunData, runB: RunData, dna: { name: string; scoreA: number; scoreB: number }[]): string[] {
  const lines: string[] = [];
  const sA = runA.summary;
  const sB = runB.summary;

  // Penalty.
  if (sA.totalPenalty < sB.totalPenalty) {
    lines.push(`Run A achieved a lower penalty (${sA.totalPenalty.toLocaleString()} vs ${sB.totalPenalty.toLocaleString()}).`);
  } else if (sB.totalPenalty < sA.totalPenalty) {
    lines.push(`Run B achieved a lower penalty (${sB.totalPenalty.toLocaleString()} vs ${sA.totalPenalty.toLocaleString()}).`);
  } else {
    lines.push('Both runs achieved identical penalties.');
  }

  // DNA differences.
  for (const m of dna) {
    const diff = m.scoreA - m.scoreB;
    if (Math.abs(diff) > 15) {
      const better = diff > 0 ? 'A' : 'B';
      lines.push(`Run ${better} scored higher on ${m.name.toLowerCase()} (${Math.max(m.scoreA, m.scoreB)} vs ${Math.min(m.scoreA, m.scoreB)}).`);
    }
  }

  // Diversity.
  const entropyA = computeEntropy(runA.nodes);
  const entropyB = computeEntropy(runB.nodes);
  const avgEntropyA = entropyA.length > 0 ? entropyA.reduce((s, e) => s + e, 0) / entropyA.length : 0;
  const avgEntropyB = entropyB.length > 0 ? entropyB.reduce((s, e) => s + e, 0) / entropyB.length : 0;
  if (avgEntropyA > avgEntropyB * 1.2) {
    lines.push('Run A maintained diversity longer.');
  } else if (avgEntropyB > avgEntropyA * 1.2) {
    lines.push('Run B maintained diversity longer.');
  }

  // Convergence.
  const globalA = runA.discoveries.filter(d => d.eventType === 'global_best');
  const globalB = runB.discoveries.filter(d => d.eventType === 'global_best');
  if (globalA.length > globalB.length * 1.3) {
    lines.push('Run A discovered more global improvements.');
  } else if (globalB.length > globalA.length * 1.3) {
    lines.push('Run B discovered more global improvements.');
  }

  // Workers.
  if (sA.totalWorkers < sB.totalWorkers * 0.8) {
    lines.push('Run A used fewer workers.');
  } else if (sB.totalWorkers < sA.totalWorkers * 0.8) {
    lines.push('Run B used fewer workers.');
  }

  // Convergence speed.
  const lastGlobalA = globalA.length > 0 ? globalA[globalA.length - 1].elapsedMs : 0;
  const lastGlobalB = globalB.length > 0 ? globalB[globalB.length - 1].elapsedMs : 0;
  if (lastGlobalA < lastGlobalB * 0.7 && lastGlobalA > 0) {
    lines.push('Run A converged earlier.');
  } else if (lastGlobalB < lastGlobalA * 0.7 && lastGlobalB > 0) {
    lines.push('Run B converged earlier.');
  }

  return lines;
}
