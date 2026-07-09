import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import type { RouteConstraintRow } from '@/lib/constraint-analysis';
import type { FeasibilitySummary } from '@/lib/feasibility-summary';

interface Props {
  problemType: 'cvrp' | 'vrptw';
  summary: FeasibilitySummary;
  rows: RouteConstraintRow[];
  timeWindowViolations?: number;
}

export default function RoutingConstraints({ problemType, summary, rows, timeWindowViolations = 0 }: Props) {
  const isVRPTW = problemType === 'vrptw';
  const title = isVRPTW ? 'VRPTW Constraints' : 'CVRP Constraints';

  return (
    <div className="space-y-4">
      <Card title={title}>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <MetricCard
            label="Feasible"
            value={(summary.kind === 'vrptw' ? summary.feasible : summary.kind === 'capacity' ? summary.feasible : false) ? '✓' : '✗'}
            color={(summary.kind === 'vrptw' ? summary.feasible : summary.kind === 'capacity' ? summary.feasible : false) ? 'green' : 'red'}
          />
          {isVRPTW && summary.kind === 'vrptw' && (
            <>
              <MetricCard
                label="TW violations"
                value={String(summary.timeWindowViolations)}
                color={summary.timeWindowViolations > 0 ? 'red' : 'green'}
              />
              <MetricCard
                label="Infeasible routes"
                value={String(summary.infeasibleRoutes)}
                color={summary.infeasibleRoutes > 0 ? 'red' : 'green'}
              />
            </>
          )}
          {summary.kind !== 'jss' && (
            <>
              <MetricCard
                label="Capacity overloads"
                value={String(summary.overloadedRoutes)}
                color={summary.overloadedRoutes > 0 ? 'red' : 'green'}
              />
              <MetricCard
                label="Peak capacity"
                value={`${summary.maxUtilisation.toFixed(0)}%`}
                color={summary.maxUtilisation > 100 ? 'red' : summary.maxUtilisation > 95 ? 'amber' : 'blue'}
              />
            </>
          )}
        </div>

        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Route</th>
              <th className="text-right p-2">Customers</th>
              <th className="text-right p-2">Load</th>
              <th className="text-right p-2">Capacity %</th>
              {isVRPTW && <th className="text-center p-2">TW OK</th>}
              <th className="text-center p-2">Feasible</th>
              <th className="text-right p-2">Distance</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr
                key={row.routeIndex}
                className={`border-t border-gray-800 ${!row.feasible ? 'bg-red-950/20' : ''}`}
              >
                <td className="p-2 font-mono">R{row.routeIndex}</td>
                <td className="text-right p-2">{row.customers}</td>
                <td className={`text-right p-2 ${row.overloaded ? 'text-red-400' : ''}`}>{row.load}</td>
                <td className={`text-right p-2 ${row.utilisation > 95 ? 'text-amber-400' : ''} ${row.overloaded ? 'text-red-400' : ''}`}>
                  {row.capacity > 0 ? `${row.utilisation.toFixed(0)}%` : '—'}
                </td>
                {isVRPTW && (
                  <td className={`text-center p-2 ${row.twOk ? 'text-emerald-400' : 'text-red-400'}`}>
                    {row.twOk ? '✓' : '✗'}
                  </td>
                )}
                <td className={`text-center p-2 ${row.feasible ? 'text-emerald-400' : 'text-red-400'}`}>
                  {row.feasible ? '✓' : '✗'}
                </td>
                <td className="text-right p-2">{row.distance}</td>
              </tr>
            ))}
          </tbody>
        </table>

        {isVRPTW && timeWindowViolations > 0 && (
          <p className="text-xs text-red-400 mt-3">
            {timeWindowViolations} time-window violation(s) reported in solution.json.
          </p>
        )}
      </Card>
    </div>
  );
}
