/**
 * CLI reference for the Getting Started lab page.
 * Keep in sync with platform/go/cmd/owp flag parsers.
 */

export type CommandFamily = 'solve' | 'tune-pfrs' | 'both';

export type FlagGroupId =
  | 'core-solve'
  | 'algorithm'
  | 'ga'
  | 'portfolio'
  | 'beam'
  | 'sa-lahc-tabu'
  | 'intelligence'
  | 'storage'
  | 'progress';

/** Getting Started shows essential first; advanced stays available but collapsed. */
export type FlagTier = 'essential' | 'advanced';

export interface CliFlag {
  flag: string;
  values: string;
  defaultValue?: string;
  summary: string;
  /** Longer explanation for graduates / hiring reviewers. */
  detail?: string;
  group: FlagGroupId;
  /** Which CLI entry points accept this flag. */
  commands: CommandFamily;
  /** Algorithms this flag applies to. Empty = all modes for that command. */
  algorithms?: string[];
  /** Other flags that must / should be set for this to matter. */
  dependsOn?: string[];
  /** Soft pairing — useful together. */
  pairsWith?: string[];
  note?: string;
  /** Presentation tier on Getting Started (all flags are kept). */
  tier: FlagTier;
}

/** Levers that appear in Quick start, worked examples, and published recipes. */
const ESSENTIAL_FLAGS = new Set([
  '--instance',
  '--mode',
  '--iterations',
  '--seed',
  '--portfolio',
  '--ga-population',
  '--pfrs-mode',
  '--pfrs-portfolio',
  '--pfrs-beam-width',
  '--pfrs-beam-seeds',
  '--pfrs-iterations-per-worker',
  '--pfrs-max-concurrent',
  '--pfrs-beam-strategy',
  '--pfrs-lookahead-weight',
  '--pfrs-final-window-weeks',
  '--worker-decision-mode',
  '--policy-mode',
  '--run-label',
  '--pfrs-run-label',
  '--storage / --pfrs-storage',
  '--progress-interval',
]);

function withTier(flags: Omit<CliFlag, 'tier'>[]): CliFlag[] {
  return flags.map((f) => ({
    ...f,
    tier: ESSENTIAL_FLAGS.has(f.flag) ? 'essential' : 'advanced',
  }));
}

export interface WorkedExample {
  id: string;
  title: string;
  domain: string;
  algorithm: string;
  timeHint: string;
  why: string;
  expected: string;
  cwd: string;
  command: string;
}

export const FLAG_GROUPS: Array<{ id: FlagGroupId; title: string; blurb: string }> = [
  {
    id: 'core-solve',
    title: 'Core solve flags',
    blurb: 'Shared by owp solve <domain> across CVRP, VRPTW, JSS, and single-week NRP.',
  },
  {
    id: 'algorithm',
    title: 'Search mode',
    blurb: 'Chooses the metaheuristic. Same Problem interface — change --mode, keep the domain.',
  },
  {
    id: 'sa-lahc-tabu',
    title: 'SA / LAHC / Tabu knobs',
    blurb: 'Only affect the named algorithm (or portfolio members that use that algorithm).',
  },
  {
    id: 'ga',
    title: 'Genetic Algorithm (GA)',
    blurb: 'Population-based mode. For owp solve use --mode ga; for NRP beam use --pfrs-mode ga or include ga in --pfrs-portfolio.',
  },
  {
    id: 'portfolio',
    title: 'Portfolio',
    blurb: 'Runs several strategies and keeps the best. Costs more compute; quality is never worse than the best member.',
  },
  {
    id: 'beam',
    title: 'NRP beam search (tune-pfrs)',
    blurb: 'Multi-week nurse rostering. These flags do nothing on owp solve cvrp|vrptw|jobshop.',
  },
  {
    id: 'intelligence',
    title: 'Search Intelligence (optional ML)',
    blurb: 'Off by default. Assist = rule checkpoints; Policies = learned JSON trees. They stack.',
  },
  {
    id: 'storage',
    title: 'Runs & storage',
    blurb: 'Labels create folders under the lab data tree; S3 is for the deployed dashboard.',
  },
  {
    id: 'progress',
    title: 'Progress & output',
    blurb: 'Console feedback during long beam runs.',
  },
];

