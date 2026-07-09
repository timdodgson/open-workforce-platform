import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import type { JSSViolation, MachineLoadRow } from '@/lib/constraint-analysis';
import type { FeasibilitySummary } from '@/lib/feasibility-summary';

interface Props {
  summary: FeasibilitySummary;
  violations: JSSViolation[];
  machineLoads: MachineLoadRow[];
}

export default function JSSConstraints({ summary, violations, machineLoads }: Props) {
  if (summary.kind !== 'jss') return null;

  return (
    <div className="space-y-4">
      <Card title="Job Shop Constraints">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
          <MetricCard label="Makespan" value={summary.makespan.toLocaleString()} color="green" />
          <MetricCard
            label="Hard violations"
            value={String(violations.length)}
            color={violations.length > 0 ? 'red' : 'green'}
          />
          <MetricCard
            label="Avg machine load"
            value={`${summary.avgMachineUtilisation.toFixed(0)}%`}
            color="blue"
          />
          <MetricCard
            label="Bottleneck"
            value={`M${summary.bottleneckMachine + 1}`}
            color={summary.maxMachineUtilisation > 95 ? 'amber' : 'default'}
          />
        </div>

        <h3 className="text-xs uppercase text-gray-500 mb-2">Machine load</h3>
        <table className="w-full text-xs mb-4">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Machine</th>
              <th className="text-right p-2">Operations</th>
              <th className="text-right p-2">Busy time</th>
              <th className="text-right p-2">Load %</th>
            </tr>
          </thead>
          <tbody>
            {machineLoads.map((row) => (
              <tr key={row.machine} className="border-t border-gray-800">
                <td className="p-2 font-mono">M{row.machine + 1}</td>
                <td className="text-right p-2">{row.operations}</td>
                <td className="text-right p-2">{row.busyTime}</td>
                <td className={`text-right p-2 ${row.utilisation > 95 ? 'text-amber-400' : ''}`}>
                  {row.utilisation.toFixed(0)}%
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {violations.length > 0 ? (
          <>
            <h3 className="text-xs uppercase text-gray-500 mb-2">Violations</h3>
            <ul className="space-y-1 text-xs">
              {violations.map((v, i) => (
                <li key={i} className="text-red-400 font-mono">
                  [{v.code}] {v.detail}
                </li>
              ))}
            </ul>
          </>
        ) : (
          <p className="text-xs text-emerald-400">
            No precedence or machine-overlap violations detected in solution.json.
          </p>
        )}
      </Card>
    </div>
  );
}
