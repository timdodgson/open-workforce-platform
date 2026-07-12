'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import Card from '@/components/Card';
import {
  ADVANCED_FLAG_COUNT,
  CLI_FLAGS,
  ESSENTIAL_FLAG_COUNT,
  FLAG_GROUPS,
  PREREQUISITES,
  WORKED_EXAMPLES,
  type CliFlag,
  type FlagGroupId,
  type FlagTier,
} from '@/lib/cli-reference';

const QUICK_START = WORKED_EXAMPLES.find((ex) => ex.id === 'cvrp-lahc')!;

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
        {flag.tier === 'advanced' && (
          <span className="text-[9px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-gray-800/80 text-gray-500">
            advanced
          </span>
        )}
      </div>
      <p className="text-sm text-gray-300 mt-1">{flag.summary}</p>
      {flag.detail && <p className="text-xs text-gray-500 mt-1 leading-relaxed">{flag.detail}</p>}
      {flag.note && <p className="text-xs text-amber-200/70 mt-1">{flag.note}</p>}
      <FlagBadges flag={flag} />
    </div>
  );
}

function FlagGroupCards({ flags }: { flags: CliFlag[] }) {
  const byGroup = useMemo(() => {
    const map = new Map<FlagGroupId, CliFlag[]>();
    for (const f of flags) {
      const list = map.get(f.group) ?? [];
      list.push(f);
      map.set(f.group, list);
    }
    return map;
  }, [flags]);

  return (
    <div className="space-y-4">
      {FLAG_GROUPS.map((g) => {
        const groupFlags = byGroup.get(g.id);
        if (!groupFlags?.length) return null;
        return (
          <Card key={g.id} title={g.title}>
            <p className="text-xs text-gray-500 mb-2 leading-relaxed">{g.blurb}</p>
            {groupFlags.map((f) => (
              <FlagRow key={f.flag} flag={f} />
            ))}
          </Card>
        );
      })}
    </div>
  );
}

