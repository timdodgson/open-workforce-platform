import type { DomainId } from '@/lib/capability-matrix';

export interface DomainGap {
  domain: DomainId;
  label: string;
  instances: number;
  atOptimal: number;
  bestGapPct: number | null;
  worstGapPct: number | null;
  avgGapPct: number | null;
}

/** Published benchmark gaps — see docs/BENCHMARKS.md Platform Performance Summary. */
export const DOMAIN_GAPS: DomainGap[] = [
  { domain: 'cvrp', label: 'CVRP', instances: 4, atOptimal: 1, bestGapPct: 0.3, worstGapPct: 3.6, avgGapPct: 1.6 },
  { domain: 'jss', label: 'JSS', instances: 3, atOptimal: 2, bestGapPct: 0, worstGapPct: 2.8, avgGapPct: 0.9 },
  { domain: 'vrptw', label: 'VRPTW', instances: 1, atOptimal: 0, bestGapPct: 0.1, worstGapPct: 0.1, avgGapPct: 0.1 },
  { domain: 'nrp', label: 'NRP', instances: 2, atOptimal: 0, bestGapPct: 13.9, worstGapPct: null, avgGapPct: null },
];

/**
 * The published problem the platform has not yet cracked.
 * NRP leads on gap-to-reference, exact-method difficulty, and open SI regressions.
 */
export const FLAGSHIP_CHALLENGE = {
  domain: 'nrp' as DomainId,
  label: 'NRP — INRC-II Nurse Rostering',
  instance: 'n012w8',
  platformBest: 3440,
  platformMode: 'Portfolio + SI hybrid + diversity 30% + fw 6M (3M/worker)',
  referenceLabel: 'HiGHS ILP feasible (reference)',
  referenceValue: 3020,
  gapPct: 13.9,
  whyHardest: [
    'Largest gap to a published reference (+13.9% vs ILP feasible on n012w8) — ILP is a yardstick, not a scalable solver.',
    'Exact methods stall: HiGHS takes hours on n012w8 and still leaves a ~56% MIP gap; larger instances are impractical for day-to-day search.',
    'Only domain where SI policy modes regress vs rules on the val-* harness (+378–486 penalty).',
    'Multi-week beam search, succession rules, and soft-constraint trade-offs — not a single-route problem.',
  ],
  nextAlgorithms: ['final-window', 'diversity', 'multi-seed'] as const,
  /** Primary flagship path — multi-week PFRS beam (cwd: platform/go). */
  tryCommand:
    'go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-beam-width 12 --pfrs-beam-seeds 42,101,202,303,404 --pfrs-iterations-per-worker 3000000 --pfrs-max-total-workers 24 --pfrs-max-concurrent 8 --pfrs-beam-strategy budget --pfrs-lookahead-weight 4.0 --pfrs-final-window-weeks 2 --pfrs-final-window-iterations 6000000 --pfrs-diversity-slots 30 --worker-decision-mode assist --policy-mode hybrid --policy-dir ../ml/policies',
  /** GA inside the same beam path (not single-week solve). */
  gaCommand:
    'go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu,ga --pfrs-beam-width 12 --pfrs-beam-seeds 42,101,202 --pfrs-iterations-per-worker 300000 --pfrs-max-concurrent 8',
  cwd: 'platform/go',
  /** HiGHS dual / MIP lower bound (not a proven optimal roster). */
  ilpLowerBound: 1845,
  /** Best feasible ILP solution found within the published time budget. */
  ilpFeasible: 3020,
} as const;

export function hardestDomain(): DomainGap {
  return DOMAIN_GAPS.reduce((worst, d) => {
    const gap = d.worstGapPct ?? d.bestGapPct ?? 0;
    const worstGap = worst.worstGapPct ?? worst.bestGapPct ?? 0;
    return gap > worstGap ? d : worst;
  });
}
