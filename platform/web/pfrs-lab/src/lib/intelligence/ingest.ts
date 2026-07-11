import type { StorageProvider } from '@/lib/storage';
import {
  detectDomain,
  parseCounterfactualCSV,
  parseDecisionCSV,
  parseLearningCSV,
  parsePolicyDecisionsCSV,
  parsePolicyLearningReport,
  parsePortfolioAssistCSV,
  parseSearchAssistCSV,
  parseWorkerAssistAsDecisions,
  parseWorkerAssistCSV,
} from '@/lib/parsers/intelligence';
import { RUN_BATCH_SIZE } from './constants';
import type { IngestAcc, IngestMode } from './types';

async function readRunFiles(
  storage: StorageProvider,
  runId: string,
  filenames: string[],
): Promise<Record<string, string | null>> {
  const entries = await Promise.all(
    filenames.map(async (name) => [name, await storage.readFile(runId, name)] as const),
  );
  return Object.fromEntries(entries);
}

async function ingestRun(
  storage: StorageProvider,
  runId: string,
  mode: IngestMode,
  acc: IngestAcc,
) {
  const domain = detectDomain(runId);

  if (mode === 'learning') {
    const files = await readRunFiles(storage, runId, [
      'worker_learning.csv',
      'worker_decisions.csv',
      'worker_assist.csv',
    ]);
    const learningContent = files['worker_learning.csv'];
    const decisionContent = files['worker_decisions.csv'];
    const workerAssistContent = files['worker_assist.csv'];

    if (learningContent) {
      const learningRows = parseLearningCSV(learningContent, runId);
      acc.learning.push(...learningRows);
      acc.decisionLearning.push(...learningRows);
    }
    if (decisionContent) acc.decisions.push(...parseDecisionCSV(decisionContent, runId));
    else if (workerAssistContent) acc.decisions.push(...parseWorkerAssistAsDecisions(workerAssistContent, runId));
    if (workerAssistContent) acc.assistRecords.push(...parseWorkerAssistCSV(workerAssistContent, runId));
    return;
  }

  const files = await readRunFiles(storage, runId, [
    'policy_decisions.csv',
    'policy_learning_report.json',
    'policy_evaluation.csv',
    'generic_search_assist.csv',
    'portfolio_assist.csv',
    'worker_assist.csv',
  ]);

  if (files['policy_decisions.csv']) {
    acc.policyDecisions.push(...parsePolicyDecisionsCSV(files['policy_decisions.csv']!, runId));
  }
  if (files['policy_learning_report.json']) {
    const report = parsePolicyLearningReport(files['policy_learning_report.json']!, runId);
    if (report) acc.policyLearningReports.push(report);
  }
  if (files['policy_evaluation.csv']) {
    const evalLines = files['policy_evaluation.csv']!.trim().split('\n').length - 1;
    if (evalLines > 0) acc.policyEvalCount += evalLines;
  }
  if (files['generic_search_assist.csv']) {
    acc.assistRecords.push(...parseSearchAssistCSV(files['generic_search_assist.csv']!, runId, domain));
  }
  if (files['portfolio_assist.csv']) {
    acc.assistRecords.push(...parsePortfolioAssistCSV(files['portfolio_assist.csv']!, runId));
  }
  if (files['worker_assist.csv']) {
    acc.assistRecords.push(...parseWorkerAssistCSV(files['worker_assist.csv']!, runId));
  }
}

export async function ingestBatches(
  store: StorageProvider,
  runIds: string[],
  mode: IngestMode,
  acc: IngestAcc,
) {
  for (let i = 0; i < runIds.length; i += RUN_BATCH_SIZE) {
    const batch = runIds.slice(i, i + RUN_BATCH_SIZE);
    await Promise.all(batch.map((runId) => ingestRun(store, runId, mode, acc)));
  }
}

export function emptyIngestAcc(): IngestAcc {
  return {
    learning: [],
    decisions: [],
    decisionLearning: [],
    assistRecords: [],
    policyDecisions: [],
    policyLearningReports: [],
    policyEvalCount: 0,
  };
}
