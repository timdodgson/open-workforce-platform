'use client';

import { useMemo, useState } from 'react';
import Card from '@/components/Card';
import { WhatIfPrediction } from './page';
import SimulationCharts from './SimulationCharts';
import WorkerReplay from './WorkerReplay';

interface Props {
  predictions: WhatIfPrediction[];
}

type RecommendationMode = 'skip' | 'reduce_budget' | 'increase_budget' | 'run';
type BudgetPolicy = 'conservative' | 'balanced' | 'aggressive';
type AlgorithmPolicy = 'keep_current' | 'ai_recommend';

interface SimControls {
  confidenceThreshold: number;
  recommendationMode: RecommendationMode;
  budgetPolicy: BudgetPolicy;
  algorithmPolicy: AlgorithmPolicy;
}

interface SimulatedWorker {
  prediction: WhatIfPrediction;
  action: 'run' | 'skip' | 'reduce' | 'increase';
  budgetMultiplier: number;
  reason: string;
}

interface SimResult {
  totalWorkers: number;
  workersRun: number;
  workersSkipped: number;
  workersReduced: number;
  workersIncreased: number;
  cpuSavedPct: number;
  expectedImprovement: number;
  actualImprovement: number;
  globalBestsMissed: number;
  totalGlobalBests: number;
  avgRoi: number;
  avgImprovement: number;
  missedImprovementTotal: number;
  safetyStatus: 'SAFE' | 'CAUTION' | 'UNSAFE';
  workers: SimulatedWorker[];
}

function simulate(predictions: WhatIfPrediction[], controls: SimControls): SimResult {
  const workers: SimulatedWorker[] = [];
  let skippedRuntime = 0;
  let totalRuntime = 0;
  let globalBestsMissed = 0;
  let totalGlobalBests = 0;
  let missedImprovementTotal = 0;

  // Estimate runtime per worker (use iteration budget as proxy).
  const avgRuntime = 1; // normalised unit

  for (const p of predictions) {
    totalRuntime += avgRuntime;

    if (p.actual.produced_global_best) totalGlobalBests++;

    // Determine what the model would recommend.
    const pImproved = p.predicted.p_improved;
    const pGlobalBest = p.predicted.p_global_best;
    const confidence = Math.max(pImproved, 1 - pImproved); // how certain the model is

    let action: 'run' | 'skip' | 'reduce' | 'increase' = 'run';
    let budgetMultiplier = 1.0;
    let reason = 'default_run';

    // Only act if confidence exceeds threshold.
    if (confidence >= controls.confidenceThreshold) {
      if (controls.recommendationMode === 'skip') {
        // Skip mode: skip workers the model says won't improve.
        if (pImproved < 0.5) {
          action = 'skip';
          reason = `P(improved)=${(pImproved * 100).toFixed(0)}% < 50%`;
        }
      } else if (controls.recommendationMode === 'reduce_budget') {
        // Reduce budget for low-value workers.
        if (pImproved < 0.3) {
          action = 'skip';
          reason = `P(improved)=${(pImproved * 100).toFixed(0)}% < 30%`;
        } else if (pImproved < 0.5) {
          action = 'reduce';
          budgetMultiplier = controls.budgetPolicy === 'aggressive' ? 0.25
            : controls.budgetPolicy === 'balanced' ? 0.5 : 0.75;
          reason = `P(improved)=${(pImproved * 100).toFixed(0)}% → reduce budget`;
        }
      } else if (controls.recommendationMode === 'increase_budget') {
        // Increase budget for high-value workers, reduce for low-value.
        if (pGlobalBest > 0.3) {
          action = 'increase';
          budgetMultiplier = controls.budgetPolicy === 'aggressive' ? 3.0
            : controls.budgetPolicy === 'balanced' ? 2.0 : 1.5;
          reason = `P(global_best)=${(pGlobalBest * 100).toFixed(0)}% → increase budget`;
        } else if (pImproved < 0.3) {
          action = 'reduce';
          budgetMultiplier = 0.5;
          reason = `P(improved)=${(pImproved * 100).toFixed(0)}% → reduce budget`;
        }
      }
      // 'run' mode = no changes, baseline.
    }

    // Protect global best candidates regardless of mode.
    if (pGlobalBest > 0.5 && action === 'skip') {
      action = 'run';
      reason = `Protected: P(global_best)=${(pGlobalBest * 100).toFixed(0)}%`;
    }

    workers.push({ prediction: p, action, budgetMultiplier, reason });

    if (action === 'skip') {
      skippedRuntime += avgRuntime;
      if (p.actual.improved) missedImprovementTotal += p.actual.improvement_amount;
      if (p.actual.produced_global_best) globalBestsMissed++;
    } else if (action === 'reduce') {
      skippedRuntime += avgRuntime * (1 - budgetMultiplier);
    }
  }

  const workersSkipped = workers.filter(w => w.action === 'skip').length;
  const workersReduced = workers.filter(w => w.action === 'reduce').length;
  const workersIncreased = workers.filter(w => w.action === 'increase').length;
  const workersRun = workers.filter(w => w.action === 'run').length;

  const cpuSavedPct = totalRuntime > 0 ? (skippedRuntime / totalRuntime) * 100 : 0;

  const actualImprovement = predictions.reduce((s, p) => s + p.actual.improvement_amount, 0);
  const expectedImprovement = actualImprovement - missedImprovementTotal;

  const runWorkers = workers.filter(w => w.action !== 'skip');
  const avgRoi = runWorkers.length > 0
    ? runWorkers.reduce((s, w) => s + w.prediction.actual.roi, 0) / runWorkers.length : 0;
  const avgImprovement = runWorkers.length > 0
    ? runWorkers.reduce((s, w) => s + w.prediction.actual.improvement_amount, 0) / runWorkers.length : 0;

  let safetyStatus: 'SAFE' | 'CAUTION' | 'UNSAFE' = 'SAFE';
  if (globalBestsMissed > 0) safetyStatus = 'UNSAFE';
  else if (cpuSavedPct > 30 || missedImprovementTotal > actualImprovement * 0.1) safetyStatus = 'CAUTION';

  return {
    totalWorkers: predictions.length,
    workersRun,
    workersSkipped,
    workersReduced,
    workersIncreased,
    cpuSavedPct,
    expectedImprovement,
    actualImprovement,
    globalBestsMissed,
    totalGlobalBests,
    avgRoi,
    avgImprovement,
    missedImprovementTotal,
    safetyStatus,
    workers,
  };
}

