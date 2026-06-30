import { loadTree, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import TreeView from './TreeView';

export default async function TreePage() {
  const [nodes, summary] = await Promise.all([loadTree(), loadRunSummary()]);
  const hasTree = nodes.length > 0;

  // Build winning lineage from audit data (available even without tree.csv)
  const lineage = summary.weeks.map((w, i) => ({
    week: w.week,
    pathID: `W${w.bestWorkerID}`,
    seed: w.seed,
    penalty: w.finalPenalty,
    cumulative: summary.cumulativePenalties[i],
    workers: w.workersStarted,
    candidates: w.candidates,
    depth: w.winningBranchDepth,
  }));

  return (
    <div>
      {/* Winning Lineage — always available from audit CSV */}
      <Card title="Winning Lineage">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="p-2">Path ID</th>
              <th className="p-2">Seed</th>
              <th className="text-right p-2">Penalty</th>
              <th className="text-right p-2">Cumulative</th>
              <th className="text-right p-2">Workers</th>
              <th className="text-right p-2">Depth</th>
            </tr>
          </thead>
          <tbody>
            {lineage.map(l => (
              <tr key={l.week} className="border-t border-gray-800">
                <td className="p-2">{l.week}</td>
                <td className="p-2 text-center">{l.pathID}</td>
                <td className="p-2 text-center">{l.seed}</td>
                <td className="text-right p-2">{l.penalty.toLocaleString()}</td>
                <td className="text-right p-2">{l.cumulative.toLocaleString()}</td>
                <td className="text-right p-2">{l.workers}</td>
                <td className="text-right p-2">{l.depth}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Full tree view — only if tree.csv exists */}
      {hasTree ? (
        <TreeView nodes={nodes} />
      ) : (
        <Card title="Search Tree Visualisation">
          <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
            <p className="mb-2">Tree data not available.</p>
            <p className="text-xs">Generate with <code className="bg-gray-800 px-1 rounded">--tree-csv</code> or beam path export.</p>
          </div>
        </Card>
      )}
    </div>
  );
}
