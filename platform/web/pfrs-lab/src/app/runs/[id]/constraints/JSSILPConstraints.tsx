import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';

interface ILPDatum {
  objective?: number;
  lowerBound?: number;
  gapPercent?: number;
  status?: string;
  runtimeSeconds?: number;
  jobs?: number;
  machines?: number;
  operations?: number;
}

interface Props {
  runMeta: Record<string, unknown>;
  benchmark: ILPDatum | null;
}

export default function JSSILPConstraints({ runMeta, benchmark }: Props) {
  const objective = Number(benchmark?.objective ?? runMeta.objective ?? runMeta.bestMakespan ?? 0);
  const bound = Number(benchmark?.lowerBound ?? runMeta.bound ?? 0);
  const gap = Number(benchmark?.gapPercent ?? runMeta.gap ?? 0);
  const status = String(benchmark?.status ?? runMeta.status ?? '—');

  return (
    <Card title="JSS ILP Proof (HiGHS)">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
        <MetricCard label="Status" value={status} color={status === 'OPTIMAL' ? 'green' : 'amber'} />
        <MetricCard label="Makespan" value={objective.toLocaleString()} color="green" />
        <MetricCard label="Lower bound" value={bound > 0 ? bound.toLocaleString() : '—'} color="blue" />
        <MetricCard label="Gap" value={bound > 0 ? `${gap.toFixed(2)}%` : '—'} color="default" />
        <MetricCard label="Jobs" value={String(benchmark?.jobs ?? runMeta.jobs ?? '—')} color="default" />
        <MetricCard label="Machines" value={String(benchmark?.machines ?? runMeta.machines ?? '—')} color="default" />
        <MetricCard
          label="Runtime"
          value={`${Number(benchmark?.runtimeSeconds ?? runMeta.runtime ?? 0).toFixed(1)}s`}
          color="default"
        />
        <MetricCard label="Solver" value="HiGHS ILP" color="default" />
      </div>
      <p className="text-xs text-amber-400">
        This run has no solution.json schedule yet. Re-run benchmark-jss-ilp to export operation timings,
        or open Summary / Gantt after solution export. The NRP S1–S8 panel does not apply to job shop runs.
      </p>
    </Card>
  );
}
