import Card from '@/components/Card';
import { getStorageProvider } from '@/lib/storage';

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

export default async function AdminPage() {
  const storage = getStorageProvider();
  const runIds = await storage.listRuns();

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

  return (
    <div className="space-y-6">
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
└── runs/
    └── <runLabel>/
        ├── run.json           # Metadata (this schema)
        ├── solution.json      # Domain-specific solution
        ├── results.csv        # Audit log / search progress
        └── discoveries.csv    # Global best improvements`}</pre>
      </Card>
    </div>
  );
}

function InfoBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-sm font-semibold text-gray-200">{value}</div>
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
