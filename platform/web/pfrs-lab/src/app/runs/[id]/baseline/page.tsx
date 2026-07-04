import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import ILPProgress from './ILPProgress';

export const dynamic = 'force-dynamic';

interface Props {
  params: Promise<{ id: string }>;
}

export default async function BaselinePage({ params }: Props) {
  const { id } = await params;
  const storage = getStorageProvider();

  const benchmarkContent = await storage.readFile(id, 'ilp-benchmark.json');
  const progressContent = await storage.readFile(id, 'ilp-progress.csv');
  const runContent = await storage.readFile(id, 'run.json');

  if (!benchmarkContent && !runContent) {
    return (
      <Card title="ILP Baseline">
        <p className="text-gray-500">No ILP benchmark data found for this run.</p>
        <p className="text-xs text-gray-600 mt-2">
          Run with: <code className="bg-gray-800 px-1 rounded">owp benchmark-ilp --instance n012w8 --weeks 8 --time-limit 7200 --storage s3</code>
        </p>
      </Card>
    );
  }

  const benchmark = benchmarkContent ? JSON.parse(benchmarkContent) : null;
  const runMeta = runContent ? JSON.parse(runContent) : null;

  // Parse progress CSV.
  const progress = parseProgressCSV(progressContent);

  return (
    <div className="space-y-6">
      <Card title="ILP Baseline Result">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Metric label="Objective" value={benchmark?.objective ?? runMeta?.objective ?? '—'} />
          <Metric label="Lower Bound" value={benchmark?.lowerBound ?? runMeta?.bound ?? '—'} />
          <Metric label="Gap" value={`${(benchmark?.gapPercent ?? runMeta?.gap ?? 0).toFixed(2)}%`} />
          <Metric label="Status" value={benchmark?.status ?? runMeta?.status ?? '—'} />
          <Metric label="Runtime" value={`${((benchmark?.runtimeSeconds ?? runMeta?.runtime ?? 0) / 60).toFixed(1)} min`} />
          <Metric label="Time Limit" value={`${((benchmark?.timeLimit ?? runMeta?.timeLimit ?? 0) / 60).toFixed(0)} min`} />
          <Metric label="Threads" value={benchmark?.threads ?? runMeta?.threads ?? '—'} />
          <Metric label="Instance" value={benchmark?.instance ?? runMeta?.instance ?? '—'} />
        </div>
        {benchmark?.notes && (
          <p className="text-xs text-yellow-400 mt-3">{benchmark.notes}</p>
        )}
      </Card>

      {progress.length > 0 && (
        <Card title="Solve Progress">
          <ILPProgress data={progress} />
        </Card>
      )}

      {benchmark?.supportedConstraints && (
        <Card title="Model Coverage">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h4 className="text-xs text-green-400 font-semibold mb-2">Supported Constraints</h4>
              <ul className="text-xs text-gray-400 space-y-1">
                {benchmark.supportedConstraints.map((c: string) => (
                  <li key={c}>✓ {c}</li>
                ))}
              </ul>
            </div>
            {benchmark.unsupportedConstraints?.length > 0 && (
              <div>
                <h4 className="text-xs text-red-400 font-semibold mb-2">Unsupported Constraints</h4>
                <ul className="text-xs text-gray-400 space-y-1">
                  {benchmark.unsupportedConstraints.map((c: string) => (
                    <li key={c}>✗ {c}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Heuristic Gap Analysis */}
      <Card title="Heuristic Gap Analysis">
        <p className="text-xs text-gray-500 mb-4">
          Compares the ILP optimal/bounded solution against the best heuristic result on the same instance.
          The gap shows how close the metaheuristic approaches get to proven-optimal solutions.
        </p>
        <HeuristicGap
          ilpObjective={benchmark?.objective ?? runMeta?.objective}
          ilpBound={benchmark?.lowerBound ?? runMeta?.bound}
          instance={benchmark?.instance ?? runMeta?.instance}
        />
      </Card>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <p className="text-[10px] uppercase text-gray-500 tracking-wider">{label}</p>
      <p className="text-lg font-bold text-white">{value}</p>
    </div>
  );
}

function HeuristicGap({ ilpObjective, ilpBound, instance }: { ilpObjective?: number; ilpBound?: number; instance?: string }) {
  if (!ilpObjective || ilpObjective === 0) {
    return (
      <div className="border-2 border-dashed border-gray-700 rounded-lg p-6 text-center text-gray-500">
        <p>No ILP objective available for gap calculation.</p>
      </div>
    );
  }

  // Known best heuristic results for common instances.
  const knownBests: Record<string, { penalty: number; config: string }> = {
    'n012w8': { penalty: 3465, config: 'portfolio+lookahead+fw2, beam 12, 1.5M iter' },
    'n005w4': { penalty: 385, config: 'SA baseline, beam 5, 500K iter' },
  };

  const instanceKey = instance || '';
  const heuristic = knownBests[instanceKey];

  if (!heuristic) {
    return (
      <div className="border-2 border-dashed border-gray-700 rounded-lg p-6 text-center text-gray-500">
        <p className="mb-2">No matching heuristic run found for instance: {instanceKey}</p>
        <p className="text-xs">Run PFRS on the same instance to populate the comparison.</p>
      </div>
    );
  }

  const gap = ((heuristic.penalty - ilpObjective) / ilpObjective * 100);
  const gapToBound = ilpBound ? ((heuristic.penalty - ilpBound) / ilpBound * 100) : null;

  return (
    <div className="space-y-3">
      <table className="w-full text-xs">
        <thead>
          <tr className="text-gray-500 uppercase">
            <th className="text-left p-2">Solver</th>
            <th className="text-right p-2">Objective</th>
            <th className="text-right p-2">Gap to ILP</th>
            <th className="text-right p-2">Gap to Bound</th>
            <th className="text-left p-2">Config</th>
          </tr>
        </thead>
        <tbody>
          <tr className="border-t border-gray-800">
            <td className="p-2 text-blue-400 font-semibold">ILP (HiGHS)</td>
            <td className="text-right p-2 text-emerald-400 font-semibold">{ilpObjective.toLocaleString()}</td>
            <td className="text-right p-2 text-gray-500">—</td>
            <td className="text-right p-2 text-gray-500">{ilpBound ? `${((ilpObjective - ilpBound) / ilpBound * 100).toFixed(1)}%` : '—'}</td>
            <td className="p-2 text-gray-500">Exact solver (time-limited)</td>
          </tr>
          <tr className="border-t border-gray-800">
            <td className="p-2 text-amber-400 font-semibold">PFRS (best)</td>
            <td className="text-right p-2">{heuristic.penalty.toLocaleString()}</td>
            <td className="text-right p-2 text-amber-400">+{gap.toFixed(1)}%</td>
            <td className="text-right p-2">{gapToBound ? `+${gapToBound.toFixed(1)}%` : '—'}</td>
            <td className="p-2 text-gray-500 text-[10px]">{heuristic.config}</td>
          </tr>
        </tbody>
      </table>
      <div className="bg-gray-800 rounded p-3">
        <p className="text-xs text-gray-400">
          <span className="text-emerald-400 font-semibold">Interpretation:</span> The heuristic achieves a solution within {gap.toFixed(1)}% of the ILP result.
          {gap < 20 && ' This is a strong result — the heuristic is finding near-optimal solutions in seconds rather than hours.'}
          {gap >= 20 && gap < 50 && ' There is room for improvement — consider higher iteration budgets or portfolio mode.'}
          {ilpBound && <> The ILP itself has a {((ilpObjective - ilpBound) / ilpBound * 100).toFixed(1)}% gap to its proven lower bound, meaning the true optimum may be even lower.</>}
        </p>
      </div>
    </div>
  );
}

interface ProgressPoint {
  elapsed: number;
  incumbent: number | null;
  bound: number | null;
  gap: number;
  nodes: number;
  lpIters: number;
}

function parseProgressCSV(content: string | null): ProgressPoint[] {
  if (!content) return [];
  const lines = content.trim().split('\n');
  if (lines.length < 2) return [];

  return lines.slice(1).map(line => {
    const [elapsed, incumbent, bound, gap, nodes, lpIters] = line.split(',');
    return {
      elapsed: parseFloat(elapsed),
      incumbent: incumbent ? parseFloat(incumbent) : null,
      bound: bound ? parseFloat(bound) : null,
      gap: parseFloat(gap),
      nodes: parseInt(nodes),
      lpIters: parseInt(lpIters),
    };
  }).filter(p => !isNaN(p.elapsed));
}
