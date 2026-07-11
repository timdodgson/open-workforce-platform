import { readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import AdminGuard from '@/components/AdminGuard';
import Card from '@/components/Card';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'ML Journey (private)',
  description: 'Personal learning notes — Search Intelligence ML maturity ladder.',
  robots: { index: false, follow: false },
};

export const dynamic = 'force-dynamic';

type HarnessReport = {
  generatedAt?: string;
  mlMaturity?: number;
  totalRuns?: number;
  modeSummaries?: Array<{
    policyMode: string;
    n: number;
    meanObjective: number;
    meanRuntimeMs: number;
  }>;
  comparisons?: Array<{
    domain: string;
    algorithm: string;
    modeB: string;
    objectiveDelta: number;
    runtimeSavedPct: number;
    verdict: string;
    roi: number;
  }>;
  gates?: {
    step2PromoteOk?: boolean;
    step4PromoteOk?: boolean;
    step4FalseStopRate?: number;
  };
};

function loadHarness(): HarnessReport | null {
  const candidates = [
    join(process.cwd(), '..', '..', '..', 'docs', 'reports', 'ml-harness', 'latest.json'),
    join(process.cwd(), 'data', 'ml-harness-latest.json'),
  ];
  for (const p of candidates) {
    if (existsSync(p)) {
      try {
        return JSON.parse(readFileSync(p, 'utf-8')) as HarnessReport;
      } catch {
        return null;
      }
    }
  }
  return null;
}

const STEPS = [
  {
    id: 0,
    level: '3–4',
    title: 'Baseline — offline trees + hybrid',
    status: 'done',
    what: 'Sklearn trees → JSON → Go inference. Hybrid = learned first, rules fallback.',
    why: 'Interpretable, fast, validated on 288 val-* runs.',
    measure: '288/288 matrix; EXP-007 stats.',
    cost: 'Low',
  },
  {
    id: 1,
    level: '5',
    title: 'Measurement harness',
    status: 'active',
    what: 'Compare rules / hybrid / learned on same val-* labels.',
    why: 'Repeatable scorecard before fancier ML.',
    measure: 'npm run compare-policy-modes',
    cost: 'Days',
  },
  {
    id: 2,
    level: '5',
    title: 'Gradient boosting (distilled)',
    status: 'done',
    what: 'Boosting on grouped CV; distill winner to deployable tree.',
    why: 'Better classical ML without new Go runtime.',
    measure: 'Retrain + re-run harness; ≥2 domain wins.',
    cost: '1–2 weeks',
  },
  {
    id: 3,
    level: '5.5',
    title: 'Per-context policies',
    status: 'done',
    what: 'Classifiers per domain × algorithm × instance; Go prefers instance-specific tree.',
    why: 'NRP ≠ CVRP stagnation shapes; A-n32-k5 ≠ X-n101-k25.',
    measure: 'Retrain policies; classifier count in JSON; re-run harness.',
    cost: 'Weeks',
  },
  {
    id: 4,
    level: '6',
    title: 'Counterfactual eval',
    status: 'active',
    what: 'Offline ex-post simulation; false-stop gate before promotion.',
    why: 'Avoid deploying policies that stop too early on new seeds.',
    measure: 'npm run evaluate-counterfactual; false_stop_rate ≤ 5%.',
    cost: 'Weeks',
  },
  {
    id: 5,
    level: '7',
    title: 'Bandits',
    status: 'planned',
    what: 'Multi-step portfolio / worker decisions.',
    why: 'Search is sequential.',
    measure: 'Episode regret.',
    cost: 'Months',
  },
  {
    id: 6,
    level: '8',
    title: 'Trajectory models',
    status: 'planned',
    what: 'Learn from full checkpoint sequences.',
    why: 'Plateau shape, not just length.',
    measure: 'Episode ROI vs bandits.',
    cost: 'Months',
  },
  {
    id: 7,
    level: '9',
    title: 'Deep (if needed)',
    status: 'planned',
    what: 'Neural policies only after trees plateau.',
    why: 'Cost must justify gain.',
    measure: 'Hard-instance wins only.',
    cost: 'High',
  },
  {
    id: 8,
    level: '10',
    title: 'Closed-loop research',
    status: 'planned',
    what: 'Auto-propose experiments, human approves.',
    why: 'Lab product mode.',
    measure: 'Cost per significant finding.',
    cost: 'Very high',
  },
] as const;

