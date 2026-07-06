import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import AssistDashboard from './AssistDashboard';

export const dynamic = 'force-dynamic';

// --- Unified assist record types ---

export type AssistArchitecture = 'worker' | 'search' | 'portfolio';

export interface WorkerAssistRecord {
  architecture: 'worker';
  domain: string;
  runId: string;
  workerId: number;
  algorithm: string;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  safetyTriggered: boolean;
  safetyRule: string;
  outcome: string;
  finalAction: string;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  runtimeMs: number;
}

export interface SearchAssistRecord {
  architecture: 'search';
  domain: string;
  runId: string;
  algorithm: string;
  checkpoint: number;
  candidates: number;
  iterationsTotal: number;
  bestPenalty: number;
  plateauLength: number;
  recommendedAction: string;
  confidence: number;
  reasons: string;
  safetyTriggered: boolean;
  safetyRule: string;
  accepted: boolean;
  finalAction: string;
  finalBestPenalty: number;
  runtimeMs: number;
}

export interface PortfolioAssistRecord {
  architecture: 'portfolio';
  domain: string;
  runId: string;
  instance: string;
  strategy: string;
  originalBudget: number;
  recommendedBudget: number;
  finalBudget: number;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  accepted: boolean;
  safetyRejected: boolean;
  safetyRule: string;
  resultObjective: number;
  strategyWon: boolean;
  runtimeMs: number;
}

export type UnifiedAssistRecord = WorkerAssistRecord | SearchAssistRecord | PortfolioAssistRecord;

// --- CSV Parsers ---

function parseWorkerAssistCSV(content: string, runId: string): WorkerAssistRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];
  const records: WorkerAssistRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const f = lines[i].split(',');
    if (f.length < 23) continue;
    records.push({
      architecture: 'worker',
      domain: 'nrp',
      runId,
      workerId: parseInt(f[0]) || 0,
      algorithm: f[3],
      recommendation: f[7],
      confidence: parseFloat(f[8]) || 0,
      reasonCodes: f[9],
      safetyTriggered: f[12] === '1',
      safetyRule: f[13],
      outcome: f[14],
      finalAction: f[15],
      improved: f[18] === '1',
      producedGlobalBest: f[19] === '1',
      improvementAmount: parseInt(f[20]) || 0,
      runtimeMs: parseInt(f[22]) || 0,
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
    records.push({
      architecture: 'search',
      domain,
      runId,
      algorithm: f[0],
      checkpoint: parseInt(f[1]) || 0,
      candidates: parseInt(f[2]) || 0,
      iterationsTotal: parseInt(f[3]) || 0,
      bestPenalty: parseInt(f[5]) || 0,
      plateauLength: parseInt(f[8]) || 0,
      recommendedAction: f[10],
      confidence: parseFloat(f[11]) || 0,
      reasons: f[12],
      safetyTriggered: f[13] === '1',
      safetyRule: f[14],
      accepted: f[15] === '1',
      finalAction: f[16],
      finalBestPenalty: parseInt(f[17]) || 0,
      runtimeMs: parseInt(f[19]) || 0,
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
    records.push({
      architecture: 'portfolio',
      domain: f[0],
      runId,
      instance: f[1],
      strategy: f[2],
      originalBudget: parseInt(f[4]) || 0,
      recommendedBudget: parseInt(f[5]) || 0,
      finalBudget: parseInt(f[6]) || 0,
      recommendation: f[7],
      confidence: parseFloat(f[8]) || 0,
      reasonCodes: f[9],
      accepted: f[10] === '1',
      safetyRejected: f[11] === '1',
      safetyRule: f[12],
      resultObjective: parseInt(f[13]) || 0,
      strategyWon: f[14] === '1',
      runtimeMs: parseInt(f[15]) || 0,
    });
  }
  return records;
}

function detectDomain(runId: string): string {
  if (runId.includes('cvrp')) return 'cvrp';
  if (runId.includes('jss') || runId.includes('jobshop')) return 'jss';
  if (runId.includes('vrptw')) return 'vrptw';
  if (runId.includes('nrp') || runId.includes('portfolio')) return 'nrp';
  return 'unknown';
}

export default async function AssistPage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

  const allRecords: UnifiedAssistRecord[] = [];

  for (const runId of runIds) {
    const domain = detectDomain(runId);

    // Worker assist (NRP beam search).
    const workerContent = await storage.readFile(runId, 'worker_assist.csv');
    if (workerContent) {
      allRecords.push(...parseWorkerAssistCSV(workerContent, runId));
    }

    // Generic search assist (single-algorithm runs).
    const searchContent = await storage.readFile(runId, 'generic_search_assist.csv');
    if (searchContent) {
      allRecords.push(...parseSearchAssistCSV(searchContent, runId, domain));
    }

    // Portfolio assist (multi-strategy runs).
    const portfolioContent = await storage.readFile(runId, 'portfolio_assist.csv');
    if (portfolioContent) {
      allRecords.push(...parsePortfolioAssistCSV(portfolioContent, runId));
    }
  }

  if (allRecords.length === 0) {
    return (
      <Card title="Search Intelligence — Assist Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No assist mode data available yet.</p>
          <p className="text-xs">
            Run experiments with <code className="text-blue-400">--worker-decision-mode shadow|assist</code>
          </p>
          <div className="mt-4 text-left inline-block">
            <p className="text-[10px] text-gray-400 mb-1">Supported across all solvers:</p>
            <ul className="text-[10px] text-gray-500 space-y-0.5">
              <li><code className="text-gray-400">tune-pfrs</code> — WorkerAssist (beam search)</li>
              <li><code className="text-gray-400">solve-cvrp</code> — SearchAssist + PortfolioAssist</li>
              <li><code className="text-gray-400">solve-jss</code> — SearchAssist + PortfolioAssist</li>
              <li><code className="text-gray-400">solve-vrptw</code> — SearchAssist + PortfolioAssist</li>
            </ul>
          </div>
        </div>
      </Card>
    );
  }

  return <AssistDashboard records={allRecords} />;
}
