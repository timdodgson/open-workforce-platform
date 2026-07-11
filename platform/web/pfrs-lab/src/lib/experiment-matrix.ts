import type { PolicyMode, ProblemType } from '@/lib/benchmark-suites';
import { BENCHMARK_POLICIES, BENCHMARK_SUITES } from '@/lib/benchmark-suites';

export type { ProblemType, PolicyMode };

export type OptionState = 'on' | 'off' | 'default' | 'na';

export type MatrixTier = 'fast' | 'deep' | 'benchmark';

export interface RunOption {
  flag: string;
  label: string;
  state: OptionState;
  value?: string;
  /** Why this state for this variation */
  rationale: string;
}

export interface CoreAlgorithm {
  id: 'sa' | 'lahc' | 'tabu';
  label: string;
  description: string;
  /** Domain-specific parameter flags when this alg is active */
  params: Array<{ flag: string; default: string; purpose: string }>;
}

export interface OptionLayer {
  id: string;
  title: string;
  appliesTo: string;
  flags: Array<{
    flag: string;
    values: string[];
    whenOn: string;
    whenOff: string;
    why: string;
  }>;
}

export interface VariationConfig {
  id: string;
  domain: ProblemType;
  tier: MatrixTier;
  title: string;
  command: 'solve' | 'tune-pfrs';
  solveDomain?: 'cvrp' | 'vrptw' | 'jobshop';
  primaryMode: string;
  instance: string;
  instanceSlug: string;
  labelPrefix: string;
  /** e.g. val-nrp-n012w8-sa-{policy}-s{seed} */
  labelPattern: string;
  iterations: string;
  policies: PolicyMode[];
  seeds: number[];
  variationsPerConfig: number;
  script: string;
  whyThisConfig: string;
  whyNotOthers: string;
  options: RunOption[];
}

export const CORE_ALGORITHMS: CoreAlgorithm[] = [
  {
    id: 'sa',
    label: 'Simulated Annealing (SA)',
    description: 'Accepts worse moves with decreasing probability — good general baseline, strong on VRPTW.',
    params: [
      { flag: '--temperature', default: '100', purpose: 'Initial acceptance width for uphill moves' },
      { flag: '--cooling-mode', default: 'adaptive', purpose: 'Auto-tune cooling; avoids per-instance rate tuning' },
    ],
  },
  {
    id: 'lahc',
    label: 'Late Acceptance Hill Climbing (LAHC)',
    description: 'Compares against a lagged reference solution — often optimal on small CVRP instances.',
    params: [
      { flag: '--lahc-length', default: '1000', purpose: 'History window; larger = more exploration' },
    ],
  },
  {
    id: 'tabu',
    label: 'Tabu Search',
    description: 'Short-term memory forbids recent moves — strongest on larger CVRP and JSS per benchmark ladder.',
    params: [
      { flag: '--tabu-tenure', default: 'domain-tuned', purpose: 'How long a move stays forbidden' },
    ],
  },
];

export const COMPOSITE_MODES = [
  {
    id: 'portfolio',
    label: 'Portfolio',
    description: 'Runs SA + LAHC + Tabu in parallel workers and keeps the best — upper-bound quality, higher compute.',
    whenUsed: 'SI validation pairs each domain with one single-strategy mode plus portfolio for contrast.',
  },
  {
    id: 'adaptive',
    label: 'Adaptive',
    description: 'Search Intelligence dynamically switches strategies during the run.',
    whenUsed: 'Exploratory / production tuning — not part of the canonical 240-run SI2 matrix.',
  },
] as const;

