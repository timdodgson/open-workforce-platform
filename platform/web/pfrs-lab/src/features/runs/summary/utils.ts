export function improvementPct(initial: number, final: number): string {
  if (initial <= 0) return '0';
  return ((initial - final) / initial * 100).toFixed(1);
}

export function runtimeSeconds(ms: number | undefined): string {
  return `${((ms || 0) / 1000).toFixed(1)}s`;
}
