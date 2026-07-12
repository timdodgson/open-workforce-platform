'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import Card from '@/components/Card';
import {
  BENCHMARK_LADDER_NOTES,
  COMPOSITE_MODES,
  CORE_ALGORITHMS,
  OPTION_LAYERS,
  VARIATION_CONFIGS,
  matrixSummary,
  optionStateClass,
  optionStateLabel,
  type MatrixTier,
  type OptionState,
  type ProblemType,
  type VariationConfig,
} from '@/lib/experiment-matrix';

export interface ConfigCoverage {
  configId: string;
  found: number;
  expected: number;
}

const DOMAIN_FILTERS: Array<{ id: 'all' | ProblemType; label: string }> = [
  { id: 'all', label: 'All domains' },
  { id: 'nrp', label: 'NRP' },
  { id: 'cvrp', label: 'CVRP' },
  { id: 'jss', label: 'JSS' },
  { id: 'vrptw', label: 'VRPTW' },
];

const TIER_FILTERS: Array<{ id: 'all' | MatrixTier; label: string }> = [
  { id: 'all', label: 'All tiers' },
  { id: 'fast', label: 'Fast soak (240)' },
  { id: 'deep', label: 'Deep soak (48)' },
];

function OptionBadge({ state }: { state: OptionState }) {
  return (
    <span className={`inline-block text-[9px] px-1.5 py-0.5 rounded border font-mono ${optionStateClass(state)}`}>
      {optionStateLabel(state)}
    </span>
  );
}

