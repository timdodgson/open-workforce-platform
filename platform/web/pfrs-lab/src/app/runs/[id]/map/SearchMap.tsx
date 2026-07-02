'use client';

import { useState, useMemo, useRef } from 'react';
import Card from '@/components/Card';
import { DiscoveryRecord } from '@/lib/types';

type ColorMode = 'worker' | 'family' | 'lineage' | 'week' | 'type';

interface SelectedPoint {
  discovery: DiscoveryRecord;
  index: number;
}

interface Props {
  discoveries: DiscoveryRecord[];
}

const WEEK_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#f97316', '#ec4899'];

function hashColor(id: number): string {
  const hue = (id * 137.508) % 360;
  return `hsl(${hue}, 70%, 55%)`;
}

export default function SearchMap({ discoveries }: Props) {
  const [colorMode, setColorMode] = useState<ColorMode>('type');
  const [xAxis, setXAxis] = useState<'order' | 'time'>('time');
  const [filterWeek, setFilterWeek] = useState<number | null>(null);
  const [filterType, setFilterType] = useState<string | null>(null);
  const [selected, setSelected] = useState<SelectedPoint | null>(null);
  const [zoom, setZoom] = useState({ xMin: 0, xMax: 1, yMin: 0, yMax: 1 });
  const svgRef = useRef<SVGSVGElement>(null);

  const weeks = useMemo(() => {
    const set = new Set(discoveries.map(d => d.week));
    return Array.from(set).sort((a, b) => a - b);
  }, [discoveries]);

  const eventTypes = useMemo(() => {
    const set = new Set(discoveries.map(d => d.eventType));
    return Array.from(set);
  }, [discoveries]);

  // Filter discoveries.
  const filtered = useMemo(() => {
    let data = discoveries;
    if (filterWeek !== null) data = data.filter(d => d.week === filterWeek);
    if (filterType !== null) data = data.filter(d => d.eventType === filterType);
    return data;
  }, [discoveries, filterWeek, filterType]);

  // Compute axis bounds.
  const bounds = useMemo(() => {
    if (filtered.length === 0) return { xMin: 0, xMax: 1, yMin: 0, yMax: 1 };
    const xValues = filtered.map((d, i) => xAxis === 'time' ? d.elapsedMs : i);
    const yValues = filtered.map(d => d.newBest);
    return {
      xMin: Math.min(...xValues),
      xMax: Math.max(...xValues),
      yMin: Math.min(...yValues),
      yMax: Math.max(...yValues),
    };
  }, [filtered, xAxis]);

  // Apply zoom.
  const viewBounds = useMemo(() => ({
    xMin: bounds.xMin + (bounds.xMax - bounds.xMin) * zoom.xMin,
    xMax: bounds.xMin + (bounds.xMax - bounds.xMin) * zoom.xMax,
    yMin: bounds.yMin + (bounds.yMax - bounds.yMin) * zoom.yMin,
    yMax: bounds.yMin + (bounds.yMax - bounds.yMin) * zoom.yMax,
  }), [bounds, zoom]);

  // Map data point to SVG coordinates.
  const WIDTH = 900;
  const HEIGHT = 400;
  const PADDING = 40;

  function toSvgX(val: number): number {
    const range = viewBounds.xMax - viewBounds.xMin || 1;
    return PADDING + ((val - viewBounds.xMin) / range) * (WIDTH - 2 * PADDING);
  }

  function toSvgY(val: number): number {
    const range = viewBounds.yMax - viewBounds.yMin || 1;
    // Invert Y (lower penalty = higher on screen).
    return PADDING + (1 - (val - viewBounds.yMin) / range) * (HEIGHT - 2 * PADDING);
  }

  function getColor(d: DiscoveryRecord): string {
    switch (colorMode) {
      case 'worker': return hashColor(d.workerID);
      case 'family': return hashColor(d.beamPath);
      case 'lineage': return hashColor(d.branchDepth * 31 + d.beamPath);
      case 'week': return WEEK_COLORS[(d.week - 1) % WEEK_COLORS.length];
      case 'type':
        return d.eventType === 'global_best' ? '#fbbf24' : '#10b981';
      default: return '#3b82f6';
    }
  }

  function getRadius(d: DiscoveryRecord): number {
    if (d.eventType === 'global_best') return 5;
    return 3;
  }

  function handleZoomIn() {
    setZoom(z => {
      const xRange = z.xMax - z.xMin;
      const yRange = z.yMax - z.yMin;
      return {
        xMin: z.xMin + xRange * 0.1,
        xMax: z.xMax - xRange * 0.1,
        yMin: z.yMin + yRange * 0.1,
        yMax: z.yMax - yRange * 0.1,
      };
    });
  }

  function handleZoomOut() {
    setZoom(z => {
      const xRange = z.xMax - z.xMin;
      const yRange = z.yMax - z.yMin;
      return {
        xMin: Math.max(0, z.xMin - xRange * 0.15),
        xMax: Math.min(1, z.xMax + xRange * 0.15),
        yMin: Math.max(0, z.yMin - yRange * 0.15),
        yMax: Math.min(1, z.yMax + yRange * 0.15),
      };
    });
  }

  function handleReset() {
    setZoom({ xMin: 0, xMax: 1, yMin: 0, yMax: 1 });
  }

  function handlePointClick(d: DiscoveryRecord, i: number) {
    setSelected({ discovery: d, index: i });
  }

  function formatMs(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card title="Search Map">
        <div className="flex flex-wrap gap-3 mb-3">
          {/* X axis */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">X:</span>
            {(['time', 'order'] as const).map(opt => (
              <button key={opt} onClick={() => setXAxis(opt)}
                className={`px-2 py-0.5 rounded text-[10px] ${
                  xAxis === opt ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                }`}>{opt}</button>
            ))}
          </div>
          {/* Color mode */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Color:</span>
            {(['type', 'worker', 'family', 'week', 'lineage'] as ColorMode[]).map(opt => (
              <button key={opt} onClick={() => setColorMode(opt)}
                className={`px-2 py-0.5 rounded text-[10px] ${
                  colorMode === opt ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                }`}>{opt}</button>
            ))}
          </div>
          {/* Week filter */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Week:</span>
            <button onClick={() => setFilterWeek(null)}
              className={`px-2 py-0.5 rounded text-[10px] ${
                filterWeek === null ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'
              }`}>All</button>
            {weeks.map(w => (
              <button key={w} onClick={() => setFilterWeek(w)}
                className={`px-2 py-0.5 rounded text-[10px] ${
                  filterWeek === w ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'
                }`}>{w}</button>
            ))}
          </div>
          {/* Type filter */}
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Type:</span>
            <button onClick={() => setFilterType(null)}
              className={`px-2 py-0.5 rounded text-[10px] ${
                filterType === null ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'
              }`}>All</button>
            {eventTypes.map(t => (
              <button key={t} onClick={() => setFilterType(t)}
                className={`px-2 py-0.5 rounded text-[10px] ${
                  filterType === t ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'
                }`}>{t.replace('_', ' ')}</button>
            ))}
          </div>
          {/* Zoom */}
          <div className="flex items-center gap-1 ml-auto">
            <button onClick={handleZoomIn} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">🔍+</button>
            <button onClick={handleZoomOut} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">🔍−</button>
            <button onClick={handleReset} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">Reset</button>
          </div>
        </div>
        <p className="text-[9px] text-gray-600 mb-2">{filtered.length.toLocaleString()} discoveries shown</p>

        {/* SVG scatter plot */}
        <svg ref={svgRef} viewBox={`0 0 ${WIDTH} ${HEIGHT}`} className="w-full h-[400px] bg-gray-900 rounded border border-gray-800">
          {/* Y-axis labels */}
          {[0, 0.25, 0.5, 0.75, 1].map(frac => {
            const val = viewBounds.yMin + (viewBounds.yMax - viewBounds.yMin) * (1 - frac);
            const y = PADDING + frac * (HEIGHT - 2 * PADDING);
            return (
              <g key={frac}>
                <line x1={PADDING} y1={y} x2={WIDTH - PADDING} y2={y} stroke="#1f2937" strokeWidth="0.5" />
                <text x={PADDING - 5} y={y} textAnchor="end" dominantBaseline="middle"
                  className="fill-gray-600 text-[8px]">{Math.round(val)}</text>
              </g>
            );
          })}
          {/* X-axis labels */}
          {[0, 0.25, 0.5, 0.75, 1].map(frac => {
            const val = viewBounds.xMin + (viewBounds.xMax - viewBounds.xMin) * frac;
            const x = PADDING + frac * (WIDTH - 2 * PADDING);
            const label = xAxis === 'time' ? formatMs(val) : Math.round(val).toString();
            return (
              <g key={frac}>
                <line x1={x} y1={PADDING} x2={x} y2={HEIGHT - PADDING} stroke="#1f2937" strokeWidth="0.5" />
                <text x={x} y={HEIGHT - PADDING + 12} textAnchor="middle"
                  className="fill-gray-600 text-[8px]">{label}</text>
              </g>
            );
          })}
          {/* Data points */}
          {filtered.map((d, i) => {
            const xVal = xAxis === 'time' ? d.elapsedMs : i;
            const yVal = d.newBest;
            const cx = toSvgX(xVal);
            const cy = toSvgY(yVal);
            // Skip points outside view.
            if (cx < PADDING || cx > WIDTH - PADDING || cy < PADDING || cy > HEIGHT - PADDING) return null;
            const isSelected = selected?.index === i;
            return (
              <circle
                key={i}
                cx={cx} cy={cy}
                r={isSelected ? 7 : getRadius(d)}
                fill={getColor(d)}
                opacity={isSelected ? 1 : 0.7}
                stroke={isSelected ? '#fff' : 'none'}
                strokeWidth={isSelected ? 2 : 0}
                className="cursor-pointer hover:opacity-100"
                onClick={() => handlePointClick(d, i)}
              />
            );
          })}
          {/* Axis labels */}
          <text x={WIDTH / 2} y={HEIGHT - 5} textAnchor="middle" className="fill-gray-500 text-[9px]">
            {xAxis === 'time' ? 'Elapsed Time' : 'Discovery Order'}
          </text>
          <text x={12} y={HEIGHT / 2} textAnchor="middle" transform={`rotate(-90, 12, ${HEIGHT / 2})`}
            className="fill-gray-500 text-[9px]">Penalty (Best)</text>
        </svg>
      </Card>

      {/* Selected point detail */}
      {selected && (
        <Card title="Selected Discovery">
          <div className="grid grid-cols-4 gap-3 text-xs">
            <div>
              <p className="text-gray-500">Worker</p>
              <p className="font-mono text-blue-400">{selected.discovery.workerID}</p>
            </div>
            <div>
              <p className="text-gray-500">Beam Path</p>
              <p className="font-mono text-purple-400">{selected.discovery.beamPath}</p>
            </div>
            <div>
              <p className="text-gray-500">Week</p>
              <p className="font-medium">{selected.discovery.week}</p>
            </div>
            <div>
              <p className="text-gray-500">Event Type</p>
              <p className={selected.discovery.eventType === 'global_best' ? 'text-yellow-400' : 'text-emerald-400'}>
                {selected.discovery.eventType.replace('_', ' ')}
              </p>
            </div>
            <div>
              <p className="text-gray-500">New Best</p>
              <p className="font-bold text-emerald-400">{selected.discovery.newBest.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-gray-500">Previous Best</p>
              <p>{selected.discovery.previousBest.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-gray-500">Improvement</p>
              <p className="text-emerald-400">−{selected.discovery.improvement}</p>
            </div>
            <div>
              <p className="text-gray-500">Elapsed</p>
              <p>{formatMs(selected.discovery.elapsedMs)}</p>
            </div>
            <div>
              <p className="text-gray-500">Branch Depth</p>
              <p>{selected.discovery.branchDepth}</p>
            </div>
            <div>
              <p className="text-gray-500">Seed</p>
              <p className="font-mono">{selected.discovery.seedUsed}</p>
            </div>
            <div>
              <p className="text-gray-500">Candidate #</p>
              <p>{selected.discovery.candidate.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-gray-500">Discovery #</p>
              <p>{selected.discovery.discoveryNumber}</p>
            </div>
          </div>
          <button onClick={() => setSelected(null)}
            className="mt-3 px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px] text-gray-400">
            Close
          </button>
        </Card>
      )}

      {/* Legend */}
      <Card title="Legend">
        <div className="flex flex-wrap gap-4 text-[10px] text-gray-400">
          {colorMode === 'type' && (
            <>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-yellow-400" />Global Best</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-emerald-500" />Local Best</span>
            </>
          )}
          {colorMode === 'week' && weeks.map(w => (
            <span key={w} className="flex items-center gap-1">
              <span className="w-3 h-3 rounded-full" style={{ background: WEEK_COLORS[(w - 1) % WEEK_COLORS.length] }} />
              Week {w}
            </span>
          ))}
          {colorMode === 'worker' && <span>Each colour = unique worker ID</span>}
          {colorMode === 'family' && <span>Each colour = beam path family</span>}
          {colorMode === 'lineage' && <span>Each colour = branch lineage</span>}
        </div>
      </Card>
    </div>
  );
}
