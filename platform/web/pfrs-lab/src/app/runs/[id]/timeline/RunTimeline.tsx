'use client';

import { useState, useMemo, useRef } from 'react';
import Card from '@/components/Card';
import { DiscoveryRecord, ImprovementEvent, WorkerLifecycle, PlateauEvent, WeekRecord } from '@/lib/types';

type EventType = 'global_best' | 'local_best' | 'worker_start' | 'worker_end' | 'plateau' | 'week_start';
type Lane = 'global' | 'discovery' | 'workers' | 'plateaus' | 'weeks';

interface TimelineEvent {
  timeMs: number;
  type: EventType;
  lane: Lane;
  label: string;
  detail: Record<string, string | number>;
  week: number;
  color: string;
  icon: string;
}

interface Props {
  discoveries: DiscoveryRecord[];
  improvements: ImprovementEvent[];
  workers: WorkerLifecycle[];
  plateaus: PlateauEvent[];
  weeks: WeekRecord[];
}

const LANE_CONFIG: Record<Lane, { label: string; y: number }> = {
  weeks: { label: 'Weeks', y: 0 },
  global: { label: 'Global Best', y: 1 },
  discovery: { label: 'Discoveries', y: 2 },
  workers: { label: 'Workers', y: 3 },
  plateaus: { label: 'Plateaus', y: 4 },
};

const TYPE_CONFIG: Record<EventType, { color: string; icon: string }> = {
  global_best: { color: '#fbbf24', icon: '🏆' },
  local_best: { color: '#10b981', icon: '💡' },
  worker_start: { color: '#3b82f6', icon: '🚀' },
  worker_end: { color: '#6b7280', icon: '🏁' },
  plateau: { color: '#f97316', icon: '🏔️' },
  week_start: { color: '#a855f7', icon: '📅' },
};

