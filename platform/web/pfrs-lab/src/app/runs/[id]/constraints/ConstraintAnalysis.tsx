'use client';

import { useState } from 'react';
import Card from '@/components/Card';
import { WeekRecord } from '@/lib/types';
import { RosterEntry } from '@/lib/data-loader';

// Constraint definitions matching INRC-II spec.
const CONSTRAINTS = [
  { id: 'S1', name: 'Optimal Coverage', weight: 30, description: 'Understaffing below optimal level' },
  { id: 'S2', name: 'Consecutive Working Days', weight: 30, description: 'Too many or too few consecutive working days' },
  { id: 'S3', name: 'Consecutive Days Off', weight: 30, description: 'Too many or too few consecutive days off' },
  { id: 'S4', name: 'Consecutive Shift Type', weight: 15, description: 'Too many or too few consecutive same-shift assignments' },
  { id: 'S5', name: 'Shift-Off Requests', weight: 10, description: 'Working when a day-off was requested' },
  { id: 'S6', name: 'Complete Weekends', weight: 30, description: 'Working Saturday XOR Sunday (incomplete)' },
  { id: 'S7', name: 'Total Assignments', weight: 20, description: 'Over or under total assignment limits' },
  { id: 'S8', name: 'Total Working Weekends', weight: 30, description: 'Exceeding maximum working weekends' },
];

interface ConstraintBreakdown {
  id: string;
  name: string;
  weight: number;
  penalty: number;
  violations: number;
  percentage: number;
  available: boolean;
}

interface ExportedConstraintRow {
  id: string;
  penalty: number;
  violations: number;
}

interface ExportedBreakdown {
  totalPenalty: number;
  numWeeks: number;
  hardViolations: number;
  constraints: ExportedConstraintRow[];
}

interface Props {
  weeks: WeekRecord[];
  totalPenalty: number;
  roster: RosterEntry[];
  numWeeks: number;
  exportedBreakdown?: ExportedBreakdown | null;
}

type SortKey = 'penalty' | 'violations' | 'id';
type SortDir = 'asc' | 'desc';

function estimateConstraintBreakdown(weeks: WeekRecord[], totalPenalty: number, roster: RosterEntry[]): ConstraintBreakdown[] {
  // Currently we don't have per-constraint breakdown from the scorer.
  // We provide what we CAN derive from the roster + results data,
  // and mark unavailable constraints clearly.

  const hasRoster = roster.length > 0;
  const totalViolations = weeks.reduce((sum, w) => sum + w.softViolations, 0);

  // S6: Complete weekends — we can compute this from the roster.
  let s6Violations = 0;
  if (hasRoster) {
    // Group by nurse + week, check if sat XOR sun worked.
    const nurseWeekWork = new Map<string, { sat: boolean; sun: boolean }>();
    for (const entry of roster) {
      const key = `${entry.nurse}_${entry.week}`;
      if (!nurseWeekWork.has(key)) {
        nurseWeekWork.set(key, { sat: false, sun: false });
      }
      const nw = nurseWeekWork.get(key)!;
      if (entry.dayIndex === 5) nw.sat = true;
      if (entry.dayIndex === 6) nw.sun = true;
    }
    for (const nw of nurseWeekWork.values()) {
      if (nw.sat !== nw.sun) s6Violations++;
    }
  }

  // Build breakdown with available/unavailable markers.
  const breakdown: ConstraintBreakdown[] = CONSTRAINTS.map(c => {
    if (c.id === 'S6' && hasRoster) {
      return {
        ...c,
        penalty: s6Violations * c.weight,
        violations: s6Violations,
        percentage: totalPenalty > 0 ? (s6Violations * c.weight / totalPenalty) * 100 : 0,
        available: true,
      };
    }
    // For others, mark as unavailable — detailed telemetry not yet exported.
    return {
      ...c,
      penalty: 0,
      violations: 0,
      percentage: 0,
      available: false,
    };
  });

  // Estimate remaining penalty distribution (rough).
  const knownPenalty = breakdown.filter(b => b.available).reduce((s, b) => s + b.penalty, 0);
  const unknownPenalty = totalPenalty - knownPenalty;
  const unavailable = breakdown.filter(b => !b.available);

  if (unavailable.length > 0 && unknownPenalty > 0) {
    // Distribute unknown proportionally by weight as rough estimate.
    const totalWeight = unavailable.reduce((s, b) => s + b.weight, 0);
    for (const b of unavailable) {
      b.penalty = Math.round((b.weight / totalWeight) * unknownPenalty);
      b.violations = b.weight > 0 ? Math.round(b.penalty / b.weight) : 0;
      b.percentage = totalPenalty > 0 ? (b.penalty / totalPenalty) * 100 : 0;
    }
  }

  return breakdown;
}

