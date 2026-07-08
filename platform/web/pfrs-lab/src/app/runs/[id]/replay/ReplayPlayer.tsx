'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import Card from '@/components/Card';
import { DiscoveryRecord, ImprovementEvent, TreeNode, WorkerLifecycle } from '@/lib/types';

interface ReplayEvent {
  timeMs: number;
  type: 'discovery' | 'global_best' | 'branch' | 'worker_start' | 'worker_end';
  week: number;
  workerID: number;
  penalty?: number;
  improvement?: number;
  pathID?: number;
  label: string;
}

interface Props {
  discoveries: DiscoveryRecord[];
  improvements: ImprovementEvent[];
  tree: TreeNode[];
  workers: WorkerLifecycle[];
}

function buildTimeline(props: Props): ReplayEvent[] {
  const events: ReplayEvent[] = [];

  // Discoveries — local best improvements.
  for (const d of props.discoveries) {
    if (d.eventType === 'local_best') {
      events.push({
        timeMs: d.elapsedMs,
        type: 'discovery',
        week: d.week,
        workerID: d.workerID,
        penalty: d.newBest,
        improvement: d.improvement,
        pathID: d.beamPath,
        label: `Worker ${d.workerID} local best ${d.newBest}`,
      });
    }
  }

  // Global best improvements.
  for (const imp of props.improvements) {
    events.push({
      timeMs: imp.elapsedMs,
      type: 'global_best',
      week: imp.week,
      workerID: imp.workerID,
      penalty: imp.newGlobalBest,
      improvement: imp.improvement,
      label: `🏆 New Global Best ${imp.newGlobalBest} (−${imp.improvement})`,
    });
  }

  // Worker starts and ends.
  for (const w of props.workers) {
    events.push({
      timeMs: w.startTimeMs,
      type: 'worker_start',
      week: w.week,
      workerID: w.workerID,
      label: `Worker ${w.workerID} started (depth ${w.depth})`,
    });
    events.push({
      timeMs: w.finishTimeMs,
      type: 'worker_end',
      week: w.week,
      workerID: w.workerID,
      penalty: w.bestPenalty,
      label: `Worker ${w.workerID} finished (best ${w.bestPenalty})`,
    });
  }

  // Sort by time.
  events.sort((a, b) => a.timeMs - b.timeMs);
  return events;
}

const SPEED_OPTIONS = [
  { label: '1x', multiplier: 1 },
  { label: '2x', multiplier: 2 },
  { label: '5x', multiplier: 5 },
  { label: '10x', multiplier: 10 },
  { label: '50x', multiplier: 50 },
  { label: 'Instant', multiplier: 10000 },
];

