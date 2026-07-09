export type DomainId = 'nrp' | 'cvrp' | 'vrptw' | 'jss';

export type CellStatus = 'done' | 'partial' | 'missing' | 'na' | 'adapted';

export interface DomainInfo {
  id: DomainId;
  label: string;
  subtitle: string;
}

export interface MatrixRow {
  id: string;
  label: string;
  notes?: string;
  cells: Record<DomainId, CellStatus>;
}

export interface LayeredAssistRow {
  id: string;
  label: string;
  description: string;
  runtime: Record<DomainId, CellStatus>;
  telemetry: Record<DomainId, CellStatus>;
  fixPhase?: string;
  fixNote?: string;
}

export interface GapItem {
  id: string;
  title: string;
  phase: string;
  effort: string;
  status: 'open' | 'in_progress' | 'done';
  why: string;
  fix: string;
  paths: string[];
}

export const DOMAINS: DomainInfo[] = [
  { id: 'nrp', label: 'NRP', subtitle: 'Nurse Rostering / INRC-II' },
  { id: 'cvrp', label: 'CVRP', subtitle: 'Capacitated Vehicle Routing' },
  { id: 'vrptw', label: 'VRPTW', subtitle: 'Vehicle Routing with Time Windows' },
  { id: 'jss', label: 'JSS', subtitle: 'Job Shop Scheduling' },
];

const all = (s: CellStatus): Record<DomainId, CellStatus> => ({
  nrp: s,
  cvrp: s,
  vrptw: s,
  jss: s,
});

export const SEARCH_ALGORITHMS: MatrixRow[] = [
  { id: 'sa', label: 'SA', notes: 'Generic metaheuristic', cells: all('done') },
  { id: 'lahc', label: 'LAHC', notes: 'Generic metaheuristic', cells: all('done') },
  { id: 'tabu', label: 'Tabu', notes: 'Generic metaheuristic', cells: all('done') },
  { id: 'portfolio', label: 'Portfolio', notes: 'Runs multiple strategies', cells: all('done') },
  { id: 'adaptive', label: 'Adaptive', notes: 'SI policy-driven behaviour', cells: all('done') },
  {
    id: 'ilp',
    label: 'ILP benchmark',
    notes: 'Reference/benchmark, not main solver',
    cells: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
  },
];

export const SI_MODES: MatrixRow[] = [
  { id: 'off', label: 'off', notes: 'Baseline, no SI', cells: all('done') },
  { id: 'shadow', label: 'shadow', notes: 'Records decisions, no behaviour change', cells: all('done') },
  { id: 'assist', label: 'assist', notes: 'Safe recommendations can affect compute', cells: all('done') },
  { id: 'adaptive', label: 'adaptive', notes: 'Live adaptive search control', cells: all('done') },
  { id: 'rules', label: 'rules policy', notes: 'v1 rule policy', cells: all('done') },
  { id: 'hybrid', label: 'hybrid policy', notes: 'Learned when confident, rules fallback', cells: all('done') },
  { id: 'learned', label: 'learned policy', notes: 'Learned decisions, safety fallback', cells: all('done') },
];

export const ASSIST_ARCHITECTURE: LayeredAssistRow[] = [
  {
    id: 'worker',
    label: 'WorkerAssist',
    description: 'Beam worker spawn / skip / budget (native on PFRS beam)',
    runtime: { nrp: 'done', cvrp: 'na', vrptw: 'na', jss: 'na' },
    telemetry: { nrp: 'done', cvrp: 'adapted', vrptw: 'adapted', jss: 'adapted' },
    fixPhase: '—',
    fixNote: 'Routing/JSS use SearchAssist natively; adapted worker_assist.csv via si_adapters.go',
  },
  {
    id: 'search',
    label: 'SearchAssist',
    description: 'Single-search stop / extend / restart checkpoints',
    runtime: { nrp: 'partial', cvrp: 'done', vrptw: 'done', jss: 'done' },
    telemetry: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
    fixPhase: 'Phase 3',
    fixNote: 'Worker-mapped policy CSVs + real AllocatedIters in search adapter',
  },
  {
    id: 'portfolio',
    label: 'PortfolioAssist',
    description: 'Strategy budget allocation across portfolio runs',
    runtime: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
    telemetry: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
  },
  {
    id: 'policy',
    label: 'PolicyExecutor',
    description: 'rules / hybrid / learned SI 2.0 execution',
    runtime: { nrp: 'partial', cvrp: 'done', vrptw: 'done', jss: 'done' },
    telemetry: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
    fixPhase: 'Phase 3',
    fixNote: 'policy_decisions/evaluation emitted from worker assist (Go)',
  },
];

export const VIEWERS: MatrixRow[] = [
  { id: 'schedule', label: 'Schedule / roster', cells: { nrp: 'done', cvrp: 'na', vrptw: 'na', jss: 'na' } },
  { id: 'routes', label: 'Route viewer', cells: { nrp: 'na', cvrp: 'done', vrptw: 'done', jss: 'na' } },
  { id: 'gantt', label: 'Gantt chart', cells: { nrp: 'na', cvrp: 'na', vrptw: 'na', jss: 'done' } },
  {
    id: 'constraints',
    label: 'Constraints',
    cells: { nrp: 'done', cvrp: 'done', vrptw: 'done', jss: 'done' },
    notes: 'NRP S1–S8 page; routing/JSS feasibility on summary from solution.json',
  },
  { id: 'summary', label: 'Summary', cells: all('done') },
  { id: 'benchmarks', label: 'Benchmark ladder', cells: all('done') },
  { id: 'statistics', label: 'Statistics', cells: all('done') },
  { id: 'compare', label: 'Compare', cells: all('done') },
  { id: 'trends', label: 'Trends', cells: all('done') },
  { id: 'intelligence', label: 'Search Intelligence', cells: all('done') },
];