export const OPTION_LAYERS: OptionLayer[] = [
  {
    id: 'algorithm',
    title: 'Search algorithm (--mode / --pfrs-mode)',
    appliesTo: 'All domains',
    flags: [
      {
        flag: '--mode',
        values: ['sa', 'lahc', 'tabu', 'portfolio', 'adaptive'],
        whenOn: 'Selects which metaheuristic (or composite) drives the search.',
        whenOff: 'N/A — a mode is always required.',
        why: 'Three core metaheuristics (SA, LAHC, Tabu) plus portfolio and adaptive composites.',
      },
    ],
  },
  {
    id: 'si-assist',
    title: 'Assist layer (--worker-decision-mode)',
    appliesTo: 'NRP beam search (tune-pfrs); optional on solve domains',
    flags: [
      {
        flag: '--worker-decision-mode',
        values: ['off', 'shadow', 'assist', 'adaptive'],
        whenOn: 'SI can observe or steer worker spawning, beam width, and iteration budgets.',
        whenOff: 'Pure metaheuristic — used for benchmark ladder runs that isolate algorithm quality.',
        why: 'Policy-layer validation turns this ON (assist) for NRP only — beam search is where worker decisions matter. Solve domains leave it OFF so policy-mode effects are isolated.',
      },
    ],
  },
  {
    id: 'si-policy',
    title: 'Policy layer (--policy-mode)',
    appliesTo: 'All domains when validating learned policies',
    flags: [
      {
        flag: '--policy-mode',
        values: ['rules', 'hybrid', 'learned'],
        whenOn: 'Checkpoint policy chooses restart / diversify / intensify actions from telemetry.',
        whenOff: 'No policy layer — baseline runs and pure algorithm benchmarks.',
        why: 'The 240-run matrix always sets a policy so every seed × domain × mode is comparable across rules, hybrid, and learned.',
      },
      {
        flag: '--policy-dir',
        values: ['../ml/policies'],
        whenOn: 'Loads trained policy artifacts for hybrid/learned modes.',
        whenOff: 'Rules-only or no policy.',
        why: 'Required path for hybrid and learned; rules mode still reads the same directory for consistency.',
      },
    ],
  },
  {
    id: 'nrp-beam',
    title: 'NRP beam options (tune-pfrs only)',
    appliesTo: 'NRP multi-week only',
    flags: [
      {
        flag: '--pfrs-iterations-per-worker',
        values: ['100K fast · 300K deep · up to 1.5M exploratory'],
        whenOn: 'Each parallel worker runs this many iterations before a decision point.',
        whenOff: 'N/A for multi-week NRP — single-week uses owp solve nrp instead.',
        why: 'Beam search scales by spawning workers; iteration budget trades runtime vs solution quality.',
      },
      {
        flag: '--pfrs-max-total-workers',
        values: ['16 fast · 32 deep · 32+ exploratory'],
        whenOn: 'Caps parallel beam width — more workers = broader search, more CPU.',
        whenOff: 'Defaults to platform cap.',
        why: 'Validation uses moderate caps for reproducible overnight runs; EXP-001 used 32 for quality ceiling.',
      },
      {
        flag: '--beam-width / --lookahead-weight / --final-window-weeks',
        values: ['defaults in tune-pfrs'],
        whenOn: 'Look-ahead scoring and final-week intensification — major NRP quality levers.',
        whenOff: 'Defaults still apply; not swept in SI2 scripts.',
        why: 'Held constant during policy validation so policy comparisons are fair; tuned separately in EXP-001.',
      },
    ],
  },
];

function si2SolveOptions(domain: ProblemType, tier: MatrixTier): RunOption[] {
  return [
    {
      flag: '--worker-decision-mode',
      label: 'Assist worker decisions',
      state: 'off',
      value: 'off (default)',
      rationale:
        'Solve-path validation isolates SI2 policy effects. Worker-level SI is NRP-specific and enabled there via tune-pfrs.',
    },
    {
      flag: '--policy-mode',
      label: 'Policy layer',
      state: 'on',
      value: 'rules | hybrid | learned',
      rationale: 'Every SI2 run sweeps all three policies so checkpoint decisions are comparable across seeds.',
    },
    {
      flag: '--policy-dir',
      label: 'Policy artifacts',
      state: 'on',
      value: '../ml/policies',
      rationale: 'Hybrid and learned require trained models; rules uses the same path for a consistent CLI surface.',
    },
    {
      flag: '--iterations',
      label: 'Iteration budget',
      state: 'on',
      rationale: tier === 'fast'
        ? 'Fast soak uses smaller instances and moderate budgets for 240 runs overnight.'
        : 'Deep soak uses larger instances and longer budgets for richer checkpoint telemetry.',
    },
    {
      flag: '--portfolio',
      label: 'Portfolio strategies',
      state: 'na',
      rationale: 'Only set when --mode portfolio; then runs SA+LAHC+Tabu internally.',
    },
  ];
}

