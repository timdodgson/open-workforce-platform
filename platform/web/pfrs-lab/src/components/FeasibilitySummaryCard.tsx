import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import type { FeasibilitySummary } from '@/lib/feasibility-summary';

export default function FeasibilitySummaryCard({ summary }: { summary: FeasibilitySummary }) {
  if (summary.kind === 'jss') {
    return (
      <Card title="Schedule Feasibility">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <MetricCard label="Makespan" value={summary.makespan.toLocaleString()} color="green" />
          <MetricCard label="Operations" value={String(summary.operations)} color="default" />
          <MetricCard
            label="Avg machine load"
            value={`${summary.avgMachineUtilisation.toFixed(0)}%`}
            color="blue"
          />
          <MetricCard
            label="Bottleneck"
            value={`M${summary.bottleneckMachine + 1} (${summary.maxMachineUtilisation.toFixed(0)}%)`}
            color={summary.maxMachineUtilisation > 95 ? 'amber' : 'default'}
          />
        </div>
        <p className="text-xs text-gray-500">
          Machine load = busy time ÷ makespan per machine. JSS schedules are precedence-feasible by construction.
        </p>
      </Card>
    );
  }

  if (summary.kind === 'vrptw') {
    return (
      <Card title="Route Feasibility">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <MetricCard
            label="Feasible"
            value={summary.feasible ? '✓' : '✗'}
            color={summary.feasible ? 'green' : 'red'}
          />
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
          <MetricCard
            label="Capacity overloads"
            value={String(summary.overloadedRoutes)}
            color={summary.overloadedRoutes > 0 ? 'red' : 'green'}
          />
          <MetricCard label="Routes" value={String(summary.routes)} color="default" />
          <MetricCard
            label="Avg capacity %"
            value={`${summary.avgUtilisation.toFixed(0)}%`}
            color="blue"
          />
          <MetricCard
            label="Peak capacity %"
            value={`${summary.maxUtilisation.toFixed(0)}%`}
            color={summary.maxUtilisation > 95 ? 'amber' : 'default'}
          />
          <MetricCard label="Total load" value={summary.totalLoad.toLocaleString()} color="default" />
        </div>
        <p className="text-xs text-gray-500">
          Time-window and capacity checks from solution.json (per-route feasible flag + timeWindowViolations).
        </p>
      </Card>
    );
  }

  return (
    <Card title="Route Feasibility">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
        <MetricCard
          label="Feasible"
          value={summary.feasible ? '✓' : '✗'}
          color={summary.feasible ? 'green' : 'red'}
        />
        <MetricCard label="Routes" value={String(summary.routes)} color="default" />
        <MetricCard
          label="Capacity overloads"
          value={String(summary.overloadedRoutes)}
          color={summary.overloadedRoutes > 0 ? 'red' : 'green'}
        />
        <MetricCard
          label="Avg capacity %"
          value={`${summary.avgUtilisation.toFixed(0)}%`}
          color="blue"
        />
        <MetricCard
          label="Peak capacity %"
          value={`${summary.maxUtilisation.toFixed(0)}%`}
          color={summary.maxUtilisation > 95 ? 'amber' : 'default'}
        />
        <MetricCard label="Total load" value={summary.totalLoad.toLocaleString()} color="default" />
      </div>
      <p className="text-xs text-gray-500">
        Capacity utilisation per route vs instance capacity from run.json.
      </p>
    </Card>
  );
}