export const CLI_FLAGS: CliFlag[] = withTier([
  // --- core solve ---
  {
    flag: '--instance',
    values: '<path|name>',
    summary: 'Problem instance to load.',
    detail:
      'CVRP/VRPTW/JSS: file path. NRP tune-pfrs: short name such as n012w8 (resolved under examples/inrc2).',
    group: 'core-solve',
    commands: 'both',
  },
  {
    flag: '--mode',
    values: 'sa | lahc | tabu | ga | portfolio | adaptive',
    defaultValue: 'sa',
    summary: 'Metaheuristic for owp solve.',
    detail: 'adaptive is mainly used on job shop; portfolio runs multiple strategies in parallel.',
    group: 'algorithm',
    commands: 'solve',
  },
  {
    flag: '--iterations',
    values: '<int>',
    defaultValue: '500000 (domain-dependent)',
    summary: 'Scored move evaluations (candidate budget).',
    detail: 'Hard-rejected moves do not consume this budget. Larger = better quality, longer runtime.',
    group: 'core-solve',
    commands: 'solve',
  },
  {
    flag: '--seed',
    values: '<int>',
    defaultValue: '42',
    summary: 'RNG seed for reproducible runs.',
    group: 'core-solve',
    commands: 'solve',
  },
  {
    flag: '--temperature',
    values: '<float>',
    defaultValue: '100',
    summary: 'Initial SA temperature (acceptance of uphill moves).',
    group: 'sa-lahc-tabu',
    commands: 'solve',
    algorithms: ['sa', 'portfolio'],
    note: 'Ignored for pure lahc / tabu / ga (except SA members inside portfolio).',
  },
  {
    flag: '--late-acceptance-length',
    values: '<int>',
    defaultValue: '1000',
    summary: 'LAHC history window length.',
    group: 'sa-lahc-tabu',
    commands: 'solve',
    algorithms: ['lahc', 'portfolio'],
  },
  {
    flag: '--tabu-tenure',
    values: '<int>',
    defaultValue: '7',
    summary: 'How long a move stays forbidden in Tabu search.',
    group: 'sa-lahc-tabu',
    commands: 'solve',
    algorithms: ['tabu', 'portfolio'],
  },
  {
    flag: '--tabu-neighbourhood',
    values: '<int>',
    defaultValue: '100',
    summary: 'Candidate moves sampled per Tabu iteration.',
    group: 'sa-lahc-tabu',
    commands: 'solve',
    algorithms: ['tabu'],
  },
  {
    flag: '--portfolio',
    values: 'sa,lahc,tabu,ga',
    defaultValue: 'sa,lahc,tabu,ga',
    summary: 'Comma-separated strategies when --mode portfolio.',
    group: 'portfolio',
    commands: 'solve',
    algorithms: ['portfolio'],
    dependsOn: ['--mode portfolio'],
  },
  {
    flag: '--ga-population',
    values: '<int>',
    defaultValue: '32 (solve) / 8 (PFRS)',
    summary: 'GA population size.',
    group: 'ga',
    commands: 'solve',
    algorithms: ['ga', 'portfolio'],
    dependsOn: ['--mode ga (or portfolio including ga)'],
  },
  {
    flag: '--ga-elite',
    values: '<int>',
    defaultValue: '2',
    summary: 'Elites preserved each generation.',
    group: 'ga',
    commands: 'solve',
    algorithms: ['ga'],
    dependsOn: ['--mode ga'],
  },
  {
    flag: '--ga-tournament',
    values: '<int>',
    defaultValue: '3',
    summary: 'Tournament selection size.',
    group: 'ga',
    commands: 'solve',
    algorithms: ['ga'],
  },
  {
    flag: '--ga-mutation-moves',
    values: '<int>',
    defaultValue: '5',
    summary: 'Greedy mutation moves per offspring.',
    group: 'ga',
    commands: 'solve',
    algorithms: ['ga'],
  },
  {
    flag: '--ga-crossover-moves',
    values: '<int>',
    defaultValue: '3',
    summary: 'Dual-parent blending moves (solve GA).',
    group: 'ga',
    commands: 'solve',
    algorithms: ['ga'],
  },

  // --- tune-pfrs / beam ---
  {
    flag: '--pfrs-mode',
    values: 'sa | lahc | tabu | ga | portfolio',
    defaultValue: 'sa',
    summary: 'Worker algorithm inside NRP PFRS / beam search.',
    group: 'beam',
    commands: 'tune-pfrs',
    note: 'Not the same flag as --mode on owp solve.',
  },
  {
    flag: '--pfrs-portfolio',
    values: 'sa,lahc,tabu,ga',
    defaultValue: 'sa,lahc,tabu,ga when mode=portfolio',
    summary: 'Strategies spawned on each branch in portfolio mode.',
    group: 'portfolio',
    commands: 'tune-pfrs',
    dependsOn: ['--pfrs-mode portfolio (implied if this flag is set)'],
    pairsWith: ['--pfrs-mode portfolio'],
  },
  {
    flag: '--pfrs-beam-width',
    values: '<int>',
    defaultValue: '1 (must set >1 for beam)',
    summary: 'How many partial multi-week paths to keep after each week.',
    detail: 'Wider beam explores more futures; compute grows roughly linearly with width × seeds.',
    group: 'beam',
    commands: 'tune-pfrs',
    dependsOn: ['beam search (≥2 width or beam seeds)'],
  },
  {
    flag: '--pfrs-beam-seeds',
    values: '42,101,202,...',
    summary: 'Seeds used to expand each retained path into the next week.',
    group: 'beam',
    commands: 'tune-pfrs',
    pairsWith: ['--pfrs-beam-width'],
  },
  {
    flag: '--pfrs-iterations-per-worker',
    values: '<int>',
    defaultValue: '500000',
    summary: 'Candidate budget for each PFRS worker.',
    detail: 'Alias: --iterations is accepted as a fallback if this flag is unset.',
    group: 'beam',
    commands: 'tune-pfrs',
  },
  {
    flag: '--pfrs-max-total-workers',
    values: '<int>',
    summary: 'Hard cap on workers started per week solve (branching stops / drops after this).',
    group: 'beam',
    commands: 'tune-pfrs',
  },
  {
    flag: '--pfrs-max-concurrent',
    values: '<int>',
    defaultValue: 'NumCPU',
    summary: 'How many workers run at the same time.',
    detail:
      'Lower this if you hit memory pressure on large portfolios (especially with GA). Quality is unchanged; wall-clock increases.',
    group: 'beam',
    commands: 'tune-pfrs',
  },
  {
    flag: '--pfrs-beam-strategy',
    values: 'none | lookahead | budget',
    defaultValue: 'none (or lookahead if weight > 0)',
    summary: 'How beam candidates are ranked across weeks.',
    detail:
      'budget penalises paths that overspend early soft-constraint “credit”; helps avoid week-8 cliffs.',
    group: 'beam',
    commands: 'tune-pfrs',
    pairsWith: ['--pfrs-lookahead-weight'],
  },
  {
    flag: '--pfrs-lookahead-weight',
    values: '<float>',
    defaultValue: '0',
    summary: 'Strength of look-ahead / budget bias in beam ranking.',
    group: 'beam',
    commands: 'tune-pfrs',
    dependsOn: ['--pfrs-beam-strategy lookahead|budget (or weight > 0 implies lookahead)'],
    note: '4.0 is aggressive; try 1.0 if mid-horizon paths look good but week 8 collapses.',
  },
  {
    flag: '--pfrs-final-window-weeks',
    values: '<int>',
    defaultValue: '1',
    summary: 'Couple the last N weeks before pruning.',
    detail:
      'With 2, weeks 7 and 8 run as a window — prune on the combined outcome. Important for NRP week-8 cost spikes.',
    group: 'beam',
    commands: 'tune-pfrs',
    pairsWith: ['--pfrs-final-window-iterations'],
  },
  {
    flag: '--pfrs-final-window-iterations',
    values: '<int>',
    summary: 'Extra iteration budget inside the final coupled window.',
    group: 'beam',
    commands: 'tune-pfrs',
    dependsOn: ['--pfrs-final-window-weeks ≥ 2'],
  },
  {
    flag: '--pfrs-diversity-slots',
    values: '<pct 0–100>',
    summary: 'Reserve a fraction of the beam for underrepresented path families.',
    group: 'beam',
    commands: 'tune-pfrs',
    pairsWith: ['--pfrs-beam-width'],
  },
  {
    flag: '--pfrs-refinement',
    values: 'none | sa | …',
    defaultValue: 'none',
    summary: 'Optional post-beam refinement pass.',
    group: 'beam',
    commands: 'tune-pfrs',
    pairsWith: ['--pfrs-refinement-iterations', '--pfrs-refinement-temperature'],
  },
  {
    flag: '--pfrs-refinement-iterations',
    values: '<int>',
    defaultValue: '100000',
    summary: 'Iteration budget for refinement.',
    group: 'beam',
    commands: 'tune-pfrs',
    dependsOn: ['--pfrs-refinement (not none)'],
  },
  {
    flag: '--pfrs-refinement-temperature',
    values: '<float>',
    defaultValue: '10',
    summary: 'SA temperature for refinement when applicable.',
    group: 'beam',
    commands: 'tune-pfrs',
    dependsOn: ['--pfrs-refinement'],
  },
  {
    flag: '--pfrs-cooling-mode',
    values: 'adaptive | fixed-rate',
    defaultValue: 'adaptive',
    summary: 'How SA temperature cools inside PFRS workers.',
    group: 'sa-lahc-tabu',
    commands: 'tune-pfrs',
    algorithms: ['sa', 'portfolio'],
  },
  {
    flag: '--pfrs-initial-temperature',
    values: '<float>',
    summary: 'Initial SA temperature for PFRS workers.',
    group: 'sa-lahc-tabu',
    commands: 'tune-pfrs',
    algorithms: ['sa', 'portfolio'],
  },
  {
    flag: '--pfrs-cooling-rate',
    values: '<float>',
    summary: 'Fixed cooling rate (only when cooling-mode is fixed-rate).',
    group: 'sa-lahc-tabu',
    commands: 'tune-pfrs',
    algorithms: ['sa'],
    dependsOn: ['--pfrs-cooling-mode fixed-rate'],
  },
  {
    flag: '--pfrs-late-acceptance-length',
    values: '<int>',
    summary: 'LAHC buffer for PFRS LAHC workers.',
    group: 'sa-lahc-tabu',
    commands: 'tune-pfrs',
    algorithms: ['lahc', 'portfolio'],
  },
  {
    flag: '--seeds',
    values: '42,123,...',
    defaultValue: '42',
    summary: 'Outer seeds for tuning sweeps (non-beam single-config uses these).',
    group: 'beam',
    commands: 'tune-pfrs',
    note: 'Beam expansion seeds are --pfrs-beam-seeds.',
  },

  // --- intelligence ---
  {
    flag: '--worker-decision-mode',
    values: 'off | shadow | assist | adaptive',
    defaultValue: 'off (unset)',
    summary: 'Assist layer: WorkerAssist / SearchAssist checkpoints.',
    detail:
      'shadow records only. assist may skip/reduce workers with safety overrides. Does nothing useful unless you care about SI telemetry or compute savings.',
    group: 'intelligence',
    commands: 'both',
    pairsWith: ['--policy-mode'],
  },
  {
    flag: '--policy-mode',
    values: 'rules | hybrid | learned',
    summary: 'Policy layer: learned stagnation / restart / budget trees in Go.',
    detail:
      'Unset = no policy layer. On NRP, hybrid/learned can still regress vs rules on some harness cells — start with rules or shadow.',
    group: 'intelligence',
    commands: 'both',
    dependsOn: ['--policy-dir (auto-resolved when mode is set)'],
    pairsWith: ['--worker-decision-mode assist'],
  },
  {
    flag: '--policy-dir',
    values: '<path>',
    defaultValue: '../ml/policies when policy-mode set',
    summary: 'Directory of policy JSON files.',
    group: 'intelligence',
    commands: 'both',
    dependsOn: ['--policy-mode'],
  },

  // --- storage / progress ---
  {
    flag: '--run-label',
    values: '<name>',
    summary: 'Label for owp solve artifacts (local run folder).',
    group: 'storage',
    commands: 'solve',
  },
  {
    flag: '--pfrs-run-label',
    values: '<name>',
    summary: 'Label for tune-pfrs artifacts under pfrs-lab/data/runs/<name>.',
    group: 'storage',
    commands: 'tune-pfrs',
  },
  {
    flag: '--storage / --pfrs-storage',
    values: 'local | s3',
    defaultValue: 'local',
    summary: 'Where run artifacts are written / uploaded.',
    group: 'storage',
    commands: 'both',
    note: 'S3 needs credentials and bucket config for the deployed lab.',
  },
  {
    flag: '--progress-interval',
    values: '30s',
    defaultValue: '10s',
    summary: 'How often tune-pfrs prints live worker progress.',
    group: 'progress',
    commands: 'tune-pfrs',
  },
]);