export default function ReplayPlayer({ discoveries, improvements, tree, workers }: Props) {
  const timeline = useRef<ReplayEvent[]>([]);
  const [events, setEvents] = useState<ReplayEvent[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);
  const [globalBest, setGlobalBest] = useState<number | null>(null);
  const [activeWorkers, setActiveWorkers] = useState<Set<number>>(new Set());
  const [currentWeek, setCurrentWeek] = useState(1);
  const [recentEvents, setRecentEvents] = useState<ReplayEvent[]>([]);
  const [totalDiscoveries, setTotalDiscoveries] = useState(0);
  const [totalBranches, setTotalBranches] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    timeline.current = buildTimeline({ discoveries, improvements, tree, workers });
    setEvents(timeline.current);
  }, [discoveries, improvements, tree, workers]);

  const advanceTo = useCallback((targetIndex: number) => {
    const evts = timeline.current;
    if (targetIndex < 0) targetIndex = 0;
    if (targetIndex >= evts.length) targetIndex = evts.length - 1;

    // Compute state at targetIndex by scanning from current position.
    let gb = globalBest;
    const aw = new Set(activeWorkers);
    let week = currentWeek;
    let disc = totalDiscoveries;
    let br = totalBranches;
    const recent: ReplayEvent[] = [];

    const start = targetIndex < currentIndex ? 0 : currentIndex;
    if (targetIndex < currentIndex) {
      // Reset state for backwards.
      gb = null;
      aw.clear();
      week = 1;
      disc = 0;
      br = 0;
    }

    for (let i = start; i <= targetIndex; i++) {
      const e = evts[i];
      if (e.type === 'global_best') {
        gb = e.penalty ?? gb;
      }
      if (e.type === 'worker_start') {
        aw.add(e.workerID);
      }
      if (e.type === 'worker_end') {
        aw.delete(e.workerID);
      }
      if (e.type === 'discovery') {
        disc++;
      }
      if (e.type === 'branch') {
        br++;
      }
      week = e.week;
    }

    // Keep last 8 events for feed.
    const feedStart = Math.max(0, targetIndex - 7);
    for (let i = feedStart; i <= targetIndex; i++) {
      recent.push(evts[i]);
    }

    setCurrentIndex(targetIndex);
    setGlobalBest(gb);
    setActiveWorkers(aw);
    setCurrentWeek(week);
    setTotalDiscoveries(disc);
    setTotalBranches(br);
    setRecentEvents(recent);
  }, [currentIndex, globalBest, activeWorkers, currentWeek, totalDiscoveries, totalBranches]);

  // Play/pause logic.
  useEffect(() => {
    if (playing && events.length > 0) {
      const evts = timeline.current;
      intervalRef.current = setInterval(() => {
        setCurrentIndex(prev => {
          const next = prev + 1;
          if (next >= evts.length) {
            setPlaying(false);
            return prev;
          }
          // Compute time delta to decide how many events to advance.
          advanceTo(next);
          return next;
        });
      }, Math.max(10, 100 / speed));
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [playing, speed, events.length, advanceTo]);

  const totalTime = events.length > 0 ? events[events.length - 1].timeMs : 0;
  const currentTime = events[currentIndex]?.timeMs ?? 0;
  const progress = events.length > 0 ? (currentIndex / (events.length - 1)) * 100 : 0;

  function formatTime(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    const s = ms / 1000;
    if (s < 60) return `${s.toFixed(1)}s`;
    const m = Math.floor(s / 60);
    const rem = (s % 60).toFixed(0);
    return `${m}m${rem}s`;
  }

  function eventIcon(type: string): string {
    switch (type) {
      case 'global_best': return '🏆';
      case 'discovery': return '💡';
      case 'worker_start': return '🚀';
      case 'worker_end': return '🏁';
      case 'branch': return '🌿';
      default: return '•';
    }
  }

  function eventColor(type: string): string {
    switch (type) {
      case 'global_best': return 'text-yellow-400';
      case 'discovery': return 'text-emerald-400';
      case 'worker_start': return 'text-blue-400';
      case 'worker_end': return 'text-gray-400';
      case 'branch': return 'text-purple-400';
      default: return 'text-gray-500';
    }
  }

  return (
    <div className="space-y-4">
      {/* Header stats */}
      <div className="grid grid-cols-5 gap-3">
        <Card title="Week">
          <p className="text-2xl font-bold text-blue-400">{currentWeek}</p>
        </Card>
        <Card title="Global Best">
          <p className="text-2xl font-bold text-emerald-400">
            {globalBest !== null ? globalBest.toLocaleString() : '—'}
          </p>
        </Card>
        <Card title="Active Workers">
          <p className="text-2xl font-bold text-orange-400">{activeWorkers.size}</p>
        </Card>
        <Card title="Discoveries">
          <p className="text-2xl font-bold text-purple-400">{totalDiscoveries}</p>
        </Card>
        <Card title="Time">
          <p className="text-2xl font-bold text-gray-300">{formatTime(currentTime)}</p>
        </Card>
      </div>

      {/* Controls */}
      <Card title="Replay Controls">
        <div className="space-y-3">
          {/* Timeline slider */}
          <div className="flex items-center gap-3">
            <span className="text-xs text-gray-500 w-16">{formatTime(currentTime)}</span>
            <input
              type="range"
              min={0}
              max={Math.max(events.length - 1, 1)}
              value={currentIndex}
              onChange={(e) => advanceTo(parseInt(e.target.value))}
              className="flex-1 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-500"
            />
            <span className="text-xs text-gray-500 w-16 text-right">{formatTime(totalTime)}</span>
          </div>

          {/* Buttons */}
          <div className="flex items-center gap-2">
            <button
              onClick={() => advanceTo(Math.max(0, currentIndex - 1))}
              className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 rounded text-sm"
              disabled={currentIndex <= 0}
            >
              ⏮ Step Back
            </button>
            <button
              onClick={() => setPlaying(!playing)}
              className={`px-4 py-1.5 rounded text-sm font-medium ${
                playing
                  ? 'bg-red-600 hover:bg-red-500 text-white'
                  : 'bg-emerald-600 hover:bg-emerald-500 text-white'
              }`}
            >
              {playing ? '⏸ Pause' : '▶ Play'}
            </button>
            <button
              onClick={() => advanceTo(Math.min(events.length - 1, currentIndex + 1))}
              className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 rounded text-sm"
              disabled={currentIndex >= events.length - 1}
            >
              Step Fwd ⏭
            </button>
            <button
              onClick={() => { setPlaying(false); advanceTo(0); }}
              className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 rounded text-sm"
            >
              ⏹ Reset
            </button>

            {/* Speed selector */}
            <div className="ml-4 flex items-center gap-1">
              <span className="text-xs text-gray-500 mr-1">Speed:</span>
              {SPEED_OPTIONS.map(opt => (
                <button
                  key={opt.label}
                  onClick={() => setSpeed(opt.multiplier)}
                  className={`px-2 py-1 rounded text-xs ${
                    speed === opt.multiplier
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>

          {/* Progress bar */}
          <div className="h-1 bg-gray-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all duration-75"
              style={{ width: `${progress}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-gray-600">
            <span>Event {currentIndex + 1} / {events.length}</span>
            <span>{progress.toFixed(1)}%</span>
          </div>
        </div>
      </Card>

      {/* Event Feed */}
      <Card title="Event Feed">
        <div className="space-y-1 max-h-64 overflow-y-auto font-mono text-xs">
          {recentEvents.length === 0 && (
            <p className="text-gray-600 italic">Press Play to start replay...</p>
          )}
          {recentEvents.map((e, i) => (
            <div
              key={`${e.timeMs}-${i}`}
              className={`flex items-center gap-2 py-0.5 px-2 rounded ${
                i === recentEvents.length - 1 ? 'bg-gray-800/50' : ''
              }`}
            >
              <span className="text-gray-600 w-14 shrink-0">{formatTime(e.timeMs)}</span>
              <span>{eventIcon(e.type)}</span>
              <span className={`${eventColor(e.type)} truncate`}>{e.label}</span>
              {e.type === 'global_best' && (
                <span className="ml-auto text-yellow-500 font-bold">
                  {e.penalty?.toLocaleString()}
                </span>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* Global Best Timeline (visual) */}
      <Card title="Global Best Over Time">
        <div className="h-32 flex items-end gap-px">
          {(() => {
            // Show global best milestones as bars.
            const gbEvents = events.filter(e => e.type === 'global_best');
            if (gbEvents.length === 0) return <p className="text-gray-600 text-xs">No improvements yet</p>;
            const maxPenalty = gbEvents[0]?.penalty ?? 1;
            const visibleEvents = gbEvents.filter((_, i) => i <= gbEvents.findIndex(e => e.timeMs >= currentTime));
            return visibleEvents.map((e, i) => {
              const height = e.penalty && maxPenalty ? ((e.penalty / maxPenalty) * 100) : 0;
              const isLatest = i === visibleEvents.length - 1;
              return (
                <div
                  key={i}
                  className={`flex-1 min-w-[3px] rounded-t transition-all duration-150 ${
                    isLatest ? 'bg-emerald-400' : 'bg-emerald-800'
                  }`}
                  style={{ height: `${Math.max(4, height)}%` }}
                  title={`${e.penalty} at ${formatTime(e.timeMs)}`}
                />
              );
            });
          })()}
        </div>
      </Card>

      {/* Beam Paths State */}
      <Card title="Beam Paths">
        <div className="grid grid-cols-8 gap-1">
          {tree.filter(t => t.week <= currentWeek).slice(-40).map(node => {
            const isActive = node.week === currentWeek;
            const isWinning = node.winning;
            return (
              <div
                key={node.pathID}
                className={`h-8 rounded text-[9px] flex items-center justify-center ${
                  isWinning ? 'bg-emerald-600 text-white' :
                  isActive ? 'bg-blue-800 text-blue-200' :
                  node.retained ? 'bg-gray-700 text-gray-400' :
                  'bg-gray-900 text-gray-700'
                }`}
                title={`Path ${node.pathID} W${node.week} pen=${node.cumulativePenalty}`}
              >
                {node.weekPenalty}
              </div>
            );
          })}
        </div>
      </Card>
    </div>
  );
}