function si2NrpOptions(tier: MatrixTier): RunOption[] {
  return [
    {
      flag: '--worker-decision-mode',
      label: 'Assist worker decisions',
      state: 'on',
      value: 'assist',
      rationale:
        'NRP beam search is where worker spawn and budget decisions matter. Assist applies safe SI recommendations during tuning.',
    },
    {
      flag: '--policy-mode',
      label: 'Policy layer',
      state: 'on',
      value: 'rules | hybrid | learned',
      rationale: 'Same policy sweep as other domains — 30 NRP runs per mode (3 policies × 10 seeds).',
    },
    {
      flag: '--pfrs-mode',
      label: 'Beam worker algorithm',
      state: 'on',
      rationale: 'SA or portfolio per config — LAHC/Tabu only appear inside portfolio mode for NRP validation.',
    },
    {
      flag: '--pfrs-iterations-per-worker',
      label: 'Iterations per worker',
      state: 'on',
      value: tier === 'fast' ? '100,000' : '300,000',
      rationale: tier === 'fast' ? 'Fast matrix: 100K/worker × 16 max workers.' : 'Deep matrix: 300K/worker × 32 max workers for richer traces.',
    },
    {
      flag: '--pfrs-max-total-workers',
      label: 'Max beam workers',
      state: 'on',
      value: tier === 'fast' ? '16' : '32',
      rationale: 'Caps parallelism so validation finishes predictably; exploratory runs may use 32+.',
    },
    {
      flag: 'owp solve nrp',
      label: 'Single-week shortcut',
      state: 'off',
      rationale: 'SI2 matrix uses tune-pfrs for multi-week beam search. solve nrp is for single-week demos and partial bench uploads.',
    },
  ];
}

const FAST_SEEDS = [42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055];
const DEEP_SEEDS = [42, 123];

