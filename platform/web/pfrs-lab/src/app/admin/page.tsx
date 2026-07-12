import Card from '@/components/Card';
import AdminGuard from '@/components/AdminGuard';
import RebuildArtifactsButton from './RebuildArtifactsButton';
import ArtifactStatusCard from './ArtifactStatusCard';
import { getStorageProvider } from '@/lib/storage';
import { getArtifactStatus } from '@/lib/intelligence';
import { getReleaseInfo } from '@/lib/release-info';
import { estimateCostUsd, loadChatUsage } from '@/lib/llm/usage';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Admin',
  description: 'Platform administration, storage layout, and run.json schema reference.',
  robots: { index: false, follow: false },
};

export const dynamic = 'force-dynamic';

const RUN_JSON_SCHEMA = {
  required: [
    { field: 'problemType', type: 'string', values: 'nrp | cvrp | vrptw | jss | ilp' },
    { field: 'mode', type: 'string', values: 'sa | lahc | tabu | portfolio | adaptive | ilp' },
    { field: 'instance', type: 'string', values: 'Instance name (e.g. A-n32-k5, n030w4, ft06, C101)' },
    { field: 'bestObjective', type: 'int', values: 'Best objective value (universal, lower is better)' },
    { field: 'runLabel', type: 'string', values: 'Unique run identifier' },
    { field: 'runtimeMs', type: 'int', values: 'Solver runtime in milliseconds' },
    { field: 'iterations', type: 'int', values: 'Iteration budget per strategy' },
    { field: 'seed', type: 'int', values: 'Random seed' },
  ],
  domainSpecific: {
    cvrp: [
      { field: 'bestDistance', type: 'int', values: 'Total travel distance' },
      { field: 'customers', type: 'int', values: 'Number of customers' },
      { field: 'capacity', type: 'int', values: 'Vehicle capacity' },
      { field: 'feasible', type: 'bool', values: 'All constraints satisfied' },
      { field: 'initialDistance', type: 'int', values: 'Constructive baseline' },
    ],
    vrptw: [
      { field: 'bestDistance', type: 'int', values: 'Total travel distance' },
      { field: 'customers', type: 'int', values: 'Number of customers' },
      { field: 'capacity', type: 'int', values: 'Vehicle capacity' },
      { field: 'vehicles', type: 'int', values: 'Max vehicles' },
      { field: 'bestVehicles', type: 'int', values: 'Vehicles used in solution' },
      { field: 'feasible', type: 'bool', values: 'Capacity + time windows satisfied' },
    ],
    jss: [
      { field: 'bestMakespan', type: 'int', values: 'Best makespan' },
      { field: 'jobs', type: 'int', values: 'Number of jobs' },
      { field: 'machines', type: 'int', values: 'Number of machines' },
      { field: 'initialMakespan', type: 'int', values: 'Constructive baseline' },
    ],
    nrp: [
      { field: 'totalPenalty', type: 'int', values: 'Sum of soft constraint penalties' },
    ],
    ilp: [
      { field: 'objective', type: 'int', values: 'Proven optimal/bound' },
      { field: 'gap', type: 'float', values: 'Optimality gap %' },
      { field: 'status', type: 'string', values: 'optimal | feasible | infeasible' },
    ],
  },
};

function formatUsd(n: number): string {
  if (n < 0.01) return `$${n.toFixed(4)}`;
  return `$${n.toFixed(2)}`;
}

function formatWhen(iso: string): string {
  if (!iso || iso === 'unknown' || iso.startsWith('1970')) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toISOString().replace('T', ' ').replace(/\.\d+Z$/, ' UTC');
}

