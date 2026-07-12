import Link from 'next/link';
import CopyBlock from '@/components/CopyBlock';
import {
  BYOA_PATH,
  BYOD_STEPS,
  BYOD_TRY_IT_CWD,
  BYOD_TRY_IT_EXAMPLES,
  BYOD_TSP_PATH,
  CUSTOM_SEARCH_MODES,
  GITHUB_REPO,
  SDK_MODULE,
  SDK_VERSION,
} from '@/lib/byod-catalog';

export default function ByodExtensionSection({ compact = false }: { compact?: boolean }) {
  const primary = BYOD_TRY_IT_EXAMPLES[0];

  return (
    <section className={compact ? 'site-section site-section--panel' : 'site-section'}>
      <h2 className="site-heading">Extend with your own domain</h2>
      <p className="site-body">
        PFRS is not locked to four problem types. External domains plug in through{' '}
        <code className="site-code">owp-sdk</code> — implement{' '}
        <code className="site-code">searchdef.Problem</code>, register with{' '}
        <code className="site-code">sdk.RegisterProblem</code>, and run through the same{' '}
        <code className="site-code">owp solve</code> path as CVRP or NRP.
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

      <div className="mt-6 p-4 rounded-lg border border-indigo-900/50 bg-indigo-950/20 space-y-3">
        <p className="text-sm font-medium text-indigo-200">Copy-paste demo: symmetric TSP</p>
        <p className="text-xs text-gray-400 leading-relaxed">
          {primary.blurb}{' '}
          <a href={`${GITHUB_REPO}/tree/main/${BYOD_TSP_PATH}`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
            {BYOD_TSP_PATH}
          </a>
          .
        </p>
        <p className="text-[10px] text-gray-600">cwd: {BYOD_TRY_IT_CWD}</p>
        <CopyBlock text={primary.command} />
        <p className="text-xs text-emerald-400/90">
          <span className="font-medium">Expect:</span> {primary.expected}
        </p>
        <p className="text-[10px] text-gray-600">
          SDK for external packages: <code className="text-gray-500">go get {SDK_MODULE}@{SDK_VERSION}</code>
          {' '}— not required for the command above (TSP is already blank-imported in this repo).
        </p>
      </div>

      {!compact && (
        <div className="mt-4">
          <p className="text-sm text-gray-400 mb-2">Custom algorithms (BYOA)</p>
          <ul className="space-y-3">
            {CUSTOM_SEARCH_MODES.map((m) => (
              <li key={m.name}>
                <p className="text-xs text-gray-500 mb-1">
                  <span className="font-mono text-emerald-400">{m.name}</span> — {m.description}
                </p>
                <p className="text-[10px] text-gray-600 mb-1">cwd: {BYOD_TRY_IT_CWD}</p>
                <CopyBlock text={m.example} />
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
