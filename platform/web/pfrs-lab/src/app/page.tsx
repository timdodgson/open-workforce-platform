import Link from 'next/link';
import { listRuns } from '@/lib/data-loader';
import Card from '@/components/Card';
import DeleteRunButton from './DeleteRunButton';

export const dynamic = 'force-dynamic';

export default function HomePage() {
  const runs = listRuns();

  if (runs.length === 0) {
    return (
      <div>
        <Card title="No Runs Available">
          <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
            <p className="mb-2">No run data found in <code className="bg-gray-800 px-1 rounded">data/runs/</code></p>
            <p className="text-xs">Run with <code className="bg-gray-800 px-1 rounded">--pfrs-run-label &lt;name&gt;</code> to save results.</p>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <Card title="Saved Runs">
        <div className="grid gap-3">
          {runs.map(run => (
            <div key={run.id} className="flex items-center bg-gray-800 border border-gray-700 hover:border-blue-500 rounded-lg transition-colors">
              <Link href={`/runs/${run.id}/summary`} className="flex-1 p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-semibold text-blue-400">{run.id}</h3>
                    {run.metadata && (
                      <p className="text-xs text-gray-500 mt-1">
                        {run.metadata.mode?.toUpperCase()} · {run.metadata.instance} · Beam {run.metadata.beamWidth} · {((run.metadata.iterationsPerWorker || 0) / 1000).toFixed(0)}K iter
                      </p>
                    )}
                  </div>
                  <span className="text-gray-600 text-lg">→</span>
                </div>
              </Link>
              <div className="pr-3">
                <DeleteRunButton runId={run.id} />
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
