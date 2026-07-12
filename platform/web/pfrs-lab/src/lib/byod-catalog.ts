export interface RegisteredSolver {
  name: string;
  title: string;
  kind: 'builtin' | 'byod';
  usage: string;
  notes?: string;
}

export interface CustomSearchMode {
  name: string;
  description: string;
  example: string;
}

export interface ByodTryItExample {
  id: string;
  title: string;
  blurb: string;
  command: string;
  expected: string;
}

/** Working copy-paste demos — verified from platform/go. */
export const BYOD_TRY_IT_CWD = 'platform/go';

export const BYOD_TRY_IT_EXAMPLES: ByodTryItExample[] = [
  {
    id: 'tsp-sa',
    title: 'BYOD — TSP with Simulated Annealing',
    blurb: 'Five-city instance shipped in examples/byod-tsp. Proves a custom domain on the same owp solve path.',
    command:
      'go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode sa --iterations 50000 --seed 42',
    expected: 'Tour length 23 (baseline 28 → 23). Feasible: true. Typically finishes in milliseconds after compile.',
  },
  {
    id: 'tsp-greedy',
    title: 'BYOA — same TSP with custom greedy mode',
    blurb: 'Same instance; --mode greedy is registered via sdk.RegisterSearch in internal/sdk/byoa.',
    command:
      'go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode greedy --iterations 20000 --seed 42',
    expected: 'Tour length 23. Feasible: true. Shows a custom algorithm without changing the domain.',
  },
];

export const BYOD_LIST_SOLVERS =
  'go run ./cmd/owp list-solvers';

export const BYOD_SEED_DEMOS =
  'powershell -ExecutionPolicy Bypass -File .\\scripts\\seed-tsp-demo-runs.ps1';

/** Mirrors `owp list-solvers` — update when builtin/byod registration changes. */
export const REGISTERED_SOLVERS: RegisteredSolver[] = [
  { name: 'cvrp', title: 'Capacitated VRP', kind: 'builtin', usage: 'go run ./cmd/owp solve cvrp --instance <path.vrp>' },
  { name: 'vrptw', title: 'VRP with time windows', kind: 'builtin', usage: 'go run ./cmd/owp solve vrptw --instance <path.txt>' },
  { name: 'jobshop', title: 'Job shop scheduling', kind: 'builtin', usage: 'go run ./cmd/owp solve jobshop --instance <path>' },
  { name: 'nrp', title: 'Nurse rostering (single week)', kind: 'builtin', usage: 'go run ./cmd/owp solve nrp --instance <name|dir>' },
  {
    name: 'tsp',
    title: 'TSP (BYOD demo)',
    kind: 'byod',
    usage:
      'go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode sa --iterations 50000 --seed 42',
    notes: 'From platform/go. Template: examples/byod-tsp',
  },
];

export const BUILTIN_SEARCH_MODES = ['sa', 'lahc', 'tabu', 'ga', 'portfolio', 'adaptive'] as const;

export const CUSTOM_SEARCH_MODES: CustomSearchMode[] = [
  {
    name: 'greedy',
    description: 'Strict hill-climb — improving moves only (BYOA demo in internal/sdk/byoa)',
    example:
      'go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode greedy --iterations 20000 --seed 42',
  },
];

export const BYOD_STEPS = [
  {
    step: '1',
    title: 'Implement searchdef.Problem',
    body: 'Solutions, moves, Evaluate, Undo — same contract as built-in domains.',
  },
  {
    step: '2',
    title: 'Register with owp-sdk',
    body: 'sdk.RegisterProblem in init() with Title, PolicyDomain, ObjectiveLabel for generic CLI display.',
  },
  {
    step: '3',
    title: 'Wire your binary',
    body: 'Blank-import your package from cmd/owp (or a custom main). No fork of optimisation required.',
  },
  {
    step: '4',
    title: 'Run and store',
    body: 'From platform/go: go run ./cmd/owp solve <name> --instance <path> --run-label <label> writes run.json + updates the local manifest.',
  },
] as const;

export const SDK_MODULE = 'github.com/timdodgson/open-workforce-platform/owp-sdk';
export const SDK_VERSION = 'v0.1.0';
export const GITHUB_REPO = 'https://github.com/timdodgson/open-workforce-platform';
export const BYOD_TSP_PATH = 'examples/byod-tsp';
export const BYOA_PATH = 'examples/byod-byoa';