export const VARIATION_CONFIGS: VariationConfig[] = [
  {
    id: 'fast-cvrp-sa',
    domain: 'cvrp',
    tier: 'fast',
    title: 'CVRP — SA baseline',
    command: 'solve',
    solveDomain: 'cvrp',
    primaryMode: 'sa',
    instance: 'A-n32-k5',
    instanceSlug: 'a32k5',
    labelPrefix: 'val-cvrp-a32k5-sa',
    labelPattern: 'val-cvrp-a32k5-sa-{policy}-s{seed}',
    iterations: '500,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'SA is the canonical single-strategy baseline for routing domains.',
    whyNotOthers:
      'LAHC omitted as standalone — it often beats SA on CVRP but is included inside portfolio mode. Tabu omitted for the same reason; portfolio covers all three.',
    options: si2SolveOptions('cvrp', 'fast'),
  },
  {
    id: 'fast-cvrp-portfolio',
    domain: 'cvrp',
    tier: 'fast',
    title: 'CVRP — Portfolio',
    command: 'solve',
    solveDomain: 'cvrp',
    primaryMode: 'portfolio',
    instance: 'A-n32-k5',
    instanceSlug: 'a32k5',
    labelPrefix: 'val-cvrp-a32k5-portfolio',
    labelPattern: 'val-cvrp-a32k5-portfolio-{policy}-s{seed}',
    iterations: '500,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Portfolio runs SA+LAHC+Tabu in parallel — quality upper bound for policy comparison.',
    whyNotOthers: 'Adaptive mode not in matrix — would confound policy vs adaptive strategy switching.',
    options: [
      ...si2SolveOptions('cvrp', 'fast'),
      {
        flag: '--mode',
        label: 'Portfolio',
        state: 'on',
        value: 'portfolio',
        rationale: 'Internally runs all three core metaheuristics; best result wins.',
      },
    ],
  },
  {
    id: 'fast-jss-tabu',
    domain: 'jss',
    tier: 'fast',
    title: 'JSS — Tabu baseline',
    command: 'solve',
    solveDomain: 'jobshop',
    primaryMode: 'tabu',
    instance: 'la01',
    instanceSlug: 'la01',
    labelPrefix: 'val-jss-la01-tabu',
    labelPattern: 'val-jss-la01-tabu-{policy}-s{seed}',
    iterations: '100,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Tabu is the strongest single-strategy mode on JSS per benchmark ladder.',
    whyNotOthers: 'SA/LAHC not used standalone — weaker on JSS; still available inside portfolio.',
    options: si2SolveOptions('jss', 'fast'),
  },
  {
    id: 'fast-jss-portfolio',
    domain: 'jss',
    tier: 'fast',
    title: 'JSS — Portfolio',
    command: 'solve',
    solveDomain: 'jobshop',
    primaryMode: 'portfolio',
    instance: 'la01',
    instanceSlug: 'la01',
    labelPrefix: 'val-jss-la01-portfolio',
    labelPattern: 'val-jss-la01-portfolio-{policy}-s{seed}',
    iterations: '100,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Multi-strategy contrast to Tabu-only baseline on la01.',
    whyNotOthers: 'Deep tier switches to ft10 for harder instance — not duplicated in fast.',
    options: si2SolveOptions('jss', 'fast'),
  },
  {
    id: 'fast-vrptw-sa',
    domain: 'vrptw',
    tier: 'fast',
    title: 'VRPTW — SA baseline',
    command: 'solve',
    solveDomain: 'vrptw',
    primaryMode: 'sa',
    instance: 'C101',
    instanceSlug: 'c101',
    labelPrefix: 'val-vrptw-c101-sa',
    labelPattern: 'val-vrptw-c101-sa-{policy}-s{seed}',
    iterations: '100,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'SA baseline on Solomon C101 — tight clustered time windows.',
    whyNotOthers: 'LAHC often edges SA on VRPTW but is inside portfolio; Tabu similar.',
    options: si2SolveOptions('vrptw', 'fast'),
  },
  {
    id: 'fast-vrptw-portfolio',
    domain: 'vrptw',
    tier: 'fast',
    title: 'VRPTW — Portfolio',
    command: 'solve',
    solveDomain: 'vrptw',
    primaryMode: 'portfolio',
    instance: 'C101',
    instanceSlug: 'c101',
    labelPrefix: 'val-vrptw-c101-portfolio',
    labelPattern: 'val-vrptw-c101-portfolio-{policy}-s{seed}',
    iterations: '100,000',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Portfolio quality ceiling for policy validation on C101.',
    whyNotOthers: 'Other Solomon classes (R/RC) reserved for benchmark ladder expansion.',
    options: si2SolveOptions('vrptw', 'fast'),
  },
  {
    id: 'fast-nrp-sa',
    domain: 'nrp',
    tier: 'fast',
    title: 'NRP — Beam SA',
    command: 'tune-pfrs',
    primaryMode: 'sa',
    instance: 'n012w8',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-nrp-n012w8-sa',
    labelPattern: 'val-nrp-n012w8-sa-{policy}-s{seed}',
    iterations: '100,000/worker',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Single-strategy beam search on the reference INRC-II instance.',
    whyNotOthers: 'solve nrp not used — multi-week beam requires tune-pfrs. LAHC/Tabu only via portfolio.',
    options: si2NrpOptions('fast'),
  },
  {
    id: 'fast-nrp-portfolio',
    domain: 'nrp',
    tier: 'fast',
    title: 'NRP — Beam Portfolio',
    command: 'tune-pfrs',
    primaryMode: 'portfolio',
    instance: 'n012w8',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-nrp-n012w8-portfolio',
    labelPattern: 'val-nrp-n012w8-portfolio-{policy}-s{seed}',
    iterations: '100,000/worker',
    policies: BENCHMARK_POLICIES,
    seeds: FAST_SEEDS,
    variationsPerConfig: 30,
    script: 'platform/go/scripts/validate-si2.ps1',
    whyThisConfig: 'Best-known platform quality on n012w8 uses portfolio beam (EXP-001).',
    whyNotOthers: 'Exploratory 1.5M-iter runs are separate from this validation grid.',
    options: si2NrpOptions('fast'),
  },
  // Deep tier (48 runs total)
  {
    id: 'deep-cvrp-sa',
    domain: 'cvrp',
    tier: 'deep',
    title: 'CVRP — SA (large instance)',
    command: 'solve',
    solveDomain: 'cvrp',
    primaryMode: 'sa',
    instance: 'A-n80-k10',
    instanceSlug: 'a80k10',
    labelPrefix: 'val-deep-cvrp-a80k10-sa',
    labelPattern: 'val-deep-cvrp-a80k10-sa-{policy}-s{seed}',
    iterations: '5,000,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'Larger CVRP instance + 5M iterations → more checkpoints for policy training.',
    whyNotOthers: 'LAHC/Tabu standalone skipped — same rationale as fast tier.',
    options: si2SolveOptions('cvrp', 'deep'),
  },
  {
    id: 'deep-cvrp-portfolio',
    domain: 'cvrp',
    tier: 'deep',
    title: 'CVRP — Portfolio (large instance)',
    command: 'solve',
    solveDomain: 'cvrp',
    primaryMode: 'portfolio',
    instance: 'A-n80-k10',
    instanceSlug: 'a80k10',
    labelPrefix: 'val-deep-cvrp-a80k10-portfolio',
    labelPattern: 'val-deep-cvrp-a80k10-portfolio-{policy}-s{seed}',
    iterations: '2,000,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'Portfolio on A-n80-k10 where Tabu leads on benchmark ladder.',
    whyNotOthers: '2M portfolio iters vs 5M SA — portfolio is more expensive per iteration.',
    options: si2SolveOptions('cvrp', 'deep'),
  },
  {
    id: 'deep-jss-tabu',
    domain: 'jss',
    tier: 'deep',
    title: 'JSS — Tabu (ft10)',
    command: 'solve',
    solveDomain: 'jobshop',
    primaryMode: 'tabu',
    instance: 'ft10',
    instanceSlug: 'ft10',
    labelPrefix: 'val-deep-jss-ft10-tabu',
    labelPattern: 'val-deep-jss-ft10-tabu-{policy}-s{seed}',
    iterations: '1,000,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'ft10 is harder than la01 — longer search surfaces more policy decision points.',
    whyNotOthers: 'la01 kept for fast tier only to limit total runtime.',
    options: si2SolveOptions('jss', 'deep'),
  },
  {
    id: 'deep-jss-portfolio',
    domain: 'jss',
    tier: 'deep',
    title: 'JSS — Portfolio (ft10)',
    command: 'solve',
    solveDomain: 'jobshop',
    primaryMode: 'portfolio',
    instance: 'ft10',
    instanceSlug: 'ft10',
    labelPrefix: 'val-deep-jss-ft10-portfolio',
    labelPattern: 'val-deep-jss-ft10-portfolio-{policy}-s{seed}',
    iterations: '500,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'Portfolio contrast on ft10 deep soak.',
    whyNotOthers: 'Same 8-config structure as fast tier for apples-to-apples policy training.',
    options: si2SolveOptions('jss', 'deep'),
  },
  {
    id: 'deep-vrptw-sa',
    domain: 'vrptw',
    tier: 'deep',
    title: 'VRPTW — SA (extended budget)',
    command: 'solve',
    solveDomain: 'vrptw',
    primaryMode: 'sa',
    instance: 'C101',
    instanceSlug: 'c101',
    labelPrefix: 'val-deep-vrptw-c101-sa',
    labelPattern: 'val-deep-vrptw-c101-sa-{policy}-s{seed}',
    iterations: '2,000,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: '20× iteration budget vs fast tier for richer VRPTW telemetry.',
    whyNotOthers: 'Same instance (C101) — instance sweep is benchmark ladder, not SI2 validation.',
    options: si2SolveOptions('vrptw', 'deep'),
  },
  {
    id: 'deep-vrptw-portfolio',
    domain: 'vrptw',
    tier: 'deep',
    title: 'VRPTW — Portfolio (extended budget)',
    command: 'solve',
    solveDomain: 'vrptw',
    primaryMode: 'portfolio',
    instance: 'C101',
    instanceSlug: 'c101',
    labelPrefix: 'val-deep-vrptw-c101-portfolio',
    labelPattern: 'val-deep-vrptw-c101-portfolio-{policy}-s{seed}',
    iterations: '1,000,000',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'Portfolio deep soak on C101.',
    whyNotOthers: '1M portfolio iters balances runtime against SA 2M.',
    options: si2SolveOptions('vrptw', 'deep'),
  },
  {
    id: 'deep-nrp-sa',
    domain: 'nrp',
    tier: 'deep',
    title: 'NRP — Beam SA (extended)',
    command: 'tune-pfrs',
    primaryMode: 'sa',
    instance: 'n012w8',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-deep-nrp-n012w8-sa',
    labelPattern: 'val-deep-nrp-n012w8-sa-{policy}-s{seed}',
    iterations: '300,000/worker',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: '3× per-worker iterations and 32 max workers for NRP checkpoint density.',
    whyNotOthers: 'Beam tuning knobs (lookahead, final window) held at defaults — see EXP-001.',
    options: si2NrpOptions('deep'),
  },
  {
    id: 'deep-nrp-portfolio',
    domain: 'nrp',
    tier: 'deep',
    title: 'NRP — Beam Portfolio (extended)',
    command: 'tune-pfrs',
    primaryMode: 'portfolio',
    instance: 'n012w8',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-deep-nrp-n012w8-portfolio',
    labelPattern: 'val-deep-nrp-n012w8-portfolio-{policy}-s{seed}',
    iterations: '300,000/worker',
    policies: BENCHMARK_POLICIES,
    seeds: DEEP_SEEDS,
    variationsPerConfig: 6,
    script: 'platform/go/scripts/validate-si2-deep.ps1',
    whyThisConfig: 'Deep portfolio beam — primary training data for NRP policies.',
    whyNotOthers: 'Partial bench uploads (bench-nrp-*) may lack roster.json — those are not this matrix.',
    options: si2NrpOptions('deep'),
  },
];