export default function MlJourneyAdminPage() {
  const harness = loadHarness();

  return (
    <AdminGuard>
      <div className="max-w-4xl mx-auto space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">ML Journey — private learning notes</h1>
          <p className="text-sm text-gray-400 mt-2">
            Login-only. Tracked path from ~3–4/10 → 10/10 with ROI gates. See also{' '}
            <code className="text-blue-400">docs/ML_JOURNEY.md</code>.
          </p>
        </div>

        <Card title="Current position">
          <div className="grid grid-cols-2 sm:grid-cols-5 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-blue-400">{harness?.mlMaturity ?? 6}/10</p>
              <p className="text-[10px] text-gray-500 uppercase">ML maturity</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-emerald-400">{harness?.totalRuns ?? '—'}</p>
              <p className="text-[10px] text-gray-500 uppercase">Harness runs</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-amber-400">4</p>
              <p className="text-[10px] text-gray-500 uppercase">Active step</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-300">
                {harness?.gates?.step2PromoteOk ? 'Yes' : 'TBD'}
              </p>
              <p className="text-[10px] text-gray-500 uppercase">Step 2 gate</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-300">
                {harness?.gates?.step4PromoteOk ? 'Yes' : harness?.gates?.step4FalseStopRate != null
                  ? `${(harness.gates.step4FalseStopRate * 100).toFixed(1)}%`
                  : 'TBD'}
              </p>
              <p className="text-[10px] text-gray-500 uppercase">Step 4 gate</p>
            </div>
          </div>
        </Card>

        {harness && (
          <Card title="Latest harness snapshot">
            <p className="text-xs text-gray-500 mb-3">
              {harness.generatedAt ? new Date(harness.generatedAt).toLocaleString() : '—'}
            </p>
            <div className="text-xs text-gray-400 space-y-1 font-mono">
              {harness.modeSummaries?.map((s) => (
                <p key={s.policyMode}>
                  {s.policyMode}: n={s.n} obj={s.meanObjective} rt={Math.round(s.meanRuntimeMs)}ms
                </p>
              ))}
              {harness.comparisons?.slice(0, 6).map((c, i) => (
                <p key={i}>
                  {c.domain}/{c.algorithm} {c.modeB}: Δobj={c.objectiveDelta} Δrt={c.runtimeSavedPct}% ROI={c.roi}
                </p>
              ))}
            </div>
            <p className="text-xs text-gray-500 mt-3">Refresh: npm run compare-policy-modes</p>
          </Card>
        )}

        <Card title="SI + ML reminder">
          <ul className="text-sm text-gray-400 space-y-2 list-disc pl-5">
            <li>Three hooks: SearchAssist, PortfolioAssist, WorkerAssist (NRP).</li>
            <li>Train offline (Python) → JSON policies → Go at solve time.</li>
            <li>Hybrid = learned first; rules if unsure; safety always wins.</li>
          </ul>
        </Card>

        <Card title="Each iteration — detail">
          <div className="space-y-4">
            {STEPS.map((step) => (
              <article
                key={step.id}
                className={`border rounded-lg p-4 ${
                  step.status === 'active'
                    ? 'border-blue-800 bg-blue-950/20'
                    : step.status === 'done'
                      ? 'border-emerald-900/50'
                      : 'border-gray-800'
                }`}
              >
                <div className="flex gap-2 mb-1">
                  <span className="text-[10px] text-gray-500">Step {step.id}</span>
                  <span className="text-[10px] text-gray-600">level {step.level}</span>
                  <span className="text-[10px] text-blue-400">{step.status}</span>
                </div>
                <h3 className="font-semibold text-gray-200">{step.title}</h3>
                <p className="text-sm text-gray-400 mt-2"><strong className="text-gray-500">What:</strong> {step.what}</p>
                <p className="text-sm text-gray-400"><strong className="text-gray-500">Why:</strong> {step.why}</p>
                <p className="text-sm text-gray-400 font-mono text-xs"><strong className="text-gray-500 font-sans">Measure:</strong> {step.measure}</p>
                <p className="text-sm text-gray-500"><strong>Cost:</strong> {step.cost}</p>
              </article>
            ))}
          </div>
        </Card>

        <Card title="Stop rules">
          <ul className="text-sm text-gray-400 space-y-1 list-disc pl-5">
            <li>Stop if next step is &gt;2× cost for &lt;2% gain.</li>
            <li>Level 6–7 is a sensible ceiling unless SI becomes the product.</li>
          </ul>
        </Card>
      </div>
    </AdminGuard>
  );
}
