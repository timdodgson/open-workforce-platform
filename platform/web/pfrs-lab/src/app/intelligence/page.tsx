import { getStorageProvider } from '@/lib/storage';
import IntelligenceShell from './IntelligenceShell';
import { LearningRecord } from '../learning/types';
import { DecisionRecord, LearningRecord as DecLearningRecord } from '../decisions/types';
import { WorkerModel } from '../feature-importance/types';
import { WhatIfPrediction } from '../what-if/types';
import type {
  UnifiedAssistRecord, WorkerAssistRecord, SearchAssistRecord, PortfolioAssistRecord,
} from '../assist/types';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Search Intelligence',
  description: 'AI advisory system for optimisation. Monitors search behaviour and allocates compute safely and measurably.',
};

export const revalidate = 60;

// --- CSV Parsers (reused from existing pages) ---

function parseLearningCSV(content: string, runId: string): LearningRecord[] {
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

function parseDecisionCSV(content: string, runId: string): DecisionRecord[] {
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

function parseWorkerAssistCSV(content: string, runId: string): WorkerAssistRecord[] {
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

function parseSearchAssistCSV(content: string, runId: string, domain: string): SearchAssistRecord[] {
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

function parsePortfolioAssistCSV(content: string, runId: string): PortfolioAssistRecord[] {
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

function detectDomain(runId: string): string {
  if (runId.includes('cvrp')) return 'cvrp';
  if (runId.includes('jss') || runId.includes('jobshop')) return 'jss';
  if (runId.includes('vrptw')) return 'vrptw';
  return 'nrp';
}

// --- Data Loading ---

export interface IntelligenceData {
  learning: LearningRecord[];
  decisions: DecisionRecord[];
  decisionLearning: DecLearningRecord[];
  model: WorkerModel | null;
  predictions: WhatIfPrediction[];
  assistRecords: UnifiedAssistRecord[];
}

export default async function IntelligencePage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

  // Load all data in parallel where possible.
  const learning: LearningRecord[] = [];
  const decisions: DecisionRecord[] = [];
  const decisionLearning: DecLearningRecord[] = [];
  const assistRecords: UnifiedAssistRecord[] = [];

  for (const runId of runIds) {
    const domain = detectDomain(runId);

    const [learningContent, decisionContent, workerAssistContent, searchAssistContent, portfolioAssistContent] =
      await Promise.all([
        storage.readFile(runId, 'worker_learning.csv'),
        storage.readFile(runId, 'worker_decisions.csv'),
        storage.readFile(runId, 'worker_assist.csv'),
        storage.readFile(runId, 'generic_search_assist.csv'),
        storage.readFile(runId, 'portfolio_assist.csv'),
      ]);

    if (learningContent) learning.push(...parseLearningCSV(learningContent, runId));
    if (decisionContent) decisions.push(...parseDecisionCSV(decisionContent, runId));
    if (learningContent) decisionLearning.push(...parseLearningCSV(learningContent, runId) as unknown as DecLearningRecord[]);
    if (workerAssistContent) assistRecords.push(...parseWorkerAssistCSV(workerAssistContent, runId));
    if (searchAssistContent) assistRecords.push(...parseSearchAssistCSV(searchAssistContent, runId, domain));
    if (portfolioAssistContent) assistRecords.push(...parsePortfolioAssistCSV(portfolioAssistContent, runId));
  }

  // Load model and predictions.
  const modelContent = await storage.readRootFile('worker_model.json');
  let model: WorkerModel | null = null;
  if (modelContent) {
    try { model = JSON.parse(modelContent); } catch { /* graceful */ }
  }

  const predContent = await storage.readRootFile('worker_predictions.json');
  let predictions: WhatIfPrediction[] = [];
  if (predContent) {
    try { predictions = JSON.parse(predContent).predictions || []; } catch { /* graceful */ }
  }

  const data: IntelligenceData = {
    learning, decisions, decisionLearning, model, predictions, assistRecords,
  };

  return <IntelligenceShell data={data} />;
}
