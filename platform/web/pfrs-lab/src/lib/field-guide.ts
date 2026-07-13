/**
 * Public field-guide copy — algorithms, domains, external standards.
 * Keep claims aligned with docs/BENCHMARKS.md and README.
 */

export interface AlgorithmGuide {
  id: string;
  name: string;
  abbr: string;
  family: string;
  strength: string;
  weakness: string;
  when: string;
  realWorld: string;
}

export interface DomainGuide {
  id: string;
  name: string;
  shortName: string;
  realWorld: string;
  whyHard: string;
  benchmark: string;
  benchmarkUrl: string;
  platformNote: string;
  relatedAlgos: string[];
}

export interface ReferenceLink {
  id: string;
  name: string;
  kind: 'society' | 'benchmark' | 'solver' | 'archive';
  url: string;
  why: string;
}

export const ALGORITHMS: AlgorithmGuide[] = [
  {
    id: 'sa',
    name: 'Simulated Annealing',
    abbr: 'SA',
    family: 'Trajectory / single-solution',
    strength: 'Accepts uphill moves early, then cools — strong general baseline across domains.',
    weakness: 'Needs a sensible temperature schedule; can wander if cooling is too slow.',
    when: 'Default first lever on a new instance; often strong on VRPTW.',
    realWorld: 'Same idea as “explore then exploit” in scheduling and logistics when the landscape is rugged.',
  },
  {
    id: 'lahc',
    name: 'Late Acceptance Hill Climbing',
    abbr: 'LAHC',
    family: 'Trajectory / history-based',
    strength: 'Escapes local optima without a temperature parameter; simple and fast.',
    weakness: 'History length matters; less memory of tabu-style forbidden moves.',
    when: 'Often hits or nears optimal on small CVRP Augerat instances.',
    realWorld: 'Useful when operators want a robust improver that is easy to reason about.',
  },
  {
    id: 'tabu',
    name: 'Tabu Search',
    abbr: 'Tabu',
    family: 'Trajectory / memory-based',
    strength: 'Forbids recent moves so search does not cycle; strong on larger neighbourhoods.',
    weakness: 'Neighbourhood evaluation is more expensive per iteration.',
    when: 'Larger CVRP and classic JSS instances in our ladder.',
    realWorld: 'Common in production rostering and shop-floor scheduling toolkits.',
  },
  {
    id: 'ga',
    name: 'Genetic Algorithm',
    abbr: 'GA',
    family: 'Population-based',
    strength: 'Maintains diversity across many candidate solutions at once.',
    weakness: 'Higher memory and clone cost under heavy parallel beam search.',
    when: 'Portfolio member on NRP; competitive on CVRP smoke tests.',
    realWorld: 'Population search is a staple of evolutionary OR and competition entries for rostering.',
  },
  {
    id: 'portfolio',
    name: 'Portfolio Mode',
    abbr: 'Portfolio',
    family: 'Algorithm portfolio',
    strength: 'Runs several strategies and keeps the best — quality ceiling of the best member.',
    weakness: 'Uses more compute; under worker caps, members share a budget.',
    when: 'When the goal is best-known quality, not minimum wall-clock.',
    realWorld: 'Mirrors how solvers hedge: do not bet the ward or fleet on one heuristic.',
  },
];

