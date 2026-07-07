'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { DecisionRecord } from './types';

interface Props {
  decisions: DecisionRecord[];
}

interface FalseSkipDetail {
  runId: string;
  workerId: number;
  week: number;
  depth: number;
  algorithm: string;
  parentObjective: number;
  globalBest: number;
  distanceFromBest: number;
  confidence: number;
  reasonCodes: string;
  improvementAmount: number;
  producedGlobalBest: boolean;
  runtimeMs: number;
  gapPct: number;
}

export default function FalseSkipForensics({ decisions }: Props) {
  // All false skips: recommended skip but worker actually improved.
  const falseSkips = useMemo((): FalseSkipDetail[] => {
    return decisions
      .filter(d => d.recommendation === 'skip' && d.improved)
      .map(d => ({
        runId: d.runId,
        workerId: d.workerId,
        week: d.week,
        depth: d.depth,
        algorithm: d.algorithm,
        parentObjective: d.parentObjective,
        globalBest: d.globalBest,
        distanceFromBest: d.distanceFromBest,
        confidence: d.confidence,
        reasonCodes: d.reasonCodes,
        improvementAmount: d.improvementAmount,
        producedGlobalBest: d.producedGlobalBest,
        runtimeMs: d.runtimeMs,
        gapPct: d.globalBest > 0 ? (d.distanceFromBest / d.globalBest) * 100 : 0,
      }))
      .sort((a, b) => b.improvementAmount - a.improvementAmount);
  }, [decisions]);

  // Missed global bests.
  const missedGlobalBests = useMemo(() => {
    return falseSkips.filter(d => d.producedGlobalBest);
  }, [falseSkips]);

  // Common reason codes across false skips.
  const commonReasonCodes = useMemo(() => {
    const map = new Map<string, number>();
    for (const d of falseSkips) {
      const codes = d.reasonCodes.split(';').filter(Boolean);
      for (const code of codes) {
        map.set(code, (map.get(code) || 0) + 1);
      }
    }
    return Array.from(map.entries())
      .map(([code, count]) => ({ code, count, pct: (count / Math.max(falseSkips.length, 1)) * 100 }))
      .sort((a, b) => b.count - a.count);
  }, [falseSkips]);

  // Common parent gap ranges.
  const gapDistribution = useMemo(() => {
    const buckets = [
      { label: '0–25%', min: 0, max: 25, count: 0 },
      { label: '25–50%', min: 25, max: 50, count: 0 },
      { label: '50–75%', min: 50, max: 75, count: 0 },
      { label: '75–100%', min: 75, max: 100, count: 0 },
      { label: '100%+', min: 100, max: Infinity, count: 0 },
    ];
    for (const d of falseSkips) {
      const bucket = buckets.find(b => d.gapPct >= b.min && d.gapPct < b.max);
      if (bucket) bucket.count++;
    }
    return buckets;
  }, [falseSkips]);

  // Week/Depth pattern.
  const depthDistribution = useMemo(() => {
    const map = new Map<string, number>();
    for (const d of falseSkips) {
      const key = `W${d.week}/D${d.depth}`;
      map.set(key, (map.get(key) || 0) + 1);
    }
    return Array.from(map.entries())
      .map(([key, count]) => ({ weekDepth: key, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);
  }, [falseSkips]);

  // Root cause: which rule caused the bad decision.
  const rootCauses = useMemo(() => {
    // The primary reason code (first in the list) is the triggering rule.
    const map = new Map<string, { count: number; missedGlobalBests: number; totalImprovement: number }>();
    for (const d of falseSkips) {
      const primaryCode = d.reasonCodes.split(';')[0] || 'unknown';
      const entry = map.get(primaryCode) || { count: 0, missedGlobalBests: 0, totalImprovement: 0 };
      entry.count++;
      entry.totalImprovement += d.improvementAmount;
      if (d.producedGlobalBest) entry.missedGlobalBests++;
      map.set(primaryCode, entry);
    }
    return Array.from(map.entries())
      .map(([rule, stats]) => ({
        rule,
        count: stats.count,
        missedGlobalBests: stats.missedGlobalBests,
        totalImprovement: stats.totalImprovement,
        avgImprovement: stats.count > 0 ? stats.totalImprovement / stats.count : 0,
      }))
      .sort((a, b) => b.count - a.count);
  }, [falseSkips]);

  if (falseSkips.length === 0) {
    return (
      <Card title="False Skip Forensics">
        <div className="border-2 border-dashed border-emerald-700 rounded-lg p-6 text-center">
          <p className="text-emerald-400 font-semibold">✓ No false skips detected</p>
          <p className="text-xs text-gray-500 mt-1">
            The rule engine has not recommended skipping any worker that later improved.
          </p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Summary Header */}
      <Card title="False Skip Forensics">
        <p className="text-xs text-gray-500 mb-4">
          Deep analysis of workers the engine recommended skipping that actually improved.
          Each false skip represents a missed opportunity. Global best misses are critical failures.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <Stat label="False Skips" value={falseSkips.length} colour="red" />
          <Stat label="Missed Global Bests" value={missedGlobalBests.length} colour="red" />
          <Stat
            label="Total Missed Δ"
            value={falseSkips.reduce((s, d) => s + d.improvementAmount, 0).toLocaleString()}
            colour="amber"
          />
          <Stat
            label="Avg Gap at Skip"
            value={`${(falseSkips.reduce((s, d) => s + d.gapPct, 0) / Math.max(falseSkips.length, 1)).toFixed(0)}%`}
            colour="amber"
          />
        </div>
      </Card>

      {/* Missed Global Bests — Critical */}
      {missedGlobalBests.length > 0 && (
        <Card title="⚠ Missed Global Bests — Critical">
          <p className="text-xs text-red-400 mb-3">
            These workers found the global best solution but the engine recommended skipping them.
            This is the most critical failure mode.
          </p>
          <div className="overflow-x-auto">
            <table className="w-full text-[10px]">
              <thead>
                <tr className="text-gray-500 uppercase">
                  <th className="text-left p-1.5">Run</th>
                  <th className="text-left p-1.5">Worker</th>
                  <th className="text-left p-1.5">Algo</th>
                  <th className="text-right p-1.5">W/D</th>
                  <th className="text-right p-1.5">Parent Obj</th>
                  <th className="text-right p-1.5">Global Best</th>
                  <th className="text-right p-1.5">Gap %</th>
                  <th className="text-right p-1.5">Confidence</th>
                  <th className="text-left p-1.5">Reason Codes</th>
                  <th className="text-right p-1.5">Δ Improvement</th>
                  <th className="text-right p-1.5">Runtime</th>
                </tr>
              </thead>
              <tbody>
                {missedGlobalBests.map((d, i) => (
                  <tr key={i} className="border-t border-red-900/50 bg-red-900/10">
                    <td className="p-1.5 text-blue-400 font-mono">{d.runId.slice(0, 16)}</td>
                    <td className="p-1.5 font-semibold text-red-400">{d.workerId}</td>
                    <td className="p-1.5 text-emerald-400">{d.algorithm}</td>
                    <td className="text-right p-1.5">W{d.week}/D{d.depth}</td>
                    <td className="text-right p-1.5">{d.parentObjective.toLocaleString()}</td>
                    <td className="text-right p-1.5">{d.globalBest.toLocaleString()}</td>
                    <td className="text-right p-1.5 text-amber-400">{d.gapPct.toFixed(0)}%</td>
                    <td className="text-right p-1.5 text-amber-400">{d.confidence.toFixed(2)}</td>
                    <td className="p-1.5 text-gray-400 font-mono">{d.reasonCodes}</td>
                    <td className="text-right p-1.5 text-red-400 font-semibold">⭐ {d.improvementAmount.toLocaleString()}</td>
                    <td className="text-right p-1.5">{d.runtimeMs}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Full False Skip Table */}
      <Card title="All False Skips — Detailed">
        <div className="overflow-x-auto">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">Run</th>
                <th className="text-left p-1.5">Worker</th>
                <th className="text-left p-1.5">Algo</th>
                <th className="text-right p-1.5">W/D</th>
                <th className="text-right p-1.5">Parent Obj</th>
                <th className="text-right p-1.5">Global Best</th>
                <th className="text-right p-1.5">Gap %</th>
                <th className="text-right p-1.5">Confidence</th>
                <th className="text-left p-1.5">Reason Codes</th>
                <th className="text-right p-1.5">Δ Improvement</th>
                <th className="text-right p-1.5">Runtime</th>
                <th className="text-center p-1.5">Global Best?</th>
              </tr>
            </thead>
            <tbody>
              {falseSkips.map((d, i) => (
                <tr key={i} className={`border-t border-gray-800 ${d.producedGlobalBest ? 'bg-red-900/10' : ''}`}>
                  <td className="p-1.5 text-blue-400 font-mono">{d.runId.slice(0, 16)}</td>
                  <td className="p-1.5">{d.workerId}</td>
                  <td className="p-1.5 text-emerald-400">{d.algorithm}</td>
                  <td className="text-right p-1.5">W{d.week}/D{d.depth}</td>
                  <td className="text-right p-1.5">{d.parentObjective.toLocaleString()}</td>
                  <td className="text-right p-1.5">{d.globalBest.toLocaleString()}</td>
                  <td className="text-right p-1.5 text-amber-400">{d.gapPct.toFixed(0)}%</td>
                  <td className="text-right p-1.5 text-amber-400">{d.confidence.toFixed(2)}</td>
                  <td className="p-1.5 text-gray-400 font-mono">{d.reasonCodes}</td>
                  <td className="text-right p-1.5 text-red-400 font-semibold">{d.improvementAmount.toLocaleString()}</td>
                  <td className="text-right p-1.5">{d.runtimeMs}ms</td>
                  <td className="text-center p-1.5">{d.producedGlobalBest ? '⭐' : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Root Cause Analysis */}
      <Card title="Root Cause — Which Rule Caused Bad Decisions">
        <p className="text-xs text-gray-500 mb-3">
          The primary reason code that triggered the SKIP recommendation for each false skip.
        </p>
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Rule</th>
              <th className="text-right p-2">False Skips</th>
              <th className="text-right p-2">Missed Global Bests</th>
              <th className="text-right p-2">Total Missed Δ</th>
              <th className="text-right p-2">Avg Missed Δ</th>
            </tr>
          </thead>
          <tbody>
            {rootCauses.map(rc => (
              <tr key={rc.rule} className="border-t border-gray-800">
                <td className="p-2 text-red-400 font-mono text-[10px]">{rc.rule}</td>
                <td className="text-right p-2">{rc.count}</td>
                <td className="text-right p-2">
                  {rc.missedGlobalBests > 0 && (
                    <span className="text-red-400 font-semibold">⭐ {rc.missedGlobalBests}</span>
                  )}
                  {rc.missedGlobalBests === 0 && <span className="text-gray-500">0</span>}
                </td>
                <td className="text-right p-2 text-amber-400">{rc.totalImprovement.toLocaleString()}</td>
                <td className="text-right p-2">{rc.avgImprovement.toFixed(0)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Pattern Analysis */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Common Reason Codes */}
        <Card title="Common Reason Codes in False Skips">
          <div className="space-y-2">
            {commonReasonCodes.map(rc => (
              <div key={rc.code} className="flex justify-between items-center">
                <span className="text-[10px] text-gray-400 font-mono">{rc.code}</span>
                <div className="flex items-center gap-2">
                  <div className="w-24 bg-gray-800 rounded-full h-2">
                    <div
                      className="bg-red-500 h-2 rounded-full"
                      style={{ width: `${rc.pct}%` }}
                    />
                  </div>
                  <span className="text-[10px] text-gray-500 w-16 text-right">
                    {rc.count} ({rc.pct.toFixed(0)}%)
                  </span>
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Parent Gap Distribution */}
        <Card title="Parent Gap Distribution (False Skips)">
          <div className="space-y-2">
            {gapDistribution.map(b => (
              <div key={b.label} className="flex justify-between items-center">
                <span className="text-xs text-gray-400">{b.label}</span>
                <div className="flex items-center gap-2">
                  <div className="w-24 bg-gray-800 rounded-full h-2">
                    <div
                      className="bg-amber-500 h-2 rounded-full"
                      style={{ width: `${falseSkips.length > 0 ? (b.count / falseSkips.length) * 100 : 0}%` }}
                    />
                  </div>
                  <span className="text-[10px] text-gray-500 w-10 text-right">{b.count}</span>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Week/Depth Pattern */}
      {depthDistribution.length > 0 && (
        <Card title="Week/Depth Pattern (False Skips)">
          <p className="text-xs text-gray-500 mb-3">
            Most common week/depth combinations where false skips occurred.
          </p>
          <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
            {depthDistribution.map(d => (
              <div key={d.weekDepth} className="bg-gray-800 rounded p-2 text-center">
                <div className="text-xs text-blue-400 font-mono">{d.weekDepth}</div>
                <div className="text-sm font-bold text-red-400">{d.count}</div>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: string | number; colour: string }) {
  const colourClass = colour === 'red' ? 'text-red-400' : colour === 'amber' ? 'text-amber-400' : 'text-blue-400';
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-lg font-bold ${colourClass}`}>{value}</div>
    </div>
  );
}
