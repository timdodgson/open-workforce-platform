'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import Card from '@/components/Card';
import {
  CLI_FLAGS,
  FLAG_GROUPS,
  PREREQUISITES,
  WORKED_EXAMPLES,
  type CliFlag,
  type FlagGroupId,
} from '@/lib/cli-reference';

function CopyBlock({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="relative group">
      <pre className="text-[11px] leading-relaxed text-gray-200 bg-gray-950 border border-gray-800 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap font-mono">
        {text}
      </pre>
      <button
        type="button"
        onClick={async () => {
          await navigator.clipboard.writeText(text);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1500);
        }}
        className="absolute top-2 right-2 text-[10px] px-2 py-1 rounded border border-gray-700 bg-gray-900 text-gray-400 hover:text-gray-200"
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  );
}

function FlagBadges({ flag }: { flag: CliFlag }) {
  return (
    <div className="flex flex-wrap gap-1.5 mt-2">
      <span className="text-[9px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-gray-800 text-gray-400">
        {flag.commands === 'both' ? 'solve + tune-pfrs' : flag.commands}
      </span>
      {flag.algorithms?.map((a) => (
        <span key={a} className="text-[9px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-indigo-950/60 text-indigo-300 border border-indigo-900/50">
          {a}
        </span>
      ))}
      {flag.dependsOn?.map((d) => (
        <span key={d} className="text-[9px] px-1.5 py-0.5 rounded bg-amber-950/50 text-amber-200/90 border border-amber-900/40">
          needs: {d}
        </span>
      ))}
      {flag.pairsWith?.map((p) => (
        <span key={p} className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-950/40 text-emerald-300/90 border border-emerald-900/40">
          pairs: {p}
        </span>
      ))}
    </div>
  );
}

function FlagRow({ flag }: { flag: CliFlag }) {
  return (
    <div className="border-b border-gray-800/80 py-3 last:border-0">
      <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <code className="text-sm text-blue-300 font-mono">{flag.flag}</code>
        <span className="text-[11px] text-gray-500 font-mono">{flag.values}</span>
        {flag.defaultValue && (
          <span className="text-[10px] text-gray-600">default {flag.defaultValue}</span>
        )}
      </div>
      <p className="text-sm text-gray-300 mt-1">{flag.summary}</p>
      {flag.detail && <p className="text-xs text-gray-500 mt-1 leading-relaxed">{flag.detail}</p>}
      {flag.note && <p className="text-xs text-amber-200/70 mt-1">{flag.note}</p>}
      <FlagBadges flag={flag} />
    </div>
  );
}

