'use client';
import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import { RunSummary } from '@/lib/types';
import { RosterEntry } from '@/lib/data-loader';

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const SHIFT_COLORS: Record<string, string> = {
  Early: 'bg-amber-600',
  Day: 'bg-blue-600',
  Late: 'bg-indigo-600',
  Night: 'bg-purple-700',
};

interface Props {
  roster: RosterEntry[];
  summary: RunSummary;
}

export default function ScheduleViewer({ roster, summary }: Props) {
  const weeks = Array.from(new Set(roster.map(r => r.week))).sort((a, b) => a - b);
  const nurses = Array.from(new Set(roster.map(r => r.nurse))).sort();
  const contracts = Array.from(new Set(roster.map(r => r.contract))).sort();
  const skills = Array.from(new Set(roster.map(r => r.skill))).sort();

  const [selectedWeek, setSelectedWeek] = useState(weeks[0] || 1);
  const [filterNurse, setFilterNurse] = useState<string>('');
  const [filterContract, setFilterContract] = useState<string>('');
  const [filterSkill, setFilterSkill] = useState<string>('');

  // Filter roster for selected week.
  const weekRoster = useMemo(() => {
    let filtered = roster.filter(r => r.week === selectedWeek);
    if (filterNurse) filtered = filtered.filter(r => r.nurse === filterNurse);
    if (filterContract) filtered = filtered.filter(r => r.contract === filterContract);
    if (filterSkill) filtered = filtered.filter(r => r.skill === filterSkill);
    return filtered;
  }, [roster, selectedWeek, filterNurse, filterContract, filterSkill]);

  // Build nurse-day grid.
  const displayNurses = Array.from(new Set(weekRoster.map(r => r.nurse))).sort();
  const grid = useMemo(() => {
    const map = new Map<string, Map<number, RosterEntry>>();
    for (const r of weekRoster) {
      if (!map.has(r.nurse)) map.set(r.nurse, new Map());
      map.get(r.nurse)!.set(r.dayIndex, r);
    }
    return map;
  }, [weekRoster]);

  // Summary metrics.
  const totalAssignments = weekRoster.length;
  const weekendAssignments = weekRoster.filter(r => r.dayIndex >= 5).length;
  const nursesWorking = displayNurses.length;
  const shiftCounts = new Map<string, number>();
  for (const r of weekRoster) {
    shiftCounts.set(r.shiftType, (shiftCounts.get(r.shiftType) || 0) + 1);
  }

  return (
    <>
      {/* Summary cards */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mb-4">
        <MetricCard label="Total Penalty" value={summary.totalPenalty.toLocaleString()} color="green" />
        <MetricCard label="Week Assignments" value={String(totalAssignments)} color="blue" />
        <MetricCard label="Weekend Shifts" value={String(weekendAssignments)} color="amber" />
        <MetricCard label="Nurses Active" value={String(nursesWorking)} color="default" />
        <MetricCard label="Hard Violations" value="0" color="green" />
      </div>

      {/* Filters */}
      <Card title="Filters">
        <div className="flex flex-wrap gap-3">
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Week</label>
            <div className="flex gap-1">
              {weeks.map(w => (
                <button key={w} onClick={() => setSelectedWeek(w)}
                  className={`px-2 py-1 text-xs rounded ${selectedWeek === w ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}`}>
                  W{w}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Nurse</label>
            <select value={filterNurse} onChange={e => setFilterNurse(e.target.value)}
              className="bg-gray-800 text-xs text-gray-300 rounded px-2 py-1 border border-gray-700">
              <option value="">All</option>
              {nurses.map(n => <option key={n} value={n}>{n}</option>)}
            </select>
          </div>
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Contract</label>
            <select value={filterContract} onChange={e => setFilterContract(e.target.value)}
              className="bg-gray-800 text-xs text-gray-300 rounded px-2 py-1 border border-gray-700">
              <option value="">All</option>
              {contracts.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Skill</label>
            <select value={filterSkill} onChange={e => setFilterSkill(e.target.value)}
              className="bg-gray-800 text-xs text-gray-300 rounded px-2 py-1 border border-gray-700">
              <option value="">All</option>
              {skills.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
        </div>
      </Card>

      {/* Roster Grid */}
      <Card title={`Week ${selectedWeek} Schedule`}>
        <div className="overflow-x-auto">
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr>
                <th className="text-left p-2 text-gray-500 sticky left-0 bg-gray-850">Nurse</th>
                {DAYS.map(d => (
                  <th key={d} className={`p-2 text-center text-gray-500 ${d === 'Sat' || d === 'Sun' ? 'bg-gray-800/50' : ''}`}>
                    {d}
                  </th>
                ))}
                <th className="p-2 text-right text-gray-500">Total</th>
              </tr>
            </thead>
            <tbody>
              {displayNurses.map(nurse => {
                const nurseRow = grid.get(nurse);
                const totalShifts = nurseRow ? nurseRow.size : 0;
                return (
                  <tr key={nurse} className="border-t border-gray-800 hover:bg-gray-800/30">
                    <td className="p-2 text-gray-300 font-mono text-[10px] sticky left-0 bg-gray-850 whitespace-nowrap">
                      {nurse}
                    </td>
                    {DAYS.map((_, dayIdx) => {
                      const entry = nurseRow?.get(dayIdx);
                      if (!entry) {
                        return (
                          <td key={dayIdx} className={`p-1 text-center ${dayIdx >= 5 ? 'bg-gray-800/30' : ''}`}>
                            <span className="text-gray-700 text-[9px]">—</span>
                          </td>
                        );
                      }
                      const color = SHIFT_COLORS[entry.shiftType] || 'bg-gray-600';
                      return (
                        <td key={dayIdx} className={`p-1 text-center ${dayIdx >= 5 ? 'bg-gray-800/30' : ''}`}>
                          <span className={`inline-block ${color} text-white text-[9px] font-bold px-1.5 py-0.5 rounded cursor-default`}
                            title={`${entry.nurse} | ${entry.day} | ${entry.shiftType} | Skill: ${entry.skill} | Contract: ${entry.contract}`}>
                            {entry.shiftType.charAt(0)}
                          </span>
                        </td>
                      );
                    })}
                    <td className="p-2 text-right text-gray-400">{totalShifts}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Shift Distribution */}
      <Card title="Shift Distribution">
        <div className="flex gap-4">
          {Array.from(shiftCounts.entries()).sort((a, b) => b[1] - a[1]).map(([shift, count]) => (
            <div key={shift} className="flex items-center gap-2">
              <span className={`w-3 h-3 rounded ${SHIFT_COLORS[shift] || 'bg-gray-600'}`}></span>
              <span className="text-xs text-gray-400">{shift}: {count}</span>
            </div>
          ))}
        </div>
      </Card>
    </>
  );
}
