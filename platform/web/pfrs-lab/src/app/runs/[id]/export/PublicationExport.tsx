'use client';

import { useState, useRef, useMemo } from 'react';
import Card from '@/components/Card';
import { RunSummary, TreeNode, DiversityRecord, DiscoveryRecord, WorkerLifecycle } from '@/lib/types';

type ExportFormat = 'svg' | 'png';
type FigureId = 'penalty_waterfall' | 'beam_health' | 'discovery_timeline' | 'entropy' | 'worker_utilisation' | 'convergence';

interface Figure {
  id: FigureId;
  title: string;
  caption: string;
  number: number;
}

interface Props {
  runId: string;
  summary: RunSummary;
  tree: TreeNode[];
  diversity: DiversityRecord[];
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
}

// Publication colour palette (colourblind-friendly).
const PALETTE = {
  primary: '#2563eb',
  secondary: '#059669',
  accent: '#d97706',
  danger: '#dc2626',
  neutral: '#6b7280',
  bg: '#ffffff',
  text: '#1f2937',
  grid: '#e5e7eb',
  font: 'Inter, system-ui, sans-serif',
};

const DPI_OPTIONS = [72, 150, 300, 600];

export default function PublicationExport({ runId, summary, tree, diversity, discoveries, workers }: Props) {
  const [format, setFormat] = useState<ExportFormat>('svg');
  const [dpi, setDpi] = useState(300);
  const [selectedFigures, setSelectedFigures] = useState<Set<FigureId>>(new Set(['penalty_waterfall', 'beam_health', 'discovery_timeline', 'entropy', 'worker_utilisation', 'convergence']));
  const svgRefs = useRef<Map<FigureId, SVGSVGElement>>(new Map());

  const figures: Figure[] = useMemo(() => {
    const figs: Figure[] = [];
    let num = 1;
    const allFigs: { id: FigureId; title: string; captionFn: () => string }[] = [
      { id: 'penalty_waterfall', title: 'Penalty by Week', captionFn: () => `Cumulative penalty growth across ${summary.numWeeks} weeks. Total penalty: ${summary.totalPenalty.toLocaleString()}.` },
      { id: 'convergence', title: 'Convergence Over Time', captionFn: () => `Global best penalty over elapsed time. ${discoveries.filter(d=>d.eventType==='GLOBAL_BEST').length} improvements found.` },
      { id: 'beam_health', title: 'Beam Diversity', captionFn: () => `Near-duplicate rate and retained path count per week.` },
      { id: 'entropy', title: 'Lineage Entropy', captionFn: () => `Shannon entropy of beam family distribution over time.` },
      { id: 'discovery_timeline', title: 'Discovery Timeline', captionFn: () => `All ${discoveries.length} discoveries plotted by elapsed time and penalty achieved.` },
      { id: 'worker_utilisation', title: 'Worker Utilisation', captionFn: () => `Distribution of ${workers.length} workers by contribution type.` },
    ];
    for (const f of allFigs) {
      if (selectedFigures.has(f.id)) {
        figs.push({ id: f.id, title: f.title, caption: f.captionFn(), number: num++ });
      }
    }
    return figs;
  }, [selectedFigures, summary, discoveries, workers]);

  function toggleFigure(id: FigureId) {
    setSelectedFigures(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  function exportSVG(id: FigureId) {
    const svg = svgRefs.current.get(id);
    if (!svg) return;
    const serializer = new XMLSerializer();
    const source = serializer.serializeToString(svg);
    const blob = new Blob([source], { type: 'image/svg+xml' });
    downloadBlob(blob, `figure_${id}.svg`);
  }

  function exportPNG(id: FigureId) {
    const svg = svgRefs.current.get(id);
    if (!svg) return;
    const serializer = new XMLSerializer();
    const source = serializer.serializeToString(svg);
    const canvas = document.createElement('canvas');
    const scale = dpi / 72;
    canvas.width = 800 * scale;
    canvas.height = 300 * scale;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const img = new Image();
    img.onload = () => {
      ctx.scale(scale, scale);
      ctx.drawImage(img, 0, 0);
      canvas.toBlob(blob => {
        if (blob) downloadBlob(blob, `figure_${id}_${dpi}dpi.png`);
      });
    };
    img.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(source)));
  }

  function downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = filename;
    document.body.appendChild(a); a.click();
    document.body.removeChild(a); URL.revokeObjectURL(url);
  }

  function exportFigure(id: FigureId) {
    if (format === 'svg') exportSVG(id);
    else exportPNG(id);
  }

  function exportAll() {
    for (const fig of figures) exportFigure(fig.id);
  }

  // SVG rendering helpers (publication style — white bg, dark text).
  const W = 800, H = 300, PL = 60, PR = 30, PT = 30, PB = 50;

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card title="Publication Export">
        <div className="flex flex-wrap gap-3 mb-3">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Format:</span>
            {(['svg', 'png'] as ExportFormat[]).map(f => (
              <button key={f} onClick={() => setFormat(f)}
                className={`px-2 py-0.5 rounded text-[10px] uppercase ${format === f ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{f}</button>
            ))}
          </div>
          {format === 'png' && (
            <div className="flex items-center gap-1">
              <span className="text-[10px] text-gray-500">DPI:</span>
              {DPI_OPTIONS.map(d => (
                <button key={d} onClick={() => setDpi(d)}
                  className={`px-2 py-0.5 rounded text-[10px] ${dpi === d ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{d}</button>
              ))}
            </div>
          )}
          <button onClick={exportAll}
            className="ml-auto px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-xs font-medium">
            Export All ({figures.length})
          </button>
        </div>
        <div className="flex flex-wrap gap-2">
          {(['penalty_waterfall', 'convergence', 'beam_health', 'entropy', 'discovery_timeline', 'worker_utilisation'] as FigureId[]).map(id => (
            <label key={id} className="flex items-center gap-1 text-[10px] text-gray-400">
              <input type="checkbox" checked={selectedFigures.has(id)} onChange={() => toggleFigure(id)} />
              {id.replace('_', ' ')}
            </label>
          ))}
        </div>
      </Card>

      {/* Figures */}
      {figures.map(fig => (
        <Card key={fig.id} title={`Figure ${fig.number}. ${fig.title}`}>
          <div className="bg-white rounded p-2 mb-2">
            <svg ref={el => { if (el) svgRefs.current.set(fig.id, el); }}
              viewBox={`0 0 ${W} ${H}`} className="w-full" style={{ height: '250px', fontFamily: PALETTE.font }}>
              <rect width={W} height={H} fill={PALETTE.bg} />
              {renderFigure(fig.id, { summary, tree, diversity, discoveries, workers })}
              <text x={W/2} y={H - 10} textAnchor="middle" fontSize="10" fill={PALETTE.text}>
                Figure {fig.number}. {fig.title}
              </text>
            </svg>
          </div>
          <p className="text-xs text-gray-400 italic mb-2">{fig.caption}</p>
          <button onClick={() => exportFigure(fig.id)}
            className="px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px] text-gray-300">
            Export {format.toUpperCase()}
          </button>
        </Card>
      ))}
    </div>
  );
}

