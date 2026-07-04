'use client';
import { useState, useMemo } from 'react';
import { ScatterChart, Scatter, XAxis, YAxis, ZAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import Card from '@/components/Card';
import { DiversityRecord, DiscoveryRecord, TreeNode } from '@/lib/types';

type ViewMode = 'scatter' | 'heatmap' | 'contour';
type ColourBy = 'penalty' | 'week' | 'family';

interface Props {
  diversity: DiversityRecord[];
  discoveries: DiscoveryRecord[];
  tree: TreeNode[];
}

export default function FitnessLandscape({ diversity, discoveries, tree }: Props) {
  const [viewMode, setViewMode] = useState<ViewMode>('scatter');
  const [colourBy, setColourBy] = useState<ColourBy>('penalty');
  const [showWinning, setShowWinning] = useState(true);
  const [showGlobalBests, setShowGlobalBests] = useState(true);

  // Build landscape points from diversity data.
  // X = hammingToBest (structural distance from best), Y = week, Z = penalty
  const landscapePoints = useMemo(() => {
    return diversity.map(d => ({
      x: d.hammingToBest * 100, // percentage
      y: d.week,
      penalty: d.cumulativePenalty,
      weekPenalty: d.weekPenalty,
      pathID: d.pathID,
      winning: d.winning,
      retained: d.retained,
      family: d.pathID % 8, // approximate family from path ID
      fingerprint: d.fingerprint,
    }));
  }, [diversity]);

  // Discovery events as points on the landscape.
  const discoveryPoints = useMemo(() => {
    return discoveries
      .filter(d => d.eventType === 'GLOBAL_BEST')
      .map(d => ({
        x: 0, // global bests are at distance 0 by definition
        y: d.week,
        penalty: d.newBest,
        workerID: d.workerID,
        candidate: d.candidate,
        improvement: d.improvement,
        elapsedMs: d.elapsedMs,
      }));
  }, [discoveries]);

  // Compute observations from the data.
  const observations = useMemo(() => {
    const obs: string[] = [];

    if (landscapePoints.length === 0) return obs;

    // Basin analysis: are points clustered or spread?
    const hammingValues = landscapePoints.map(p => p.x);
    const maxHamming = Math.max(...hammingValues);
    const avgHamming = hammingValues.reduce((a, b) => a + b, 0) / hammingValues.length;

    if (maxHamming < 10) {
      obs.push('Search explored a single narrow basin — all paths are structurally similar (< 10% Hamming).');
    } else if (avgHamming > 30) {
      obs.push('Search explored multiple competing valleys — high structural diversity (avg > 30% Hamming).');
    } else {
      obs.push(`Moderate structural diversity (avg ${avgHamming.toFixed(1)}% Hamming) — search found nearby alternatives.`);
    }

    // Penalty spread.
    const penalties = landscapePoints.map(p => p.penalty);
    const minPenalty = Math.min(...penalties);
    const maxPenalty = Math.max(...penalties);
    const penaltySpread = maxPenalty - minPenalty;

    if (penaltySpread > minPenalty * 0.5) {
      obs.push(`Wide penalty range (${minPenalty} – ${maxPenalty}) — landscape has significant variation.`);
    } else {
      obs.push(`Tight penalty range (${minPenalty} – ${maxPenalty}) — landscape is relatively flat in this region.`);
    }

    // Winning path position.
    const winningPoints = landscapePoints.filter(p => p.winning);
    if (winningPoints.length > 0) {
      const winningHamming = winningPoints.map(p => p.x);
      const avgWinningHamming = winningHamming.reduce((a, b) => a + b, 0) / winningHamming.length;
      if (avgWinningHamming < 5) {
        obs.push('Winning path stayed very close to the best-known structure throughout.');
      } else {
        obs.push(`Winning path diverged structurally (avg ${avgWinningHamming.toFixed(1)}% from best per week).`);
      }
    }

    // Global best frequency.
    if (discoveryPoints.length > 0) {
      obs.push(`${discoveryPoints.length} global best improvements found across the search.`);
    }

    return obs;
  }, [landscapePoints, discoveryPoints]);

  // Colour mapping.
  const getColour = (point: typeof landscapePoints[0]) => {
    switch (colourBy) {
      case 'penalty': {
        const penalties = landscapePoints.map(p => p.penalty);
        const min = Math.min(...penalties);
        const max = Math.max(...penalties);
        const t = max > min ? (point.penalty - min) / (max - min) : 0;
        // Blue (good) → Red (bad)
        const r = Math.round(t * 255);
        const b = Math.round((1 - t) * 255);
        return `rgb(${r}, 50, ${b})`;
      }
      case 'week': {
        const hue = (point.y / 8) * 300; // weeks 1-8 mapped across hue
        return `hsl(${hue}, 70%, 50%)`;
      }
      case 'family': {
        const familyHues = [0, 45, 90, 135, 180, 225, 270, 315];
        return `hsl(${familyHues[point.family % 8]}, 70%, 50%)`;
      }
    }
  };

  return (
    <div className="space-y-6">
      {/* Controls */}
      <Card title="Fitness Landscape">
        <p className="text-xs text-gray-500 mb-3">
          Structural distance (Hamming %) vs penalty for all retained beam paths. Points near the left at low penalty represent high-quality similar solutions. Wide horizontal spread indicates the algorithm explored structurally diverse solutions — good for avoiding local optima. The winning path (green) should reach both low distance and low penalty.
        </p>
        <div className="flex flex-wrap gap-4 mb-4">
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">View</label>
            <div className="flex gap-1">
              {(['scatter', 'heatmap', 'contour'] as ViewMode[]).map(mode => (
                <button key={mode} onClick={() => setViewMode(mode)}
                  className={`px-2 py-1 text-xs rounded ${viewMode === mode ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-400'}`}>
                  {mode}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">Colour</label>
            <div className="flex gap-1">
              {(['penalty', 'week', 'family'] as ColourBy[]).map(cb => (
                <button key={cb} onClick={() => setColourBy(cb)}
                  className={`px-2 py-1 text-xs rounded ${colourBy === cb ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-400'}`}>
                  {cb}
                </button>
              ))}
            </div>
          </div>
          <div className="flex gap-3 items-end">
            <label className="flex items-center gap-1 text-xs text-gray-400 cursor-pointer">
              <input type="checkbox" checked={showWinning} onChange={e => setShowWinning(e.target.checked)} />
              Winning path
            </label>
            <label className="flex items-center gap-1 text-xs text-gray-400 cursor-pointer">
              <input type="checkbox" checked={showGlobalBests} onChange={e => setShowGlobalBests(e.target.checked)} />
              Global bests
            </label>
          </div>
        </div>

        {/* Main visualisation */}
        {viewMode === 'scatter' && (
          <ResponsiveContainer width="100%" height={500}>
            <ScatterChart margin={{ top: 20, right: 20, bottom: 40, left: 40 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                type="number"
                dataKey="x"
                name="Hamming Distance"
                unit="%"
                stroke="#6B7280"
                fontSize={10}
                label={{ value: 'Structural Distance (Hamming %)', position: 'bottom', fill: '#9CA3AF', fontSize: 11 }}
              />
              <YAxis
                type="number"
                dataKey="penalty"
                name="Penalty"
                stroke="#6B7280"
                fontSize={10}
                label={{ value: 'Cumulative Penalty', angle: -90, position: 'insideLeft', fill: '#9CA3AF', fontSize: 11 }}
              />
              <ZAxis type="number" dataKey="y" range={[30, 200]} name="Week" />
              <Tooltip
                contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151', fontSize: 11 }}
                formatter={(value: any, name: any) => [value, name]}
                labelFormatter={() => ''}
                content={({ payload }) => {
                  if (!payload || payload.length === 0) return null;
                  const d = payload[0].payload;
                  return (
                    <div className="bg-gray-800 border border-gray-600 rounded p-2 text-xs space-y-1">
                      <div><span className="text-gray-400">Path:</span> {d.pathID}</div>
                      <div><span className="text-gray-400">Week:</span> {d.y}</div>
                      <div><span className="text-gray-400">Penalty:</span> {d.penalty.toLocaleString()}</div>
                      <div><span className="text-gray-400">Hamming:</span> {d.x.toFixed(1)}%</div>
                      {d.winning && <div className="text-emerald-400 font-bold">★ Winning path</div>}
                    </div>
                  );
                }}
              />
              {/* All exploration points */}
              <Scatter data={landscapePoints.filter(p => !p.winning)} name="Explored">
                {landscapePoints.filter(p => !p.winning).map((point, i) => (
                  <Cell key={i} fill={getColour(point)} opacity={0.6} />
                ))}
              </Scatter>
              {/* Winning lineage */}
              {showWinning && (
                <Scatter data={landscapePoints.filter(p => p.winning)} name="Winning">
                  {landscapePoints.filter(p => p.winning).map((_, i) => (
                    <Cell key={i} fill="#10B981" opacity={1} />
                  ))}
                </Scatter>
              )}
              {/* Global bests */}
              {showGlobalBests && discoveryPoints.length > 0 && (
                <Scatter data={discoveryPoints} name="Global Best">
                  {discoveryPoints.map((_, i) => (
                    <Cell key={i} fill="#F59E0B" opacity={1} />
                  ))}
                </Scatter>
              )}
            </ScatterChart>
          </ResponsiveContainer>
        )}

        {viewMode === 'heatmap' && (
          <HeatmapView points={landscapePoints} />
        )}

        {viewMode === 'contour' && (
          <ContourView points={landscapePoints} />
        )}
      </Card>

      {/* Observations */}
      {observations.length > 0 && (
        <Card title="Landscape Observations">
          <ul className="space-y-2">
            {observations.map((obs, i) => (
              <li key={i} className="flex gap-2 text-sm text-gray-300">
                <span className="text-blue-400">●</span>
                {obs}
              </li>
            ))}
          </ul>
          <p className="text-[10px] text-gray-600 mt-3 italic">
            Observations derived from measured telemetry only. No interpolation applied.
          </p>
        </Card>
      )}

      {/* Stats summary */}
      <Card title="Landscape Statistics">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Stat label="Data Points" value={landscapePoints.length} />
          <Stat label="Global Bests" value={discoveryPoints.length} />
          <Stat label="Max Hamming" value={`${Math.max(...landscapePoints.map(p => p.x), 0).toFixed(1)}%`} />
          <Stat label="Penalty Range" value={`${Math.min(...landscapePoints.map(p => p.penalty))} – ${Math.max(...landscapePoints.map(p => p.penalty))}`} />
        </div>
      </Card>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <p className="text-[10px] uppercase text-gray-500 tracking-wider">{label}</p>
      <p className="text-sm font-bold text-white">{value}</p>
    </div>
  );
}

// Heatmap view: bin data into grid cells and show density/penalty.
function HeatmapView({ points }: { points: { x: number; y: number; penalty: number }[] }) {
  const gridSize = 20;
  const grid = useMemo(() => {
    if (points.length === 0) return [];
    const maxX = Math.max(...points.map(p => p.x), 1);
    const maxWeek = Math.max(...points.map(p => p.y), 1);

    const cells: { col: number; row: number; count: number; avgPenalty: number }[] = [];
    const bins: Map<string, { penalties: number[]; count: number }> = new Map();

    for (const p of points) {
      const col = Math.min(Math.floor((p.x / maxX) * gridSize), gridSize - 1);
      const row = Math.min(Math.floor(((p.y - 1) / maxWeek) * gridSize), gridSize - 1);
      const key = `${col},${row}`;
      const bin = bins.get(key) || { penalties: [], count: 0 };
      bin.penalties.push(p.penalty);
      bin.count++;
      bins.set(key, bin);
    }

    for (const [key, bin] of bins) {
      const [col, row] = key.split(',').map(Number);
      const avgPenalty = bin.penalties.reduce((a, b) => a + b, 0) / bin.penalties.length;
      cells.push({ col, row, count: bin.count, avgPenalty });
    }

    return cells;
  }, [points]);

  if (grid.length === 0) return <p className="text-gray-500 text-sm">No data for heatmap.</p>;

  const maxCount = Math.max(...grid.map(c => c.count));
  const cellSize = 100 / gridSize;

  return (
    <div className="relative w-full" style={{ paddingBottom: '100%' }}>
      <svg viewBox="0 0 100 100" className="absolute inset-0 w-full h-full">
        {grid.map((cell, i) => {
          const intensity = cell.count / maxCount;
          const r = Math.round(intensity * 255);
          return (
            <rect
              key={i}
              x={cell.col * cellSize}
              y={(gridSize - 1 - cell.row) * cellSize}
              width={cellSize}
              height={cellSize}
              fill={`rgba(59, 130, 246, ${intensity})`}
              stroke="#1F2937"
              strokeWidth="0.2"
            >
              <title>{`Hamming bin: ${cell.col}, Week bin: ${cell.row + 1}, Count: ${cell.count}, Avg penalty: ${cell.avgPenalty.toFixed(0)}`}</title>
            </rect>
          );
        })}
      </svg>
      <div className="absolute bottom-0 left-0 right-0 text-center text-[10px] text-gray-500">
        Structural Distance →
      </div>
      <div className="absolute top-0 left-0 bottom-0 flex items-center">
        <span className="text-[10px] text-gray-500 transform -rotate-90">Week →</span>
      </div>
    </div>
  );
}

// Contour view: simplified contour using binned penalty averages.
function ContourView({ points }: { points: { x: number; y: number; penalty: number }[] }) {
  const gridSize = 15;
  const contourData = useMemo(() => {
    if (points.length === 0) return [];
    const maxX = Math.max(...points.map(p => p.x), 1);
    const maxWeek = Math.max(...points.map(p => p.y), 1);
    const minPenalty = Math.min(...points.map(p => p.penalty));
    const maxPenalty = Math.max(...points.map(p => p.penalty));

    const cells: { col: number; row: number; avgPenalty: number; normalised: number }[] = [];
    const bins: Map<string, number[]> = new Map();

    for (const p of points) {
      const col = Math.min(Math.floor((p.x / maxX) * gridSize), gridSize - 1);
      const row = Math.min(Math.floor(((p.y - 1) / maxWeek) * gridSize), gridSize - 1);
      const key = `${col},${row}`;
      const bin = bins.get(key) || [];
      bin.push(p.penalty);
      bins.set(key, bin);
    }

    for (const [key, penalties] of bins) {
      const [col, row] = key.split(',').map(Number);
      const avg = penalties.reduce((a, b) => a + b, 0) / penalties.length;
      const normalised = maxPenalty > minPenalty ? (avg - minPenalty) / (maxPenalty - minPenalty) : 0;
      cells.push({ col, row, avgPenalty: avg, normalised });
    }

    return cells;
  }, [points]);

  if (contourData.length === 0) return <p className="text-gray-500 text-sm">No data for contour.</p>;

  const cellSize = 100 / gridSize;

  return (
    <div className="relative w-full" style={{ paddingBottom: '100%' }}>
      <svg viewBox="0 0 100 100" className="absolute inset-0 w-full h-full">
        {contourData.map((cell, i) => {
          // Blue = low penalty (valley), Red = high penalty (peak)
          const r = Math.round(cell.normalised * 220);
          const g = Math.round((1 - Math.abs(cell.normalised - 0.5) * 2) * 100);
          const b = Math.round((1 - cell.normalised) * 220);
          return (
            <rect
              key={i}
              x={cell.col * cellSize}
              y={(gridSize - 1 - cell.row) * cellSize}
              width={cellSize}
              height={cellSize}
              fill={`rgb(${r}, ${g}, ${b})`}
              stroke="none"
            >
              <title>{`Avg penalty: ${cell.avgPenalty.toFixed(0)} (${(cell.normalised * 100).toFixed(0)}% of range)`}</title>
            </rect>
          );
        })}
        {/* Contour lines at 25%, 50%, 75% */}
        {[0.25, 0.5, 0.75].map(level => (
          <text key={level} x="2" y={100 - level * 95} fill="rgba(255,255,255,0.3)" fontSize="3">
            {level * 100}%
          </text>
        ))}
      </svg>
      <div className="flex justify-between mt-2 text-[10px] text-gray-500 px-2">
        <span>🔵 Valley (low penalty)</span>
        <span>🔴 Peak (high penalty)</span>
      </div>
    </div>
  );
}
