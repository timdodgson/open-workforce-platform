import { loadRunMetadata } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import RouteViewer from './RouteViewer';

export const dynamic = 'force-dynamic';

interface CVRPRoute {
  customers: number[];
  load: number;
  distance: number;
}

interface CVRPSolution {
  routes: CVRPRoute[];
  totalCost: number;
  vehicles: number;
  feasible: boolean;
}

export default async function RoutesPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [metadata, solutionContent] = await Promise.all([
    loadRunMetadata(id),
    getStorageProvider().readFile(id, 'solution.json'),
  ]);

  if (!solutionContent) {
    return (
      <Card title="Route Viewer">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">No route data available.</p>
          <p className="text-xs">Run solve-cvrp with --run-label to generate solution.json.</p>
        </div>
      </Card>
    );
  }

  let solution: CVRPSolution;
  try {
    solution = JSON.parse(solutionContent) as CVRPSolution;
  } catch {
    return (
      <Card title="Route Viewer">
        <p className="text-red-400">Error parsing solution.json</p>
      </Card>
    );
  }

  const capacity = (metadata as unknown as Record<string, unknown>)?.capacity as number || 0;

  return (
    <div className="space-y-4">
      {/* Summary metrics */}
      <Card title="CVRP Solution">
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3">
          <MetricCard label="Total Distance" value={solution.totalCost.toLocaleString()} color="green" />
          <MetricCard label="Vehicles Used" value={String(solution.vehicles)} color="blue" />
          <MetricCard label="Feasible" value={solution.feasible ? '✓' : '✗'} color={solution.feasible ? 'green' : 'red'} />
          <MetricCard label="Instance" value={String((metadata as unknown as Record<string, unknown>)?.instance || '—')} color="default" />
          <MetricCard label="Capacity" value={String(capacity)} color="default" />
          <MetricCard label="Customers" value={String((metadata as unknown as Record<string, unknown>)?.customers || '—')} color="default" />
        </div>
      </Card>

      {/* Route table */}
      <Card title="Routes">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Route</th>
              <th className="text-right p-2">Customers</th>
              <th className="text-right p-2">Load</th>
              <th className="text-right p-2">Capacity %</th>
              <th className="text-right p-2">Distance</th>
              <th className="text-left p-2">Visit Order</th>
            </tr>
          </thead>
          <tbody>
            {solution.routes.map((route, i) => {
              const utilisation = capacity > 0 ? (route.load / capacity * 100) : 0;
              const overloaded = capacity > 0 && route.load > capacity;
              return (
                <tr key={i} className="border-t border-gray-800">
                  <td className="p-2 font-mono">R{i + 1}</td>
                  <td className="text-right p-2">{route.customers.length}</td>
                  <td className={`text-right p-2 ${overloaded ? 'text-red-400' : ''}`}>
                    {route.load}
                  </td>
                  <td className={`text-right p-2 ${utilisation > 95 ? 'text-amber-400' : ''} ${overloaded ? 'text-red-400' : ''}`}>
                    {utilisation.toFixed(0)}%
                  </td>
                  <td className="text-right p-2">{route.distance}</td>
                  <td className="p-2 text-gray-400 font-mono text-[10px] truncate max-w-[200px]">
                    D → {route.customers.join(' → ')} → D
                  </td>
                </tr>
              );
            })}
          </tbody>
          <tfoot>
            <tr className="border-t border-gray-600 font-semibold">
              <td className="p-2">Total</td>
              <td className="text-right p-2">
                {solution.routes.reduce((s, r) => s + r.customers.length, 0)}
              </td>
              <td className="text-right p-2">
                {solution.routes.reduce((s, r) => s + r.load, 0)}
              </td>
              <td className="text-right p-2">—</td>
              <td className="text-right p-2">{solution.totalCost}</td>
              <td className="p-2">—</td>
            </tr>
          </tfoot>
        </table>
      </Card>

      {/* Visual route map */}
      <RouteViewer solution={solution} />
    </div>
  );
}
