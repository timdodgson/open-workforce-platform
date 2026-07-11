'use client';

import Link from 'next/link';
import RunList from '@/app/RunList';
import Card from '@/components/Card';
import type { RunListEntry } from '@/lib/data-loader';

export interface LabStats {
  totalRuns: number;
  valRuns: number;
  domains: Record<string, number>;
}

const LAB_ENTRIES = [
  { href: '/benchmarks', label: 'Benchmarks', desc: 'Algorithm leaderboard by instance and domain', icon: '🏆' },
  { href: '/experiment-matrix', label: 'Experiment Matrix', desc: 'Every run variation — options on/off and why', icon: '🧪' },
  { href: '/lab/byod', label: 'BYOD / BYOA', desc: 'Solver registry, owp-sdk contract, TSP demo', icon: '🔌' },
  { href: '/statistics', label: 'Statistics', desc: 'Welch t-test, effect sizes, SI comparisons', icon: '📊' },
  { href: '/runs', label: 'All Runs', desc: 'Browse and filter every stored experiment', icon: '📂' },
  { href: '/intelligence', label: 'Search Intelligence', desc: 'Policies, telemetry, training artifacts', icon: '🧠' },
  { href: '/capabilities', label: 'Capabilities', desc: 'Platform matrix — what works per domain', icon: '✅' },
  { href: '/compare', label: 'Compare', desc: 'Side-by-side run comparison', icon: '🔀' },
  { href: '/knowledge', label: 'Knowledge Base', desc: 'Documentation and reference material', icon: '📚' },
] as const;

export default function LabHub({ runs, stats }: { runs: RunListEntry[]; stats: LabStats }) {
  const domainEntries = Object.entries(stats.domains).sort((a, b) => b[1] - a[1]);

  return (
    <div className="space-y-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-[10px] uppercase tracking-wider text-blue-500 font-semibold mb-1">Research Lab</p>
          <h1 className="text-2xl font-bold text-gray-100">Metrics, runs &amp; validation</h1>
          <p className="text-sm text-gray-400 mt-2 max-w-2xl">
            This is the working lab — benchmarks, statistical tests, experiment configs, and live run data.
            The public site at{' '}
            <Link href="/" className="text-blue-400 hover:underline">pfrs-lab.com</Link>
            {' '}is the front door; everything operational lives here.
          </p>
        </div>
        <Link
          href="/"
          className="text-xs text-gray-500 hover:text-gray-300 border border-gray-800 rounded-lg px-3 py-2 transition-colors"
        >
          ← Public site
        </Link>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <Card title="Total runs">
          <p className="text-3xl font-bold text-gray-100">{stats.totalRuns.toLocaleString()}</p>
          <p className="text-[10px] text-gray-500 mt-1">in storage manifest</p>
        </Card>
        <Card title="SI validation">
          <p className="text-3xl font-bold text-gray-100">{stats.valRuns.toLocaleString()}</p>
          <p className="text-[10px] text-gray-500 mt-1">val-* canonical runs</p>
        </Card>
        <Card title="Domains">
          <p className="text-3xl font-bold text-gray-100">{domainEntries.length}</p>
          <p className="text-[10px] text-gray-500 mt-1">problem types represented</p>
        </Card>
        <Card title="Experiment grid">
          <p className="text-3xl font-bold text-gray-100">288</p>
          <p className="text-[10px] text-gray-500 mt-1">canonical fast + deep variations</p>
        </Card>
      </div>

      {domainEntries.length > 0 && (
        <Card title="Runs by domain">
          <div className="flex flex-wrap gap-3">
            {domainEntries.map(([domain, count]) => (
              <div key={domain} className="px-3 py-2 rounded-lg bg-gray-900 border border-gray-800">
                <span className="text-xs font-bold text-gray-200 uppercase">{domain}</span>
                <span className="text-[10px] text-gray-500 ml-2">{count.toLocaleString()} runs</span>
              </div>
            ))}
          </div>
        </Card>
      )}

      <div>
        <h2 className="text-sm font-semibold text-gray-300 mb-3">Lab tools</h2>
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {LAB_ENTRIES.map(({ href, label, desc, icon }) => (
            <Link
              key={href}
              href={href}
              className="group p-4 rounded-lg border border-gray-800 bg-gray-900/50 hover:border-gray-600 hover:bg-gray-900 transition-colors"
            >
              <span className="text-lg" aria-hidden>{icon}</span>
              <p className="text-sm font-medium text-gray-200 mt-2 group-hover:text-blue-300 transition-colors">
                {label}
              </p>
              <p className="text-[11px] text-gray-500 mt-1 leading-relaxed">{desc}</p>
            </Link>
          ))}
        </div>
      </div>

      {runs.length > 0 && (
        <section>
          <RunList runs={runs} />
        </section>
      )}
    </div>
  );
}
