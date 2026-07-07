'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { DecisionRecord, LearningRecord } from './types';
import ConfusionMatrix from './ConfusionMatrix';
import ConfidenceScatter from './ConfidenceScatter';
import RuleEffectivenessChart from './RuleEffectivenessChart';
import ROIComparison from './ROIComparison';
import FalseSkipForensics from './FalseSkipForensics';

interface Props {
  decisions: DecisionRecord[];
  learning: LearningRecord[];
}

export default function WorkerDecisionDashboard({ decisions, learning }: Props) {
  // --- Core classification ---
  // "skip" recommendations: the engine said don't run this worker.
  // "run" / "increase_budget" / "reduce_budget" / "change_algorithm": engine said run (with variation).
  // Actual outcome: improved = true means the worker was useful.
  const classification = useMemo(() => {
    const skipRecs = decisions.filter(d => d.recommendation === 'skip');
    const runRecs = decisions.filter(d => d.recommendation !== 'skip');

    // True Positive: recommended skip, worker did NOT improve (correct skip).
    const tp = skipRecs.filter(d => !d.improved).length;
    // False Positive: recommended skip, worker DID improve (would have missed improvement).
    const fp = skipRecs.filter(d => d.improved).length;
    // True Negative: recommended run, worker DID improve (correct run).
    const tn = runRecs.filter(d => d.improved).length;
    // False Negative: recommended run, worker did NOT improve (wasted CPU).
    const fn = runRecs.filter(d => !d.improved).length;

    const precision = tp + fp > 0 ? tp / (tp + fp) : 0;
    const recall = tp + fn > 0 ? tp / (tp + fn) : 0;
    const accuracy = decisions.length > 0 ? (tp + tn) / decisions.length : 0;
    const f1 = precision + recall > 0 ? 2 * (precision * recall) / (precision + recall) : 0;

    return { tp, fp, tn, fn, precision, recall, accuracy, f1, skipRecs, runRecs };
  }, [decisions]);

  // --- Recommendation breakdown ---
  const recommendationBreakdown = useMemo(() => {
    const map = new Map<string, number>();
    for (const d of decisions) {
      map.set(d.recommendation, (map.get(d.recommendation) || 0) + 1);
    }
    return Array.from(map.entries())
      .map(([rec, count]) => ({ recommendation: rec, count, pct: (count / decisions.length) * 100 }))
      .sort((a, b) => b.count - a.count);
  }, [decisions]);

  // --- Confidence distribution ---
  const confidenceDistribution = useMemo(() => {
    const buckets = [
      { label: '0.0–0.3', min: 0, max: 0.3, count: 0 },
      { label: '0.3–0.5', min: 0.3, max: 0.5, count: 0 },
      { label: '0.5–0.7', min: 0.5, max: 0.7, count: 0 },
      { label: '0.7–0.9', min: 0.7, max: 0.9, count: 0 },
      { label: '0.9–1.0', min: 0.9, max: 1.01, count: 0 },
    ];
    for (const d of decisions) {
      const bucket = buckets.find(b => d.confidence >= b.min && d.confidence < b.max);
      if (bucket) bucket.count++;
    }
    return buckets;
  }, [decisions]);

  // --- Predicted useful vs actually useful ---
  const predictedVsActual = useMemo(() => {
    const predictedUseful = decisions.filter(d => d.recommendation !== 'skip').length;
    const actuallyUseful = decisions.filter(d => d.improved).length;
    const predictedUseless = decisions.filter(d => d.recommendation === 'skip').length;
    const actuallyUseless = decisions.filter(d => !d.improved).length;
    return { predictedUseful, actuallyUseful, predictedUseless, actuallyUseless };
  }, [decisions]);

  // --- Rule effectiveness ---
  const ruleEffectiveness = useMemo(() => {
    const ruleMap = new Map<string, { total: number; correct: number; improved: number; runtimeMs: number }>();

    for (const d of decisions) {
      const rules = d.reasonCodes.split(';').filter(Boolean);
      for (const rule of rules) {
        const entry = ruleMap.get(rule) || { total: 0, correct: 0, improved: 0, runtimeMs: 0 };
        entry.total++;
        entry.runtimeMs += d.runtimeMs;
        if (d.improved) entry.improved++;

        // Correct: skip recommendation + didn't improve, OR run recommendation + did improve.
        const isSkipRec = d.recommendation === 'skip';
        if ((isSkipRec && !d.improved) || (!isSkipRec && d.improved)) {
          entry.correct++;
        }
        ruleMap.set(rule, entry);
      }
    }

    return Array.from(ruleMap.entries())
      .map(([rule, stats]) => ({
        rule,
        total: stats.total,
        correct: stats.correct,
        accuracy: stats.total > 0 ? (stats.correct / stats.total) * 100 : 0,
        improvedPct: stats.total > 0 ? (stats.improved / stats.total) * 100 : 0,
        avgRuntimeMs: stats.total > 0 ? stats.runtimeMs / stats.total : 0,
      }))
      .sort((a, b) => b.total - a.total);
  }, [decisions]);

  // --- Accuracy by algorithm ---
  const byAlgorithm = useMemo(() => {
    const map = new Map<string, { total: number; correct: number }>();
    for (const d of decisions) {
      const entry = map.get(d.algorithm) || { total: 0, correct: 0 };
      entry.total++;
      const isSkipRec = d.recommendation === 'skip';
      if ((isSkipRec && !d.improved) || (!isSkipRec && d.improved)) {
        entry.correct++;
      }
      map.set(d.algorithm, entry);
    }
    return Array.from(map.entries())
      .map(([algo, stats]) => ({
        algorithm: algo,
        total: stats.total,
        accuracy: stats.total > 0 ? (stats.correct / stats.total) * 100 : 0,
      }))
      .sort((a, b) => b.accuracy - a.accuracy);
  }, [decisions]);

  // --- Accuracy by week/depth ---
  const byWeekDepth = useMemo(() => {
    const map = new Map<string, { total: number; correct: number }>();
    for (const d of decisions) {
      const key = `W${d.week}/D${d.depth}`;
      const entry = map.get(key) || { total: 0, correct: 0 };
      entry.total++;
      const isSkipRec = d.recommendation === 'skip';
      if ((isSkipRec && !d.improved) || (!isSkipRec && d.improved)) {
        entry.correct++;
      }
      map.set(key, entry);
    }
    return Array.from(map.entries())
      .map(([key, stats]) => ({
        weekDepth: key,
        total: stats.total,
        accuracy: stats.total > 0 ? (stats.correct / stats.total) * 100 : 0,
      }))
      .sort((a, b) => a.weekDepth.localeCompare(b.weekDepth));
  }, [decisions]);

  // --- CPU savings estimate ---
  const cpuSavings = useMemo(() => {
    const skipRecs = decisions.filter(d => d.recommendation === 'skip');
    const totalRuntimeMs = decisions.reduce((s, d) => s + d.runtimeMs, 0);
    const skippableRuntimeMs = skipRecs.reduce((s, d) => s + d.runtimeMs, 0);
    const savingsPct = totalRuntimeMs > 0 ? (skippableRuntimeMs / totalRuntimeMs) * 100 : 0;

    // False positives: skipped workers that actually improved.
    const missedImprovements = skipRecs.filter(d => d.improved);
    const missedImprovementTotal = missedImprovements.reduce((s, d) => s + d.improvementAmount, 0);
    const missedGlobalBests = missedImprovements.filter(d => d.producedGlobalBest).length;

    return {
      totalRuntimeMs,
      skippableRuntimeMs,
      savingsPct,
      missedImprovementCount: missedImprovements.length,
      missedImprovementTotal,
      missedGlobalBests,
    };
  }, [decisions]);

  // --- Skipped-but-useful analysis (false positives detail) ---
  const skippedButUseful = useMemo(() => {
    return decisions
      .filter(d => d.recommendation === 'skip' && d.improved)
      .sort((a, b) => b.improvementAmount - a.improvementAmount)
      .slice(0, 20);
  }, [decisions]);

  // --- Run vs Skip ROI ---
  const roiComparison = useMemo(() => {
    const runRecs = decisions.filter(d => d.recommendation !== 'skip');
    const skipRecs = decisions.filter(d => d.recommendation === 'skip');

    const runImproved = runRecs.filter(d => d.improved);
    const skipImproved = skipRecs.filter(d => d.improved);

    const runAvgImprovement = runImproved.length > 0
      ? runImproved.reduce((s, d) => s + d.improvementAmount, 0) / runImproved.length : 0;
    const skipAvgImprovement = skipImproved.length > 0
      ? skipImproved.reduce((s, d) => s + d.improvementAmount, 0) / skipImproved.length : 0;

    const runAvgRuntime = runRecs.length > 0
      ? runRecs.reduce((s, d) => s + d.runtimeMs, 0) / runRecs.length : 0;
    const skipAvgRuntime = skipRecs.length > 0
      ? skipRecs.reduce((s, d) => s + d.runtimeMs, 0) / skipRecs.length : 0;

    const runSuccessRate = runRecs.length > 0 ? (runImproved.length / runRecs.length) * 100 : 0;
    const skipSuccessRate = skipRecs.length > 0 ? (skipImproved.length / skipRecs.length) * 100 : 0;

    return {
      run: { count: runRecs.length, successRate: runSuccessRate, avgImprovement: runAvgImprovement, avgRuntime: runAvgRuntime },
      skip: { count: skipRecs.length, successRate: skipSuccessRate, avgImprovement: skipAvgImprovement, avgRuntime: skipAvgRuntime },
    };
  }, [decisions]);

  // --- Confidence vs correctness data for scatter ---
  const confidenceVsCorrectness = useMemo(() => {
    return decisions.map(d => {
      const isSkipRec = d.recommendation === 'skip';
      const correct = (isSkipRec && !d.improved) || (!isSkipRec && d.improved);
      return {
        confidence: d.confidence,
        correct,
        recommendation: d.recommendation,
        improved: d.improved,
        improvementAmount: d.improvementAmount,
      };
    });
  }, [decisions]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title="Worker Decision Analysis — Shadow Mode">
        <p className="text-xs text-gray-500 mb-4">
          Evaluates shadow-mode recommendations against actual worker outcomes.
          Goal: determine whether the rule engine is trustworthy enough for future assist mode.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <Stat label="Total Decisions" value={decisions.length} colour="blue" />
          <Stat label="Accuracy" value={`${(classification.accuracy * 100).toFixed(1)}%`} colour="emerald" />
          <Stat label="Precision" value={`${(classification.precision * 100).toFixed(1)}%`} colour="emerald" />
          <Stat label="Recall" value={`${(classification.recall * 100).toFixed(1)}%`} colour="amber" />
          <Stat label="F1 Score" value={classification.f1.toFixed(3)} colour="amber" />
          <Stat label="CPU Saveable" value={`${cpuSavings.savingsPct.toFixed(1)}%`} colour="blue" />
        </div>
      </Card>

      {/* Recommendation Breakdown */}
      <Card title="Recommendation Breakdown">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 mb-4">
          {recommendationBreakdown.map(r => (
            <div key={r.recommendation} className="bg-gray-800 rounded p-3">
              <div className="text-[9px] text-gray-500 uppercase">{r.recommendation}</div>
              <div className="text-lg font-bold text-blue-400">{r.count}</div>
              <div className="text-[10px] text-gray-500">{r.pct.toFixed(1)}%</div>
            </div>
          ))}
        </div>
      </Card>

      {/* Confidence Distribution */}
      <Card title="Confidence Distribution">
        <div className="grid grid-cols-5 gap-2">
          {confidenceDistribution.map(b => (
            <div key={b.label} className="bg-gray-800 rounded p-3 text-center">
              <div className="text-[9px] text-gray-500">{b.label}</div>
              <div className="text-lg font-bold text-amber-400">{b.count}</div>
              <div className="text-[10px] text-gray-500">
                {decisions.length > 0 ? ((b.count / decisions.length) * 100).toFixed(0) : 0}%
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* Predicted vs Actual */}
      <Card title="Predicted Useful vs Actually Useful">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <Stat label="Predicted Useful" value={predictedVsActual.predictedUseful} colour="blue" />
          <Stat label="Actually Useful" value={predictedVsActual.actuallyUseful} colour="emerald" />
          <Stat label="Predicted Useless" value={predictedVsActual.predictedUseless} colour="amber" />
          <Stat label="Actually Useless" value={predictedVsActual.actuallyUseless} colour="amber" />
        </div>
      </Card>

      {/* Confusion Matrix */}
      <ConfusionMatrix
        tp={classification.tp}
        fp={classification.fp}
        tn={classification.tn}
        fn={classification.fn}
      />

      {/* Confidence vs Correctness Scatter */}
      <ConfidenceScatter data={confidenceVsCorrectness} />

      {/* Rule Effectiveness */}
      <RuleEffectivenessChart rules={ruleEffectiveness} />

      {/* Accuracy by Algorithm */}
      <Card title="Accuracy by Algorithm">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Algorithm</th>
              <th className="text-right p-2">Decisions</th>
              <th className="text-right p-2">Accuracy</th>
            </tr>
          </thead>
          <tbody>
            {byAlgorithm.map((a, i) => (
              <tr key={a.algorithm} className={`border-t border-gray-800 ${i === 0 ? 'bg-emerald-900/10' : ''}`}>
                <td className="p-2 text-blue-400 font-semibold">{a.algorithm.toUpperCase()}</td>
                <td className="text-right p-2">{a.total}</td>
                <td className="text-right p-2">
                  <span className={a.accuracy >= 70 ? 'text-emerald-400' : a.accuracy >= 50 ? 'text-amber-400' : 'text-red-400'}>
                    {a.accuracy.toFixed(1)}%
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Accuracy by Week/Depth */}
      <Card title="Accuracy by Week / Depth">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week/Depth</th>
              <th className="text-right p-2">Decisions</th>
              <th className="text-right p-2">Accuracy</th>
            </tr>
          </thead>
          <tbody>
            {byWeekDepth.map(wd => (
              <tr key={wd.weekDepth} className="border-t border-gray-800">
                <td className="p-2 text-blue-400">{wd.weekDepth}</td>
                <td className="text-right p-2">{wd.total}</td>
                <td className="text-right p-2">
                  <span className={wd.accuracy >= 70 ? 'text-emerald-400' : wd.accuracy >= 50 ? 'text-amber-400' : 'text-red-400'}>
                    {wd.accuracy.toFixed(1)}%
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* CPU Savings Estimate */}
      <Card title="CPU Savings Estimate (If Assist Mode Were Active)">
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 mb-4">
          <Stat label="Total Runtime" value={`${(cpuSavings.totalRuntimeMs / 1000).toFixed(1)}s`} colour="blue" />
          <Stat label="Skippable Runtime" value={`${(cpuSavings.skippableRuntimeMs / 1000).toFixed(1)}s`} colour="emerald" />
          <Stat label="Savings" value={`${cpuSavings.savingsPct.toFixed(1)}%`} colour="emerald" />
        </div>
        <div className="border-t border-gray-800 pt-3">
          <p className="text-xs text-gray-500 mb-2">Risk Assessment (improvements that would have been missed):</p>
          <div className="grid grid-cols-3 gap-3">
            <Stat label="Missed Improvements" value={cpuSavings.missedImprovementCount} colour="amber" />
            <Stat label="Missed Total Δ" value={cpuSavings.missedImprovementTotal.toLocaleString()} colour="amber" />
            <Stat label="Missed Global Bests" value={cpuSavings.missedGlobalBests} colour="amber" />
          </div>
        </div>
      </Card>

      {/* Run vs Skip ROI */}
      <ROIComparison run={roiComparison.run} skip={roiComparison.skip} />

      {/* Skipped-but-useful analysis */}
      {skippedButUseful.length > 0 && (
        <Card title="Skipped-but-Useful Analysis (False Positives)">
          <p className="text-xs text-gray-500 mb-3">
            Workers the engine recommended skipping that actually improved. These represent missed opportunities.
          </p>
          <div className="overflow-x-auto">
            <table className="w-full text-[10px]">
              <thead>
                <tr className="text-gray-500 uppercase">
                  <th className="text-left p-1.5">Run</th>
                  <th className="text-left p-1.5">Worker</th>
                  <th className="text-left p-1.5">Algo</th>
                  <th className="text-right p-1.5">W/D</th>
                  <th className="text-right p-1.5">Confidence</th>
                  <th className="text-left p-1.5">Reason</th>
                  <th className="text-right p-1.5">Δ Improvement</th>
                  <th className="text-right p-1.5">Runtime</th>
                  <th className="text-center p-1.5">Global Best?</th>
                </tr>
              </thead>
              <tbody>
                {skippedButUseful.map((d, i) => (
                  <tr key={i} className="border-t border-gray-800">
                    <td className="p-1.5 text-blue-400 font-mono">{d.runId.slice(0, 16)}</td>
                    <td className="p-1.5">{d.workerId}</td>
                    <td className="p-1.5 text-emerald-400">{d.algorithm}</td>
                    <td className="text-right p-1.5">W{d.week}/D{d.depth}</td>
                    <td className="text-right p-1.5 text-amber-400">{d.confidence.toFixed(2)}</td>
                    <td className="p-1.5 text-gray-400">{d.reasonCodes}</td>
                    <td className="text-right p-1.5 text-red-400 font-semibold">{d.improvementAmount.toLocaleString()}</td>
                    <td className="text-right p-1.5">{d.runtimeMs}ms</td>
                    <td className="text-center p-1.5">{d.producedGlobalBest ? '⭐' : '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* False Skip Forensics — deep analysis of bad skip decisions */}
      <FalseSkipForensics decisions={decisions} />

      {/* Historical Rule Analysis */}
      <Card title="Historical Rule Analysis">
        <p className="text-[10px] text-gray-500 mb-2">
          These results are from the original worker rule engine. Current Search Intelligence
          includes WorkerAssist, SearchAssist, PortfolioAssist, and Adaptive Mode — evaluated
          separately on the Assist Validation page.
        </p>
        <TrustworthinessAssessment
          accuracy={classification.accuracy}
          precision={classification.precision}
          recall={classification.recall}
          f1={classification.f1}
          missedGlobalBests={cpuSavings.missedGlobalBests}
          savingsPct={cpuSavings.savingsPct}
          totalDecisions={decisions.length}
        />
      </Card>
    </div>
  );
}

function Stat({ label, value, colour }: { label: string; value: string | number; colour: string }) {
  const colourClass = colour === 'blue' ? 'text-blue-400' : colour === 'emerald' ? 'text-emerald-400' : 'text-amber-400';
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-lg font-bold ${colourClass}`}>{value}</div>
    </div>
  );
}

function TrustworthinessAssessment({
  accuracy, precision, recall, f1, missedGlobalBests, savingsPct, totalDecisions,
}: {
  accuracy: number; precision: number; recall: number; f1: number;
  missedGlobalBests: number; savingsPct: number; totalDecisions: number;
}) {
  const observations: { text: string; severity: 'good' | 'warning' | 'danger' }[] = [];

  if (totalDecisions < 50) {
    observations.push({ text: `Only ${totalDecisions} decisions recorded. Need ≥100 for reliable assessment.`, severity: 'warning' });
  }

  if (accuracy >= 0.8) {
    observations.push({ text: `Accuracy ${(accuracy * 100).toFixed(1)}% — strong overall correctness.`, severity: 'good' });
  } else if (accuracy >= 0.6) {
    observations.push({ text: `Accuracy ${(accuracy * 100).toFixed(1)}% — moderate, needs improvement before assist mode.`, severity: 'warning' });
  } else {
    observations.push({ text: `Accuracy ${(accuracy * 100).toFixed(1)}% — too low for assist mode.`, severity: 'danger' });
  }

  if (precision >= 0.9) {
    observations.push({ text: `Precision ${(precision * 100).toFixed(1)}% — very few false skips.`, severity: 'good' });
  } else if (precision >= 0.7) {
    observations.push({ text: `Precision ${(precision * 100).toFixed(1)}% — some useful workers would be skipped.`, severity: 'warning' });
  } else {
    observations.push({ text: `Precision ${(precision * 100).toFixed(1)}% — too many useful workers would be skipped.`, severity: 'danger' });
  }

  if (missedGlobalBests > 0) {
    observations.push({ text: `${missedGlobalBests} global best(s) would have been missed — critical concern.`, severity: 'danger' });
  } else {
    observations.push({ text: 'No global bests would have been missed.', severity: 'good' });
  }

  if (savingsPct >= 10) {
    observations.push({ text: `Potential CPU savings: ${savingsPct.toFixed(1)}% — meaningful reduction.`, severity: 'good' });
  } else {
    observations.push({ text: `Potential CPU savings: ${savingsPct.toFixed(1)}% — marginal benefit.`, severity: 'warning' });
  }

  const goodCount = observations.filter(o => o.severity === 'good').length;
  const dangerCount = observations.filter(o => o.severity === 'danger').length;

  let verdict: string;
  let verdictColour: string;
  if (dangerCount > 0) {
    verdict = 'Original rules had critical limitations. These drove the evolution to the current Search Intelligence system.';
    verdictColour = 'text-amber-400';
  } else if (goodCount >= 3) {
    verdict = 'Original rules performed well — formed the basis for current assist/adaptive modes.';
    verdictColour = 'text-emerald-400';
  } else {
    verdict = 'Mixed performance — additional integration styles were needed.';
    verdictColour = 'text-amber-400';
  }

  return (
    <div>
      <ul className="space-y-2 mb-4">
        {observations.map((obs, i) => (
          <li key={i} className="text-sm flex gap-2">
            <span className={
              obs.severity === 'good' ? 'text-emerald-400' :
              obs.severity === 'warning' ? 'text-amber-400' : 'text-red-400'
            }>
              {obs.severity === 'good' ? '✓' : obs.severity === 'warning' ? '⚠' : '✗'}
            </span>
            <span className="text-gray-300">{obs.text}</span>
          </li>
        ))}
      </ul>
      <div className="border-t border-gray-800 pt-3">
        <p className="text-xs text-gray-500 uppercase mb-1">Verdict</p>
        <p className={`text-sm font-semibold ${verdictColour}`}>{verdict}</p>
      </div>
    </div>
  );
}