export const VALIDATION_STATUS: Array<{
  domain: DomainId;
  status: CellStatus;
  notes: string;
}> = [
  { domain: 'nrp', status: 'done', notes: 'WorkerAssist + portfolio; within stochastic variance' },
  { domain: 'cvrp', status: 'done', notes: 'Same quality, major compute saving' },
  { domain: 'vrptw', status: 'done', notes: 'Better quality through budget extension' },
  { domain: 'jss', status: 'done', notes: 'Earlier rule issues; learned/hybrid policy addresses them' },
];

export const GAP_ROADMAP: GapItem[] = [
  {
    id: 'docs-clarity',
    title: 'Split matrix: runtime vs telemetry vs UX',
    phase: 'Phase 0',
    effort: '½ day',
    status: 'done',
    why: 'False reds when adapted CSV parity exists but native hook differs',
    fix: 'docs/CAPABILITIES.md + this page',
    paths: ['docs/CAPABILITIES.md', 'platform/web/pfrs-lab/src/lib/capability-matrix.ts'],
  },
  {
    id: 'capabilities-page',
    title: 'Capabilities page on dashboard',
    phase: 'Phase 1',
    effort: '1 day',
    status: 'done',
    why: 'Reviewers need single honest product matrix',
    fix: '/capabilities route with live registry badge',
    paths: ['platform/web/pfrs-lab/src/app/capabilities/'],
  },
  {
    id: 'vrptw-routes',
    title: 'VRPTW route viewer polish',
    phase: 'Phase 2',
    effort: '1 day',
    status: 'done',
    why: 'Page copy is CVRP-biased; TW violations not shown',
    fix: 'Branch routes/page.tsx on problemType; per-route feasible + TW violations',
    paths: ['platform/web/pfrs-lab/src/app/runs/[id]/routes/'],
  },
  {
    id: 'constraint-summaries',
    title: 'Feasibility summaries on non-NRP runs',
    phase: 'Phase 2',
    effort: '2–3 days',
    status: 'done',
    why: 'No dedicated constraint viewer for CVRP/VRPTW/JSS',
    fix: 'FeasibilitySummaryCard on summary from solution.json (capacity, TW, machine load)',
    paths: ['platform/web/pfrs-lab/src/app/runs/[id]/summary/'],
  },
  {
    id: 'nrp-search-assist',
    title: 'NRP SearchAssist + policy hook fidelity',
    phase: 'Phase 3',
    effort: '2–4 days',
    status: 'done',
    why: 'Adapter lacked real budget fields; NRP policy CSVs stubbed in Go',
    fix: 'si_adapters.go AllocatedIters + emitNRPPolicyCSVs in cli_telemetry.go',
    paths: ['platform/go/cmd/owp/si_adapters.go', 'platform/go/cmd/owp/cli_telemetry.go'],
  },
  {
    id: 'vrptw-ilp',
    title: 'VRPTW ILP benchmark command',
    phase: 'Phase 4',
    effort: '1–2 weeks',
    status: 'done',
    why: 'Benchmarks use Solomon BKS, not HiGHS proof',
    fix: 'vrptw/ilp package + owp benchmark-vrptw-ilp',
    paths: ['platform/go/internal/infrastructure/vrptw/ilp/'],
  },
  {
    id: 'jss-ilp',
    title: 'JSS ILP benchmark command',
    phase: 'Phase 5',
    effort: '1–2 weeks',
    status: 'done',
    why: 'Benchmarks use published optima, not solver proof',
    fix: 'jobshop/ilp package + owp benchmark-jss-ilp',
    paths: ['platform/go/internal/infrastructure/jobshop/ilp/'],
  },
  {
    id: 'domain-constraints',
    title: 'Domain constraint viewer pages',
    phase: 'Phase 6',
    effort: '3–5 days',
    status: 'open',
    why: 'constraints/ page is NRP-only',
    fix: 'CVRP capacity, VRPTW TW, JSS machine load panels',
    paths: ['platform/web/pfrs-lab/src/app/runs/[id]/constraints/'],
  },
];

export function statusLabel(s: CellStatus): string {
  switch (s) {
    case 'done': return '✅';
    case 'partial': return '⚠️';
    case 'missing': return '❌';
    case 'na': return '—';
    case 'adapted': return '🔀';
    default: return '?';
  }
}

export function statusTitle(s: CellStatus): string {
  switch (s) {
    case 'done': return 'Complete';
    case 'partial': return 'Partial';
    case 'missing': return 'Missing';
    case 'na': return 'Not applicable';
    case 'adapted': return 'Adapted (telemetry parity)';
    default: return s;
  }
}

export function statusClass(s: CellStatus): string {
  switch (s) {
    case 'done': return 'text-emerald-400';
    case 'partial': return 'text-amber-400';
    case 'missing': return 'text-red-400';
    case 'na': return 'text-gray-600';
    case 'adapted': return 'text-blue-400';
    default: return 'text-gray-400';
  }
}
