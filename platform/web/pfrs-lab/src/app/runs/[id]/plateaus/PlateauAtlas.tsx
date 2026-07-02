'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { PlateauEvent } from '@/lib/types';

interface Props {
  plateaus: PlateauEvent[];
  numWeeks: number;
}

export default function PlateauAtlas({ plateaus, numWeeks }: Props) {
  const [filterWeek, setFilterWeek] = useState<number | null>(null);
  const [filterWorker, setFilterWorker] = useState<number | null>(null);

  const weeks = useMemo(() => Array.from(new Set(plateaus.map(p => p.week))).sort((a, b) => a - b), [plateaus]);
  const workers = useMemo(() => Array.from(new Set(plateaus.map(p => p.workerID))).sort((a, b) => a - b), [plateaus]);

  const filtered = useMemo(() => {
    let data = plateaus;
    if (filterWeek !== null) data = data.filter(p => p.week === filterWeek);
    if (filterWorker !== null) data = data.filter(p => p.workerID === filterWorker);
    return data;
  }, [plateaus, filterWeek, filterWorker]);

  // Statistics.
  const durations = filtered.map(p => p.candsSinceImprove);
  const longest = Math.max(...durations, 0);
  const avg = durations.length > 0 ? Math.round(durations.reduce((s, d) => s + d, 0) / durations.length) : 0;
  const sorted = [...durations].sort((a, b) => a - b);
  const median = sorted.length > 0 ? sorted[Math.floor(sorted.length / 2)] : 0;

  // Per-week counts.
  const perWeek = useMemo(() => {
    const counts = new Map<number, number>();
    for (const p of filtered) counts.set(p.week, (counts.get(p.week) || 0) + 1);
    return counts;
  }, [filtered]);

  const mostCommonWeek = useMemo(() => {
    let best = 0, bestCount = 0;
    for (const [w, c] of perWeek) { if (c > bestCount) { best = w; bestCount = c; } }
    return { week: best, count: bestCount };
  }, [perWeek]);

  // Per-worker counts.
  const perWorker = useMemo(() => {
    const counts = new Map<number, number>();
    for (const p of filtered) counts.set(p.workerID, (counts.get(p.workerID) || 0) + 1);
    return counts;
  }, [filtered]);

  const mostCommonWorker = useMemo(() => {
    let best = 0, bestCount = 0;
    for (const [w, c] of perWorker) { if (c > bestCount) { best = w; bestCount = c; } }
    return { worker: best, count: bestCount };
  }, [perWorker]);

  // Duration histogram bins.
  const histBins = useMemo(() => {
    if (durations.length === 0) return [];
    const numBins = 25;
    const binSize = Math.max(longest / numBins, 1);
    const bins = Array(numBins).fill(0);
    for (const d of durations) {
      const bin = Math.min(Math.floor(d / binSize), numBins - 1);
      bins[bin]++;
    }
    return bins;
  }, [durations, longest]);

  // Long plateaus (> 75th percentile).
  const p75 = sorted.length > 0 ? sorted[Math.floor(sorted.length * 0.75)] : 0;
  const longPlateaus = filtered.filter(p => p.candsSinceImprove > p75);
  const longPerWeek = new Map<number, number>();
  for (const p of longPlateaus) longPerWeek.set(p.week, (longPerWeek.get(p.week) || 0) + 1);

  // Observations.
  const observations = useMemo(() => {
    const obs: string[] = [];
    if (mostCommonWeek.count > 0) {
      const pct = (mostCommonWeek.count / filtered.length * 100).toFixed(0);
      obs.push(`Week ${mostCommonWeek.week} has the most plateaus (${mostCommonWeek.count}, ${pct}% of total).`);
    }
    // Long plateau concentration.
    if (longPlateaus.length > 0) {
      let worstWeek = 0, worstCount = 0;
      for (const [w, c] of longPerWeek) { if (c > worstCount) { worstWeek = w; worstCount = c; } }
      if (worstCount > longPlateaus.length * 0.3) {
        obs.push(`Week ${worstWeek} contains ${((worstCount / longPlateaus.length) * 100).toFixed(0)}% of all long plateaus.`);
      }
    }
    if (longest > 0) {
      obs.push(`Longest plateau: ${longest.toLocaleString()} iterations without improvement.`);
    }
    if (avg > 1000) {
      obs.push(`Average plateau duration is high (${avg.toLocaleString()} iterations) — search may benefit from more aggressive perturbation.`);
    }
    const escapedLong = filtered.filter(p => p.candsSinceImprove > 1200);
    if (escapedLong.length === 0 && filtered.length > 0) {
      obs.push(`No worker experienced a plateau longer than 1,200 iterations.`);
    }
    return obs;
  }, [filtered, longPlateaus, longPerWeek, mostCommonWeek, longest, avg]);

  // Heatmap data: week × depth → count.
  const heatmapData = useMemo(() => {
    const maxDepth = Math.max(...filtered.map(p => p.depth), 0);
    const grid: number[][] = [];
    for (let w = 0; w <= numWeeks; w++) {
      grid.push(Array(maxDepth + 1).fill(0));
    }
    for (const p of filtered) {
      if (p.week < grid.length && p.depth < grid[0].length) {
        grid[p.week][p.depth]++;
      }
    }
    return { grid, maxDepth };
  }, [filtered, numWeeks]);

  function heatColor(count: number, maxCount: number): string {
    if (count === 0) return '#111827';
    const intensity = Math.min(count / Math.max(maxCount, 1), 1);
    const r = Math.round(239 * intensity);
    const g = Math.round(68 * (1 - intensity * 0.5));
    const b = Math.round(68 * (1 - intensity * 0.5));
    return `rgb(${r}, ${g}, ${b})`;
  }

  const heatMax = Math.max(...heatmapData.grid.flat(), 1);

  return (
    <div className="space-y-4">
      {/* Filters */}
      <Card title="Filters">
        <div className="flex flex-wrap gap-2">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Week:</span>
            <button onClick={() => setFilterWeek(null)}
              className={`px-2 py-0.5 rounded text-[10px] ${filterWeek === null ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>All</button>
            {weeks.map(w => (
              <button key={w} onClick={() => setFilterWeek(w)}
                className={`px-2 py-0.5 rounded text-[10px] ${filterWeek === w ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{w}</button>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Worker:</span>
            <button onClick={() => setFilterWorker(null)}
              className={`px-2 py-0.5 rounded text-[10px] ${filterWorker === null ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>All</button>
            {workers.slice(0, 12).map(w => (
              <button key={w} onClick={() => setFilterWorker(w)}
                className={`px-2 py-0.5 rounded text-[10px] ${filterWorker === w ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{w}</button>
            ))}
            {workers.length > 12 && <span className="text-[9px] text-gray-600">+{workers.length - 12}</span>}
          </div>
        </div>
      </Card>

      {/* Stats */}
      <div className="grid grid-cols-5 gap-3">
        <Card title="Total Plateaus"><p className="text-2xl font-bold text-orange-400">{filtered.length.toLocaleString()}</p></Card>
        <Card title="Longest"><p className="text-2xl font-bold text-red-400">{longest.toLocaleString()}</p></Card>
        <Card title="Average"><p className="text-2xl font-bold text-blue-400">{avg.toLocaleString()}</p></Card>
        <Card title="Median"><p className="text-2xl font-bold text-gray-300">{median.toLocaleString()}</p></Card>
        <Card title="Most Common Week"><p className="text-2xl font-bold text-purple-400">W{mostCommonWeek.week}</p></Card>
      </div>

      {/* Duration histogram */}
      <Card title="Plateau Duration Distribution">
        <div className="flex items-end gap-px h-28">
          {histBins.map((count, i) => {
            const maxBin = Math.max(...histBins, 1);
            return (
              <div key={i} className="flex-1 flex flex-col justify-end">
                <div
                  className="bg-orange-500 rounded-t min-w-[2px]"
                  style={{ height: `${(count / maxBin) * 100}%`, minHeight: count > 0 ? '2px' : '0' }}
                  title={`${count} plateaus`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Short</span>
          <span>Long ({longest.toLocaleString()} max)</span>
        </div>
      </Card>

      {/* Heatmap: week × depth */}
      <Card title="Heatmap: Week × Depth">
        <div className="overflow-x-auto">
          <div className="inline-grid gap-px" style={{
            gridTemplateColumns: `40px repeat(${heatmapData.maxDepth + 1}, 1fr)`,
          }}>
            {/* Header */}
            <div className="text-[8px] text-gray-600" />
            {Array.from({ length: heatmapData.maxDepth + 1 }, (_, d) => (
              <div key={d} className="text-[8px] text-gray-600 text-center">D{d}</div>
            ))}
            {/* Rows */}
            {weeks.map(w => (
              <>
                <div key={`l-${w}`} className="text-[8px] text-gray-500 flex items-center">W{w}</div>
                {Array.from({ length: heatmapData.maxDepth + 1 }, (_, d) => {
                  const count = w < heatmapData.grid.length && d < heatmapData.grid[0].length
                    ? heatmapData.grid[w][d] : 0;
                  return (
                    <div key={`${w}-${d}`}
                      className="w-6 h-5 rounded-sm"
                      style={{ background: heatColor(count, heatMax) }}
                      title={`Week ${w}, Depth ${d}: ${count} plateaus`}
                    />
                  );
                })}
              </>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2 mt-2 text-[9px] text-gray-500">
          <span>Low</span>
          <div className="flex gap-px">
            {[0, 0.25, 0.5, 0.75, 1].map(f => (
              <div key={f} className="w-4 h-3 rounded-sm"
                style={{ background: heatColor(f * heatMax, heatMax) }} />
            ))}
          </div>
          <span>High</span>
        </div>
      </Card>

      {/* Bubble chart: week vs worker, size = duration */}
      <Card title="Bubble: Week × Worker (size = duration)">
        <svg viewBox="0 0 600 200" className="w-full h-48">
          {filtered.slice(0, 500).map((p, i) => {
            const x = weeks.length > 1 ? 40 + ((p.week - weeks[0]) / (weeks[weeks.length - 1] - weeks[0])) * 520 : 300;
            const workerIdx = workers.indexOf(p.workerID);
            const y = workers.length > 1 ? 20 + (workerIdx / (workers.length - 1)) * 160 : 100;
            const r = Math.min(Math.max(p.candsSinceImprove / longest * 12, 2), 12);
            const opacity = Math.min(0.3 + (p.candsSinceImprove / longest) * 0.7, 1);
            return (
              <circle key={i} cx={x} cy={y} r={r}
                fill="#f97316" opacity={opacity}
              >
                <title>{`W${p.week} Worker ${p.workerID}: ${p.candsSinceImprove} iterations`}</title>
              </circle>
            );
          })}
          {/* Axis labels */}
          <text x="300" y="198" textAnchor="middle" className="fill-gray-600 text-[8px]">Week</text>
          <text x="8" y="100" textAnchor="middle" transform="rotate(-90, 8, 100)" className="fill-gray-600 text-[8px]">Worker</text>
        </svg>
      </Card>

      {/* Observations */}
      <Card title="Observations">
        {observations.length === 0 ? (
          <p className="text-gray-600 text-sm italic">No notable patterns detected.</p>
        ) : (
          <div className="space-y-2">
            {observations.map((obs, i) => (
              <p key={i} className="text-sm text-gray-300">{obs}</p>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}