export const DOMAINS: DomainGuide[] = [
  {
    id: 'nrp',
    name: 'Nurse Rostering (NRP)',
    shortName: 'NRP',
    realWorld:
      'Hospitals and care providers must cover shifts with skilled staff while respecting contracts, succession rules, weekends, and fairness. Bad rosters drive overtime cost, burnout, and unsafe coverage.',
    whyHard:
      'Combinatorial assignment over nurses × days × skills with hard feasibility and many soft trade-offs. Multi-week horizons couple decisions through history.',
    benchmark: 'INRC-II — International Nurse Rostering Competition II',
    benchmarkUrl: 'http://mobiz.vives.be/inrc2/',
    platformNote:
      'Flagship challenge on this platform: largest gap to reference on n012w8; solved via parallel beam search (PFRS), not a single local-search call.',
    relatedAlgos: ['SA', 'LAHC', 'Tabu', 'GA', 'Portfolio'],
  },
  {
    id: 'cvrp',
    name: 'Capacitated Vehicle Routing (CVRP)',
    shortName: 'CVRP',
    realWorld:
      'Depots dispatch capacity-limited vehicles to customers. Distance and fleet size drive fuel, labour, and service level in retail, parcel, and field service.',
    whyHard:
      'Classic NP-hard routing: visit each customer once, respect vehicle capacity, minimise distance.',
    benchmark: 'CVRPLIB — Augerat and related published instances',
    benchmarkUrl: 'http://vrp.atd-lab.inf.puc-rio.br/',
    platformNote:
      'Several Augerat Set A instances within a few percent of proven optima; LAHC/Tabu often lead.',
    relatedAlgos: ['LAHC', 'Tabu', 'SA', 'GA', 'Portfolio'],
  },
  {
    id: 'vrptw',
    name: 'Vehicle Routing with Time Windows (VRPTW)',
    shortName: 'VRPTW',
    realWorld:
      'Deliveries and home visits arrive in time windows — pharmacies, grocery slots, technicians. Late arrivals are infeasible, not just expensive.',
    whyHard:
      'CVRP plus hard time windows: feasible geography can still be infeasible in time.',
    benchmark: 'Solomon benchmark instances (SINTEF / TOP)',
    benchmarkUrl: 'https://www.sintef.no/projectweb/top/vrptw/solomon-benchmark/',
    platformNote:
      'Solomon C101 used as the primary instance; platform results sit close to published BKS distance.',
    relatedAlgos: ['SA', 'LAHC', 'Tabu', 'GA', 'Portfolio'],
  },
  {
    id: 'jss',
    name: 'Job Shop Scheduling (JSS)',
    shortName: 'JSS',
    realWorld:
      'Factories sequence operations on machines so jobs finish on time. Makespan and bottlenecks dominate throughput and WIP.',
    whyHard:
      'Permutation of operations with machine conflicts; small instance changes can scramble the schedule.',
    benchmark: 'OR-Library / Fisher & Thompson / Lawrence instances',
    benchmarkUrl: 'http://jobshop.jjvh.nl/',
    platformNote:
      'la01 and ft06 often optimal under Tabu/LAHC; harder instances (ft10) remain a measured gap.',
    relatedAlgos: ['Tabu', 'LAHC', 'SA', 'GA', 'Portfolio'],
  },
];

export const REFERENCES: ReferenceLink[] = [
  {
    id: 'or-society',
    name: 'The OR Society (UK)',
    kind: 'society',
    url: 'https://www.theorsociety.com/',
    why: 'Professional body for operational research in the UK — practice, education, and community.',
  },
  {
    id: 'informs',
    name: 'INFORMS',
    kind: 'society',
    url: 'https://www.informs.org/',
    why: 'International society for analytics and OR; journals and conferences that define the field.',
  },
  {
    id: 'euro',
    name: 'EURO — Association of European OR Societies',
    kind: 'society',
    url: 'https://www.euro-online.org/',
    why: 'European OR umbrella; links national societies and EURO conferences.',
  },
  {
    id: 'inrc2',
    name: 'INRC-II Nurse Rostering Competition',
    kind: 'benchmark',
    url: 'http://mobiz.vives.be/inrc2/',
    why: 'Published competition instances and scoring model for multi-week nurse rostering.',
  },
  {
    id: 'cvrplib',
    name: 'CVRPLIB',
    kind: 'benchmark',
    url: 'http://vrp.atd-lab.inf.puc-rio.br/',
    why: 'Canonical capacitated VRP instance library with published optima / BKS.',
  },
  {
    id: 'solomon',
    name: 'Solomon VRPTW benchmarks (SINTEF TOP)',
    kind: 'benchmark',
    url: 'https://www.sintef.no/projectweb/top/vrptw/solomon-benchmark/',
    why: 'Standard time-window routing instances used across academia and industry papers.',
  },
  {
    id: 'jobshop',
    name: 'Job shop instance archive',
    kind: 'archive',
    url: 'http://jobshop.jjvh.nl/',
    why: 'Classic JSS instances (ft*, la*) with known optima for reproducible comparison.',
  },
  {
    id: 'or-library',
    name: 'OR-Library (Brunel)',
    kind: 'archive',
    url: 'https://people.brunel.ac.uk/~mastjjb/jeb/info.html',
    why: 'Long-standing public OR datasets spanning scheduling, routing, and more.',
  },
  {
    id: 'highs',
    name: 'HiGHS — open-source LP/MIP solver',
    kind: 'solver',
    url: 'https://highs.dev/',
    why: 'Yardstick only: ILP/MIP reference for gap reporting on small instances. Does not replace heuristics — exact search does not scale to large NRP/CVRP.',
  },
  {
    id: 'coin-or',
    name: 'COIN-OR',
    kind: 'solver',
    url: 'https://www.coin-or.org/',
    why: 'Open-source computational infrastructure for OR — solvers and modelling tools.',
  },
];

export const REFERENCE_KIND_LABEL: Record<ReferenceLink['kind'], string> = {
  society: 'Societies & professional bodies',
  benchmark: 'Benchmark competitions & libraries',
  archive: 'Instance archives',
  solver: 'Solvers & open infrastructure',
};
