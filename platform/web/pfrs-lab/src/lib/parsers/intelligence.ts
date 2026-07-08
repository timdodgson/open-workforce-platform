import type {
  CounterfactualRow,
  DecisionRecord,
  LearningRecord,
  PolicyDecisionRecord,
  PolicyLearningReport,
  PortfolioAssistRecord,
  SearchAssistRecord,
  WorkerAssistRecord,
} from '@/lib/types/intelligence';

export function parseLearningCSV(content: string, runId: string): LearningRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: LearningRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const fields = lines[i].split(',');
    if (fields.length < 38) continue;
    records.push({
      runId, problemType: fields[0], instance: fields[1], algorithm: fields[2],
      seed: parseInt(fields[3]) || 0, week: parseInt(fields[4]) || 0,
      depth: parseInt(fields[6]) || 0, temperature: parseFloat(fields[14]) || 0,
      iterationsAlloc: parseInt(fields[17]) || 0, globalBest: parseInt(fields[18]) || 0,
      parentObjective: parseInt(fields[19]) || 0, distanceFromBest: parseInt(fields[20]) || 0,
      improved: fields[25] === '1', producedGlobalBest: fields[26] === '1',
      improvementAmount: parseInt(fields[27]) || 0, finalObjective: parseInt(fields[28]) || 0,
      runtimeMs: parseInt(fields[29]) || 0, candidatesEval: parseInt(fields[30]) || 0,
      accepted: parseInt(fields[31]) || 0, rejected: parseInt(fields[32]) || 0,
      roi: parseFloat(fields[35]) || 0, improvPerCPU: parseFloat(fields[36]) || 0,
      improvPer100K: parseFloat(fields[37]) || 0,
    });
  }
  return records;
}

export function parseDecisionCSV(content: string, runId: string): DecisionRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: DecisionRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const fields = lines[i].split(',');
    if (fields.length < 23) continue;
    records.push({
      runId, workerId: parseInt(fields[0]) || 0, week: parseInt(fields[1]) || 0,
      depth: parseInt(fields[2]) || 0, algorithm: fields[3],
      parentObjective: parseInt(fields[4]) || 0, globalBest: parseInt(fields[5]) || 0,
      distanceFromBest: parseInt(fields[6]) || 0, beamRank: parseInt(fields[7]) || 0,
      entropy: parseFloat(fields[8]) || 0, beamHealth: parseFloat(fields[9]) || 0,
      recentImprovRate: parseFloat(fields[10]) || 0, allocatedIters: parseInt(fields[11]) || 0,
      recommendation: fields[12], confidence: parseFloat(fields[13]) || 0,
      reasonCodes: fields[14], suggestedAlgorithm: fields[15],
      suggestedBudget: parseInt(fields[16]) || 0, improved: fields[17] === '1',
      producedGlobalBest: fields[18] === '1', improvementAmount: parseInt(fields[19]) || 0,
      finalObjective: parseInt(fields[20]) || 0, runtimeMs: parseInt(fields[21]) || 0,
      roi: parseFloat(fields[22]) || 0,
    });
  }
  return records;
}

export function parseWorkerAssistAsDecisions(content: string, runId: string): DecisionRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: DecisionRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 23) continue;
    records.push({
      runId,
      workerId: parseInt(f[0]) || 0,
      week: parseInt(f[1]) || 0,
      depth: parseInt(f[2]) || 0,
      algorithm: f[3],
      parentObjective: parseInt(f[4]) || 0,
      globalBest: parseInt(f[5]) || 0,
      distanceFromBest: parseInt(f[6]) || 0,
      beamRank: 0, entropy: 0, beamHealth: 0, recentImprovRate: 0,
      allocatedIters: parseInt(f[11]) || 0,
      recommendation: f[7],
      confidence: parseFloat(f[8]) || 0,
      reasonCodes: f[9],
      suggestedAlgorithm: f[10],
      suggestedBudget: parseInt(f[11]) || 0,
      improved: f[18] === '1',
      producedGlobalBest: f[19] === '1',
      improvementAmount: parseInt(f[20]) || 0,
      finalObjective: parseInt(f[21]) || 0,
      runtimeMs: parseInt(f[22]) || 0,
      roi: 0,
    });
  }
  return records;
}

