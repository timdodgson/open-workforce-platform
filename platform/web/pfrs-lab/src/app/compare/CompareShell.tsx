'use client';

import { useState, useMemo, useRef, useEffect } from 'react';
import Card from '@/components/Card';
import { RunInfo } from './page';

function formatRunLabel(r: RunInfo): string {
  const parts = [r.instance || r.id.split('-').slice(1, 3).join('-')];
  parts.push(r.mode);
  if (r.decisionMode && r.decisionMode !== 'off' && r.decisionMode !== '') {
    parts.push(r.decisionMode);
  }
  parts.push(r.objective.toLocaleString());
  return parts.join(' | ');
}

function objLabel(problemType: string): string {
  if (problemType === 'cvrp' || problemType === 'vrptw') return 'Distance';
  if (problemType === 'jss') return 'Makespan';
  return 'Penalty';
}

export default function CompareShell({ runs }: { runs: RunInfo[] }) {
  const [runAId, setRunAId] = useState(runs[0]?.id || '');
  const [runBId, setRunBId] = useState(runs[1]?.id || '');
  const [relevantOnly, setRelevantOnly] = useState(true);

  const runA = runs.find(r => r.id === runAId);
  const runB = runs.find(r => r.id === runBId);

  // Context-aware Run B options.
  const runBOptions = useMemo(() => {
    if (!runA) return runs;
    if (!relevantOnly) return runs.filter(r => r.id !== runAId);

    const sameInstance = runs.filter(r => r.id !== runAId && r.problemType === runA.problemType && r.instance === runA.instance);
    const sameDomain = runs.filter(r => r.id !== runAId && r.problemType === runA.problemType && r.instance !== runA.instance);
    const other = runs.filter(r => r.id !== runAId && r.problemType !== runA.problemType);

    return [
      ...sameInstance.map(r => ({ ...r, _group: 'Same instance' as const })),
      ...sameDomain.map(r => ({ ...r, _group: 'Same domain' as const })),
      ...other.map(r => ({ ...r, _group: 'Other domains' as const })),
    ];
  }, [runs, runA, runAId, relevantOnly]);

  // Auto-select Run B when Run A changes.
  useEffect(() => {
    if (runBOptions.length > 0 && !runBOptions.find(r => r.id === runBId)) {
      setRunBId(runBOptions[0].id);
    }
  }, [runBOptions, runBId]);

  return (
    <div className="space-y-4">
      <Card title="Head-to-Head Comparison">
        <div className="grid grid-cols-1 md:grid-cols-[1fr_auto_1fr] gap-4 items-start">
          {/* Run A */}
          <div>
            <label className="text-[10px] text-blue-400 uppercase block mb-1">Run A</label>
            <SearchableSelect
              options={runs}
              value={runAId}
              onChange={setRunAId}
              colour="blue"
            />
          </div>

          <span className="text-gray-600 text-lg self-center text-center hidden md:block mt-5">vs</span>

          {/* Run B */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-[10px] text-rose-400 uppercase">Run B</label>
              <button
                onClick={() => setRelevantOnly(!relevantOnly)}
                className={`text-[9px] px-2 py-0.5 rounded ${relevantOnly ? 'bg-blue-900/50 text-blue-300' : 'bg-gray-700 text-gray-400'}`}
              >
                {relevantOnly ? 'Relevant only' : 'Show all'}
              </button>
            </div>
            <SearchableSelect
              options={runBOptions}
              value={runBId}
              onChange={setRunBId}
              colour="rose"
              grouped={relevantOnly}
            />
            {relevantOnly && runBOptions.length === 0 && (
              <p className="text-[9px] text-gray-500 mt-1">
                No relevant comparisons found.{' '}
                <button onClick={() => setRelevantOnly(false)} className="text-blue-400 hover:underline">Show all</button>
              </p>
            )}
          </div>
        </div>
      </Card>

      {runA && runB && <ComparisonResult runA={runA} runB={runB} />}
    </div>
  );
}

// --- Searchable Select ---

interface SearchableSelectProps {
  options: (RunInfo & { _group?: string })[];
  value: string;
  onChange: (id: string) => void;
  colour: 'blue' | 'rose';
  grouped?: boolean;
}

function SearchableSelect({ options, value, onChange, colour, grouped }: SearchableSelectProps) {
  const [search, setSearch] = useState('');
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    if (!search) return options;
    const q = search.toLowerCase();
    return options.filter(r =>
      r.id.toLowerCase().includes(q) ||
      r.problemType.toLowerCase().includes(q) ||
      r.instance.toLowerCase().includes(q) ||
      r.mode.toLowerCase().includes(q)
    );
  }, [options, search]);

  // Close on outside click.
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const selected = options.find(r => r.id === value);
  const borderColour = colour === 'blue' ? 'border-blue-700 focus:border-blue-500' : 'border-rose-700 focus:border-rose-500';
  const textColour = colour === 'blue' ? 'text-blue-400' : 'text-rose-400';

  // Group options for display.
  const groups = useMemo(() => {
    if (!grouped) return [{ label: '', items: filtered }];
    const map = new Map<string, typeof filtered>();
    for (const r of filtered) {
      const g = (r as { _group?: string })._group || '';
      const arr = map.get(g) || [];
      arr.push(r);
      map.set(g, arr);
    }
    return Array.from(map.entries()).map(([label, items]) => ({ label, items }));
  }, [filtered, grouped]);

  return (
    <div ref={ref} className="relative">
      {/* Display selected */}
      <button
        onClick={() => setOpen(!open)}
        className={`w-full bg-gray-800 border ${borderColour} rounded px-2 py-1.5 text-left text-xs ${textColour} truncate`}
      >
        {selected ? formatRunLabel(selected) : 'Select a run...'}
      </button>

      {/* Dropdown */}
      {open && (
        <div className="absolute z-50 mt-1 w-full bg-gray-800 border border-gray-700 rounded shadow-lg max-h-64 overflow-hidden flex flex-col">
          {/* Search */}
          <input
            type="text"
            placeholder="Search runs..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="px-2 py-1.5 bg-gray-900 border-b border-gray-700 text-xs text-gray-200 outline-none"
            autoFocus
          />
          {/* Options */}
          <div className="overflow-y-auto flex-1">
            {groups.map(group => (
              <div key={group.label}>
                {group.label && (
                  <div className="px-2 py-1 text-[9px] text-gray-500 uppercase bg-gray-850 sticky top-0">{group.label}</div>
                )}
                {group.items.map(r => (
                  <button
                    key={r.id}
                    onClick={() => { onChange(r.id); setOpen(false); setSearch(''); }}
                    className={`w-full text-left px-2 py-1.5 text-[11px] hover:bg-gray-700 ${r.id === value ? 'bg-gray-700 text-white' : 'text-gray-300'}`}
                  >
                    <span className="text-gray-500 mr-1">{r.problemType.toUpperCase()}</span>
                    {formatRunLabel(r)}
                  </button>
                ))}
              </div>
            ))}
            {filtered.length === 0 && (
              <p className="px-2 py-3 text-[10px] text-gray-500 text-center">No matching runs</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// --- Comparison Result ---

function ComparisonResult({ runA, runB }: { runA: RunInfo; runB: RunInfo }) {
  const label = objLabel(runA.problemType);
  const diff = runA.objective - runB.objective;
  const diffPct = runB.objective > 0 ? ((diff / runB.objective) * 100).toFixed(1) : '0';
  const winner = diff < 0 ? 'A' : diff > 0 ? 'B' : 'Tie';
  const crossDomain = runA.problemType !== runB.problemType;

  return (
    <>
      {/* Cross-domain warning */}
      {crossDomain && (
        <Card title="⚠️ Cross-Domain Comparison">
          <p className="text-xs text-amber-400">
            Comparing {runA.problemType.toUpperCase()} ({objLabel(runA.problemType)}) with {runB.problemType.toUpperCase()} ({objLabel(runB.problemType)}).
            Objectives use different units and are not directly comparable.
          </p>
        </Card>
      )}

      {/* Conclusion */}
      <Card title="Conclusion">
        {winner === 'Tie' ? (
          <p className="text-sm text-gray-300">Both runs achieved identical objectives ({runA.objective.toLocaleString()}).</p>
        ) : (
          <p className="text-sm text-gray-300">
            <span className={winner === 'A' ? 'text-blue-400' : 'text-rose-400'}>Run {winner}</span>{' '}
            is better by {Math.abs(diff).toLocaleString()} {!crossDomain && `(${Math.abs(Number(diffPct))}%)`}.
          </p>
        )}
      </Card>

      {/* Evidence table */}
      <Card title="Comparison">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Metric</th>
              <th className="text-right p-2 text-blue-400">Run A</th>
              <th className="text-right p-2 text-rose-400">Run B</th>
              <th className="text-right p-2">Difference</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-t border-gray-800">
              <td className="p-2">{crossDomain ? 'Objective' : label}</td>
              <td className="text-right p-2 text-blue-400">{runA.objective.toLocaleString()}</td>
              <td className="text-right p-2 text-rose-400">{runB.objective.toLocaleString()}</td>
              <td className={`text-right p-2 ${diff < 0 ? 'text-emerald-400' : diff > 0 ? 'text-red-400' : 'text-gray-500'}`}>
                {diff === 0 ? '=' : `${diff > 0 ? '+' : ''}${diff.toLocaleString()}`}
              </td>
            </tr>
            <tr className="border-t border-gray-800">
              <td className="p-2">Domain</td>
              <td className="text-right p-2 text-blue-400">{runA.problemType.toUpperCase()}</td>
              <td className="text-right p-2 text-rose-400">{runB.problemType.toUpperCase()}</td>
              <td className="text-right p-2 text-gray-500">{crossDomain ? '⚠️ Different' : 'Same'}</td>
            </tr>
            <tr className="border-t border-gray-800">
              <td className="p-2">Instance</td>
              <td className="text-right p-2 text-blue-400">{runA.instance || '—'}</td>
              <td className="text-right p-2 text-rose-400">{runB.instance || '—'}</td>
              <td className="text-right p-2 text-gray-500">{runA.instance === runB.instance ? 'Same' : 'Different'}</td>
            </tr>
            <tr className="border-t border-gray-800">
              <td className="p-2">Algorithm</td>
              <td className="text-right p-2 text-blue-400">{runA.mode}</td>
              <td className="text-right p-2 text-rose-400">{runB.mode}</td>
              <td className="text-right p-2 text-gray-500">{runA.mode === runB.mode ? 'Same' : 'Different'}</td>
            </tr>
            {(runA.decisionMode || runB.decisionMode) && (
              <tr className="border-t border-gray-800">
                <td className="p-2">Decision Mode</td>
                <td className="text-right p-2 text-blue-400">{runA.decisionMode || 'off'}</td>
                <td className="text-right p-2 text-rose-400">{runB.decisionMode || 'off'}</td>
                <td className="text-right p-2 text-gray-500">{runA.decisionMode === runB.decisionMode ? 'Same' : 'Different'}</td>
              </tr>
            )}
          </tbody>
        </table>
      </Card>
    </>
  );
}