/** Benchmark-ladder runs: pure algorithm comparison, typically no policy sweep */
export const BENCHMARK_LADDER_NOTES: Record<ProblemType, string> = {
  nrp: 'EXP-001 style runs use portfolio + beam tuning without val-* labels. policy-mode OFF unless explicitly testing SI.',
  cvrp: 'EXP-002/003/004 compare SA, LAHC, Tabu on Augerat instances with --worker-decision-mode off and no --policy-mode.',
  jss: 'Taillard/OR-Library ladder uses tabu/portfolio; SI flags off for pure algorithm gaps.',
  vrptw: 'Solomon C101 ladder (EXP-005+) — LAHC often wins; no policy sweep in baseline benchmarks.',
};

export function expandLabel(config: VariationConfig, policy: PolicyMode, seed: number): string {
  return config.labelPattern.replace('{policy}', policy).replace('{seed}', String(seed));
}

export function allLabelsForConfig(config: VariationConfig): string[] {
  const labels: string[] = [];
  for (const policy of config.policies) {
    for (const seed of config.seeds) {
      labels.push(expandLabel(config, policy, seed));
    }
  }
  return labels;
}

export function countMatchingRuns(runIds: string[], config: VariationConfig): number {
  const labels = new Set(allLabelsForConfig(config));
  return runIds.filter((id) => labels.has(id)).length;
}

