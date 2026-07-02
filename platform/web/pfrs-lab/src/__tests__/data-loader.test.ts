/**
 * Tests for parseAuditCSV — the core results parser.
 */
import { parseAuditCSV } from '@/lib/csv-parser';

// Fields: instance,seed,mode,iterationsPerWorker,maxTotalWorkers,maxConcurrent,initialTemperature,coolingRate,coolingMode,effectiveCoolingRate,minTemperature,lateAcceptanceLen,week,startPenalty,finalPenalty,improvement,hardViolations,softViolations,candidates,accepted,rejected,acceptanceRate,bestIteration,bestWorkerID,workersStarted,branchesCreated,branchesDropped,maxQueueDepth,maxConcurrentSeen,durationMs,saFinalTemp,saTempAtBest,saAcceptedBetter,saAcceptedWorse,saRejectedByProb,lahcByCurrent,lahcByLate,lahcRejectedByLate,branchesQueued,branchesStarted,branchesCompleted,winningBranchDepth,workersImproved,workersProducedBest,rejectedNoop,rejectedSkill,rejectedSuccession,rejectedHistory
const HEADER = 'instance,seed,mode,iterationsPerWorker,maxTotalWorkers,maxConcurrent,initialTemperature,coolingRate,coolingMode,effectiveCoolingRate,minTemperature,lateAcceptanceLen,week,startPenalty,finalPenalty,improvement,hardViolations,softViolations,candidates,accepted,rejected,acceptanceRate,bestIteration,bestWorkerID,workersStarted,branchesCreated,branchesDropped,maxQueueDepth,maxConcurrentSeen,durationMs,saFinalTemp,saTempAtBest,saAcceptedBetter,saAcceptedWorse,saRejectedByProb,lahcByCurrent,lahcByLate,lahcRejectedByLate,branchesQueued,branchesStarted,branchesCompleted,winningBranchDepth,workersImproved,workersProducedBest,rejectedNoop,rejectedSkill,rejectedSuccession,rejectedHistory';
const ROW1 = 'n012w8,42,sa,500000,32,16,2.0,0.001,adaptive,0.001,0.0001,1000,1,1000,500,500,0,10,500000,100,400000,0.02,250000,5,10,8,2,4,8,1200,0.001,0.5,50,30,420000,0,0,0,8,8,8,2,5,2,380000,10000,5000,5000';
const ROW2 = 'n012w8,42,sa,500000,32,16,2.0,0.001,adaptive,0.001,0.0001,1000,2,800,400,400,0,8,500000,120,380000,0.024,200000,3,12,6,1,3,10,1100,0.0005,0.3,60,25,375000,0,0,0,6,6,6,1,4,1,360000,8000,4000,8000';

describe('parseAuditCSV', () => {
  it('parses multiple weeks', () => {
    const weeks = parseAuditCSV(`${HEADER}\n${ROW1}\n${ROW2}`);
    expect(weeks).toHaveLength(2);
  });

  it('extracts week numbers', () => {
    const weeks = parseAuditCSV(`${HEADER}\n${ROW1}\n${ROW2}`);
    expect(weeks[0].week).toBe(1);
    expect(weeks[1].week).toBe(2);
  });

  it('extracts penalties', () => {
    const weeks = parseAuditCSV(`${HEADER}\n${ROW1}`);
    expect(weeks[0].startPenalty).toBe(1000);
    expect(weeks[0].finalPenalty).toBe(500);
    expect(weeks[0].improvement).toBe(500);
  });

  it('extracts violations', () => {
    const weeks = parseAuditCSV(`${HEADER}\n${ROW1}`);
    expect(weeks[0].hardViolations).toBe(0);
    expect(weeks[0].softViolations).toBe(10);
  });

  it('extracts workers and candidates', () => {
    const weeks = parseAuditCSV(`${HEADER}\n${ROW1}`);
    expect(weeks[0].workersStarted).toBe(10);
    expect(weeks[0].candidates).toBe(500000);
  });

  it('returns empty for empty input', () => {
    expect(parseAuditCSV('')).toHaveLength(0);
  });

  it('returns empty for header only', () => {
    expect(parseAuditCSV(HEADER)).toHaveLength(0);
  });
});
