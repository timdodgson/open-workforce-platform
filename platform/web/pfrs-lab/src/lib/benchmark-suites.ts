export type ProblemType = 'nrp' | 'cvrp' | 'jss' | 'vrptw';

export type PolicyMode = 'rules' | 'hybrid' | 'learned';

export interface SiValidationSpec {
  /** Run label prefix, e.g. val-cvrp or val-deep-cvrp */
  labelPrefix: string;
  instances: string[];
  modes: string[];
  seeds: number[];
  script: string;
  scriptHint: string;
}

export interface BenchmarkSuite {
  id: ProblemType;
  title: string;
  subtitle: string;
  referenceLabel: string;
  instances: string[];
  algorithms: Array<'sa' | 'lahc' | 'tabu' | 'portfolio' | 'adaptive'>;
  seeds: number[];
  settings: Record<string, string>;
  siFast: SiValidationSpec;
  siDeep: SiValidationSpec;
}

const FAST_SEEDS = [42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055];
const DEEP_SEEDS = [42, 123];
const POLICIES: PolicyMode[] = ['rules', 'hybrid', 'learned'];

export const BENCHMARK_POLICIES = POLICIES;

export const BENCHMARK_SUITES: BenchmarkSuite[] = [
  {
    id: 'nrp',
    title: 'Nurse Rostering (INRC-II)',
    subtitle: 'Objective: minimise soft constraint penalty.',
    referenceLabel: 'Reference: ILP baseline (when available)',
    instances: ['n012w8'],
    algorithms: ['sa', 'portfolio', 'adaptive'],
    seeds: FAST_SEEDS,
    settings: {
      'PFRS mode': 'SA / Portfolio',
      'Iterations/worker': '100,000 (fast) · 300,000 (deep)',
      'Max workers': '16 (fast) · 32 (deep)',
      'SI policies': 'rules | hybrid | learned',
    },
    siFast: {
      labelPrefix: 'val-nrp',
      instances: ['n012w8'],
      modes: ['sa', 'portfolio'],
      seeds: FAST_SEEDS,
      script: 'platform/go/scripts/validate-si2.ps1',
      scriptHint: '240 runs · all domains · ~overnight',
    },
    siDeep: {
      labelPrefix: 'val-deep-nrp',
      instances: ['n012w8'],
      modes: ['sa', 'portfolio'],
      seeds: DEEP_SEEDS,
      script: 'platform/go/scripts/validate-si2-deep.ps1',
      scriptHint: '48 runs · all domains · richer telemetry',
    },
  },
  {
    id: 'cvrp',
    title: 'Vehicle Routing (CVRPLIB)',
    subtitle: 'Objective: minimise total travel distance.',
    referenceLabel: 'Reference: CVRPLIB best-known/optimal',
    instances: ['A-n32-k5', 'A-n80-k10'],
    algorithms: ['sa', 'lahc', 'tabu', 'portfolio', 'adaptive'],
    seeds: FAST_SEEDS,
    settings: {
      'Iterations (SA/Tabu)': '500K (fast A-n32-k5) · 5M (deep A-n80-k10)',
      'Iterations (Portfolio)': '500K (fast) · 2M (deep)',
      'SI policies': 'rules | hybrid | learned',
    },
    siFast: {
      labelPrefix: 'val-cvrp',
      instances: ['A-n32-k5'],
      modes: ['sa', 'portfolio'],
      seeds: FAST_SEEDS,
      script: 'platform/go/scripts/validate-si2.ps1',
      scriptHint: 'Includes val-cvrp-a32k5-* (60 triplets for this domain)',
    },
    siDeep: {
      labelPrefix: 'val-deep-cvrp',
      instances: ['A-n80-k10'],
      modes: ['sa', 'portfolio'],
      seeds: DEEP_SEEDS,
      script: 'platform/go/scripts/validate-si2-deep.ps1',
      scriptHint: 'Includes val-deep-cvrp-a80k10-* (12 triplets for this domain)',
    },
  },
  {
    id: 'jss',
    title: 'Job Shop Scheduling (Taillard / OR-Library)',
    subtitle: 'Objective: minimise makespan.',
    referenceLabel: 'Reference: published optimal (where known)',
    instances: ['la01', 'ft10'],
    algorithms: ['sa', 'lahc', 'tabu', 'portfolio', 'adaptive'],
    seeds: FAST_SEEDS,
    settings: {
      'Iterations (Tabu)': '100K (fast la01) · 1M (deep ft10)',
      'Iterations (Portfolio)': '100K (fast) · 500K (deep)',
      'SI policies': 'rules | hybrid | learned',
    },
    siFast: {
      labelPrefix: 'val-jss',
      instances: ['la01'],
      modes: ['tabu', 'portfolio'],
      seeds: FAST_SEEDS,
      script: 'platform/go/scripts/validate-si2.ps1',
      scriptHint: 'Includes val-jss-la01-* (60 triplets for this domain)',
    },
    siDeep: {
      labelPrefix: 'val-deep-jss',
      instances: ['ft10'],
      modes: ['tabu', 'portfolio'],
      seeds: DEEP_SEEDS,
      script: 'platform/go/scripts/validate-si2-deep.ps1',
      scriptHint: 'Includes val-deep-jss-ft10-* (12 triplets for this domain)',
    },
  },
  {
    id: 'vrptw',
    title: 'Vehicle Routing with Time Windows (Solomon)',
    subtitle: 'Objective: minimise distance under time windows.',
    referenceLabel: 'Reference: Solomon best-known (distance only)',
    instances: ['C101'],
    algorithms: ['sa', 'lahc', 'tabu', 'portfolio', 'adaptive'],
    seeds: FAST_SEEDS,
    settings: {
      'Iterations (SA)': '100K (fast) · 2M (deep)',
      'Iterations (Portfolio)': '100K (fast) · 1M (deep)',
      'SI policies': 'rules | hybrid | learned',
    },
    siFast: {
      labelPrefix: 'val-vrptw',
      instances: ['C101'],
      modes: ['sa', 'portfolio'],
      seeds: FAST_SEEDS,
      script: 'platform/go/scripts/validate-si2.ps1',
      scriptHint: 'Includes val-vrptw-c101-* (60 triplets for this domain)',
    },
    siDeep: {
      labelPrefix: 'val-deep-vrptw',
      instances: ['C101'],
      modes: ['sa', 'portfolio'],
      seeds: DEEP_SEEDS,
      script: 'platform/go/scripts/validate-si2-deep.ps1',
      scriptHint: 'Includes val-deep-vrptw-c101-* (12 triplets for this domain)',
    },
  },
];

/** Expected complete rules/hybrid/learned triplets for one SI validation tier. */
export function expectedSiTriplets(spec: SiValidationSpec): number {
  return spec.instances.length * spec.modes.length * spec.seeds.length;
}
