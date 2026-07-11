import Link from 'next/link';
import {
  BYOA_PATH,
  BYOD_STEPS,
  BYOD_TSP_PATH,
  CUSTOM_SEARCH_MODES,
  GITHUB_REPO,
  SDK_MODULE,
  SDK_VERSION,
} from '@/lib/byod-catalog';

export default function ByodExtensionSection({ compact = false }: { compact?: boolean }) {
  return (
    <section className={compact ? 'site-section site-section--panel' : 'site-section'}>
      <h2 className="site-heading">Extend with your own domain</h2>
      <p className="site-body">
        PFRS is not locked to four problem types. External domains plug in through{' '}
        <code className="text-indigo-300 text-sm">owp-sdk</code> — implement{' '}
        <code className="text-indigo-300 text-sm">searchdef.Problem</code>, register with{' '}
        <code className="text-indigo-300 text-sm">sdk.RegisterProblem</code>, and run through the same{' '}
        <code className="text-indigo-300 text-sm">owp solve</code> path as CVRP or NRP.
      </p>

      <div className="site-byod-steps">
        {BYOD_STEPS.map((s) => (
          <div key={s.step} className="site-byod-step">
            <span className="site-byod-step-num">{s.step}</span>
            <div>
              <p className="site-byod-step-title">{s.title}</p>
              <p className="site-byod-step-body">{s.body}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-6 p-4 rounded-lg border border-indigo-900/50 bg-indigo-950/20">
        <p className="text-sm font-medium text-indigo-200">Working demo: symmetric TSP</p>
        <p className="text-xs text-gray-400 mt-1 leading-relaxed">
          <a href={`${GITHUB_REPO}/tree/main/${BYOD_TSP_PATH}`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
            {BYOD_TSP_PATH}
          </a>
          {' '}registers domain <strong className="text-gray-300">tsp</strong> via blank import. Generic solve produces{' '}
          <code className="text-xs text-gray-500">run.json</code> and <code className="text-xs text-gray-500">solution.json</code> without custom cmd hooks.
        </p>
        <p className="text-xs font-mono text-gray-500 mt-2">
          go get {SDK_MODULE}@{SDK_VERSION}
        </p>
      </div>

      {!compact && (
        <div className="mt-4">
          <p className="text-sm text-gray-400 mb-2">Custom algorithms (BYOA)</p>
          <ul className="space-y-2">
            {CUSTOM_SEARCH_MODES.map((m) => (
              <li key={m.name} className="text-xs text-gray-500">
                <span className="font-mono text-emerald-400">{m.name}</span> — {m.description}
              </li>
            ))}
          </ul>
          <p className="text-xs text-gray-600 mt-2">
            <a href={`${GITHUB_REPO}/tree/main/${BYOA_PATH}`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
              {BYOA_PATH} →
            </a>
          </p>
        </div>
      )}

      <div className="flex flex-wrap gap-4 mt-6 text-sm">
        <Link href="/lab/byod" className="site-inline-link">Lab: solver registry →</Link>
        <a href={`${GITHUB_REPO}/tree/main/${BYOD_TSP_PATH}`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
          TSP example source →
        </a>
        <a href={`${GITHUB_REPO}/tree/main/platform/owp-sdk`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
          owp-sdk module →
        </a>
      </div>
    </section>
  );
}