export const ESSENTIAL_FLAG_COUNT = CLI_FLAGS.filter((f) => f.tier === 'essential').length;
export const ADVANCED_FLAG_COUNT = CLI_FLAGS.filter((f) => f.tier === 'advanced').length;
export const WORKED_EXAMPLES: WorkedExample[] = [
  {
    id: 'cvrp-lahc',
    title: 'CVRP — LAHC on A-n32-k5',
    domain: 'CVRP',
    algorithm: 'LAHC',
    timeHint: '~1–5 seconds',
    why: 'Small published routing instance. Optimal distance is 784 — a good smoke test that the solver is wired correctly.',
    expected: 'Feasible distance in the high 700s / low 800s depending on iterations.',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode lahc --iterations 50000 --seed 42',
  },
  {
    id: 'jss-tabu',
    title: 'Job Shop — Tabu on la01',
    domain: 'JSS',
    algorithm: 'Tabu',
    timeHint: '~seconds to a minute',
    why: 'Classic OR-Library instance. Optimal makespan is 666 — Tabu often hits it.',
    expected: 'Makespan 666 (optimal) or a small gap on short budgets.',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp solve jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --mode tabu --iterations 100000 --seed 42',
  },
  {
    id: 'vrptw-sa',
    title: 'VRPTW — SA on Solomon C101',
    domain: 'VRPTW',
    algorithm: 'SA',
    timeHint: '~seconds',
    why: 'Time-window routing. Shows hard feasibility constraints beyond plain CVRP.',
    expected: 'Feasible tour near published BKS distance (~828).',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp solve vrptw --instance ../../examples/vrptw/C101.txt --mode sa --iterations 100000 --seed 42',
  },
  {
    id: 'cvrp-ga',
    title: 'CVRP — Genetic Algorithm',
    domain: 'CVRP',
    algorithm: 'GA',
    timeHint: '~seconds',
    why: 'Population-based search on the same Problem interface — compare against LAHC on the same instance.',
    expected: 'Competitive with LAHC on A-n32-k5 at similar budgets.',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode ga --iterations 50000 --seed 42 --ga-population 32',
  },
  {
    id: 'nrp-beam-fast',
    title: 'NRP — Fast beam (portfolio)',
    domain: 'NRP',
    algorithm: 'Portfolio + beam',
    timeHint: '~1–5 minutes',
    why: 'Flagship multi-week rostering. This is the “real” NRP path (not single-week solve).',
    expected: 'Valid 8-week roster; penalty typically mid–high thousands on a short budget.',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-beam-width 5 --pfrs-beam-seeds 42,101,202 --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --pfrs-max-concurrent 8 --pfrs-beam-strategy budget --pfrs-lookahead-weight 4.0 --pfrs-final-window-weeks 2 --pfrs-run-label my-first-nrp --progress-interval 15s',
  },
  {
    id: 'nrp-si-optional',
    title: 'NRP — Same beam + Assist (optional ML)',
    domain: 'NRP',
    algorithm: 'Portfolio + Assist',
    timeHint: 'Similar wall-clock; may save worker compute',
    why: 'Shows how to turn Search Intelligence on. Start here before hybrid/learned policies.',
    expected: 'worker_assist / decision CSVs appear in the run folder; quality should stay close to baseline.',
    cwd: 'platform/go',
    command:
      'go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-beam-width 5 --pfrs-beam-seeds 42,101,202 --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --worker-decision-mode assist --policy-mode rules --policy-dir ../ml/policies --pfrs-run-label my-first-nrp-si --progress-interval 15s',
  },
];

export const PREREQUISITES = [
  {
    title: 'Go 1.22+',
    why: 'The solver is a concurrent Go program: PFRS spawns many workers that share a global best via atomics and channels. Go’s goroutine runtime is why portfolio + beam can use your CPU without a heavyweight job framework.',
    how: 'Install from https://go.dev/dl/ then confirm with go version.',
  },
  {
    title: 'Git + this repository',
    why: 'Instances, policies, and the lab dashboard live in-tree. Paths in the examples below assume the monorepo layout.',
    how: 'git clone <repo> && cd open-workforce-platform',
  },
  {
    title: 'Terminal in platform/go',
    why: 'go run ./cmd/owp resolves modules from that module root; relative instance paths are documented from here.',
    how: 'cd platform/go',
  },
  {
    title: 'Optional: Node 20+ for the dashboard',
    why: 'Only needed to browse runs locally. The CLI works without the Next.js app.',
    how: 'cd platform/web/pfrs-lab && npm install && npm run dev',
  },
] as const;