export default function WhatIfLab({ predictions }: Props) {
  const [controls, setControls] = useState<SimControls>({
    confidenceThreshold: 0.6,
    recommendationMode: 'skip',
    budgetPolicy: 'balanced',
    algorithmPolicy: 'keep_current',
  });
  const [selectedWorkerIdx, setSelectedWorkerIdx] = useState<number | null>(null);

  const result = useMemo(() => simulate(predictions, controls), [predictions, controls]);

  // Generate threshold curve data.
  const thresholdCurve = useMemo(() => {
    const points = [];
    for (let t = 0.5; t <= 0.99; t += 0.05) {
      const r = simulate(predictions, { ...controls, confidenceThreshold: t });
      points.push({
        threshold: parseFloat(t.toFixed(2)),
        cpuSaved: parseFloat(r.cpuSavedPct.toFixed(1)),
        globalBestsMissed: r.globalBestsMissed,
        improvement: parseFloat(((r.expectedImprovement / Math.max(r.actualImprovement, 1)) * 100).toFixed(1)),
        skipped: r.workersSkipped,
      });
    }
    return points;
  }, [predictions, controls]);

  const selectedWorker = selectedWorkerIdx !== null ? result.workers[selectedWorkerIdx] : null;

  return (
    <div className="space-y-4">
      {/* Header + Safety Banner */}
      <Card title="What-If Lab — Offline Simulation">
        <p className="text-xs text-gray-500 mb-3">
          Simulates how the optimiser would have behaved if the Worker Value Model had been trusted.
          This is offline analysis only — no optimiser behaviour changes.
        </p>

        {/* Safety Status */}
        <div className={`rounded-lg p-4 mb-4 border ${
          result.safetyStatus === 'SAFE' ? 'bg-emerald-900/20 border-emerald-700' :
          result.safetyStatus === 'CAUTION' ? 'bg-amber-900/20 border-amber-700' :
          'bg-red-900/20 border-red-700'
        }`}>
          <div className="flex items-center gap-3">
            <span className="text-2xl">
              {result.safetyStatus === 'SAFE' ? '✓' : result.safetyStatus === 'CAUTION' ? '⚠' : '✗'}
            </span>
            <div>
              <p className={`text-sm font-bold ${
                result.safetyStatus === 'SAFE' ? 'text-emerald-400' :
                result.safetyStatus === 'CAUTION' ? 'text-amber-400' : 'text-red-400'
              }`}>
                {result.safetyStatus}
              </p>
              <p className="text-xs text-gray-400">
                {result.safetyStatus === 'SAFE' && 'No global bests missed. CPU savings appear safe.'}
                {result.safetyStatus === 'CAUTION' && 'Significant savings but verify no critical improvements are lost.'}
                {result.safetyStatus === 'UNSAFE' && `${result.globalBestsMissed} global best(s) would have been missed. Do not trust at this threshold.`}
              </p>
            </div>
            <div className="ml-auto text-right">
              <p className="text-xs text-gray-500">CPU Saved</p>
              <p className="text-lg font-bold text-blue-400">{result.cpuSavedPct.toFixed(1)}%</p>
            </div>
          </div>
        </div>

        {/* Summary Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <Stat label="Workers Skipped" value={result.workersSkipped} colour="amber" />
          <Stat label="Workers Reduced" value={result.workersReduced} colour="blue" />
          <Stat label="Workers Increased" value={result.workersIncreased} colour="emerald" />
          <Stat label="Global Bests Missed" value={result.globalBestsMissed} colour={result.globalBestsMissed > 0 ? 'red' : 'emerald'} />
          <Stat label="Missed Improvement" value={result.missedImprovementTotal.toLocaleString()} colour="amber" />
        </div>
      </Card>

      {/* Simulation Controls */}
      <Card title="Simulation Controls">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {/* Confidence Threshold */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">
              Confidence Threshold: {controls.confidenceThreshold.toFixed(2)}
            </label>
            <input
              type="range"
              min="0.50"
              max="0.99"
              step="0.01"
              value={controls.confidenceThreshold}
              onChange={e => setControls(c => ({ ...c, confidenceThreshold: parseFloat(e.target.value) }))}
              className="w-full accent-blue-500"
            />
            <div className="flex justify-between text-[8px] text-gray-600">
              <span>0.50 (aggressive)</span>
              <span>0.99 (conservative)</span>
            </div>
          </div>

          {/* Recommendation Mode */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Recommendation Mode</label>
            <select
              className="w-full bg-gray-800 border border-gray-700 rounded text-xs p-1.5 text-gray-200"
              value={controls.recommendationMode}
              onChange={e => setControls(c => ({ ...c, recommendationMode: e.target.value as RecommendationMode }))}
            >
              <option value="run">Run (baseline — no changes)</option>
              <option value="skip">Skip low-value workers</option>
              <option value="reduce_budget">Reduce budget for low-value</option>
              <option value="increase_budget">Increase budget for high-value</option>
            </select>
          </div>

          {/* Budget Policy */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Budget Policy</label>
            <select
              className="w-full bg-gray-800 border border-gray-700 rounded text-xs p-1.5 text-gray-200"
              value={controls.budgetPolicy}
              onChange={e => setControls(c => ({ ...c, budgetPolicy: e.target.value as BudgetPolicy }))}
            >
              <option value="conservative">Conservative (75% / 1.5×)</option>
              <option value="balanced">Balanced (50% / 2×)</option>
              <option value="aggressive">Aggressive (25% / 3×)</option>
            </select>
          </div>

          {/* Algorithm Policy */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Algorithm Policy</label>
            <select
              className="w-full bg-gray-800 border border-gray-700 rounded text-xs p-1.5 text-gray-200"
              value={controls.algorithmPolicy}
              onChange={e => setControls(c => ({ ...c, algorithmPolicy: e.target.value as AlgorithmPolicy }))}
            >
              <option value="keep_current">Keep Current</option>
              <option value="ai_recommend">Allow Search Intelligence to Recommend</option>
            </select>
          </div>
        </div>
      </Card>

      {/* Charts */}
      <SimulationCharts result={result} thresholdCurve={thresholdCurve} />

      {/* Research Questions */}
      <Card title="Research Questions">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <QuestionCard
            question="What confidence threshold gives the best trade-off?"
            answer={(() => {
              const safe = thresholdCurve.filter(p => p.globalBestsMissed === 0);
              if (safe.length === 0) return 'No safe threshold found — model cannot reliably skip workers.';
              const best = safe.reduce((a, b) => a.cpuSaved > b.cpuSaved ? a : b);
              return `Threshold ${best.threshold} saves ${best.cpuSaved}% CPU with zero global bests missed.`;
            })()}
          />
          <QuestionCard
            question="How much CPU could have been saved?"
            answer={`At current settings: ${result.cpuSavedPct.toFixed(1)}% CPU saved (${result.workersSkipped} workers skipped, ${result.workersReduced} reduced).`}
          />
          <QuestionCard
            question="Would any global bests have been lost?"
            answer={result.globalBestsMissed === 0
              ? `No. All ${result.totalGlobalBests} global best discoveries are preserved at this threshold.`
              : `Yes — ${result.globalBestsMissed} of ${result.totalGlobalBests} global bests would have been missed.`}
          />
          <QuestionCard
            question="Where does the model become unsafe?"
            answer={(() => {
              const unsafe = thresholdCurve.find(p => p.globalBestsMissed > 0);
              if (!unsafe) return 'The model appears safe at all tested thresholds.';
              return `First global best loss occurs at threshold ${unsafe.threshold} (${unsafe.cpuSaved}% CPU saved).`;
            })()}
          />
        </div>
      </Card>

      {/* Worker Replay */}
      <Card title="Worker Replay">
        <p className="text-xs text-gray-500 mb-3">
          Select a worker to see what would have happened under this simulation.
        </p>
        <div className="overflow-x-auto max-h-[300px] overflow-y-auto mb-4">
          <table className="w-full text-[10px]">
            <thead className="sticky top-0 bg-gray-850">
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">#</th>
                <th className="text-left p-1.5">Instance</th>
                <th className="text-left p-1.5">Algo</th>
                <th className="text-center p-1.5">Action</th>
                <th className="text-right p-1.5">P(Impr)</th>
                <th className="text-center p-1.5">Actually?</th>
                <th className="text-center p-1.5">GB?</th>
                <th className="text-left p-1.5">Reason</th>
              </tr>
            </thead>
            <tbody>
              {result.workers.slice(0, 100).map((w, i) => (
                <tr
                  key={i}
                  className={`border-t border-gray-800 cursor-pointer hover:bg-gray-800/50 ${
                    selectedWorkerIdx === i ? 'bg-blue-900/20' : ''
                  } ${w.action === 'skip' && w.prediction.actual.improved ? 'bg-red-900/10' : ''}`}
                  onClick={() => setSelectedWorkerIdx(i)}
                >
                  <td className="p-1.5 text-gray-500">{w.prediction.index}</td>
                  <td className="p-1.5 text-blue-400 truncate max-w-[80px]">{w.prediction.instance}</td>
                  <td className="p-1.5 text-emerald-400">{w.prediction.algorithm}</td>
                  <td className="text-center p-1.5">
                    <ActionBadge action={w.action} />
                  </td>
                  <td className="text-right p-1.5">{(w.prediction.predicted.p_improved * 100).toFixed(0)}%</td>
                  <td className="text-center p-1.5">
                    {w.prediction.actual.improved ? <span className="text-emerald-400">✓</span> : <span className="text-gray-600">✗</span>}
                  </td>
                  <td className="text-center p-1.5">
                    {w.prediction.actual.produced_global_best ? <span className="text-amber-400">⭐</span> : '—'}
                  </td>
                  <td className="p-1.5 text-gray-500 truncate max-w-[150px]">{w.reason}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {result.workers.length > 100 && (
            <p className="text-[10px] text-gray-500 p-2 text-center">Showing first 100.</p>
          )}
        </div>

        {selectedWorker && <WorkerReplay worker={selectedWorker} />}
      </Card>
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: string | number; colour: string }) {
  const colourMap: Record<string, string> = {
    blue: 'text-blue-400', emerald: 'text-emerald-400', amber: 'text-amber-400', red: 'text-red-400',
  };
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-lg font-bold ${colourMap[colour] || 'text-gray-300'}`}>{value}</div>
    </div>
  );
}

function QuestionCard({ question, answer }: { question: string; answer: string }) {
  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <p className="text-xs text-blue-400 font-semibold mb-2">{question}</p>
      <p className="text-xs text-gray-300">{answer}</p>
    </div>
  );
}

function ActionBadge({ action }: { action: string }) {
  const styles: Record<string, string> = {
    run: 'bg-gray-700 text-gray-300',
    skip: 'bg-red-900/50 text-red-300',
    reduce: 'bg-amber-900/50 text-amber-300',
    increase: 'bg-emerald-900/50 text-emerald-300',
  };
  return (
    <span className={`text-[9px] px-1.5 py-0.5 rounded ${styles[action] || styles.run}`}>
      {action}
    </span>
  );
}