export function matrixSummary() {
  const fast = VARIATION_CONFIGS.filter((c) => c.tier === 'fast');
  const deep = VARIATION_CONFIGS.filter((c) => c.tier === 'deep');
  return {
    domains: 4,
    coreAlgorithms: 3,
    fastConfigs: fast.length,
    deepConfigs: deep.length,
    fastVariations: fast.reduce((n, c) => n + c.variationsPerConfig, 0),
    deepVariations: deep.reduce((n, c) => n + c.variationsPerConfig, 0),
    suites: BENCHMARK_SUITES,
  };
}

export function optionStateClass(state: OptionState): string {
  switch (state) {
    case 'on':
      return 'text-emerald-400 bg-emerald-950/50 border-emerald-800';
    case 'off':
      return 'text-gray-400 bg-gray-900/80 border-gray-700';
    case 'default':
      return 'text-blue-300 bg-blue-950/40 border-blue-800';
    case 'na':
      return 'text-gray-600 bg-gray-950 border-gray-800';
    default:
      return 'text-gray-400';
  }
}

export function optionStateLabel(state: OptionState): string {
  switch (state) {
    case 'on':
      return 'ON';
    case 'off':
      return 'OFF';
    case 'default':
      return 'DEFAULT';
    case 'na':
      return 'N/A';
    default:
      return state;
  }
}
