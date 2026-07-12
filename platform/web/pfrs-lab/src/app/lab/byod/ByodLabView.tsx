'use client';

import Link from 'next/link';
import Card from '@/components/Card';
import CopyBlock from '@/components/CopyBlock';
import {
  BUILTIN_SEARCH_MODES,
  BYOA_PATH,
  BYOD_LIST_SOLVERS,
  BYOD_SEED_DEMOS,
  BYOD_STEPS,
  BYOD_TRY_IT_CWD,
  BYOD_TRY_IT_EXAMPLES,
  BYOD_TSP_PATH,
  CUSTOM_SEARCH_MODES,
  GITHUB_REPO,
  REGISTERED_SOLVERS,
  SDK_MODULE,
  SDK_VERSION,
} from '@/lib/byod-catalog';

export default function ByodLabView() {
  const builtin = REGISTERED_SOLVERS.filter((s) => s.kind === 'builtin');
  const byod = REGISTERED_SOLVERS.filter((s) => s.kind === 'byod');

  return (
    <div className="space-y-8 max-w-4xl">
      <div>
        <p className="text-[10px] uppercase tracking-wider text-indigo-400 font-semibold mb-1">BYOD / BYOA</p>
        <h1 className="text-2xl font-bold text-gray-100">Solver registry &amp; extension contract</h1>
        <p className="text-sm text-gray-400 mt-2 max-w-3xl leading-relaxed">
          Bring your own domain or algorithm onto the same <code className="text-gray-500">owp solve</code> path.
          Start with the copy-paste demos below — then fork the TSP template when you are ready to extend.
        </p>
      </div>

      <Card title="Try it now (copy-paste)">
        <p className="text-xs text-gray-400 mb-3 leading-relaxed">
          Requires Go 1.22+. Run from <code className="text-gray-300">{BYOD_TRY_IT_CWD}</code>
          {' '}(same cwd as Getting Started). Commands use <code className="text-gray-300">go run ./cmd/owp</code>
          {' '}— there is no global <code className="text-gray-500">owp</code> binary required.
        </p>

        <p className="text-[10px] text-gray-600 mb-1">Optional sanity check</p>
        <CopyBlock text={BYOD_LIST_SOLVERS} />
        <p className="text-[10px] text-gray-500 mt-1 mb-4">
          Should list <code className="text-gray-400">tsp</code> under domains and <code className="text-gray-400">greedy</code> under custom modes.
        </p>

        <div className="space-y-5">
          {BYOD_TRY_IT_EXAMPLES.map((ex) => (
            <div key={ex.id}>
              <p className="text-sm font-medium text-gray-200">{ex.title}</p>
              <p className="text-xs text-gray-500 mt-0.5 mb-2 leading-relaxed">{ex.blurb}</p>
              <p className="text-[10px] text-gray-600 mb-1">cwd: {BYOD_TRY_IT_CWD}</p>
              <CopyBlock text={ex.command} />
              <p className="text-xs text-emerald-400/90 mt-2">
                <span className="font-medium">Expect:</span> {ex.expected}
              </p>
            </div>
          ))}
        </div>

        <p className="text-[11px] text-gray-500 mt-4 leading-relaxed">
          Optional: add <code className="text-gray-400">--run-label demo-tsp-5city-sa-s42</code> to write artifacts under{' '}
          <code className="text-gray-400">platform/web/pfrs-lab/data/runs/</code>, then open{' '}
          <Link href="/runs" className="text-blue-400 hover:underline">All Runs</Link>.
        </p>
      </Card>

      <div className="grid md:grid-cols-2 gap-4">
        <Card title="Built-in domains">
          <ul className="space-y-2">
            {builtin.map((s) => (
              <li key={s.name} className="text-xs border-b border-gray-800/60 pb-2 last:border-0">
                <span className="font-mono text-blue-300">{s.name}</span>
                <span className="text-gray-500 ml-2">{s.title}</span>
                <p className="font-mono text-[10px] text-gray-600 mt-1">{s.usage}</p>
              </li>
            ))}
          </ul>
        </Card>
        <Card title="BYOD domains">
          <ul className="space-y-2">
            {byod.map((s) => (
              <li key={s.name} className="text-xs border-b border-gray-800/60 pb-2 last:border-0">
                <span className="font-mono text-emerald-300">{s.name}</span>
                <span className="text-gray-500 ml-2">{s.title}</span>
                <p className="font-mono text-[10px] text-gray-600 mt-1 break-all">{s.usage}</p>
                {s.notes && <p className="text-[10px] text-gray-500 mt-0.5">{s.notes}</p>}
              </li>
            ))}
          </ul>
          <a
            href={`${GITHUB_REPO}/tree/main/${BYOD_TSP_PATH}`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-blue-400 hover:underline mt-3 inline-block"
          >
            Fork the TSP template →
          </a>
        </Card>
      </div>

      <Card title="Search modes">
        <p className="text-xs text-gray-500 mb-3">Built-in (always available via optimisation engine)</p>
        <div className="flex flex-wrap gap-2 mb-4">
          {BUILTIN_SEARCH_MODES.map((m) => (
            <span key={m} className="text-xs font-mono px-2 py-1 rounded bg-gray-800 text-gray-300 border border-gray-700">
              {m}
            </span>
          ))}
        </div>
        <p className="text-xs text-gray-500 mb-2">Custom (BYOA — sdk.RegisterSearch)</p>
        {CUSTOM_SEARCH_MODES.map((m) => (
          <div key={m.name} className="p-3 rounded-lg bg-gray-900 border border-gray-800 mb-2">
            <p className="font-mono text-sm text-emerald-400">{m.name}</p>
            <p className="text-xs text-gray-400 mt-1">{m.description}</p>
            <p className="text-[10px] text-gray-600 mt-2 mb-1">cwd: {BYOD_TRY_IT_CWD}</p>
            <CopyBlock text={m.example} />
          </div>
        ))}
        <a
          href={`${GITHUB_REPO}/tree/main/${BYOA_PATH}`}
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-blue-400 hover:underline"
        >
          BYOA README →
        </a>
      </Card>

      <Card title="Registration steps">
        <div className="space-y-3">
          {BYOD_STEPS.map((s) => (
            <div key={s.step} className="flex gap-3 text-xs">
              <span className="text-indigo-400 font-bold">{s.step}.</span>
              <div>
                <p className="text-gray-200 font-medium">{s.title}</p>
                <p className="text-gray-500 mt-0.5">{s.body}</p>
              </div>
            </div>
          ))}
        </div>
      </Card>

      <Card title="owp-sdk module">
        <p className="text-sm text-gray-300 font-mono">{SDK_MODULE}@{SDK_VERSION}</p>
        <pre className="mt-3 p-3 rounded bg-gray-950 border border-gray-800 text-[11px] text-gray-400 overflow-x-auto">{`go get ${SDK_MODULE}@${SDK_VERSION}`}</pre>
        <p className="text-xs text-gray-500 mt-3">
          Monorepo development uses a <code>replace</code> directive to <code>platform/owp-sdk</code>.
          Tag <code>{SDK_VERSION}</code> on GitHub for external consumers. The try-it commands above already
          use the blank-imported TSP demo inside this repo — you do not need <code>go get</code> for those.
        </p>
        <div className="flex flex-wrap gap-3 mt-4 text-xs">
          <Link href="/runs" className="text-blue-400 hover:underline">Browse runs (incl. demo-tsp-*) →</Link>
          <Link href="/getting-started" className="text-blue-400 hover:underline">Getting started →</Link>
          <Link href="/research" className="text-blue-400 hover:underline">Technical reference →</Link>
        </div>
      </Card>

      <Card title="Seed all three demo runs">
        <p className="text-xs text-gray-400 mb-2">
          From <code className="text-gray-300">{BYOD_TRY_IT_CWD}</code> — SA, LAHC, and greedy into local lab storage:
        </p>
        <CopyBlock text={BYOD_SEED_DEMOS} />
        <p className="text-[10px] text-gray-600 mt-2">
          Creates <code>demo-tsp-5city-sa-s42</code>, <code>demo-tsp-5city-lahc-s42</code>,{' '}
          <code>demo-tsp-5city-greedy-s42</code> under <code>data/runs/</code> and updates manifest.
        </p>
      </Card>
    </div>
  );
}