export default async function AdminPage() {
  const storage = getStorageProvider();
  const release = getReleaseInfo();
  const [runIds, artifactStatus, chatUsage] = await Promise.all([
    storage.listRuns(),
    getArtifactStatus(storage),
    loadChatUsage(storage),
  ]);

  // Count runs by domain.
  const domainCounts: Record<string, number> = {};
  for (const id of runIds.slice(0, 200)) {
    const content = await storage.readFile(id, 'run.json');
    if (content) {
      try {
        const meta = JSON.parse(content);
        const domain = meta.problemType || 'unknown';
        domainCounts[domain] = (domainCounts[domain] || 0) + 1;
      } catch { /* skip */ }
    }
  }

  const storageType = process.env.STORAGE_PROVIDER || 'local';
  const bucket = process.env.PFRS_S3_BUCKET || 'pfrs-research-lab-data';
  const region = process.env.AWS_REGION || 'eu-west-1';

  const todayKey = new Date().toISOString().slice(0, 10);
  const todayUsage = chatUsage.byDay[todayKey] ?? { requests: 0, inputTokens: 0, outputTokens: 0 };
  const totalCost = estimateCostUsd(chatUsage.totals.inputTokens, chatUsage.totals.outputTokens);
  const todayCost = estimateCostUsd(todayUsage.inputTokens, todayUsage.outputTokens);
  const dayKeys = Object.keys(chatUsage.byDay).sort().reverse().slice(0, 7);

  return (
    <AdminGuard>
    <div className="space-y-6">
      <Card title="Release">
        <p className="text-xs text-gray-400 mb-4">
          Build metadata from the last OpenNext deploy (semantic-release tag + git SHA).
        </p>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mb-2">
          <InfoBox label="App version" value={release.version} />
          <InfoBox label="Git SHA" value={release.gitShaShort} />
          <InfoBox label="Deployed at" value={formatWhen(release.deployedAt)} />
          <InfoBox label="LLM provider" value={release.llmProvider} />
          <InfoBox label="LLM model" value={release.llmModel} />
          <InfoBox label="Full SHA" value={release.gitSha} />
        </div>
      </Card>

      <Card title="Assistant token usage">
        <p className="text-xs text-gray-400 mb-4">
          Metered from Anthropic/Bedrock responses and stored in <code className="text-gray-300">chat_usage.json</code>.
          Cost is an estimate (default Haiku list rates); check Anthropic Console for billable totals.
        </p>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
          <InfoBox label="Total requests" value={String(chatUsage.totals.requests)} />
          <InfoBox label="Input tokens" value={chatUsage.totals.inputTokens.toLocaleString()} />
          <InfoBox label="Output tokens" value={chatUsage.totals.outputTokens.toLocaleString()} />
          <InfoBox label="Est. cost (all-time)" value={formatUsd(totalCost)} />
          <InfoBox label="Today requests" value={String(todayUsage.requests)} />
          <InfoBox label="Today input" value={todayUsage.inputTokens.toLocaleString()} />
          <InfoBox label="Today output" value={todayUsage.outputTokens.toLocaleString()} />
          <InfoBox label="Est. cost (today)" value={formatUsd(todayCost)} />
        </div>

        {dayKeys.length > 0 && (
          <>
            <h4 className="text-xs font-semibold text-gray-300 mb-2">Last 7 days</h4>
            <table className="w-full text-[10px] mb-6">
              <thead>
                <tr className="text-gray-500 uppercase">
                  <th className="text-left p-1.5">Day</th>
                  <th className="text-right p-1.5">Requests</th>
                  <th className="text-right p-1.5">Input</th>
                  <th className="text-right p-1.5">Output</th>
                  <th className="text-right p-1.5">Est. $</th>
                </tr>
              </thead>
              <tbody>
                {dayKeys.map((day) => {
                  const row = chatUsage.byDay[day];
                  return (
                    <tr key={day} className="border-t border-gray-800">
                      <td className="p-1.5 font-mono text-gray-300">{day}</td>
                      <td className="p-1.5 text-right text-gray-400">{row.requests}</td>
                      <td className="p-1.5 text-right text-gray-400">{row.inputTokens.toLocaleString()}</td>
                      <td className="p-1.5 text-right text-gray-400">{row.outputTokens.toLocaleString()}</td>
                      <td className="p-1.5 text-right text-emerald-400">
                        {formatUsd(estimateCostUsd(row.inputTokens, row.outputTokens))}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </>
        )}

        {chatUsage.recent.length > 0 ? (
          <>
            <h4 className="text-xs font-semibold text-gray-300 mb-2">Recent calls</h4>
            <table className="w-full text-[10px]">
              <thead>
                <tr className="text-gray-500 uppercase">
                  <th className="text-left p-1.5">When</th>
                  <th className="text-left p-1.5">User</th>
                  <th className="text-right p-1.5">In</th>
                  <th className="text-right p-1.5">Out</th>
                  <th className="text-left p-1.5">Model</th>
                </tr>
              </thead>
              <tbody>
                {chatUsage.recent.slice(0, 15).map((r, i) => (
                  <tr key={`${r.at}-${i}`} className="border-t border-gray-800">
                    <td className="p-1.5 text-gray-400 whitespace-nowrap">{formatWhen(r.at)}</td>
                    <td className="p-1.5 text-gray-500 truncate max-w-[140px]">{r.user || '—'}</td>
                    <td className="p-1.5 text-right text-gray-300">{r.inputTokens.toLocaleString()}</td>
                    <td className="p-1.5 text-right text-gray-300">{r.outputTokens.toLocaleString()}</td>
                    <td className="p-1.5 font-mono text-gray-500 truncate max-w-[180px]">{r.model}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="text-[10px] text-gray-600 mt-2">
              Last updated {formatWhen(chatUsage.updatedAt)}. Usage starts after this deploy — older chats were not recorded.
            </p>
          </>
        ) : (
          <p className="text-xs text-gray-500">
            No assistant calls recorded yet. Send a message on /experiments/chat after this deploy to start metering.
          </p>
        )}
      </Card>

      <Card title="Platform Admin">
        <p className="text-xs text-gray-400 mb-4">System configuration and data contract reference.</p>

        {/* System Info */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
          <InfoBox label="Storage" value={storageType.toUpperCase()} />
          <InfoBox label="S3 Bucket" value={bucket} />
          <InfoBox label="Region" value={region} />
          <InfoBox label="Total Runs" value={String(runIds.length)} />
        </div>

        {/* Runs by Domain */}
        <h4 className="text-xs font-semibold text-gray-300 mb-2">Runs by Domain</h4>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
          {Object.entries(domainCounts).sort((a, b) => b[1] - a[1]).map(([domain, count]) => (
            <InfoBox key={domain} label={domain.toUpperCase()} value={String(count)} />
          ))}
        </div>
      </Card>

      <Card title="ML Journey (private)">
        <p className="text-xs text-gray-400 mb-3">
          Personal learning notes — measured path from ~3–4/10 to 10/10 ML maturity, harness snapshots, and per-step detail.
        </p>
        <a href="/admin/ml-journey" className="text-sm text-blue-400 hover:text-blue-300">
          Open ML Journey →
        </a>
      </Card>

      <ArtifactStatusCard status={artifactStatus} />
      <RebuildArtifactsButton stale={artifactStatus.stale} />

      {/* Schema Reference */}
      <Card title="run.json Schema">
        <p className="text-xs text-gray-400 mb-4">
          Contract between Go CLI (producer) and dashboard (consumer). Every run must produce a run.json with these fields.
        </p>

        <h4 className="text-xs font-semibold text-emerald-400 mb-2">Required Fields (all domains)</h4>
        <SchemaTable fields={RUN_JSON_SCHEMA.required} />

        {Object.entries(RUN_JSON_SCHEMA.domainSpecific).map(([domain, fields]) => (
          <div key={domain} className="mt-4">
            <h4 className="text-xs font-semibold text-blue-400 mb-2">{domain.toUpperCase()} — Domain-Specific</h4>
            <SchemaTable fields={fields} />
          </div>
        ))}
      </Card>

      {/* Reading Priority */}
      <Card title="Objective Reading Priority">
        <p className="text-xs text-gray-400 mb-3">
          The dashboard reads the objective value in this order (first non-zero wins):
        </p>
        <ol className="list-decimal list-inside text-xs text-gray-300 space-y-1">
          <li><code className="text-emerald-400">bestObjective</code> — universal (preferred)</li>
          <li><code className="text-blue-400">bestDistance</code> — CVRP / VRPTW</li>
          <li><code className="text-blue-400">bestMakespan</code> — JSS</li>
          <li><code className="text-blue-400">totalPenalty</code> — NRP</li>
          <li><code className="text-blue-400">objective</code> — ILP</li>
          <li><code className="text-gray-500">summary.totalPenalty</code> — fallback (CSV parsing)</li>
        </ol>
      </Card>

      {/* S3 Structure */}
      <Card title="S3 Storage Layout">
        <pre className="text-[10px] text-gray-400 bg-gray-900 p-3 rounded overflow-x-auto">{`${bucket}/
├── manifest.json              # Run index (read on every page load)
├── chat_usage.json            # Assistant token metering (Admin)
└── runs/
    └── <runLabel>/
        ├── run.json           # Metadata (this schema)
        ├── solution.json      # Domain-specific solution
        ├── results.csv        # Audit log / search progress
        └── discoveries.csv    # Global best improvements`}</pre>
      </Card>
    </div>
    </AdminGuard>
  );
}

function InfoBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-sm font-semibold text-gray-200 break-all">{value}</div>
    </div>
  );
}

function SchemaTable({ fields }: { fields: { field: string; type: string; values: string }[] }) {
  return (
    <table className="w-full text-[10px]">
      <thead>
        <tr className="text-gray-500 uppercase">
          <th className="text-left p-1.5">Field</th>
          <th className="text-left p-1.5">Type</th>
          <th className="text-left p-1.5">Description</th>
        </tr>
      </thead>
      <tbody>
        {fields.map(f => (
          <tr key={f.field} className="border-t border-gray-800">
            <td className="p-1.5 font-mono text-emerald-400">{f.field}</td>
            <td className="p-1.5 text-gray-500">{f.type}</td>
            <td className="p-1.5 text-gray-400">{f.values}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
