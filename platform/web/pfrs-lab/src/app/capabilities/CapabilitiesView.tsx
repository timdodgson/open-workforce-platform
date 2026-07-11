'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import Card from '@/components/Card';
import {
  ASSIST_ARCHITECTURE,
  DOMAINS,
  GAP_ROADMAP,
  SEARCH_ALGORITHMS,
  SI_MODES,
  VALIDATION_STATUS,
  VIEWERS,
  statusClass,
  statusLabel,
  statusTitle,
  type CellStatus,
  type DomainId,
  type MatrixRow,
} from '@/lib/capability-matrix';

export interface RegistrySummary {
  total: number;
  promotionReady: number;
  domains: DomainId[];
}

function StatusCell({ status }: { status: CellStatus }) {
  return (
    <span className={statusClass(status)} title={statusTitle(status)}>
      {statusLabel(status)}
    </span>
  );
}

function SimpleMatrixTable({
  title,
  subtitle,
  rows,
  showNotes = true,
}: {
  title: string;
  subtitle?: string;
  rows: MatrixRow[];
  showNotes?: boolean;
}) {
  return (
    <Card title={title}>
      {subtitle && <p className="text-xs text-gray-500 mb-3">{subtitle}</p>}
      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase border-b border-gray-800">
              <th className="text-left p-2">Capability</th>
              {DOMAINS.map((d) => (
                <th key={d.id} className="text-center p-2">{d.label}</th>
              ))}
              {showNotes && <th className="text-left p-2">Notes</th>}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id} className="border-b border-gray-800/60 hover:bg-gray-800/30">
                <td className="p-2 text-gray-200 font-medium">{row.label}</td>
                {DOMAINS.map((d) => (
                  <td key={d.id} className="p-2 text-center">
                    <StatusCell status={row.cells[d.id]} />
                  </td>
                ))}
                {showNotes && <td className="p-2 text-gray-500 text-[10px]">{row.notes ?? ''}</td>}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

export default function CapabilitiesView({ registry }: { registry: RegistrySummary | null }) {
  const [phaseFilter, setPhaseFilter] = useState<string>('all');

  const phases = useMemo(() => {
    const set = new Set(GAP_ROADMAP.map((g) => g.phase));
    return ['all', ...Array.from(set)];
  }, []);

  const filteredGaps = useMemo(
    () => (phaseFilter === 'all' ? GAP_ROADMAP : GAP_ROADMAP.filter((g) => g.phase === phaseFilter)),
    [phaseFilter],
  );

  const openGaps = GAP_ROADMAP.filter((g) => g.status === 'open').length;

  return (
    <div className="space-y-6">
      <div className="p-4 rounded-lg bg-gray-900/80 border border-gray-800">
        <h1 className="text-lg font-semibold text-gray-100 mb-1">Platform Capabilities</h1>
        <p className="text-sm text-gray-400 mb-3">
          Honest status across four optimisation domains. Cells distinguish{' '}
          <span className="text-emerald-400">native runtime</span>,{' '}
          <span className="text-blue-400">adapted telemetry (🔀)</span>, and{' '}
          <span className="text-amber-400">dashboard UX</span>.
        </p>
        <div className="flex flex-wrap gap-2 text-[10px]">
          <LegendChip label="✅ Complete" className="text-emerald-400" />
          <LegendChip label="⚠️ Partial" className="text-amber-400" />
          <LegendChip label="❌ Missing" className="text-red-400" />
          <LegendChip label="🔀 Adapted" className="text-blue-400" />
          <LegendChip label="— N/A" className="text-gray-600" />
        </div>
        {registry && (
          <div className="mt-3 flex flex-wrap gap-3 text-xs">
            <span className="px-2 py-1 rounded bg-emerald-900/40 text-emerald-300 border border-emerald-800">
              Policies: {registry.promotionReady}/{registry.total} promotion-ready
            </span>
            <span className="px-2 py-1 rounded bg-gray-800 text-gray-400">
              Domains in registry: {registry.domains.map((d) => d.toUpperCase()).join(', ')}
            </span>
            <Link href="/intelligence" className="text-blue-400 hover:underline">Search Intelligence →</Link>
            <Link href="/benchmarks" className="text-blue-400 hover:underline">Benchmarks →</Link>
            <Link href="/experiment-matrix" className="text-blue-400 hover:underline">Experiment Matrix →</Link>
          </div>
        )}
      </div>

      <Card title="Domains">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
          {DOMAINS.map((d) => (
            <div key={d.id} className="bg-gray-800 rounded p-3">
              <p className="text-sm font-semibold text-gray-200">{d.label}</p>
              <p className="text-[10px] text-gray-500">{d.subtitle}</p>
            </div>
          ))}
        </div>
      </Card>

      <SimpleMatrixTable
        title="Search algorithms"
        subtitle="CLI solvers available per domain"
        rows={SEARCH_ALGORITHMS}
      />

      <SimpleMatrixTable
        title="Search Intelligence modes"
        subtitle="--worker-decision-mode and --policy-mode"
        rows={SI_MODES}
      />

      <Card title="Assist architecture by domain">
        <p className="text-xs text-gray-500 mb-3">
          Runtime = native hook during search. Telemetry = CSV contract on labelled runs (may use adapters).
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase border-b border-gray-800">
                <th className="text-left p-2">Assist type</th>
                <th className="text-left p-2">What it controls</th>
                <th colSpan={4} className="text-center p-2 border-l border-gray-800">Runtime native</th>
                <th colSpan={4} className="text-center p-2 border-l border-gray-800">Telemetry CSV</th>
                <th className="text-left p-2 border-l border-gray-800">Fix</th>
              </tr>
              <tr className="text-gray-600 border-b border-gray-800">
                <th colSpan={2} />
                {DOMAINS.map((d) => (
                  <th key={`r-${d.id}`} className="p-1 text-center font-normal">{d.label}</th>
                ))}
                {DOMAINS.map((d) => (
                  <th key={`t-${d.id}`} className="p-1 text-center font-normal border-l border-gray-800/50">{d.label}</th>
                ))}
                <th />
              </tr>
            </thead>
            <tbody>
              {ASSIST_ARCHITECTURE.map((row) => (
                <tr key={row.id} className="border-b border-gray-800/60 hover:bg-gray-800/30">
                  <td className="p-2 text-gray-200 font-medium">{row.label}</td>
                  <td className="p-2 text-gray-500">{row.description}</td>
                  {DOMAINS.map((d) => (
                    <td key={`r-${row.id}-${d.id}`} className="p-2 text-center">
                      <StatusCell status={row.runtime[d.id]} />
                    </td>
                  ))}
                  {DOMAINS.map((d) => (
                    <td key={`t-${row.id}-${d.id}`} className="p-2 text-center border-l border-gray-800/30">
                      <StatusCell status={row.telemetry[d.id]} />
                    </td>
                  ))}
                  <td className="p-2 text-gray-500 border-l border-gray-800/30">
                    {row.fixPhase && row.fixPhase !== '—' && (
                      <span className="text-amber-400">{row.fixPhase}</span>
                    )}
                    {row.fixNote && <p className="mt-0.5">{row.fixNote}</p>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      <SimpleMatrixTable
        title="Problem-specific viewers & platform pages"
        subtitle="Dashboard UX per domain"
        rows={VIEWERS}
      />

      <Card title="Validation status">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2">
          {VALIDATION_STATUS.map((v) => (
            <div key={v.domain} className="bg-gray-800 rounded p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-sm font-semibold text-gray-200">{v.domain.toUpperCase()}</span>
                <StatusCell status={v.status} />
              </div>
              <p className="text-[10px] text-gray-500">{v.notes}</p>
            </div>
          ))}
        </div>
      </Card>

      <Card title={`Gap roadmap (${openGaps} open)`}>
        <p className="text-xs text-gray-500 mb-3">
          Plan to address ⚠️ and ❌ cells. Phases 0–1 complete.
        </p>
        <div className="flex flex-wrap gap-2 mb-3">
          {phases.map((p) => (
            <button
              key={p}
              type="button"
              onClick={() => setPhaseFilter(p)}
              className={`text-[10px] px-2 py-1 rounded border ${
                phaseFilter === p
                  ? 'bg-blue-900/50 text-blue-300 border-blue-700'
                  : 'bg-gray-800 text-gray-400 border-gray-700'
              }`}
            >
              {p === 'all' ? 'All phases' : p}
            </button>
          ))}
        </div>
        <div className="space-y-2">
          {filteredGaps.map((gap) => (
            <div
              key={gap.id}
              className={`p-3 rounded border text-[11px] ${
                gap.status === 'in_progress'
                  ? 'bg-blue-950/30 border-blue-900'
                  : gap.status === 'done'
                    ? 'bg-emerald-950/20 border-emerald-900/50'
                    : 'bg-gray-900 border-gray-800'
              }`}
            >
              <div className="flex flex-wrap items-center gap-2 mb-1">
                <span className="font-semibold text-gray-200">{gap.title}</span>
                <span className="text-gray-500">{gap.phase}</span>
                <span className="text-gray-600">{gap.effort}</span>
                <span
                  className={`px-1.5 py-0.5 rounded text-[9px] uppercase ${
                    gap.status === 'in_progress'
                      ? 'bg-blue-900 text-blue-300'
                      : gap.status === 'done'
                        ? 'bg-emerald-900 text-emerald-300'
                        : 'bg-gray-800 text-gray-500'
                  }`}
                >
                  {gap.status.replace('_', ' ')}
                </span>
              </div>
              <p className="text-gray-500 mb-1"><span className="text-gray-600">Why:</span> {gap.why}</p>
              <p className="text-gray-400 mb-1"><span className="text-gray-600">Fix:</span> {gap.fix}</p>
              <p className="text-gray-600 font-mono text-[9px]">{gap.paths.join(' · ')}</p>
            </div>
          ))}
        </div>
      </Card>

      <p className="text-[10px] text-gray-600">
        Source: <code className="text-gray-500">docs/CAPABILITIES.md</code> (SI telemetry) +{' '}
        <code className="text-gray-500">capability-matrix.ts</code> (product matrix). Last updated with Phase 0–1.
      </p>
    </div>
  );
}

function LegendChip({ label, className }: { label: string; className: string }) {
  return (
    <span className={`px-2 py-0.5 rounded bg-gray-800 border border-gray-700 ${className}`}>
      {label}
    </span>
  );
}
