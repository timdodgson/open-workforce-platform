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
