import { loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import WorkersCharts from './WorkersCharts';

function fmt(n: number): string { return n.toLocaleString('en-US'); }

export default async function WorkersPage() {
  const d = await loadRunSummary();

  const chartData = d.weeks.map(w => ({
    week: `W${w.week}`,
    workers: w.workersStarted,
    branches: w.branchesCreated,
    dropped: w.branchesDropped,
    depth: w.winningBranchDepth,
  }));

  return (
    <div>
      <WorkersCharts data={chartData} />

      <Card title="Worker Detail">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Workers</th>
              <th className="text-right p-2">Branches</th>
              <th className="text-right p-2">Dropped</th>
              <th className="text-right p-2">Depth</th>
              <th className="text-right p-2">Improved</th>
              <th className="text-right p-2">Produced Best</th>
              <th className="text-right p-2">Duration</th>
              <th className="text-right p-2">Cands/sec</th>
              <th className="text-right p-2">Avg Cands/Worker</th>
            </tr>
          </thead>
          <tbody>
            {d.weeks.map(w => {
              const candsPerSec = w.durationMs > 0 ? Math.round(w.candidates / (w.durationMs / 1000)) : 0;
              const avgCandsPerWorker = w.workersStarted > 0 ? Math.round(w.candidates / w.workersStarted) : 0;
              return (
                <tr key={w.week} className="border-t border-gray-800">
                  <td className="p-2">{w.week}</td>
                  <td className="text-right p-2">{w.workersStarted}</td>
                  <td className="text-right p-2">{w.branchesCreated}</td>
                  <td className="text-right p-2">{w.branchesDropped}</td>
                  <td className="text-right p-2">{w.winningBranchDepth}</td>
                  <td className="text-right p-2">{w.workersImproved}</td>
                  <td className="text-right p-2">{w.workersProducedBest}</td>
                  <td className="text-right p-2">{(w.durationMs / 1000).toFixed(1)}s</td>
                  <td className="text-right p-2">{fmt(candsPerSec)}</td>
                  <td className="text-right p-2">{fmt(avgCandsPerWorker)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
