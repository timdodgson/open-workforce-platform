import {
  VARIATION_CONFIGS,
  allLabelsForConfig,
  countMatchingRuns,
  expandLabel,
  matrixSummary,
} from '@/lib/experiment-matrix';

describe('experiment-matrix', () => {
  it('fast tier sums to 240 variations', () => {
    const summary = matrixSummary();
    expect(summary.fastVariations).toBe(240);
    expect(summary.deepVariations).toBe(48);
  });

  it('expands label patterns', () => {
    const cfg = VARIATION_CONFIGS.find((c) => c.id === 'fast-cvrp-sa')!;
    expect(expandLabel(cfg, 'hybrid', 42)).toBe('val-cvrp-a32k5-sa-hybrid-s42');
  });

  it('counts matching run ids', () => {
    const cfg = VARIATION_CONFIGS.find((c) => c.id === 'fast-nrp-sa')!;
    const labels = allLabelsForConfig(cfg);
    expect(labels).toHaveLength(30);
    expect(countMatchingRuns([labels[0], 'other-run'], cfg)).toBe(1);
  });
});
