'use client';

import { useState, useMemo } from 'react';
import Link from 'next/link';
import Card from '@/components/Card';
import DeleteRunButton from './DeleteRunButton';
import { RunMetadata } from '@/lib/types';

interface RunListEntry {
  id: string;
  metadata: RunMetadata | null;
}

type FilterType = 'all' | 'nrp' | 'cvrp' | 'jss' | 'ilp';
type SortBy = 'name' | 'date' | 'type';

export default function RunList({ runs }: { runs: RunListEntry[] }) {
  const [filter, setFilter] = useState<FilterType>('all');
  const [sortBy, setSortBy] = useState<SortBy>('date');
  const [expanded, setExpanded] = useState(false);

  const enrichedRuns = useMemo(() => {
    return runs.map(run => {
      const meta = run.metadata as unknown as Record<string, unknown> | null;
      const problemType = String(meta?.problemType || 'nrp');
      const mode = String(meta?.mode || run.metadata?.mode || '—');
      const instance = String(meta?.instance || run.metadata?.instance || '—');
      // Try to extract timestamp from manifest or run label.
      const timestamp = meta?.timestamp as string || meta?.createdAt as string || '';

      return { ...run, problemType, mode, instance, timestamp };
    });
  }, [runs]);

  const filtered = useMemo(() => {
    let result = enrichedRuns;
    if (filter !== 'all') {
      result = result.filter(r => {
        if (filter === 'ilp') return r.mode === 'ilp';
        if (filter === 'jss') return r.problemType === 'jss' && r.mode !== 'ilp';
        if (filter === 'cvrp') return r.problemType === 'cvrp' && r.mode !== 'ilp';
        return r.problemType === 'nrp' && r.mode !== 'ilp';
      });
    }
    return result;
  }, [enrichedRuns, filter]);

  const sorted = useMemo(() => {
    const copy = [...filtered];
    switch (sortBy) {
      case 'name':
        return copy.sort((a, b) => a.id.localeCompare(b.id));
      case 'type':
        return copy.sort((a, b) => a.problemType.localeCompare(b.problemType) || a.id.localeCompare(b.id));
      case 'date':
        // Sort by timestamp if available, otherwise by name (newest-looking labels first).
        return copy.sort((a, b) => {
          if (a.timestamp && b.timestamp) return b.timestamp.localeCompare(a.timestamp);
          return b.id.localeCompare(a.id);
        });
      default:
        return copy;
    }
  }, [filtered, sortBy]);

  // Count per type.
  const counts = useMemo(() => {
    const c = { all: enrichedRuns.length, nrp: 0, cvrp: 0, jss: 0, ilp: 0 };
    for (const r of enrichedRuns) {
      if (r.mode === 'ilp') c.ilp++;
      else if (r.problemType === 'cvrp') c.cvrp++;
      else if (r.problemType === 'jss') c.jss++;
      else c.nrp++;
    }
    return c;
  }, [enrichedRuns]);

  return (
    <Card title={`Saved Runs (${sorted.length}${filter !== 'all' ? ` of ${runs.length}` : ''})`}>
      {/* Filter + Sort controls */}
      <div className="flex flex-wrap items-center gap-2 mb-3">
        <span className="text-[9px] text-gray-500 uppercase">Filter:</span>
        {(['all', 'nrp', 'cvrp', 'jss', 'ilp'] as FilterType[]).map(f => (
          <button key={f} onClick={() => setFilter(f)}
            className={`px-2 py-0.5 rounded text-[10px] font-semibold transition-colors ${
              filter === f
                ? f === 'nrp' ? 'bg-purple-900 text-purple-400'
                  : f === 'cvrp' ? 'bg-emerald-900 text-emerald-400'
                  : f === 'jss' ? 'bg-amber-900 text-amber-400'
                  : f === 'ilp' ? 'bg-blue-900 text-blue-400'
                  : 'bg-gray-700 text-white'
                : 'bg-gray-800 text-gray-500 hover:text-gray-300'
            }`}>
            {f.toUpperCase()} ({counts[f]})
          </button>
        ))}

        <span className="text-gray-700 mx-1">|</span>
        <span className="text-[9px] text-gray-500 uppercase">Sort:</span>
        {(['date', 'name', 'type'] as SortBy[]).map(s => (
          <button key={s} onClick={() => setSortBy(s)}
            className={`px-2 py-0.5 rounded text-[10px] ${
              sortBy === s ? 'bg-gray-700 text-white' : 'bg-gray-800 text-gray-500 hover:text-gray-300'
            }`}>
            {s}
          </button>
        ))}
      </div>

      {/* Run list */}
      <div className="grid gap-2">
        {(expanded ? sorted : sorted.slice(0, 10)).map(run => {
          const badge = run.mode === 'ilp' ? 'bg-blue-900 text-blue-400'
            : run.problemType === 'cvrp' ? 'bg-emerald-900 text-emerald-400'
            : run.problemType === 'jss' ? 'bg-amber-900 text-amber-400'
            : 'bg-purple-900 text-purple-400';
          const badgeLabel = run.mode === 'ilp' ? 'ILP' : run.problemType.toUpperCase();

          return (
            <div key={run.id} className="flex items-center bg-gray-800 border border-gray-700 hover:border-blue-500 rounded-lg transition-colors">
              <Link href={`/runs/${run.id}/summary`} className="flex-1 p-3">
                <div className="flex items-center gap-3">
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-semibold ${badge}`}>{badgeLabel}</span>
                  <div className="flex-1 min-w-0">
                    <h3 className="text-sm font-semibold text-blue-400 truncate">{run.id}</h3>
                    <p className="text-[10px] text-gray-500 truncate">
                      {run.mode.toUpperCase()} · {run.instance}
                      {run.metadata?.beamWidth ? ` · Beam ${run.metadata.beamWidth}` : ''}
                      {run.metadata?.iterationsPerWorker ? ` · ${(run.metadata.iterationsPerWorker / 1000).toFixed(0)}K iter` : ''}
                    </p>
                  </div>
                  {run.timestamp && (
                    <span className="text-[9px] text-gray-600 whitespace-nowrap">
                      {formatDate(run.timestamp)}
                    </span>
                  )}
                  <span className="text-gray-600 text-sm">→</span>
                </div>
              </Link>
              <div className="pr-3">
                <DeleteRunButton runId={run.id} />
              </div>
            </div>
          );
        })}
      </div>

      {sorted.length > 10 && (
        <div className="text-center mt-3 pt-2 border-t border-gray-800">
          <p className="text-[9px] text-gray-500 mb-1">
            Showing {expanded ? `all ${sorted.length}` : `10 of ${sorted.length}`} runs
          </p>
          <button
            onClick={() => setExpanded(!expanded)}
            className="text-[10px] text-blue-400 hover:text-blue-300 bg-gray-800 hover:bg-gray-700 px-3 py-1 rounded transition-colors"
          >
            {expanded ? 'Show fewer' : `Show all ${sorted.length}`}
          </button>
        </div>
      )}

      {sorted.length === 0 && (
        <p className="text-center text-gray-500 text-xs py-4">No runs match the current filter.</p>
      )}
    </Card>
  );
}

function formatDate(timestamp: string): string {
  if (!timestamp) return '';
  try {
    const d = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    if (diffDays === 0) return 'today';
    if (diffDays === 1) return 'yesterday';
    if (diffDays < 7) return `${diffDays}d ago`;
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
  } catch {
    return '';
  }
}
