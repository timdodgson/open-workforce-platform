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
  { domain: 'nrp', label: 'NRP', instances: 2, atOptimal: 0, bestGapPct: 14.7, worstGapPct: null, avgGapPct: null },
];

/**
 * The published problem the platform has not yet cracked.
 * NRP leads on gap-to-reference, exact-method difficulty, and open SI regressions.
 */
export const FLAGSHIP_CHALLENGE = {
  domain: 'nrp' as DomainId,
  label: 'NRP — INRC-II Nurse Rostering',
  instance: 'n012w8',
  platformBest: 3465,
  platformMode: 'Portfolio (PFRS beam)',
  referenceLabel: 'HiGHS ILP bound',
  referenceValue: 3020,
  gapPct: 14.7,
  whyHardest: [
    'Largest gap to a published reference (+14.7% vs ILP bound on n012w8).',
    'Exact methods stall: HiGHS leaves a 56% MIP gap within the time limit.',
    'Only domain where SI policy modes regress vs rules on the val-* harness (+378–486 penalty).',
    'Multi-week beam search, succession rules, and soft-constraint trade-offs — not a single-route problem.',
  ],
  nextAlgorithms: ['ga', 'portfolio'] as const,
  tryCommand:
    'owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-beam-width 12 --pfrs-beam-seeds 42,101,202',
  gaCommand:
    'owp solve nrp --instance n012w8 --mode ga --iterations 500000 --seed 42',
} as const;

export function hardestDomain(): DomainGap {
  return DOMAIN_GAPS.reduce((worst, d) => {
    const gap = d.worstGapPct ?? d.bestGapPct ?? 0;
    const worstGap = worst.worstGapPct ?? worst.bestGapPct ?? 0;
    return gap > worstGap ? d : worst;
  });
}
