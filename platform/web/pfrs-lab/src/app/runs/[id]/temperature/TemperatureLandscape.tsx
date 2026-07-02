'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { DiscoveryRecord, WorkerLifecycle, RunMetadata } from '@/lib/types';

interface Props {
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
  metadata: RunMetadata | null;
}

export default function TemperatureLandscape({ discoveries, workers, metadata }: Props) {
  // Temperature data from discoveries.
  const tempData = useMemo(() => {
    return discoveries
      .filter(d => d.temperatureAtEvent > 0)
      .map(d => ({
        temp: d.temperatureAtEvent,
        penalty: d.newBest,
        improvement: d.improvement,
        elapsedMs: d.elapsedMs,
        type: d.eventType,
        worker: d.workerID,
        week: d.week,
      }));
  }, [discoveries]);

  // Worker temperature ranges.
  const workerTemps = useMemo(() => {
    return workers.map(w => ({
      workerID: w.workerID,
      week: w.week,
      initial: w.initialTemperature,
      final: w.finalTemperature,
      atBest: w.temperatureAtBest,
      producedGlobal: w.producedGlobalBest,
    }));
  }, [workers]);

  // Temperature bins for heatmap.
  const heatmap = useMemo(() => {
    if (tempData.length === 0) return { bins: [], maxTemp: 0, numBins: 0 };
    const maxTemp = Math.max(...tempData.map(d => d.temp));
    const numBins = 20;
    const binSize = maxTemp / numBins;
    const bins = Array.from({ length: numBins }, () => ({
      count: 0, improvements: 0, totalImprovement: 0, globalBests: 0,
    }));
    for (const d of tempData) {
      const bin = Math.min(Math.floor(d.temp / binSize), numBins - 1);
      bins[bin].count++;
      bins[bin].improvements++;
      bins[bin].totalImprovement += d.improvement;
      if (d.type === 'global_best') bins[bin].globalBests++;
    }
    return { bins, maxTemp, numBins };
  }, [tempData]);

  // Find optimal temperature range (where most improvements happened).
  const optimalRange = useMemo(() => {
    if (heatmap.bins.length === 0) return { low: 0, high: 0 };
    const binSize = heatmap.maxTemp / heatmap.numBins;
    let bestBin = 0, bestCount = 0;
    for (let i = 0; i < heatmap.bins.length; i++) {
      if (heatmap.bins[i].totalImprovement > bestCount) {
        bestCount = heatmap.bins[i].totalImprovement;
        bestBin = i;
      }
    }
    return { low: +(bestBin * binSize).toFixed(2), high: +((bestBin + 1) * binSize).toFixed(2) };
  }, [heatmap]);

  // Acceptance collapse point: where discoveries stop.
  const collapseTemp = useMemo(() => {
    if (tempData.length === 0) return 0;
    const sorted = [...tempData].sort((a, b) => a.temp - b.temp);
    // Find lowest temperature with a discovery.
    return sorted[0]?.temp || 0;
  }, [tempData]);

  // Cooling curve: temperature over time.
  const coolingCurve = useMemo(() => {
    if (tempData.length === 0) return [];
    const sorted = [...tempData].sort((a, b) => a.elapsedMs - b.elapsedMs);
    // Sample every nth point for chart.
    const step = Math.max(1, Math.floor(sorted.length / 100));
    return sorted.filter((_, i) => i % step === 0);
  }, [tempData]);

  // Phase detection.
  const phases = useMemo(() => {
    if (tempData.length === 0) return { explorationEnd: 0, exploitationStart: 0 };
    const maxTemp = heatmap.maxTemp;
    const threshold = maxTemp * 0.3;
    const explorationEnd = tempData.findIndex(d => d.temp < threshold);
    return {
      explorationEnd: explorationEnd >= 0 ? tempData[explorationEnd].elapsedMs : 0,
      exploitationStart: explorationEnd >= 0 ? tempData[explorationEnd].elapsedMs : 0,
    };
  }, [tempData, heatmap.maxTemp]);

  // Observations.
  const observations = useMemo(() => {
    const obs: string[] = [];
    if (optimalRange.high > 0) {
      obs.push(`Most improvements occurred between temperatures ${optimalRange.low} and ${optimalRange.high}.`);
    }
    if (collapseTemp > 0) {
      obs.push(`Lowest temperature with a discovery: ${collapseTemp.toFixed(4)}.`);
    }
    if (phases.explorationEnd > 0) {
      obs.push(`Exploration phase ended around ${(phases.explorationEnd / 1000).toFixed(1)}s (temp dropped below 30% of initial).`);
    }
    const globalByTemp = tempData.filter(d => d.type === 'global_best');
    if (globalByTemp.length > 0) {
      const avgTemp = globalByTemp.reduce((s, d) => s + d.temp, 0) / globalByTemp.length;
      obs.push(`Average temperature at global best discovery: ${avgTemp.toFixed(4)}.`);
    }
    const highTempDisc = tempData.filter(d => d.temp > heatmap.maxTemp * 0.5).length;
    const lowTempDisc = tempData.filter(d => d.temp <= heatmap.maxTemp * 0.5).length;
    if (highTempDisc > lowTempDisc * 2) {
      obs.push('Most discoveries happened during the hot (exploration) phase.');
    } else if (lowTempDisc > highTempDisc * 2) {
      obs.push('Most discoveries happened during the cold (exploitation) phase.');
    }
    return obs;
  }, [tempData, optimalRange, collapseTemp, phases, heatmap.maxTemp]);

  const maxTemp = heatmap.maxTemp || 1;
  const maxImp = Math.max(...tempData.map(d => d.improvement), 1);

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-4 gap-3">
        <Card title="Initial Temp">
          <p className="text-2xl font-bold text-orange-400">{metadata?.initialTemperature?.toFixed(2) || '—'}</p>
        </Card>
        <Card title="Cooling Mode">
          <p className="text-2xl font-bold text-blue-400">{metadata?.coolingMode || '—'}</p>
        </Card>
        <Card title="Optimal Range">
          <p className="text-lg font-bold text-emerald-400">{optimalRange.low}–{optimalRange.high}</p>
        </Card>
        <Card title="Discoveries">
          <p className="text-2xl font-bold text-purple-400">{tempData.length}</p>
        </Card>
      </div>

      {/* Cooling curve */}
      <Card title="Cooling Curve (Temperature over Time)">
        <svg viewBox="0 0 700 160" className="w-full h-40 bg-gray-900 rounded border border-gray-800">
          {/* Phase background */}
          {phases.exploitationStart > 0 && coolingCurve.length > 0 && (
            <>
              <rect x="40" y="10" width={`${(phases.exploitationStart / (coolingCurve[coolingCurve.length-1]?.elapsedMs || 1)) * 620}%`} height="140" fill="rgba(239, 68, 68, 0.05)" />
              <text x="50" y="25" className="fill-red-400 text-[7px]">Exploration</text>
              <text x={45 + (phases.exploitationStart / (coolingCurve[coolingCurve.length-1]?.elapsedMs || 1)) * 620} y="25" className="fill-blue-400 text-[7px]">Exploitation</text>
            </>
          )}
          {/* Curve */}
          {coolingCurve.length > 1 && (
            <polyline
              points={coolingCurve.map(d => {
                const x = 40 + (d.elapsedMs / (coolingCurve[coolingCurve.length-1]?.elapsedMs || 1)) * 620;
                const y = 150 - (d.temp / maxTemp) * 130;
                return `${x},${y}`;
              }).join(' ')}
              fill="none" stroke="#f97316" strokeWidth="1.5"
            />
          )}
          <text x="350" y="158" textAnchor="middle" className="fill-gray-600 text-[8px]">Time</text>
          <text x="15" y="80" className="fill-gray-600 text-[8px]" transform="rotate(-90, 15, 80)">Temp</text>
        </svg>
      </Card>

      {/* Temperature heatmap: discoveries per temperature band */}
      <Card title="Discovery Heatmap (by Temperature)">
        <div className="flex items-end gap-px h-28">
          {heatmap.bins.map((bin, i) => {
            const maxCount = Math.max(...heatmap.bins.map(b => b.totalImprovement), 1);
            const height = (bin.totalImprovement / maxCount) * 100;
            const hasGlobal = bin.globalBests > 0;
            return (
              <div key={i} className="flex-1 flex flex-col justify-end items-center">
                {hasGlobal && <span className="text-[8px] text-yellow-400 mb-0.5">★</span>}
                <div
                  className={`w-full rounded-t ${hasGlobal ? 'bg-yellow-500' : 'bg-orange-500'}`}
                  style={{ height: `${Math.max(height, bin.count > 0 ? 3 : 0)}%` }}
                  title={`Temp ${(i * maxTemp / heatmap.numBins).toFixed(2)}–${((i+1) * maxTemp / heatmap.numBins).toFixed(2)}: ${bin.count} disc, imp=${bin.totalImprovement}`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Cold (0)</span>
          <span>Hot ({maxTemp.toFixed(2)})</span>
        </div>
      </Card>

      {/* Scatter: temperature vs improvement */}
      <Card title="Temperature vs Improvement (Scatter)">
        <svg viewBox="0 0 700 200" className="w-full h-48 bg-gray-900 rounded border border-gray-800">
          {tempData.slice(0, 1000).map((d, i) => {
            const x = 40 + (d.temp / maxTemp) * 620;
            const y = 180 - (d.improvement / maxImp) * 160;
            const isGlobal = d.type === 'global_best';
            return (
              <circle key={i} cx={x} cy={y}
                r={isGlobal ? 4 : 2}
                fill={isGlobal ? '#fbbf24' : '#f97316'}
                opacity={0.6}
              >
                <title>{`T=${d.temp.toFixed(4)} imp=${d.improvement} W${d.week}`}</title>
              </circle>
            );
          })}
          <text x="350" y="198" textAnchor="middle" className="fill-gray-600 text-[8px]">Temperature</text>
          <text x="15" y="100" className="fill-gray-600 text-[8px]" transform="rotate(-90, 15, 100)">Improvement</text>
        </svg>
      </Card>

      {/* Worker temperature ranges */}
      <Card title="Worker Temperature Ranges">
        <div className="space-y-0.5 max-h-48 overflow-y-auto">
          {workerTemps.slice(0, 30).map(w => {
            const startPct = (w.initial / maxTemp) * 100;
            const endPct = (w.final / maxTemp) * 100;
            const bestPct = (w.atBest / maxTemp) * 100;
            return (
              <div key={`${w.week}-${w.workerID}`} className="flex items-center gap-2">
                <span className="w-16 text-[9px] font-mono text-gray-500">W{w.week}#{w.workerID}</span>
                <div className="flex-1 h-3 bg-gray-800 rounded relative">
                  <div className="absolute h-full bg-gradient-to-r from-red-500 to-blue-500 rounded opacity-30"
                    style={{ left: `${Math.min(endPct, startPct)}%`, width: `${Math.abs(startPct - endPct)}%` }} />
                  <div className="absolute w-1 h-full bg-emerald-400 rounded"
                    style={{ left: `${bestPct}%` }} title={`Best at T=${w.atBest.toFixed(4)}`} />
                </div>
                {w.producedGlobal && <span className="text-[8px] text-yellow-400">★</span>}
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Cold</span><span>Hot</span>
        </div>
        <p className="text-[9px] text-gray-600 mt-1">Green bar = temperature at best discovery</p>
      </Card>

      {/* Observations */}
      <Card title="Observations">
        <div className="space-y-2">
          {observations.map((obs, i) => (
            <p key={i} className="text-sm text-gray-300">{obs}</p>
          ))}
        </div>
      </Card>
    </div>
  );
}