function renderFigure(id: FigureId, data: { summary: RunSummary; tree: TreeNode[]; diversity: DiversityRecord[]; discoveries: DiscoveryRecord[]; workers: WorkerLifecycle[] }): React.ReactElement {
  const W = 800, H = 300, PL = 60, PR = 30, PT = 30, PB = 50;
  const plotW = W - PL - PR, plotH = H - PT - PB;

  switch (id) {
    case 'penalty_waterfall': {
      const weeks = data.summary.weeks;
      const maxP = Math.max(...weeks.map(w => w.finalPenalty), 1);
      return (
        <g>
          {weeks.map((w, i) => {
            const barW = plotW / weeks.length * 0.7;
            const x = PL + (i / weeks.length) * plotW + barW * 0.15;
            const h = (w.finalPenalty / maxP) * plotH;
            return (
              <g key={i}>
                <rect x={x} y={PT + plotH - h} width={barW} height={h} fill={PALETTE.primary} />
                <text x={x + barW/2} y={PT + plotH + 15} textAnchor="middle" fontSize="9" fill={PALETTE.text}>W{w.week}</text>
                <text x={x + barW/2} y={PT + plotH - h - 5} textAnchor="middle" fontSize="8" fill={PALETTE.neutral}>{w.finalPenalty}</text>
              </g>
            );
          })}
        </g>
      );
    }
    case 'convergence': {
      const globals = data.discoveries.filter(d => d.eventType === 'GLOBAL_BEST');
      if (globals.length === 0) return <text x={W/2} y={H/2} textAnchor="middle" fill={PALETTE.neutral}>No data</text>;
      const maxT = Math.max(...globals.map(d => d.elapsedMs), 1);
      const maxP = Math.max(...globals.map(d => d.newBest), 1);
      const minP = Math.min(...globals.map(d => d.newBest));
      const range = maxP - minP || 1;
      return (
        <g>
          <polyline points={globals.map(d => {
            const x = PL + (d.elapsedMs / maxT) * plotW;
            const y = PT + (1 - (d.newBest - minP) / range) * plotH;
            return `${x},${y}`;
          }).join(' ')} fill="none" stroke={PALETTE.secondary} strokeWidth="2" />
          {globals.map((d, i) => (
            <circle key={i} cx={PL + (d.elapsedMs / maxT) * plotW} cy={PT + (1 - (d.newBest - minP) / range) * plotH} r={3} fill={PALETTE.secondary} />
          ))}
        </g>
      );
    }
    case 'entropy': {
      const weeks = [...new Set(data.tree.map(t => t.week))].sort((a, b) => a - b);
      const parentMap = new Map<number, number>();
      for (const t of data.tree) parentMap.set(t.pathID, t.parentID);
      const entropies = weeks.map(week => {
        const retained = data.tree.filter(t => t.week === week && t.retained);
        if (retained.length <= 1) return 0;
        const fam = new Map<number, number>();
        for (const t of retained) { let c = t.pathID; let i = 0; while (parentMap.has(c) && parentMap.get(c)! >= 0 && i < 100) { c = parentMap.get(c)!; i++; } fam.set(c, (fam.get(c)||0)+1); }
        let e = 0; for (const cnt of fam.values()) { const p = cnt/retained.length; if (p > 0) e -= p * Math.log2(p); } return e;
      });
      const maxE = Math.max(...entropies, 0.01);
      return (
        <g>
          {entropies.map((e, i) => {
            const barW = plotW / entropies.length * 0.7;
            const x = PL + (i / entropies.length) * plotW + barW * 0.15;
            const h = (e / maxE) * plotH;
            return <rect key={i} x={x} y={PT + plotH - h} width={barW} height={h} fill={PALETTE.primary} />;
          })}
        </g>
      );
    }
    case 'beam_health': {
      // Near-duplicate rate and retained path count per week.
      const weeks = [...new Set(data.diversity.map(d => d.week))].sort((a, b) => a - b);
      if (weeks.length === 0) return <text x={W/2} y={H/2} textAnchor="middle" fill={PALETTE.neutral}>No diversity data</text>;
      const weekData = weeks.map(w => {
        const recs = data.diversity.filter(d => d.week === w);
        const retained = recs.filter(r => r.retained).length;
        const nearDups = recs.filter(r => r.nearDuplicate).length;
        const dupRate = recs.length > 0 ? nearDups / recs.length : 0;
        return { week: w, retained, dupRate };
      });
      const maxRetained = Math.max(...weekData.map(d => d.retained), 1);
      return (
        <g>
          {weekData.map((d, i) => {
            const barW = plotW / weekData.length * 0.35;
            const x = PL + (i / weekData.length) * plotW;
            // Retained bar (blue).
            const h1 = (d.retained / maxRetained) * plotH;
            // Dup rate bar (red, scaled to plotH).
            const h2 = d.dupRate * plotH;
            return (
              <g key={i}>
                <rect x={x} y={PT + plotH - h1} width={barW} height={h1} fill={PALETTE.primary} opacity={0.8} />
                <rect x={x + barW + 2} y={PT + plotH - h2} width={barW} height={h2} fill={PALETTE.danger} opacity={0.7} />
                <text x={x + barW} y={PT + plotH + 15} textAnchor="middle" fontSize="8" fill={PALETTE.text}>W{d.week}</text>
              </g>
            );
          })}
          {/* Legend */}
          <rect x={PL} y={PT - 15} width={10} height={10} fill={PALETTE.primary} />
          <text x={PL + 14} y={PT - 7} fontSize="8" fill={PALETTE.text}>Retained</text>
          <rect x={PL + 70} y={PT - 15} width={10} height={10} fill={PALETTE.danger} />
          <text x={PL + 84} y={PT - 7} fontSize="8" fill={PALETTE.text}>Near-duplicate %</text>
        </g>
      );
    }
    case 'discovery_timeline': {
      // All discoveries plotted by elapsed time and penalty.
      if (data.discoveries.length === 0) return <text x={W/2} y={H/2} textAnchor="middle" fill={PALETTE.neutral}>No discovery data</text>;
      const maxT = Math.max(...data.discoveries.map(d => d.elapsedMs), 1);
      const maxP = Math.max(...data.discoveries.map(d => d.newBest), 1);
      const minP = Math.min(...data.discoveries.map(d => d.newBest));
      const range = maxP - minP || 1;
      const globals = data.discoveries.filter(d => d.eventType === 'GLOBAL_BEST');
      const locals = data.discoveries.filter(d => d.eventType === 'LOCAL_BEST');
      return (
        <g>
          {locals.map((d, i) => (
            <circle key={`l${i}`} cx={PL + (d.elapsedMs / maxT) * plotW} cy={PT + (1 - (d.newBest - minP) / range) * plotH} r={2} fill={PALETTE.accent} opacity={0.5} />
          ))}
          {globals.map((d, i) => (
            <circle key={`g${i}`} cx={PL + (d.elapsedMs / maxT) * plotW} cy={PT + (1 - (d.newBest - minP) / range) * plotH} r={4} fill={PALETTE.secondary} />
          ))}
          {/* Legend */}
          <circle cx={PL + 5} cy={PT - 10} r={4} fill={PALETTE.secondary} />
          <text x={PL + 14} y={PT - 7} fontSize="8" fill={PALETTE.text}>Global best</text>
          <circle cx={PL + 100} cy={PT - 10} r={2} fill={PALETTE.accent} />
          <text x={PL + 107} y={PT - 7} fontSize="8" fill={PALETTE.text}>Local best</text>
        </g>
      );
    }
    case 'worker_utilisation': {
      // Distribution of workers by contribution type.
      if (data.workers.length === 0) return <text x={W/2} y={H/2} textAnchor="middle" fill={PALETTE.neutral}>No worker data</text>;
      const improved = data.workers.filter(w => w.bestPenalty < w.startPenalty).length;
      const producedBest = data.workers.filter(w => w.producedGlobalBest).length;
      const plateaued = data.workers.filter(w => w.plateauCount > 0).length;
      const total = data.workers.length;
      const categories = [
        { label: 'Produced Global Best', count: producedBest, color: PALETTE.secondary },
        { label: 'Improved (not global)', count: improved - producedBest, color: PALETTE.primary },
        { label: 'Plateaued', count: plateaued, color: PALETTE.accent },
        { label: 'No improvement', count: total - improved, color: PALETTE.neutral },
      ].filter(c => c.count > 0);
      const maxCount = Math.max(...categories.map(c => c.count), 1);
      return (
        <g>
          {categories.map((c, i) => {
            const barH = plotH / categories.length * 0.7;
            const y = PT + (i / categories.length) * plotH;
            const w = (c.count / maxCount) * plotW * 0.7;
            return (
              <g key={i}>
                <rect x={PL} y={y} width={w} height={barH} fill={c.color} rx={3} />
                <text x={PL + w + 5} y={y + barH / 2 + 4} fontSize="9" fill={PALETTE.text}>{c.count}</text>
                <text x={PL - 5} y={y + barH / 2 + 4} textAnchor="end" fontSize="8" fill={PALETTE.neutral}>{c.label}</text>
              </g>
            );
          })}
        </g>
      );
    }
    default:
      return <text x={W/2} y={H/2} textAnchor="middle" fill={PALETTE.neutral} fontSize="12">Chart: {id}</text>;
  }
}
