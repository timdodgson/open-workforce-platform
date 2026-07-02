/**
 * Tests for CSV parsing functions.
 * Verifies that telemetry CSV files are parsed correctly into typed objects.
 */
import { parseTreeCSV, parsePlateauCSV, parseImprovementsCSV, parseDiversityCSV, parseDiscoveriesCSV, parseWorkerLifecycleCSV } from '@/lib/csv-parser';

describe('parseTreeCSV', () => {
  it('parses tree CSV with all fields', () => {
    const csv = `path_id,parent_id,week,seed,week_penalty,cumulative_penalty,workers_started,candidates,accepted,rejected,sa_accepted_better,sa_accepted_worse,sa_rejected_by_prob,hard_reject_rate,duration_ms,retained,retained_rank,winning
1,-1,1,42,385,385,10,500000,100,400000,50,30,420000,80.0,1200,1,1,1
2,1,2,42,245,630,8,400000,80,320000,40,25,335000,80.0,900,1,2,0`;

    const nodes = parseTreeCSV(csv);
    expect(nodes).toHaveLength(2);
    expect(nodes[0].pathID).toBe(1);
    expect(nodes[0].parentID).toBe(-1);
    expect(nodes[0].week).toBe(1);
    expect(nodes[0].weekPenalty).toBe(385);
    expect(nodes[0].cumulativePenalty).toBe(385);
    expect(nodes[0].retained).toBe(true);
    expect(nodes[0].winning).toBe(true);
    expect(nodes[1].parentID).toBe(1);
    expect(nodes[1].winning).toBe(false);
  });

  it('returns empty array for empty input', () => {
    expect(parseTreeCSV('')).toHaveLength(0);
  });
});

describe('parseImprovementsCSV', () => {
  it('parses improvement events', () => {
    const csv = `run_id,instance,seed,beam_width,iterations,temperature,cooling_mode,timestamp,week,worker_id,candidate,temperature_at_event,old_global_best,new_global_best,improvement,elapsed_ms
test,n012w8,42,5,500000,2.0,adaptive,2024-01-01,1,5,250000,0.5,1000,800,200,600`;

    const events = parseImprovementsCSV(csv);
    expect(events).toHaveLength(1);
    expect(events[0].week).toBe(1);
    expect(events[0].workerID).toBe(5);
    expect(events[0].oldGlobalBest).toBe(1000);
    expect(events[0].newGlobalBest).toBe(800);
    expect(events[0].improvement).toBe(200);
    expect(events[0].elapsedMs).toBe(600);
  });
});

describe('parseDiversityCSV', () => {
  it('parses diversity records', () => {
    const csv = `run_id,instance,seed,beam_width,iterations,temperature,cooling_mode,timestamp,week,path_id,fingerprint,hamming_to_best,hamming_to_parent,beam_spread,near_duplicate,retained,retained_rank,winning,cumulative_penalty,week_penalty
test,n012w8,42,5,500000,2.0,adaptive,2024-01-01,1,1,abc123,5,3,80,0,1,1,1,385,385`;

    const records = parseDiversityCSV(csv);
    expect(records).toHaveLength(1);
    expect(records[0].week).toBe(1);
    expect(records[0].pathID).toBe(1);
    expect(records[0].fingerprint).toBe('abc123');
    expect(records[0].hammingToBest).toBe(5);
    expect(records[0].nearDuplicate).toBe(false);
    expect(records[0].retained).toBe(true);
    expect(records[0].winning).toBe(true);
  });
});
