import { loadRunMetadata } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import { buildRoutingFeasibility } from '@/lib/feasibility-summary';
import { isRoutingSolution, type RoutingProblemType, type RoutingSolution } from '@/lib/solution-types';
import RunPageShell from '@/features/runs/RunPageShell';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import RouteViewer from './RouteViewer';

export const dynamic = 'force-dynamic';

function problemTypeFromMeta(meta: Record<string, unknown>): RoutingProblemType {
  const pt = String(meta.problemType || '').toLowerCase();
  return pt === 'vrptw' ? 'vrptw' : 'cvrp';
}

export default async function RoutesPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const [metadata, solutionContent] = await Promise.all([
      loadRunMetadata(id),
      getStorageProvider().readFile(id, 'solution.json'),
    ]);

    const meta = metadata as unknown as Record<string, unknown>;
    const problemType = problemTypeFromMeta(meta);
    const isVRPTW = problemType === 'vrptw';
    const solverCmd = isVRPTW ? 'solve-vrptw' : 'solve-cvrp';

    if (!solutionContent) {
      return (
        <RunPageShell
          title={isVRPTW ? 'VRPTW Route Viewer' : 'Route Viewer'}
          empty
          emptyMessage={`No route data available. Run ${solverCmd} with --run-label to generate solution.json.`}
        >
          {null}
        </RunPageShell>
      );
    }

    let parsed: unknown;
    try {
      parsed = JSON.parse(solutionContent);
    } catch {
      return (
        <RunPageShell title={isVRPTW ? 'VRPTW Route Viewer' : 'Route Viewer'} error="Error parsing solution.json">
          {null}
        </RunPageShell>
      );
    }

    if (!isRoutingSolution(parsed)) {
      return (
        <RunPageShell
          title={isVRPTW ? 'VRPTW Route Viewer' : 'Route Viewer'}
          empty
          emptyMessage="solution.json is not a routing solution (expected routes[]). Open Gantt for JSS runs."
        >
          {null}
        </RunPageShell>
      );
    }

    const solution = parsed as RoutingSolution;
    const capacity = Number(meta.capacity) || 0;
    const feasibility = buildRoutingFeasibility(solution, problemType, capacity);
    const title = isVRPTW ? 'VRPTW Route Viewer' : 'CVRP Route Viewer';
    const cardTitle = isVRPTW ? 'VRPTW Solution' : 'CVRP Solution';

    return (
      <RunPageShell title={title}>
        <div className="space-y-4">
          <Card title={cardTitle}>
            <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3">
              <MetricCard label="Total Distance" value={solution.totalCost.toLocaleString()} color="green" />
              <MetricCard label="Vehicles Used" value={String(solution.vehicles)} color="blue" />
              <MetricCard
                label="Feasible"
                value={solution.feasible ? '✓' : '✗'}
                color={solution.feasible ? 'green' : 'red'}
              />
              {isVRPTW && (
                <MetricCard
                  label="TW Violations"
                  value={String(solution.timeWindowViolations ?? 0)}
                  color={(solution.timeWindowViolations ?? 0) > 0 ? 'red' : 'green'}
                />
              )}
              <MetricCard label="Instance" value={String(meta.instance || '—')} color="default" />
              <MetricCard label="Capacity" value={capacity > 0 ? String(capacity) : '—'} color="default" />
              <MetricCard label="Customers" value={String(meta.customers || '—')} color="default" />
              {isVRPTW && meta.vehicles != null && (
                <MetricCard label="Max Vehicles" value={String(meta.vehicles)} color="default" />
              )}
            </div>
          </Card>

          <Card title="Routes">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-gray-500 uppercase">
                  <th className="text-left p-2">Route</th>
                  <th className="text-right p-2">Customers</th>
                  <th className="text-right p-2">Load</th>
                  <th className="text-right p-2">Capacity %</th>
                  {isVRPTW && <th className="text-center p-2">Feasible</th>}
                  <th className="text-right p-2">Distance</th>
                  <th className="text-left p-2">Visit Order</th>
                </tr>
              </thead>
              <tbody>
                {solution.routes.map((route, i) => {
                  const utilisation = capacity > 0 ? (route.load / capacity * 100) : 0;
                  const overloaded = capacity > 0 && route.load > capacity;
                  const routeBad = isVRPTW && route.feasible === false;
                  return (
                    <tr key={i} className={`border-t border-gray-800 ${routeBad ? 'bg-red-950/20' : ''}`}>
                      <td className="p-2 font-mono">R{i + 1}</td>
                      <td className="text-right p-2">{route.customers.length}</td>
                      <td className={`text-right p-2 ${overloaded ? 'text-red-400' : ''}`}>
                        {route.load}
                      </td>
                      <td className={`text-right p-2 ${utilisation > 95 ? 'text-amber-400' : ''} ${overloaded ? 'text-red-400' : ''}`}>
                        {capacity > 0 ? `${utilisation.toFixed(0)}%` : '—'}
                      </td>
                      {isVRPTW && (
                        <td className={`text-center p-2 ${routeBad ? 'text-red-400' : 'text-emerald-400'}`}>
                          {route.feasible === false ? '✗' : '✓'}
                        </td>
                      )}
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
                  <td className="text-right p-2">
                    {capacity > 0 ? `${feasibility.avgUtilisation.toFixed(0)}% avg` : '—'}
                  </td>
                  {isVRPTW && (
                    <td className="text-center p-2">
                      {feasibility.kind === 'vrptw' && feasibility.infeasibleRoutes > 0
                        ? <span className="text-red-400">{feasibility.infeasibleRoutes} bad</span>
                        : '✓'}
                    </td>
                  )}
                  <td className="text-right p-2">{solution.totalCost}</td>
                  <td className="p-2">—</td>
                </tr>
              </tfoot>
            </table>
          </Card>

          <RouteViewer solution={solution} problemType={problemType} capacity={capacity} />
        </div>
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Route Viewer" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
