import type { RosterEntry } from '@/lib/data-loader';

/** NRP roster.json / solution.json — array of shift assignments. */
export function isRosterSolution(value: unknown): value is RosterEntry[] {
  if (!Array.isArray(value) || value.length === 0) return false;
  const first = value[0];
  if (!first || typeof first !== 'object') return false;
  const row = first as Record<string, unknown>;
  return typeof row.nurse === 'string' && typeof row.shiftType === 'string';
}