function buildTimeline(props: Props): TimelineEvent[] {
  const events: TimelineEvent[] = [];

  // Week transitions (derived from workers start times per week).
  const weekStarts = new Map<number, number>();
  for (const w of props.workers) {
    const existing = weekStarts.get(w.week) ?? Infinity;
    if (w.startTimeMs < existing) weekStarts.set(w.week, w.startTimeMs);
  }
  for (const [week, time] of weekStarts) {
    events.push({
      timeMs: time, type: 'week_start', lane: 'weeks',
      label: `Week ${week} begins`, week,
      color: TYPE_CONFIG.week_start.color, icon: TYPE_CONFIG.week_start.icon,
      detail: { week, time: `${time}ms` },
    });
  }

  // Global best improvements.
  for (const imp of props.improvements) {
    events.push({
      timeMs: imp.elapsedMs, type: 'global_best', lane: 'global',
      label: `Global Best → ${imp.newGlobalBest}`,
      week: imp.week,
      color: TYPE_CONFIG.global_best.color, icon: TYPE_CONFIG.global_best.icon,
      detail: {
        worker: imp.workerID, oldBest: imp.oldGlobalBest,
        newBest: imp.newGlobalBest, improvement: imp.improvement,
        candidate: imp.candidate, temperature: imp.temperatureAtEvent.toFixed(4),
      },
    });
  }

  // Local best discoveries.
  for (const d of props.discoveries) {
    if (d.eventType !== 'local_best') continue;
    events.push({
      timeMs: d.elapsedMs, type: 'local_best', lane: 'discovery',
      label: `Local best ${d.newBest} (W${d.workerID})`,
      week: d.week,
      color: TYPE_CONFIG.local_best.color, icon: TYPE_CONFIG.local_best.icon,
      detail: {
        worker: d.workerID, beamPath: d.beamPath, penalty: d.newBest,
        improvement: d.improvement, candidate: d.candidate,
        depth: d.branchDepth, seed: d.seedUsed,
      },
    });
  }

  // Worker starts/ends.
  for (const w of props.workers) {
    events.push({
      timeMs: w.startTimeMs, type: 'worker_start', lane: 'workers',
      label: `Worker ${w.workerID} started`,
      week: w.week,
      color: TYPE_CONFIG.worker_start.color, icon: TYPE_CONFIG.worker_start.icon,
      detail: { worker: w.workerID, depth: w.depth, parent: w.parentWorkerID, seed: w.seed },
    });
    events.push({
      timeMs: w.finishTimeMs, type: 'worker_end', lane: 'workers',
      label: `Worker ${w.workerID} finished (best ${w.bestPenalty})`,
      week: w.week,
      color: TYPE_CONFIG.worker_end.color, icon: TYPE_CONFIG.worker_end.icon,
      detail: {
        worker: w.workerID, bestPenalty: w.bestPenalty, finalPenalty: w.finalPenalty,
        candidates: w.finishCandidate, branches: w.branchCount,
        producedGlobal: w.producedGlobalBest ? 'yes' : 'no',
      },
    });
  }

  // Plateaus (sample — take first occurrence per worker/week).
  const seenPlateaus = new Set<string>();
  for (const p of props.plateaus) {
    const key = `${p.week}-${p.workerID}`;
    if (seenPlateaus.has(key)) continue;
    seenPlateaus.add(key);
    // Estimate time from candidate number (rough).
    const worker = props.workers.find(w => w.workerID === p.workerID && w.week === p.week);
    const timeEstimate = worker ? worker.startTimeMs + (p.candidate / Math.max(worker.finishCandidate, 1)) * (worker.finishTimeMs - worker.startTimeMs) : 0;
    events.push({
      timeMs: timeEstimate, type: 'plateau', lane: 'plateaus',
      label: `Plateau W${p.workerID} (${p.candsSinceImprove} stale)`,
      week: p.week,
      color: TYPE_CONFIG.plateau.color, icon: TYPE_CONFIG.plateau.icon,
      detail: {
        worker: p.workerID, candidate: p.candidate, staleFor: p.candsSinceImprove,
        currentPenalty: p.currentPenalty, globalBest: p.globalBest,
        temperature: p.temperature.toFixed(4),
      },
    });
  }

  events.sort((a, b) => a.timeMs - b.timeMs);
  return events;
}

