'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { DiscoveryRecord, ImprovementEvent, WorkerLifecycle, PlateauEvent, TreeNode } from '@/lib/types';

type EventCategory = 'global_best' | 'local_best' | 'worker_start' | 'worker_end' | 'plateau' | 'pruning';

interface CausalEvent {
  timeMs: number;
  category: EventCategory;
  label: string;
  detail: Record<string, string | number>;
  week: number;
  icon: string;
  color: string;
}

interface Props {
  discoveries: DiscoveryRecord[];
  improvements: ImprovementEvent[];
  workers: WorkerLifecycle[];
  plateaus: PlateauEvent[];
  tree: TreeNode[];
}

const CATEGORY_CONFIG: Record<EventCategory, { icon: string; color: string; label: string }> = {
  global_best: { icon: '🏆', color: '#fbbf24', label: 'Global Best' },
  local_best: { icon: '💡', color: '#10b981', label: 'Local Best' },
  worker_start: { icon: '🚀', color: '#3b82f6', label: 'Worker Start' },
  worker_end: { icon: '🏁', color: '#6b7280', label: 'Worker End' },
  plateau: { icon: '🏔️', color: '#f97316', label: 'Plateau' },
  pruning: { icon: '✂️', color: '#ef4444', label: 'Beam Pruning' },
};

function buildAllEvents(props: Props): CausalEvent[] {
  const events: CausalEvent[] = [];

  for (const imp of props.improvements) {
    events.push({
      timeMs: imp.elapsedMs, category: 'global_best',
      label: `Global Best → ${imp.newGlobalBest} (−${imp.improvement})`,
      detail: { worker: imp.workerID, oldBest: imp.oldGlobalBest, newBest: imp.newGlobalBest, improvement: imp.improvement, week: imp.week, temperature: imp.temperatureAtEvent.toFixed(4) },
      week: imp.week, icon: '🏆', color: '#fbbf24',
    });
  }

  for (const d of props.discoveries) {
    if (d.eventType === 'local_best') {
      events.push({
        timeMs: d.elapsedMs, category: 'local_best',
        label: `Local best ${d.newBest} (W${d.workerID})`,
        detail: { worker: d.workerID, penalty: d.newBest, improvement: d.improvement, week: d.week, beamPath: d.beamPath, depth: d.branchDepth },
        week: d.week, icon: '💡', color: '#10b981',
      });
    }
  }

  for (const w of props.workers) {
    events.push({
      timeMs: w.startTimeMs, category: 'worker_start',
      label: `Worker ${w.workerID} started (depth ${w.depth})`,
      detail: { worker: w.workerID, depth: w.depth, parent: w.parentWorkerID, week: w.week, seed: w.seed },
      week: w.week, icon: '🚀', color: '#3b82f6',
    });
    events.push({
      timeMs: w.finishTimeMs, category: 'worker_end',
      label: `Worker ${w.workerID} finished (best ${w.bestPenalty})`,
      detail: { worker: w.workerID, bestPenalty: w.bestPenalty, branches: w.branchCount, global: w.producedGlobalBest ? 'yes' : 'no', week: w.week },
      week: w.week, icon: '🏁', color: '#6b7280',
    });
  }

  // Plateaus with estimated time.
  const workerMap = new Map(props.workers.map(w => [`${w.week}-${w.workerID}`, w]));
  for (const p of props.plateaus) {
    const worker = workerMap.get(`${p.week}-${p.workerID}`);
    const timeEst = worker ? worker.startTimeMs + (p.candidate / Math.max(worker.finishCandidate, 1)) * (worker.finishTimeMs - worker.startTimeMs) : 0;
    events.push({
      timeMs: timeEst, category: 'plateau',
      label: `Plateau W${p.workerID} (${p.candsSinceImprove} stale)`,
      detail: { worker: p.workerID, staleFor: p.candsSinceImprove, penalty: p.currentPenalty, globalBest: p.globalBest, temperature: p.temperature.toFixed(4), week: p.week, depth: p.depth },
      week: p.week, icon: '🏔️', color: '#f97316',
    });
  }

  // Pruning events (nodes not retained).
  const pruned = props.tree.filter(t => !t.retained && t.week > 1);
  for (const t of pruned) {
    // Estimate time from the week's worker data.
    const weekWorkers = props.workers.filter(w => w.week === t.week);
    const avgFinish = weekWorkers.length > 0 ? weekWorkers.reduce((s, w) => s + w.finishTimeMs, 0) / weekWorkers.length : 0;
    events.push({
      timeMs: avgFinish, category: 'pruning',
      label: `Path ${t.pathID} pruned (pen=${t.cumulativePenalty})`,
      detail: { pathID: t.pathID, week: t.week, penalty: t.cumulativePenalty, parent: t.parentID },
      week: t.week, icon: '✂️', color: '#ef4444',
    });
  }

  events.sort((a, b) => a.timeMs - b.timeMs);
  return events;
}

