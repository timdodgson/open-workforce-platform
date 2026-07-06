import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import WorkerDecisionDashboard from './WorkerDecisionDashboard';

export const dynamic = 'force-dynamic';

export interface DecisionRecord {
  runId: string;
  workerId: number;
  week: number;
  depth: number;
  algorithm: string;
  parentObjective: number;
  globalBest: number;
  distanceFromBest: number;
  beamRank: number;
  entropy: number;
  beamHealth: number;
  recentImprovRate: number;
  allocatedIters: number;
  recommendation: string;
  confidence: number;
  reasonCodes: string;
  suggestedAlgorithm: string;
  suggestedBudget: number;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  finalObjective: number;
  runtimeMs: number;
  roi: number;
}

export interface LearningRecord {
  runId: string;
  problemType: string;
  instance: string;
  algorithm: string;
  seed: number;
  week: number;
  depth: number;
  iterationsAlloc: number;
  globalBest: number;
  parentObjective: number;
  distanceFromBest: number;
  improved: boolean;
  producedGlobalBest: boolean;
  improvementAmount: number;
  finalObjective: number;
  runtimeMs: number;
  candidatesEval: number;
  roi: number;
}

function parseDecisionCSV(content: string, runId: string): DecisionRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const records: DecisionRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const fields = lines[i].split(',');
    if (fields.length < 23) continue;

    records.push({
      runId,
      workerId: parseInt(fields[0]) || 0,
      week: parseInt(fields[1]) || 0,
      depth: parseInt(fields[2]) || 0,
      algorithm: fields[3],
      parentObjective: parseInt(fields[4]) || 0,
      globalBest: parseInt(fields[5]) || 0,
      distanceFromBest: parseInt(fields[6]) || 0,
      beamRank: parseInt(fields[7]) || 0,
      entropy: parseFloat(fields[8]) || 0,
      beamHealth: parseFloat(fields[9]) || 0,
      recentImprovRate: parseFloat(fields[10]) || 0,
      allocatedIters: parseInt(fields[11]) || 0,
      recommendation: fields[12],
      confidence: parseFloat(fields[13]) || 0,
      reasonCodes: fields[14],
      suggestedAlgorithm: fields[15],
      suggestedBudget: parseInt(fields[16]) || 0,
      improved: fields[17] === '1',
      producedGlobalBest: fields[18] === '1',
      improvementAmount: parseInt(fields[19]) || 0,
      finalObjective: parseInt(fields[20]) || 0,
      runtimeMs: parseInt(fields[21]) || 0,
      roi: parseFloat(fields[22]) || 0,
    });
  }
  return records;
}

function parseLearningCSV(content: string, runId: string): LearningRecord[] {
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  const records: LearningRecord[] = [];
  for (let i = 1; i < lines.length; i++) {
    const fields = lines[i].split(',');
    if (fields.length < 38) continue;

    records.push({
      runId,
      problemType: fields[0],
      instance: fields[1],
      algorithm: fields[2],
      seed: parseInt(fields[3]) || 0,
      week: parseInt(fields[4]) || 0,
      depth: parseInt(fields[6]) || 0,
      iterationsAlloc: parseInt(fields[17]) || 0,
      globalBest: parseInt(fields[18]) || 0,
      parentObjective: parseInt(fields[19]) || 0,
      distanceFromBest: parseInt(fields[20]) || 0,
      improved: fields[25] === '1',
      producedGlobalBest: fields[26] === '1',
      improvementAmount: parseInt(fields[27]) || 0,
      finalObjective: parseInt(fields[28]) || 0,
      runtimeMs: parseInt(fields[29]) || 0,
      candidatesEval: parseInt(fields[30]) || 0,
      roi: parseFloat(fields[35]) || 0,
    });
  }
  return records;
}

export default async function DecisionsPage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

  const allDecisions: DecisionRecord[] = [];
  const allLearning: LearningRecord[] = [];

  for (const runId of runIds) {
    const decisionContent = await storage.readFile(runId, 'worker_decisions.csv');
    if (decisionContent) {
      allDecisions.push(...parseDecisionCSV(decisionContent, runId));
    }

    const learningContent = await storage.readFile(runId, 'worker_learning.csv');
    if (learningContent) {
      allLearning.push(...parseLearningCSV(learningContent, runId));
    }
  }

  if (allDecisions.length === 0) {
    return (
      <Card title="Worker Decision Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No worker decision data available yet.</p>
          <p className="text-xs">
            Run experiments with <code className="text-blue-400">--worker-decision-mode shadow</code> to generate worker_decisions.csv.
          </p>
        </div>
      </Card>
    );
  }

  return <WorkerDecisionDashboard decisions={allDecisions} learning={allLearning} />;
}