function buildFromExported(exported: ExportedBreakdown): ConstraintBreakdown[] {
  const totalPenalty = exported.totalPenalty || 1;
  return CONSTRAINTS.map((c) => {
    const row = exported.constraints.find((r) => r.id === c.id);
    const penalty = row?.penalty ?? 0;
    const violations = row?.violations ?? 0;
    return {
      ...c,
      penalty,
      violations,
      percentage: totalPenalty > 0 ? (penalty / totalPenalty) * 100 : 0,
      available: true,
    };
  });
}

export default function ConstraintAnalysis({
  weeks,
  totalPenalty,
  roster,
  numWeeks,
  exportedBreakdown,
}: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('penalty');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  const breakdown = exportedBreakdown
    ? buildFromExported(exportedBreakdown)
    : estimateConstraintBreakdown(weeks, totalPenalty, roster);
  const hasDetailedData = breakdown.some(b => b.available);
  const allEstimated = !hasDetailedData;

  // Sort.
  const sorted = [...breakdown].sort((a, b) => {
    const dir = sortDir === 'desc' ? -1 : 1;
    if (sortKey === 'penalty') return (a.penalty - b.penalty) * dir;
    if (sortKey === 'violations') return (a.violations - b.violations) * dir;
    return a.id.localeCompare(b.id) * dir;
  });

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir(sortDir === 'desc' ? 'asc' : 'desc');
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  }

  // Find dominant constraint.
  const dominant = sorted[0];
  const maxPenalty = Math.max(...breakdown.map(b => b.penalty), 1);

  // Colour by constraint type.
  function barColor(id: string): string {
    const colors: Record<string, string> = {
      S1: 'bg-blue-500', S2: 'bg-emerald-500', S3: 'bg-cyan-500',
      S4: 'bg-purple-500', S5: 'bg-amber-500', S6: 'bg-rose-500',
      S7: 'bg-orange-500', S8: 'bg-red-500',
    };
    return colors[id] || 'bg-gray-500';
  }

  function textColor(id: string): string {
    const colors: Record<string, string> = {
      S1: 'text-blue-400', S2: 'text-emerald-400', S3: 'text-cyan-400',
      S4: 'text-purple-400', S5: 'text-amber-400', S6: 'text-rose-400',
      S7: 'text-orange-400', S8: 'text-red-400',
    };
    return colors[id] || 'text-gray-400';
  }

  return (
    <div className="space-y-4">
      {/* Data availability notice */}
      {allEstimated && (
        <Card title="⚠️ Estimated Breakdown">
          <p className="text-xs text-amber-400">
            Detailed per-constraint scoring data is not yet exported by the solver.
            Values below are <strong>estimated</strong> based on penalty weights and available roster data.
          </p>
        </Card>
      )}

      {exportedBreakdown && (
        <Card title="ILP Constraint Breakdown">
          <p className="text-xs text-emerald-400">
            Exact S1–S8 penalties from the official INRC-II scorer ({exportedBreakdown.numWeeks} weeks,{' '}
            {exportedBreakdown.hardViolations} hard violations).
          </p>
        </Card>
      )}

      {/* Summary stats */}
      <Card title="Penalty Summary">
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div className="text-center">
            <p className="text-2xl font-bold text-emerald-400">{totalPenalty.toLocaleString()}</p>
            <p className="text-xs text-gray-500">Total Penalty</p>
          </div>
          <div className="text-center">
            <p className={`text-2xl font-bold ${textColor(dominant?.id || 'S1')}`}>
              {dominant?.id || '—'}
            </p>
            <p className="text-xs text-gray-500">Dominant Constraint</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-gray-300">{numWeeks}</p>
            <p className="text-xs text-gray-500">Weeks</p>
          </div>
        </div>
      </Card>

      {/* Stacked bar chart */}
      <Card title="Penalty Distribution">
        {/* Horizontal stacked bar */}
        <div className="mb-4">
          <div className="h-8 rounded-lg overflow-hidden flex">
            {sorted.filter(b => b.penalty > 0).map(b => (
              <div
                key={b.id}
                className={`${barColor(b.id)} relative group`}
                style={{ width: `${b.percentage}%`, minWidth: b.percentage > 0 ? '2px' : '0' }}
                title={`${b.id}: ${b.penalty} (${b.percentage.toFixed(1)}%)`}
              >
                {b.percentage > 8 && (
                  <span className="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-white">
                    {b.id}
                  </span>
                )}
              </div>
            ))}
          </div>
          {/* Legend */}
          <div className="flex flex-wrap gap-3 mt-2">
            {sorted.filter(b => b.penalty > 0).map(b => (
              <div key={b.id} className="flex items-center gap-1">
                <div className={`w-3 h-3 rounded-sm ${barColor(b.id)}`} />
                <span className="text-[10px] text-gray-400">
                  {b.id} ({b.percentage.toFixed(1)}%)
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Individual bars */}
        <div className="space-y-2">
          {sorted.map(b => (
            <div key={b.id} className="flex items-center gap-2">
              <span className={`w-8 text-xs font-mono ${textColor(b.id)}`}>{b.id}</span>
              <div className="flex-1 h-5 bg-gray-800 rounded overflow-hidden relative">
                <div
                  className={`h-full ${barColor(b.id)} transition-all duration-300 ${
                    !b.available ? 'opacity-40' : ''
                  }`}
                  style={{ width: `${(b.penalty / maxPenalty) * 100}%` }}
                />
                <span className="absolute right-2 top-0 h-full flex items-center text-[10px] text-gray-300">
                  {b.penalty.toLocaleString()}
                </span>
              </div>
              <span className="w-12 text-right text-[10px] text-gray-500">
                {b.percentage.toFixed(1)}%
              </span>
              {!b.available && (
                <span className="text-[9px] text-amber-600">est</span>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* Sortable detail table */}
      <Card title="Constraint Details">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2 cursor-pointer hover:text-white" onClick={() => toggleSort('id')}>
                Constraint {sortKey === 'id' ? (sortDir === 'desc' ? '↓' : '↑') : ''}
              </th>
              <th className="text-left p-2">Description</th>
              <th className="text-right p-2">Weight</th>
              <th className="text-right p-2 cursor-pointer hover:text-white" onClick={() => toggleSort('violations')}>
                Violations {sortKey === 'violations' ? (sortDir === 'desc' ? '↓' : '↑') : ''}
              </th>
              <th className="text-right p-2 cursor-pointer hover:text-white" onClick={() => toggleSort('penalty')}>
                Penalty {sortKey === 'penalty' ? (sortDir === 'desc' ? '↓' : '↑') : ''}
              </th>
              <th className="text-right p-2">% of Total</th>
              <th className="text-center p-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map(b => (
              <tr key={b.id} className={`border-t border-gray-800 ${
                b === dominant ? 'bg-gray-800/30' : ''
              }`}>
                <td className={`p-2 font-mono font-bold ${textColor(b.id)}`}>
                  {b.id}
                  {b === dominant && <span className="ml-1 text-[9px] text-yellow-500">★</span>}
                </td>
                <td className="p-2 text-gray-400">{b.name}</td>
                <td className="text-right p-2 text-gray-500">{b.weight}</td>
                <td className="text-right p-2">{b.violations.toLocaleString()}</td>
                <td className="text-right p-2 font-medium">{b.penalty.toLocaleString()}</td>
                <td className="text-right p-2 text-gray-400">{b.percentage.toFixed(1)}%</td>
                <td className="text-center p-2">
                  {b.available ? (
                    <span className="text-emerald-400 text-[10px]">✓ exact</span>
                  ) : (
                    <span className="text-amber-500 text-[10px]">≈ est</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr className="border-t-2 border-gray-600 font-medium">
              <td className="p-2" colSpan={3}>Total</td>
              <td className="text-right p-2">
                {sorted.reduce((s, b) => s + b.violations, 0).toLocaleString()}
              </td>
              <td className="text-right p-2">{totalPenalty.toLocaleString()}</td>
              <td className="text-right p-2">100%</td>
              <td />
            </tr>
          </tfoot>
        </table>
      </Card>

      {/* Per-week soft violations (what we have) */}
      <Card title="Soft Violations by Week">
        <div className="flex items-end gap-1 h-24">
          {weeks.map(w => {
            const maxViol = Math.max(...weeks.map(wk => wk.softViolations), 1);
            const height = (w.softViolations / maxViol) * 100;
            return (
              <div key={w.week} className="flex-1 flex flex-col items-center gap-1">
                <span className="text-[9px] text-gray-500">{w.softViolations}</span>
                <div
                  className="w-full bg-purple-600 rounded-t"
                  style={{ height: `${Math.max(2, height)}%` }}
                  title={`Week ${w.week}: ${w.softViolations} violations, ${w.finalPenalty} penalty`}
                />
                <span className="text-[9px] text-gray-600">W{w.week}</span>
              </div>
            );
          })}
        </div>
      </Card>

      {/* Future data placeholder */}
      <Card title="🔮 Future: Detailed Constraint Telemetry">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-6 text-center">
          <p className="text-gray-500 text-sm mb-2">
            Per-constraint breakdown will be available when the Go scorer exports detailed results.
          </p>
          <p className="text-gray-600 text-xs">
            Planned fields: constraint_type, nurse, week, day, penalty_amount, violation_details
          </p>
        </div>
      </Card>
    </div>
  );
}