export default function GettingStartedView() {
  const [query, setQuery] = useState('');
  const [groupFilter, setGroupFilter] = useState<FlagGroupId | 'all'>('all');
  const [tierFilter, setTierFilter] = useState<FlagTier | 'all'>('essential');
  const [advancedOpen, setAdvancedOpen] = useState(false);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return CLI_FLAGS.filter((f) => {
      if (groupFilter !== 'all' && f.group !== groupFilter) return false;
      if (tierFilter !== 'all' && f.tier !== tierFilter) return false;
      if (!q) return true;
      const hay = [f.flag, f.summary, f.detail, f.values, ...(f.algorithms ?? []), ...(f.dependsOn ?? [])]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      return hay.includes(q);
    });
  }, [query, groupFilter, tierFilter]);

  const essentialFiltered = useMemo(
    () => filtered.filter((f) => f.tier === 'essential'),
    [filtered],
  );
  const advancedFiltered = useMemo(
    () => filtered.filter((f) => f.tier === 'advanced'),
    [filtered],
  );

  const searching = query.trim().length > 0;
  const showSplit = tierFilter === 'all' && !searching;
  const showEssentialBlock = tierFilter === 'essential' || showSplit;

  const deeperExamples = WORKED_EXAMPLES.filter((ex) => ex.id !== 'cvrp-lahc');

  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">Onboarding</p>
        <h1 className="site-title site-title--single">Getting started</h1>
        <p className="site-lead">
          First win in about five minutes: one published CVRP instance, one distance to check.
          Everything below is optional depth — examples, NRP beam, and the full CLI reference.
        </p>
        <div className="site-hero-actions site-hero-actions--center">
          <a href="#quick-start" className="site-btn-primary">Quick start</a>
          <a href="#paths" className="site-btn-secondary">Then go deeper</a>
          <a href="#flags" className="site-btn-secondary">CLI reference</a>
        </div>
      </section>

      <section id="quick-start" className="site-section site-section--panel scroll-mt-24">
        <h2 className="site-heading">Quick start (~5 minutes)</h2>
        <p className="site-body">
          Prove the install. You need Go 1.22+, a clone of this repo, and a terminal in{' '}
          <code className="site-code">platform/go</code>. Details under{' '}
          <a href="#prerequisites" className="site-inline-link">Prerequisites</a> if anything fails.
        </p>

        <div className="site-quick-steps">
          <div className="site-byod-step">
            <span className="site-byod-step-num">1</span>
            <div>
              <p className="site-byod-step-title">Install &amp; enter the module</p>
              <p className="site-byod-step-body">
                <code className="site-code">go version</code> should print 1.22+.
                Then <code className="site-code">cd platform/go</code>.
              </p>
            </div>
          </div>
          <div className="site-byod-step">
            <span className="site-byod-step-num">2</span>
            <div>
              <p className="site-byod-step-title">Run one command</p>
              <p className="site-byod-step-body">
                CVRPLIB A-n32-k5 with LAHC — typically finishes in a few seconds.
              </p>
            </div>
          </div>
          <div className="site-byod-step">
            <span className="site-byod-step-num">3</span>
            <div>
              <p className="site-byod-step-title">Check the number</p>
              <p className="site-byod-step-body">
                Published optimum distance is <strong>784</strong>. A short run should land in the high 700s / low 800s.
              </p>
            </div>
          </div>
        </div>

        <p className="text-[10px] text-gray-600 mb-1">cwd: {QUICK_START.cwd}</p>
        <CopyBlock text={QUICK_START.command} />

        <div className="site-quick-success">
          <p>
            <strong>Success:</strong> the solver prints a feasible distance near 784–820.
            If you got a number, the platform is working — stop here if that was all you needed.
          </p>
          <p>
            Browse live runs anytime in the{' '}
            <Link href="/lab" className="site-inline-link">lab</Link>
            {' '}· cite or reproduce via{' '}
            <Link href="/reproduce" className="site-inline-link">/reproduce</Link>.
          </p>
        </div>
      </section>

      <section id="paths" className="site-section scroll-mt-24">
        <h2 className="site-heading">Where to go next</h2>
        <p className="site-body">
          Same codebase, three common routes. Pick one — you do not need the full switch list first.
        </p>
        <div className="site-path-grid">
          <article className="site-path-card">
            <h3>Students</h3>
            <p>More short domain examples (JSS, VRPTW, GA) and how solve differs from NRP beam.</p>
            <a href="#examples" className="site-inline-link">Worked examples →</a>
          </article>
          <article className="site-path-card">
            <h3>Researchers</h3>
            <p>Flagship multi-week NRP, Search Intelligence optional, citation and experiment ladder.</p>
            <a href="#nrp" className="site-inline-link">NRP beam path →</a>
            {' · '}
            <Link href="/reproduce" className="site-inline-link">Reproduce</Link>
          </article>
          <article className="site-path-card">
            <h3>Engineers</h3>
            <p>Dependency-aware CLI / PFRS parameter reference when you change one lever at a time.</p>
            <a href="#flags" className="site-inline-link">Switch reference →</a>
          </article>
        </div>
      </section>

      <section id="prerequisites" className="site-section scroll-mt-24 space-y-4">
        <h2 className="site-heading">Prerequisites</h2>
        <p className="site-body">
          Only if Quick start failed or you want the “why” behind each tool.
        </p>
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
        <div className="text-xs text-gray-500 leading-relaxed border border-gray-800 rounded-lg p-3 bg-gray-900/50">
          <strong className="text-gray-300">Windows note:</strong> PowerShell accepts line continuations with{' '}
          <code className="site-code">`</code>. Bash/macOS/Linux use <code className="site-code">\</code>.
          All examples below are single-line so either shell works.
        </div>
      </section>

      <section id="mental-model" className="site-section scroll-mt-24 space-y-4">
        <h2 className="site-heading">Two commands, different jobs</h2>
        <p className="site-body">
          Domains (NRP, CVRP, JSS, VRPTW) share one move / evaluate / undo interface. Algorithms
          (SA, LAHC, Tabu, GA, Portfolio) plug into that interface. Search Intelligence is off unless you set flags.
        </p>
        <div className="grid md:grid-cols-2 gap-3">
          <Card title="owp solve &lt;domain&gt;">
            <p className="text-sm text-gray-400 leading-relaxed">
              Single-instance metaheuristic search. Use for CVRP, VRPTW, JSS, and single-week NRP
              experiments. Fast feedback loop for algorithm comparison.
            </p>
            <p className="text-[11px] text-gray-500 mt-2 font-mono">go run ./cmd/owp solve cvrp|vrptw|jobshop|nrp …</p>
          </Card>
          <Card title="owp tune-pfrs">
            <p className="text-sm text-gray-400 leading-relaxed">
              Multi-week NRP beam search (PFRS). This is the flagship path — workers branch on
              improving rosters across an 8-week horizon. Beam flags only apply here.
            </p>
            <p className="text-[11px] text-gray-500 mt-2 font-mono">go run ./cmd/owp tune-pfrs --instance n012w8 …</p>
          </Card>
        </div>
      </section>

      <section id="examples" className="site-section scroll-mt-24 space-y-4">
        <h2 className="site-heading">More worked examples</h2>
        <p className="site-body">
          You already ran CVRP + LAHC above. These add Job Shop, VRPTW, GA, and NRP beam from{' '}
          <code className="site-code">platform/go</code>.
        </p>
        <div className="space-y-4">
          {deeperExamples.map((ex, i) => {
            const isFirstNrp = ex.id.startsWith('nrp')
              && !deeperExamples.slice(0, i).some((e) => e.id.startsWith('nrp'));
            return (
              <div key={ex.id} id={isFirstNrp ? 'nrp' : undefined} className={isFirstNrp ? 'scroll-mt-24' : undefined}>
                <Card title={`${i + 1}. ${ex.title}`}>
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
              </div>
            );
          })}
        </div>
      </section>

      <section id="flags" className="site-section scroll-mt-24 space-y-4">
        <h2 className="site-heading">Switch reference</h2>
        <p className="site-body">
          Start with <strong className="text-gray-300">essential</strong> levers — the ones in Quick start
          and published recipes. Advanced knobs (SA/LAHC/Tabu/GA internals, refinement, rare beam
          options) are still here; leave their defaults unless you are running an ablation.
          Amber badges mean a dependency; indigo means algorithm-specific.
        </p>

        <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
          <div className="flex flex-wrap gap-2">
            {(
              [
                { id: 'essential' as const, label: `Essential (${ESSENTIAL_FLAG_COUNT})` },
                { id: 'advanced' as const, label: `Advanced (${ADVANCED_FLAG_COUNT})` },
                { id: 'all' as const, label: `All (${CLI_FLAGS.length})` },
              ]
            ).map((opt) => (
              <button
                key={opt.id}
                type="button"
                onClick={() => {
                  setTierFilter(opt.id);
                  if (opt.id === 'advanced') setAdvancedOpen(true);
                }}
                className={`text-xs px-3 py-1.5 rounded-lg border transition-colors ${
                  tierFilter === opt.id
                    ? 'border-blue-600 bg-blue-950/40 text-blue-200'
                    : 'border-gray-800 text-gray-400 hover:border-gray-600 hover:text-gray-200'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
          <input
            type="search"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              if (e.target.value.trim()) setTierFilter('all');
            }}
            placeholder="Filter flags (e.g. beam, ga, policy)…"
            className="flex-1 min-w-[12rem] text-sm bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-gray-200 placeholder:text-gray-600 focus:outline-none focus:border-blue-700"
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

        <p className="text-[11px] text-gray-600">
          {filtered.length} switches shown
          {tierFilter === 'essential' ? ' · defaults are fine for most work' : ''}
        </p>

        {showEssentialBlock && essentialFiltered.length > 0 && (
          <div className="space-y-3">
            {showSplit && (
              <h3 className="text-sm font-semibold text-gray-200">Essential</h3>
            )}
            <FlagGroupCards flags={essentialFiltered} />
          </div>
        )}

        {tierFilter === 'essential' && !searching && (
          <div className="rounded-lg border border-gray-800 bg-gray-900/40">
            <button
              type="button"
              onClick={() => setAdvancedOpen((o) => !o)}
              className="w-full flex items-center justify-between gap-3 px-4 py-3 text-left text-sm text-gray-300 hover:text-gray-100"
              aria-expanded={advancedOpen}
            >
              <span>
                Advanced knobs
                <span className="text-gray-500 font-normal"> — {ADVANCED_FLAG_COUNT} switches, leave defaults unless tuning</span>
              </span>
              <span className="text-xs text-gray-500 shrink-0">{advancedOpen ? 'Hide' : 'Show'}</span>
            </button>
            {advancedOpen && (
              <div className="px-4 pb-4 space-y-3 border-t border-gray-800 pt-3">
                <FlagGroupCards
                  flags={CLI_FLAGS.filter((f) => {
                    if (f.tier !== 'advanced') return false;
                    if (groupFilter !== 'all' && f.group !== groupFilter) return false;
                    return true;
                  })}
                />
              </div>
            )}
          </div>
        )}

        {tierFilter === 'advanced' && (
          <FlagGroupCards flags={advancedFiltered} />
        )}

        {tierFilter === 'all' && searching && (
          <FlagGroupCards flags={filtered} />
        )}

        {showSplit && advancedFiltered.length > 0 && (
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-gray-200">Advanced</h3>
            <FlagGroupCards flags={advancedFiltered} />
          </div>
        )}
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">Where results land</h2>
        <ul className="site-list">
          <li>
            <code className="site-code">tune-pfrs --pfrs-run-label NAME</code> →{' '}
            <code className="site-code">platform/web/pfrs-lab/data/runs/NAME/</code>
          </li>
          <li>
            Open the lab{' '}
            <Link href="/runs" className="site-inline-link">All Runs</Link>
            {' '}view after a labelled run to inspect workers, discoveries, and official penalty.
          </li>
          <li>
            Deeper algorithms and SI design:{' '}
            <Link href="/research" className="site-inline-link">/research</Link>
            {' '}· reproducibility ladder:{' '}
            <Link href="/reproduce" className="site-inline-link">/reproduce</Link>
          </li>
        </ul>
      </section>
    </div>
  );
}