export default function GettingStartedView() {
  const [query, setQuery] = useState('');
  const [groupFilter, setGroupFilter] = useState<FlagGroupId | 'all'>('all');

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return CLI_FLAGS.filter((f) => {
      if (groupFilter !== 'all' && f.group !== groupFilter) return false;
      if (!q) return true;
      const hay = [f.flag, f.summary, f.detail, f.values, ...(f.algorithms ?? []), ...(f.dependsOn ?? [])]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      return hay.includes(q);
    });
  }, [query, groupFilter]);

  const byGroup = useMemo(() => {
    const map = new Map<FlagGroupId, CliFlag[]>();
    for (const f of filtered) {
      const list = map.get(f.group) ?? [];
      list.push(f);
      map.set(f.group, list);
    }
    return map;
  }, [filtered]);

  return (
    <div className="space-y-8 max-w-4xl">
      <div>
        <p className="text-[10px] uppercase tracking-wider text-blue-500 font-semibold mb-1">Onboarding</p>
        <h1 className="text-2xl font-bold text-gray-100">Getting started</h1>
        <p className="text-sm text-gray-400 mt-2 max-w-2xl leading-relaxed">
          Run the platform end-to-end: install tooling with clear reasons, execute worked examples
          across domains, then use the switch reference when you change algorithms or turn on
          Search Intelligence. Written for hiring reviewers and for graduates who can follow a
          serious research codebase.
        </p>
        <nav className="flex flex-wrap gap-3 mt-4 text-xs">
          <a href="#prerequisites" className="text-blue-400 hover:underline">Prerequisites</a>
          <a href="#examples" className="text-blue-400 hover:underline">Worked examples</a>
          <a href="#flags" className="text-blue-400 hover:underline">Switch reference</a>
          <a href="#mental-model" className="text-blue-400 hover:underline">Mental model</a>
          <Link href="/algorithms" className="text-blue-400 hover:underline">Algorithms →</Link>
          <Link href="/domains" className="text-blue-400 hover:underline">Domains →</Link>
          <Link href="/references" className="text-blue-400 hover:underline">References →</Link>
          <Link href="/research" className="text-blue-400 hover:underline">Research depth →</Link>
        </nav>
      </div>

      <Card title="What you are running">
        <p className="text-sm text-gray-400 leading-relaxed">
          Domains (NRP, CVRP, JSS, VRPTW) implement one generic search interface. Algorithms
          (SA, LAHC, Tabu, GA, Portfolio) plug into that interface. Optional Search Intelligence
          layers can guide compute — they are off unless you set flags. The lab dashboard visualises
          run folders produced by the CLI.
        </p>
      </Card>

      <section id="prerequisites" className="scroll-mt-20">
        <h2 className="text-lg font-semibold text-gray-100 mb-3">1. Prerequisites</h2>
        <div className="space-y-3">
          {PREREQUISITES.map((p) => (
            <Card key={p.title} title={p.title}>
              <p className="text-xs text-gray-500 mb-2">
                <span className="text-gray-400 font-medium">Why: </span>
                {p.why}
              </p>
              <p className="text-xs text-gray-400">
                <span className="text-gray-300 font-medium">How: </span>
                {p.how}
              </p>
            </Card>
          ))}
        </div>
        <div className="mt-3 text-xs text-gray-500 leading-relaxed border border-gray-800 rounded-lg p-3 bg-gray-900/50">
          <strong className="text-gray-300">Windows note:</strong> PowerShell accepts line continuations with{' '}
          <code className="text-blue-300">`</code>. Bash/macOS/Linux use <code className="text-blue-300">\</code>.
          All examples below are single-line so either shell works.
        </div>
      </section>

      <section id="mental-model" className="scroll-mt-20">
        <h2 className="text-lg font-semibold text-gray-100 mb-3">2. Two commands, different jobs</h2>
        <div className="grid md:grid-cols-2 gap-3">
          <Card title="owp solve &lt;domain&gt;">
            <p className="text-sm text-gray-400 leading-relaxed">
              Single-instance metaheuristic search. Use for CVRP, VRPTW, JSS, and single-week NRP
              experiments. Fast feedback loop for algorithm comparison.
            </p>
            <p className="text-[11px] text-gray-500 mt-2 font-mono">owp solve cvrp|vrptw|jobshop|nrp …</p>
          </Card>
          <Card title="owp tune-pfrs">
            <p className="text-sm text-gray-400 leading-relaxed">
              Multi-week NRP beam search (PFRS). This is the flagship path — workers branch on
              improving rosters across an 8-week horizon. Beam flags only apply here.
            </p>
            <p className="text-[11px] text-gray-500 mt-2 font-mono">owp tune-pfrs --instance n012w8 …</p>
          </Card>
        </div>
      </section>

      <section id="examples" className="scroll-mt-20">
        <h2 className="text-lg font-semibold text-gray-100 mb-1">3. Worked examples</h2>
        <p className="text-sm text-gray-500 mb-4">
          Run these from <code className="text-gray-400">platform/go</code>. Start with CVRP; graduate to NRP beam when comfortable.
        </p>
        <div className="space-y-4">
          {WORKED_EXAMPLES.map((ex, i) => (
            <Card key={ex.id} title={`${i + 1}. ${ex.title}`}>
              <div className="flex flex-wrap gap-2 mb-2 text-[10px]">
                <span className="px-1.5 py-0.5 rounded bg-gray-800 text-gray-300">{ex.domain}</span>
                <span className="px-1.5 py-0.5 rounded bg-gray-800 text-gray-300">{ex.algorithm}</span>
                <span className="px-1.5 py-0.5 rounded bg-gray-800 text-gray-500">{ex.timeHint}</span>
              </div>
              <p className="text-xs text-gray-400 mb-1"><span className="text-gray-300">Why this example: </span>{ex.why}</p>
              <p className="text-xs text-gray-500 mb-3"><span className="text-gray-400">What “good” looks like: </span>{ex.expected}</p>
              <p className="text-[10px] text-gray-600 mb-1">cwd: {ex.cwd}</p>
              <CopyBlock text={ex.command} />
            </Card>
          ))}
        </div>
      </section>

      <section id="flags" className="scroll-mt-20">
        <h2 className="text-lg font-semibold text-gray-100 mb-1">4. Switch reference</h2>
        <p className="text-sm text-gray-500 mb-4">
          Amber badges mean a dependency. Indigo badges mean algorithm-specific. Prefer changing one
          lever at a time when comparing to a published baseline.
        </p>

        <div className="flex flex-col sm:flex-row gap-2 mb-4">
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter flags (e.g. beam, ga, policy)…"
            className="flex-1 text-sm bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-gray-200 placeholder:text-gray-600 focus:outline-none focus:border-blue-700"
          />
          <select
            value={groupFilter}
            onChange={(e) => setGroupFilter(e.target.value as FlagGroupId | 'all')}
            className="text-sm bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-gray-300"
          >
            <option value="all">All groups</option>
            {FLAG_GROUPS.map((g) => (
              <option key={g.id} value={g.id}>{g.title}</option>
            ))}
          </select>
        </div>

        <p className="text-[11px] text-gray-600 mb-4">{filtered.length} switches shown</p>

        {FLAG_GROUPS.map((g) => {
          const flags = byGroup.get(g.id);
          if (!flags?.length) return null;
          return (
            <Card key={g.id} title={g.title}>
              <p className="text-xs text-gray-500 mb-2 leading-relaxed">{g.blurb}</p>
              {flags.map((f) => (
                <FlagRow key={f.flag} flag={f} />
              ))}
            </Card>
          );
        })}
      </section>

      <Card title="Where results land">
        <ul className="text-sm text-gray-400 space-y-2 list-disc pl-5">
          <li>
            <code className="text-gray-300">tune-pfrs --pfrs-run-label NAME</code> →{' '}
            <code className="text-gray-500">platform/web/pfrs-lab/data/runs/NAME/</code>
          </li>
          <li>
            Open the lab{' '}
            <Link href="/runs" className="text-blue-400 hover:underline">All Runs</Link>
            {' '}view after a labelled run to inspect workers, discoveries, and official penalty.
          </li>
          <li>
            Deeper algorithms and SI design:{' '}
            <Link href="/research" className="text-blue-400 hover:underline">/research</Link>
            {' '}· reproducibility ladder:{' '}
            <Link href="/reproduce" className="text-blue-400 hover:underline">/reproduce</Link>
          </li>
        </ul>
      </Card>
    </div>
  );
}