export default function CausalityExplorer({ discoveries, improvements, workers, plateaus, tree }: Props) {
  const allEvents = useMemo(
    () => buildAllEvents({ discoveries, improvements, workers, plateaus, tree }),
    [discoveries, improvements, workers, plateaus, tree]
  );

  const [selected, setSelected] = useState<CausalEvent | null>(null);
  const [windowMs, setWindowMs] = useState(5000); // 5 second default window
  const [filterCategories, setFilterCategories] = useState<Set<EventCategory>>(
    new Set(['global_best', 'local_best', 'worker_start', 'worker_end', 'plateau', 'pruning'])
  );

  // Find correlated events within the time window of selected event.
  const correlated = useMemo(() => {
    if (!selected) return [];
    const tMin = selected.timeMs - windowMs;
    const tMax = selected.timeMs + windowMs;
    return allEvents.filter(e =>
      e !== selected &&
      e.timeMs >= tMin && e.timeMs <= tMax &&
      filterCategories.has(e.category)
    );
  }, [selected, windowMs, allEvents, filterCategories]);

  // Group correlated by category.
  const correlatedGroups = useMemo(() => {
    const groups = new Map<EventCategory, CausalEvent[]>();
    for (const e of correlated) {
      const existing = groups.get(e.category) || [];
      existing.push(e);
      groups.set(e.category, existing);
    }
    return groups;
  }, [correlated]);

  // Before/after split.
  const before = correlated.filter(e => e.timeMs < selected!?.timeMs);
  const after = correlated.filter(e => e.timeMs >= (selected?.timeMs ?? 0));

  function toggleCategory(cat: EventCategory) {
    setFilterCategories(prev => {
      const next = new Set(prev);
      if (next.has(cat)) next.delete(cat); else next.add(cat);
      return next;
    });
  }

  function formatMs(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  // Pick interesting events to display as seed selections.
  const seedEvents = useMemo(() => {
    const seeds: CausalEvent[] = [];
    // First global best.
    const firstGlobal = allEvents.find(e => e.category === 'global_best');
    if (firstGlobal) seeds.push(firstGlobal);
    // A plateau.
    const bigPlateau = allEvents.filter(e => e.category === 'plateau').sort((a, b) => (b.detail.staleFor as number || 0) - (a.detail.staleFor as number || 0))[0];
    if (bigPlateau) seeds.push(bigPlateau);
    // A pruning event.
    const prune = allEvents.find(e => e.category === 'pruning');
    if (prune) seeds.push(prune);
    return seeds;
  }, [allEvents]);

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card title="Causality Explorer">
        <div className="flex flex-wrap gap-3 mb-3">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Window:</span>
            {[1000, 2000, 5000, 10000, 30000].map(ms => (
              <button key={ms} onClick={() => setWindowMs(ms)}
                className={`px-2 py-0.5 rounded text-[10px] ${windowMs === ms ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                {ms >= 1000 ? `${ms/1000}s` : `${ms}ms`}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Show:</span>
            {(Object.entries(CATEGORY_CONFIG) as [EventCategory, typeof CATEGORY_CONFIG[EventCategory]][]).map(([cat, cfg]) => (
              <button key={cat} onClick={() => toggleCategory(cat)}
                className={`px-2 py-0.5 rounded text-[10px] ${filterCategories.has(cat) ? 'bg-gray-700 text-white' : 'bg-gray-900 text-gray-600'}`}>
                {cfg.icon} {cfg.label}
              </button>
            ))}
          </div>
        </div>
        <p className="text-[9px] text-gray-600">
          Select an event below to see correlated events within ±{formatMs(windowMs)}.
          {allEvents.length > 0 && ` (${allEvents.length.toLocaleString()} total events)`}
        </p>
      </Card>

      {/* Seed events: quick-select interesting events */}
      <Card title="Investigate">
        <p className="text-[10px] text-gray-500 mb-2">Click an event to explore its causal context:</p>
        <div className="space-y-1">
          {seedEvents.map((e, i) => (
            <button key={i} onClick={() => setSelected(e)}
              className={`w-full text-left px-3 py-2 rounded border transition ${
                selected === e ? 'border-blue-500 bg-blue-900/20' : 'border-gray-800 hover:border-gray-600 bg-gray-800/30'
              }`}>
              <span className="mr-2">{e.icon}</span>
              <span className="text-xs" style={{ color: e.color }}>{e.label}</span>
              <span className="text-[9px] text-gray-600 ml-2">at {formatMs(e.timeMs)}</span>
            </button>
          ))}
        </div>
        {/* Recent events list for browsing */}
        <div className="mt-3 max-h-48 overflow-y-auto space-y-0.5">
          {allEvents.filter(e => filterCategories.has(e.category)).slice(0, 50).map((e, i) => (
            <button key={i} onClick={() => setSelected(e)}
              className={`w-full text-left px-2 py-0.5 rounded text-[10px] transition ${
                selected === e ? 'bg-gray-700' : 'hover:bg-gray-800'
              }`}>
              <span className="text-gray-600 w-14 inline-block">{formatMs(e.timeMs)}</span>
              <span className="mr-1">{e.icon}</span>
              <span style={{ color: e.color }}>{e.label}</span>
            </button>
          ))}
        </div>
      </Card>

      {/* Selected event + correlation panel */}
      {selected && (
        <>
          <Card title={`${selected.icon} Selected: ${selected.label}`}>
            <div className="grid grid-cols-3 gap-3 text-xs mb-3">
              <div><p className="text-gray-500">Time</p><p className="font-mono">{formatMs(selected.timeMs)}</p></div>
              <div><p className="text-gray-500">Week</p><p>{selected.week}</p></div>
              <div><p className="text-gray-500">Category</p><p style={{ color: selected.color }}>{selected.category.replace('_', ' ')}</p></div>
              {Object.entries(selected.detail).map(([k, v]) => (
                <div key={k}><p className="text-gray-500">{k}</p><p className="font-mono">{String(v)}</p></div>
              ))}
            </div>
          </Card>

          {/* Correlation results */}
          <Card title={`Correlated Events (±${formatMs(windowMs)})`}>
            <p className="text-[10px] text-gray-500 mb-2">
              {correlated.length} events found within {formatMs(windowMs)} of selection.
            </p>

            {/* Category breakdown */}
            <div className="flex gap-2 mb-3 flex-wrap">
              {Array.from(correlatedGroups.entries()).map(([cat, events]) => (
                <span key={cat} className="px-2 py-1 rounded bg-gray-800 text-[10px]" style={{ color: CATEGORY_CONFIG[cat].color }}>
                  {CATEGORY_CONFIG[cat].icon} {events.length} {CATEGORY_CONFIG[cat].label}
                </span>
              ))}
            </div>

            {/* Before/After split */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <p className="text-[10px] text-gray-500 uppercase mb-1">Before ({before.length})</p>
                <div className="space-y-0.5 max-h-40 overflow-y-auto">
                  {before.slice(-10).map((e, i) => (
                    <div key={i} className="text-[10px] flex items-center gap-1 px-1 py-0.5 rounded hover:bg-gray-800 cursor-pointer"
                      onClick={() => setSelected(e)}>
                      <span className="text-gray-600 w-12">{formatMs(selected.timeMs - e.timeMs)} ago</span>
                      <span>{e.icon}</span>
                      <span className="truncate" style={{ color: e.color }}>{e.label}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <p className="text-[10px] text-gray-500 uppercase mb-1">After ({after.length})</p>
                <div className="space-y-0.5 max-h-40 overflow-y-auto">
                  {after.slice(0, 10).map((e, i) => (
                    <div key={i} className="text-[10px] flex items-center gap-1 px-1 py-0.5 rounded hover:bg-gray-800 cursor-pointer"
                      onClick={() => setSelected(e)}>
                      <span className="text-gray-600 w-12">+{formatMs(e.timeMs - selected.timeMs)}</span>
                      <span>{e.icon}</span>
                      <span className="truncate" style={{ color: e.color }}>{e.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </Card>

          {/* Causal narrative */}
          <Card title="Causal Interpretation">
            <div className="text-sm text-gray-300 space-y-1">
              {generateCausalNarrative(selected, before, after, correlatedGroups)}
            </div>
          </Card>
        </>
      )}
    </div>
  );
}

function generateCausalNarrative(
  selected: CausalEvent,
  before: CausalEvent[],
  after: CausalEvent[],
  groups: Map<EventCategory, CausalEvent[]>
): React.ReactElement[] {
  const lines: string[] = [];

  if (selected.category === 'global_best') {
    const workerStarts = before.filter(e => e.category === 'worker_start');
    if (workerStarts.length > 0) {
      lines.push(`${workerStarts.length} workers were active before this discovery.`);
    }
    const plateausBefore = before.filter(e => e.category === 'plateau');
    if (plateausBefore.length > 0) {
      lines.push(`${plateausBefore.length} plateaus preceded this improvement — the solver may have been stuck then escaped.`);
    }
    const branchesAfter = after.filter(e => e.category === 'worker_start');
    if (branchesAfter.length > 0) {
      lines.push(`${branchesAfter.length} new workers spawned after this — likely branching from the improvement.`);
    }
  }

  if (selected.category === 'plateau') {
    const discBefore = before.filter(e => e.category === 'local_best' || e.category === 'global_best');
    const discAfter = after.filter(e => e.category === 'local_best' || e.category === 'global_best');
    if (discBefore.length === 0) {
      lines.push('No discoveries preceded this plateau — the solver entered stagnation without recent progress.');
    } else {
      lines.push(`${discBefore.length} discoveries preceded this plateau.`);
    }
    if (discAfter.length > 0) {
      lines.push(`${discAfter.length} discoveries followed — the plateau was eventually escaped.`);
    } else {
      lines.push('No discoveries followed within the window — the plateau may not have been escaped.');
    }
  }

  if (selected.category === 'pruning') {
    const globalAfter = after.filter(e => e.category === 'global_best');
    const pruningNearby = (groups.get('pruning') || []).length;
    if (pruningNearby > 1) {
      lines.push(`${pruningNearby} paths were pruned in this window — beam selection was active.`);
    }
    if (globalAfter.length > 0) {
      lines.push(`A global best was found after pruning — diversity reduction may have focused the search.`);
    }
  }

  if (selected.category === 'worker_end') {
    const globalNearby = (groups.get('global_best') || []).length;
    if (globalNearby > 0) {
      lines.push(`A global best was found near this worker's end — it may have contributed.`);
    }
  }

  if (lines.length === 0) {
    lines.push('Select different events or adjust the time window to explore causal relationships.');
  }

  return lines.map((line, i) => <p key={i}>{line}</p>);
}
