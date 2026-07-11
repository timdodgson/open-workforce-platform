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

/** Mirrors `owp list-solvers` — update when builtin/byod registration changes. */
export const REGISTERED_SOLVERS: RegisteredSolver[] = [
  { name: 'cvrp', title: 'Capacitated VRP', kind: 'builtin', usage: 'owp solve cvrp --instance <path.vrp>' },
  { name: 'vrptw', title: 'VRP with time windows', kind: 'builtin', usage: 'owp solve vrptw --instance <path.txt>' },
  { name: 'jobshop', title: 'Job shop scheduling', kind: 'builtin', usage: 'owp solve jobshop --instance <path>' },
  { name: 'nrp', title: 'Nurse rostering (single week)', kind: 'builtin', usage: 'owp solve nrp --instance <name|dir>' },
  {
    name: 'tsp',
    title: 'TSP (BYOD demo)',
    kind: 'byod',
    usage: 'owp solve tsp --instance <path.json>',
    notes: 'examples/byod-tsp — symmetric TSP via owp-sdk',
  },
];

export const BUILTIN_SEARCH_MODES = ['sa', 'lahc', 'tabu', 'portfolio', 'adaptive'] as const;

export const CUSTOM_SEARCH_MODES: CustomSearchMode[] = [
  {
    name: 'greedy',
    description: 'Strict hill-climb — improving moves only (BYOA demo in internal/sdk/byoa)',
    example: 'owp solve tsp --mode greedy --instance ../../examples/byod-tsp/instances/tsp-5city.json',
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
    body: 'owp solve <name> --instance <path> --run-label <label> writes run.json + updates local manifest.',
  },
] as const;

export const SDK_MODULE = 'github.com/timdodgson/open-workforce-platform/owp-sdk';
export const SDK_VERSION = 'v0.1.0';
export const GITHUB_REPO = 'https://github.com/timdodgson/open-workforce-platform';
export const BYOD_TSP_PATH = 'examples/byod-tsp';
export const BYOA_PATH = 'examples/byod-byoa';
