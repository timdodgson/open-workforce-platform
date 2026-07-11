import type { Metadata } from 'next';
import { listRunsAsync } from '@/lib/data-loader';
import {
  VARIATION_CONFIGS,
  countMatchingRuns,
} from '@/lib/experiment-matrix';
import ExperimentMatrixView, { type ConfigCoverage } from './ExperimentMatrixView';

export const metadata: Metadata = {
  title: 'Experiment Matrix',
  description:
    'Catalog of every standard run variation across NRP, CVRP, JSS, and VRPTW — algorithms, SI options, and why each flag is on or off.',
};

export const dynamic = 'force-dynamic';

async function loadCoverage(): Promise<ConfigCoverage[]> {
  try {
    const runs = await listRunsAsync();
    const ids = runs.map((r) => r.id);
    return VARIATION_CONFIGS.map((config) => ({
      configId: config.id,
      found: countMatchingRuns(ids, config),
      expected: config.variationsPerConfig,
    }));
  } catch {
    return VARIATION_CONFIGS.map((config) => ({
      configId: config.id,
      found: 0,
      expected: config.variationsPerConfig,
    }));
  }
}

export default async function ExperimentMatrixPage() {
  const coverage = await loadCoverage();
  return <ExperimentMatrixView coverage={coverage} />;
}