export default function RunTimeline({ discoveries, improvements, workers, plateaus, weeks }: Props) {
  const allEvents = useMemo(
    () => buildTimeline({ discoveries, improvements, workers, plateaus, weeks }),
    [discoveries, improvements, workers, plateaus, weeks]
  );

  const [filters, setFilters] = useState<Set<EventType>>(new Set([
    'global_best', 'local_best', 'worker_start', 'worker_end', 'plateau', 'week_start'
  ]));
  const [selected, setSelected] = useState<TimelineEvent | null>(null);
  const [viewStart, setViewStart] = useState(0); // fraction 0-1
  const [viewEnd, setViewEnd] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(
    () => allEvents.filter(e => filters.has(e.type)),
    [allEvents, filters]
  );

  const totalTimeMs = allEvents.length > 0 ? allEvents[allEvents.length - 1].timeMs : 1;
  const viewMinMs = totalTimeMs * viewStart;
  const viewMaxMs = totalTimeMs * viewEnd;

  const visible = useMemo(
    () => filtered.filter(e => e.timeMs >= viewMinMs && e.timeMs <= viewMaxMs),
    [filtered, viewMinMs, viewMaxMs]
  );

  function toggleFilter(type: EventType) {
    setFilters(prev => {
      const next = new Set(prev);
      if (next.has(type)) next.delete(type); else next.add(type);
      return next;
    });
  }

  function zoomIn() {
    const range = viewEnd - viewStart;
    const mid = (viewStart + viewEnd) / 2;
    setViewStart(Math.max(0, mid - range * 0.3));
    setViewEnd(Math.min(1, mid + range * 0.3));
  }

  function zoomOut() {
    const range = viewEnd - viewStart;
    const mid = (viewStart + viewEnd) / 2;
    setViewStart(Math.max(0, mid - range * 0.8));
    setViewEnd(Math.min(1, mid + range * 0.8));
  }

  function resetZoom() {
    setViewStart(0);
    setViewEnd(1);
  }

  function formatMs(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    const s = ms / 1000;
    if (s < 60) return `${s.toFixed(1)}s`;
    return `${Math.floor(s / 60)}m${Math.round(s % 60)}s`;
  }

  const LANE_HEIGHT = 36;
  const lanes = Object.keys(LANE_CONFIG) as Lane[];
  const SVG_HEIGHT = lanes.length * LANE_HEIGHT + 40;
  const SVG_WIDTH = 1100;
  const PADDING_L = 90;
  const PADDING_R = 20;

  function timeToX(ms: number): number {
    const range = viewMaxMs - viewMinMs || 1;
    return PADDING_L + ((ms - viewMinMs) / range) * (SVG_WIDTH - PADDING_L - PADDING_R);
  }

  return (
    <div className="space-y-4">
      {/* Filters + controls */}
      <Card title="Run Timeline">
        <p className="text-xs text-gray-500 mb-3">
          All search events plotted on a timeline. 🏆 Global bests show when the overall best solution improved. 🚀/🏁 Worker starts/ends show parallelism. 🏔️ Plateaus indicate periods where workers stalled. A healthy run shows discoveries spread across time, not all clustered at the start.
        </p>
        <div className="flex flex-wrap gap-2 mb-3">
          {(Object.entries(TYPE_CONFIG) as [EventType, { color: string; icon: string }][]).map(([type, cfg]) => (
            <button key={type} onClick={() => toggleFilter(type)}
              className={`flex items-center gap-1 px-2 py-1 rounded text-[10px] border ${
                filters.has(type)
                  ? 'border-gray-500 bg-gray-800 text-white'
                  : 'border-gray-800 bg-gray-900 text-gray-600'
              }`}>
              <span>{cfg.icon}</span>
              <span>{type.replace('_', ' ')}</span>
            </button>
          ))}
          <div className="ml-auto flex gap-1">
            <button onClick={zoomIn} className="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">🔍+</button>
            <button onClick={zoomOut} className="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">🔍−</button>
            <button onClick={resetZoom} className="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">Reset</button>
          </div>
        </div>

        {/* Zoom range slider */}
        <div className="flex items-center gap-2 mb-3">
          <span className="text-[9px] text-gray-500">{formatMs(viewMinMs)}</span>
          <input type="range" min={0} max={100} value={viewStart * 100}
            onChange={e => setViewStart(Math.min(parseInt(e.target.value) / 100, viewEnd - 0.05))}
            className="flex-1 h-1 accent-blue-500" />
          <input type="range" min={0} max={100} value={viewEnd * 100}
            onChange={e => setViewEnd(Math.max(parseInt(e.target.value) / 100, viewStart + 0.05))}
            className="flex-1 h-1 accent-blue-500" />
          <span className="text-[9px] text-gray-500">{formatMs(viewMaxMs)}</span>
        </div>

        <p className="text-[9px] text-gray-600 mb-2">
          {visible.length.toLocaleString()} events in view ({filtered.length.toLocaleString()} total)
        </p>

        {/* SVG Timeline */}
        <div ref={containerRef} className="overflow-x-auto">
          <svg viewBox={`0 0 ${SVG_WIDTH} ${SVG_HEIGHT}`} className="w-full min-w-[800px]" style={{ height: `${SVG_HEIGHT}px` }}>
            {/* Lane backgrounds and labels */}
            {lanes.map((lane, i) => {
              const y = i * LANE_HEIGHT + 20;
              return (
                <g key={lane}>
                  <rect x={0} y={y} width={SVG_WIDTH} height={LANE_HEIGHT}
                    fill={i % 2 === 0 ? '#111827' : '#0f172a'} />
                  <text x={5} y={y + LANE_HEIGHT / 2} dominantBaseline="middle"
                    className="fill-gray-500 text-[9px]">{LANE_CONFIG[lane].label}</text>
                  <line x1={PADDING_L} y1={y + LANE_HEIGHT} x2={SVG_WIDTH - PADDING_R} y2={y + LANE_HEIGHT}
                    stroke="#1f2937" strokeWidth="0.5" />
                </g>
              );
            })}

            {/* Time axis ticks */}
            {[0, 0.2, 0.4, 0.6, 0.8, 1].map(frac => {
              const ms = viewMinMs + (viewMaxMs - viewMinMs) * frac;
              const x = timeToX(ms);
              return (
                <g key={frac}>
                  <line x1={x} y1={15} x2={x} y2={SVG_HEIGHT - 5}
                    stroke="#1f2937" strokeWidth="0.5" strokeDasharray="2,4" />
                  <text x={x} y={12} textAnchor="middle"
                    className="fill-gray-600 text-[7px]">{formatMs(ms)}</text>
                </g>
              );
            })}

            {/* Events */}
            {visible.slice(0, 2000).map((event, i) => {
              const x = timeToX(event.timeMs);
              const laneIdx = lanes.indexOf(event.lane);
              const y = laneIdx * LANE_HEIGHT + 20 + LANE_HEIGHT / 2;
              const isSelected = selected === event;

              return (
                <g key={i} className="cursor-pointer" onClick={() => setSelected(event)}>
                  <circle cx={x} cy={y}
                    r={isSelected ? 6 : event.type === 'global_best' ? 5 : 3}
                    fill={event.color}
                    opacity={isSelected ? 1 : 0.8}
                    stroke={isSelected ? '#fff' : 'none'}
                    strokeWidth={isSelected ? 1.5 : 0}
                  />
                  {/* Tooltip on hover via title */}
                  <title>{`${formatMs(event.timeMs)} — ${event.label}`}</title>
                </g>
              );
            })}
          </svg>
        </div>
        {visible.length > 2000 && (
          <p className="text-[9px] text-amber-500 mt-1">Showing first 2,000 of {visible.length} events. Zoom in for details.</p>
        )}
      </Card>

      {/* Selected event detail */}
      {selected && (
        <Card title={`${selected.icon} ${selected.label}`}>
          <div className="grid grid-cols-3 gap-3 text-xs">
            <div>
              <p className="text-gray-500">Time</p>
              <p className="font-mono">{formatMs(selected.timeMs)}</p>
            </div>
            <div>
              <p className="text-gray-500">Week</p>
              <p>{selected.week}</p>
            </div>
            <div>
              <p className="text-gray-500">Type</p>
              <p style={{ color: selected.color }}>{selected.type.replace('_', ' ')}</p>
            </div>
            {Object.entries(selected.detail).map(([key, val]) => (
              <div key={key}>
                <p className="text-gray-500">{key}</p>
                <p className="font-mono">{String(val)}</p>
              </div>
            ))}
          </div>
          <button onClick={() => setSelected(null)}
            className="mt-3 px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px] text-gray-400">
            Close
          </button>
        </Card>
      )}

      {/* Event feed (last 15 visible events) */}
      <Card title="Event Feed">
        <div className="space-y-0.5 max-h-64 overflow-y-auto font-mono text-[10px]">
          {visible.slice(-15).reverse().map((e, i) => (
            <div key={i}
              className={`flex items-center gap-2 py-0.5 px-2 rounded cursor-pointer hover:bg-gray-800 ${
                selected === e ? 'bg-gray-800' : ''
              }`}
              onClick={() => setSelected(e)}
            >
              <span className="text-gray-600 w-14 shrink-0">{formatMs(e.timeMs)}</span>
              <span>{e.icon}</span>
              <span className="truncate" style={{ color: e.color }}>{e.label}</span>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