function ConfigCard({
  config,
  coverage,
  expanded,
  onToggle,
}: {
  config: VariationConfig;
  coverage?: ConfigCoverage;
  expanded: boolean;
  onToggle: () => void;
}) {
  const pct = coverage && coverage.expected > 0
    ? Math.round((coverage.found / coverage.expected) * 100)
    : null;

  const exampleLabels = useMemo(() => {
    const examples = [
      config.labelPattern.replace('{policy}', 'rules').replace('{seed}', String(config.seeds[0])),
      config.labelPattern.replace('{policy}', 'learned').replace('{seed}', String(config.seeds[config.seeds.length - 1])),
    ];
    return examples;
  }, [config]);

  return (
    <div className="border border-gray-800 rounded-lg overflow-hidden bg-gray-900/40">
      <button
        type="button"
        onClick={onToggle}
        className="w-full text-left px-4 py-3 flex flex-wrap items-center gap-2 hover:bg-gray-800/40 transition-colors"
      >
        <span className="text-[10px] uppercase tracking-wider text-gray-500 w-12">{config.domain}</span>
        <span className="text-[10px] uppercase tracking-wider text-blue-500/80 w-14">{config.tier}</span>
        <span className="text-sm text-gray-100 font-medium flex-1 min-w-[200px]">{config.title}</span>
        <span className="text-[10px] font-mono text-gray-500">{config.primaryMode}</span>
        <span className="text-[10px] text-gray-600">{config.variationsPerConfig} runs</span>
        {coverage && (
          <span
            className={`text-[10px] px-2 py-0.5 rounded ${
              coverage.found === coverage.expected
                ? 'bg-emerald-950 text-emerald-400'
                : coverage.found > 0
                  ? 'bg-amber-950 text-amber-400'
                  : 'bg-gray-800 text-gray-500'
            }`}
          >
            {coverage.found}/{coverage.expected} in storage{pct !== null ? ` (${pct}%)` : ''}
          </span>
        )}
        <span className="text-gray-600 text-xs">{expanded ? '▾' : '▸'}</span>
      </button>

      {expanded && (
        <div className="px-4 pb-4 space-y-4 border-t border-gray-800/80">
          <div className="grid md:grid-cols-2 gap-4 pt-3">
            <div>
              <p className="text-[10px] uppercase text-gray-500 mb-1">Why this config</p>
              <p className="text-xs text-gray-300">{config.whyThisConfig}</p>
            </div>
            <div>
              <p className="text-[10px] uppercase text-gray-500 mb-1">Why not other modes</p>
              <p className="text-xs text-gray-400">{config.whyNotOthers}</p>
            </div>
          </div>

          <div className="text-xs text-gray-500 font-mono bg-gray-950/80 rounded p-3 border border-gray-800">
            <p>
              <span className="text-gray-600">CLI: </span>
              owp {config.command === 'solve' ? `solve ${config.solveDomain}` : 'tune-pfrs'}
              {' '}--instance {config.instance} --{config.command === 'solve' ? 'mode' : 'pfrs-mode'} {config.primaryMode}
              {' '}… {config.iterations} iterations
            </p>
            <p className="mt-1">
              <span className="text-gray-600">Script: </span>
              {config.script}
            </p>
          </div>

          <div>
            <p className="text-[10px] uppercase text-gray-500 mb-2">Options for this variation</p>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-gray-500 border-b border-gray-800">
                    <th className="text-left p-2">Flag</th>
                    <th className="text-left p-2">State</th>
                    <th className="text-left p-2">Value</th>
                    <th className="text-left p-2">Why</th>
                  </tr>
                </thead>
                <tbody>
                  {config.options.map((opt) => (
                    <tr key={opt.flag} className="border-b border-gray-800/50">
                      <td className="p-2 font-mono text-gray-300">{opt.flag}</td>
                      <td className="p-2">
                        <OptionBadge state={opt.state} />
                      </td>
                      <td className="p-2 text-gray-400 font-mono text-[10px]">{opt.value ?? '—'}</td>
                      <td className="p-2 text-gray-400">{opt.rationale}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div>
            <p className="text-[10px] uppercase text-gray-500 mb-1">Run label pattern</p>
            <p className="font-mono text-[11px] text-blue-300/90">{config.labelPattern}</p>
            <p className="text-[10px] text-gray-600 mt-1">
              Examples:{' '}
              {exampleLabels.map((l, i) => (
                <span key={l}>
                  {i > 0 && ' · '}
                  <Link href={`/runs/${l}`} className="text-blue-400 hover:underline">{l}</Link>
                </span>
              ))}
            </p>
            <p className="text-[10px] text-gray-600 mt-1">
              Policies: {config.policies.join(', ')} · Seeds: {config.seeds.length} ({config.seeds.slice(0, 3).join(', ')}
              {config.seeds.length > 3 ? ', …' : ''})
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

export default function ExperimentMatrixView({ coverage }: { coverage: ConfigCoverage[] }) {
  const [domainFilter, setDomainFilter] = useState<'all' | ProblemType>('all');
  const [tierFilter, setTierFilter] = useState<'all' | MatrixTier>('all');
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const summary = matrixSummary();
  const coverageMap = useMemo(
    () => new Map(coverage.map((c) => [c.configId, c])),
    [coverage],
  );

  const filtered = useMemo(() => {
    return VARIATION_CONFIGS.filter((c) => {
      if (domainFilter !== 'all' && c.domain !== domainFilter) return false;
      if (tierFilter !== 'all' && c.tier !== tierFilter) return false;
      return true;
    });
  }, [domainFilter, tierFilter]);

  const totalFound = coverage.reduce((n, c) => n + c.found, 0);
  const totalExpected = coverage.reduce((n, c) => n + c.expected, 0);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-gray-100">Experiment Matrix</h1>
        <p className="text-sm text-gray-400 mt-1 max-w-3xl">
          Canonical catalog of every standard run variation: {summary.domains} domains, {summary.coreAlgorithms} core
          algorithms (SA, LAHC, Tabu, GA), plus portfolio and adaptive composites. Each row explains which CLI options are
          ON, OFF, or held at defaults — and why.
        </p>
        <div className="flex flex-wrap gap-3 mt-3 text-xs">
          <Link href="/capabilities" className="text-blue-400 hover:underline">Capabilities →</Link>
          <Link href="/benchmarks" className="text-blue-400 hover:underline">Benchmark ladder →</Link>
          <Link href="/runs" className="text-blue-400 hover:underline">All runs →</Link>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <Card title="Fast soak">
          <p className="text-2xl font-semibold text-gray-100">{summary.fastVariations}</p>
          <p className="text-[10px] text-gray-500">{summary.fastConfigs} configs × 3 policies × 10 seeds</p>
        </Card>
        <Card title="Deep soak">
          <p className="text-2xl font-semibold text-gray-100">{summary.deepVariations}</p>
          <p className="text-[10px] text-gray-500">{summary.deepConfigs} configs × 3 policies × 2 seeds</p>
        </Card>
        <Card title="In storage">
          <p className="text-2xl font-semibold text-gray-100">{totalFound}</p>
          <p className="text-[10px] text-gray-500">of {totalExpected} canonical val-* labels</p>
        </Card>
        <Card title="Domains">
          <p className="text-2xl font-semibold text-gray-100">4</p>
          <p className="text-[10px] text-gray-500">NRP · CVRP · JSS · VRPTW</p>
        </Card>
      </div>

      <Card title="Four core algorithms + composites">
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          {CORE_ALGORITHMS.map((alg) => (
            <div key={alg.id} className="p-3 rounded border border-gray-800 bg-gray-950/50">
              <p className="text-sm font-medium text-gray-200">{alg.label}</p>
              <p className="text-[11px] text-gray-500 mt-1">{alg.description}</p>
              <ul className="mt-2 space-y-1">
                {alg.params.map((p) => (
                  <li key={p.flag} className="text-[10px] text-gray-600 font-mono">
                    {p.flag} <span className="text-gray-500">({p.default})</span> — {p.purpose}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div className="grid md:grid-cols-2 gap-3">
          {COMPOSITE_MODES.map((m) => (
            <div key={m.id} className="p-3 rounded border border-gray-800/60">
              <p className="text-sm text-gray-300">{m.label}</p>
              <p className="text-[11px] text-gray-500 mt-1">{m.description}</p>
              <p className="text-[10px] text-gray-600 mt-1">{m.whenUsed}</p>
            </div>
          ))}
        </div>
      </Card>

      <Card title="Option layers — when ON vs OFF">
        <div className="space-y-4">
          {OPTION_LAYERS.map((layer) => (
            <div key={layer.id} className="border-b border-gray-800/60 pb-4 last:border-0 last:pb-0">
              <p className="text-sm font-medium text-gray-200">{layer.title}</p>
              <p className="text-[10px] text-gray-600 mb-2">{layer.appliesTo}</p>
              {layer.flags.map((f) => (
                <div key={f.flag} className="mb-3 pl-2 border-l-2 border-gray-800">
                  <p className="font-mono text-xs text-blue-300/90">{f.flag}</p>
                  <p className="text-[10px] text-gray-500">Values: {f.values.join(' · ')}</p>
                  <div className="grid md:grid-cols-3 gap-2 mt-1 text-[11px]">
                    <p><span className="text-emerald-600">ON: </span>{f.whenOn}</p>
                    <p><span className="text-gray-500">OFF: </span>{f.whenOff}</p>
                    <p><span className="text-gray-600">Why: </span>{f.why}</p>
                  </div>
                </div>
              ))}
            </div>
          ))}
        </div>
      </Card>

      <Card title="Benchmark ladder (non-val runs)">
        <p className="text-xs text-gray-500 mb-3">
          Algorithm leaderboard runs (EXP-002+) typically leave SI flags OFF to measure pure search quality.
        </p>
        <div className="grid md:grid-cols-2 gap-3">
          {(Object.entries(BENCHMARK_LADDER_NOTES) as Array<[ProblemType, string]>).map(([domain, note]) => (
            <div key={domain} className="text-xs p-3 rounded bg-gray-950/60 border border-gray-800">
              <span className="uppercase text-[10px] text-gray-500">{domain}</span>
              <p className="text-gray-400 mt-1">{note}</p>
            </div>
          ))}
        </div>
      </Card>

      <div>
        <div className="flex flex-wrap items-center gap-3 p-3 rounded-lg bg-gray-900/80 border border-gray-800 mb-4">
          <span className="text-[10px] text-gray-500 uppercase">Filter</span>
          {DOMAIN_FILTERS.map((d) => (
            <button
              key={d.id}
              type="button"
              onClick={() => setDomainFilter(d.id)}
              className={`text-xs px-3 py-1 rounded border transition-colors ${
                domainFilter === d.id
                  ? 'bg-blue-900/50 text-blue-300 border-blue-700'
                  : 'bg-gray-800 text-gray-400 border-gray-700 hover:border-gray-600'
              }`}
            >
              {d.label}
            </button>
          ))}
          <span className="text-gray-700">|</span>
          {TIER_FILTERS.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setTierFilter(t.id)}
              className={`text-xs px-3 py-1 rounded border transition-colors ${
                tierFilter === t.id
                  ? 'bg-purple-900/40 text-purple-300 border-purple-700'
                  : 'bg-gray-800 text-gray-400 border-gray-700 hover:border-gray-600'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        <p className="text-[10px] uppercase text-gray-500 mb-2">
          Canonical variations ({filtered.length} configs · {filtered.reduce((n, c) => n + c.variationsPerConfig, 0)} runs)
        </p>
        <div className="space-y-2">
          {filtered.map((config) => (
            <ConfigCard
              key={config.id}
              config={config}
              coverage={coverageMap.get(config.id)}
              expanded={expandedId === config.id}
              onToggle={() => setExpandedId(expandedId === config.id ? null : config.id)}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
