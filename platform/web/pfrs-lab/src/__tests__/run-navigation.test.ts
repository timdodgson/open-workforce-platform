import { isRosterSolution } from '@/lib/roster-types';
import { resolveProblemType } from '@/lib/resolve-problem-type';
import { deriveRunMode } from '@/features/runs/run-mode';

describe('isRosterSolution', () => {
  it('accepts NRP roster arrays', () => {
    expect(isRosterSolution([{
      week: 1,
      nurse: 'P001',
      day: 'Mon',
      dayIndex: 0,
      shiftType: 'Early',
      skill: 'Nurse',
      contract: 'FullTime',
      nurseSkills: ['Nurse'],
    }])).toBe(true);
  });

  it('rejects routing solutions', () => {
    expect(isRosterSolution({ routes: [], totalCost: 0, vehicles: 1, feasible: true })).toBe(false);
  });
});

describe('resolveProblemType', () => {
  it('detects si2 and val-deep prefixes from run id', () => {
    expect(resolveProblemType('si2-cvrp-hybrid-s42', null)).toBe('cvrp');
    expect(resolveProblemType('val-deep-jss-ft10-tabu-rules-s42', null)).toBe('jss');
    expect(resolveProblemType('val-deep-nrp-n012w8-sa-rules-s42', null)).toBe('nrp');
    expect(resolveProblemType('bench-nrp-n012w8-portfolio-off-s42', null)).toBe('nrp');
  });

  it('prefers run.json problemType when set', () => {
    expect(resolveProblemType('misc-run', { problemType: 'vrptw' })).toBe('vrptw');
  });
});

describe('deriveRunMode', () => {
  it('maps domains to sidebar modes', () => {
    expect(deriveRunMode({ problemType: 'cvrp', mode: 'sa' }, 'si2-cvrp-x')).toBe('cvrp');
    expect(deriveRunMode({ problemType: 'nrp', mode: 'portfolio' }, 'bench-nrp-x')).toBe('pfrs');
    expect(deriveRunMode({ problemType: 'nrp', mode: 'ilp' }, 'ilp-n012w8')).toBe('ilp');
  });
});
