import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';
import WorkerLearningDashboard from './WorkerLearningDashboard';

export const dynamic = 'force-dynamic';

export interface LearningRecord {
  runId: string;
  problemType: string;
  instance: string;
  algorithm: string;
  seed: number;
  week: number;
  depth: number;
  temperature: number;
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
  accepted: number;
  rejected: number;
  roi: number;
  improvPerCPU: number;
  improvPer100K: number;
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
      depth: parseInt(fields[5]) || 0,
      temperature: parseFloat(fields[14]) || 0,
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
      accepted: parseInt(fields[31]) || 0,
      rejected: parseInt(fields[32]) || 0,
      roi: parseFloat(fields[35]) || 0,
      improvPerCPU: parseFloat(fields[36]) || 0,
      improvPer100K: parseFloat(fields[37]) || 0,
    });
  }
  return records;
}

export default async function LearningPage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

  const allRecords: LearningRecord[] = [];

  for (const runId of runIds) {
    const content = await storage.readFile(runId, 'worker_learning.csv');
    if (!content) continue;
    const records = parseLearningCSV(content, runId);
    allRecords.push(...records);
  }

  if (allRecords.length === 0) {
    return (
      <Card title="Worker Learning">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No worker learning data available yet.</p>
          <p className="text-xs">Run experiments with the latest code to generate worker_learning.csv telemetry.</p>
        </div>
      </Card>
    );
  }

  return <WorkerLearningDashboard records={allRecords} />;
}
