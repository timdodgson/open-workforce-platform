import type { IntelligenceSection } from './types';

/** Keep S3 reads low — Lambda/ALB timeout risk on large scans. */
export const MAX_LEARNING_RUNS = 80;
export const MAX_POLICY_RUNS = 80;
export const MAX_COUNTERFACTUAL_RUNS = 80;
export const RUN_BATCH_SIZE = 8;
export const DEFAULT_PAGE_LIMIT = 100;

export const SECTION_CACHE_TTL: Partial<Record<IntelligenceSection, number>> = {
  summary: 120_000,
  promotion: 120_000,
  model: 120_000,
  'continuous-learning': 90_000,
  counterfactual: 90_000,
  policies: 60_000,
  assist: 60_000,
  learning: 45_000,
  decisions: 45_000,
};