export function parseWorkerAssistCSV(content: string, runId: string): WorkerAssistRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: WorkerAssistRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 23) continue;
    records.push({
      architecture: 'worker', domain: 'nrp', runId,
      workerId: parseInt(f[0]) || 0, algorithm: f[3], recommendation: f[7],
      confidence: parseFloat(f[8]) || 0, reasonCodes: f[9],
      safetyTriggered: f[12] === '1', safetyRule: f[13], outcome: f[14],
      finalAction: f[15], improved: f[18] === '1', producedGlobalBest: f[19] === '1',
      improvementAmount: parseInt(f[20]) || 0, runtimeMs: parseInt(f[22]) || 0,
    });
  }
  return records;
}

export function parseSearchAssistCSV(content: string, runId: string, domain: string): SearchAssistRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: SearchAssistRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 20) continue;
    const reasons = f[12] || '';
    records.push({
      architecture: 'search', domain, runId, algorithm: f[0],
      checkpoint: parseInt(f[1]) || 0, candidates: parseInt(f[2]) || 0,
      iterationsTotal: parseInt(f[3]) || 0, bestPenalty: parseInt(f[5]) || 0,
      plateauLength: parseInt(f[8]) || 0, recommendedAction: f[10],
      confidence: parseFloat(f[11]) || 0, reasons,
      safetyTriggered: f[13] === '1', safetyRule: f[14],
      accepted: f[15] === '1', finalAction: f[16],
      finalBestPenalty: parseInt(f[17]) || 0, runtimeMs: parseInt(f[19]) || 0,
      isAdaptive: reasons.includes('adaptive_'),
    });
  }
  return records;
}

export function parsePortfolioAssistCSV(content: string, runId: string): PortfolioAssistRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: PortfolioAssistRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 16) continue;
    const reasonCodes = f[9] || '';
    records.push({
      architecture: 'portfolio', domain: f[0], runId, instance: f[1], strategy: f[2],
      originalBudget: parseInt(f[4]) || 0, recommendedBudget: parseInt(f[5]) || 0,
      finalBudget: parseInt(f[6]) || 0, recommendation: f[7],
      confidence: parseFloat(f[8]) || 0, reasonCodes,
      accepted: f[10] === '1', safetyRejected: f[11] === '1', safetyRule: f[12],
      resultObjective: parseInt(f[13]) || 0, strategyWon: f[14] === '1',
      runtimeMs: parseInt(f[15]) || 0,
      usedLearned: reasonCodes.includes('learned_'),
      fallbackReason: (reasonCodes.match(/fallback:([^;]+)/) || [])[1] || '',
    });
  }
  return records;
}

export function parsePolicyDecisionsCSV(content: string, runId: string): PolicyDecisionRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: PolicyDecisionRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 8) continue;
    records.push({
      runId,
      checkpoint: parseInt(f[0]) || 0,
      candidates: parseInt(f[1]) || 0,
      policyMode: f[2],
      policyUsed: f[3],
      action: f[4],
      confidence: parseFloat(f[5]) || 0,
      fallbackReason: f[6],
      safetyOverride: f[7] === '1',
    });
  }
  return records;
}

export function parsePolicyLearningReport(content: string, runId: string): PolicyLearningReport | null {
  try {
    const parsed = JSON.parse(content);
    return {
      runId,
      action: parsed.learning_recommendation?.action || '',
      reason: parsed.learning_recommendation?.reason || '',
      samplesAdded: parsed.samples_added || 0,
    };
  } catch {
    return null;
  }
}

export function parseCounterfactualCSV(content: string, runId: string): CounterfactualRow[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const rows: CounterfactualRow[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 25) continue;
    rows.push({
      timestamp: f[0],
      runId,
      decisionType: f[2],
      domain: f[3],
      instance: f[4],
      algorithm: f[5],
      actualAction: f[6],
      confidence: parseFloat(f[7]) || 0,
      policyId: f[8],
      policyVersion: f[9],
      policyType: f[10],
      expectedValue: parseFloat(f[11]) || 0,
      actualOutcome: parseFloat(f[21]) || 0,
      outcomeMetric: f[22],
      regret: parseFloat(f[23]) || 0,
      bestAlternative: f[24],
    });
  }
  return rows;
}

export function detectDomain(runId: string): string {
  if (runId.includes('cvrp')) return 'cvrp';
  if (runId.includes('jss') || runId.includes('jobshop')) return 'jss';
  if (runId.includes('vrptw')) return 'vrptw';
  return 'nrp';
}
