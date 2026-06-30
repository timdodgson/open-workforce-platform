import { loadRunSummary, loadPlateaus, loadBranches, loadWorkerLifecycles } from '@/lib/data-loader';
import Card from '@/components/Card';
import SACharts from './SACharts';
import TemperatureCurve from './TemperatureCurve';
import PlateauDetection from './PlateauDetection';
import WorkerLifecycles from './WorkerLifecycles';

export default async function SAPage() {
  const [d, plateaus, branches, workerLifecycles] = await Promise.all([
    loadRunSummary(), loadPlateaus(), loadBranches(), loadWorkerLifecycles()
  ]);

  const chartData = d.weeks.map(w => ({
    week: `W${w.week}`,
    better: w.saAcceptedBetter,
    worse: w.saAcceptedWorse,
    rejectedProb: w.saRejectedByProb,
    worsePct: w.candidates > 0 ? +(w.saAcceptedWorse / w.candidates * 100).toFixed(2) : 0,
  }));

  // Temperature curve config — from run.json metadata or CSV first row.
  const meta = d.metadata;
  const first = d.weeks[0];
  const initialTemp = meta?.initialTemperature ?? first?.initialTemperature ?? 100;
  const coolingMode = meta?.coolingMode ?? first?.coolingMode ?? 'adaptive';
  const effectiveRate = meta?.effectiveCoolingRate ?? first?.effectiveCoolingRate ?? 0;
  const iterations = meta?.iterationsPerWorker ?? first?.iterationsPerWorker ?? 500000;
  const minTemp = first?.minTemperature ?? 0.0001;

  return (
    <div>
      <SACharts data={chartData} />

      <Card title="Temperature Profile">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Initial</th>
              <th className="text-right p-2">At Best</th>
              <th className="text-right p-2">Final</th>
              <th className="text-right p-2">Better</th>
              <th className="text-right p-2">Worse</th>
              <th className="text-right p-2">Rej Prob</th>
              <th className="text-right p-2">Worse %</th>
            </tr>
          </thead>
          <tbody>
            {d.weeks.map(w => (
              <tr key={w.week} className="border-t border-gray-800">
                <td className="p-2">{w.week}</td>
                <td className="text-right p-2">{w.initialTemperature.toFixed(1)}</td>
                <td className="text-right p-2">{w.saTempAtBest.toFixed(4)}</td>
                <td className="text-right p-2">{w.saFinalTemp.toFixed(6)}</td>
                <td className="text-right p-2">{w.saAcceptedBetter.toLocaleString()}</td>
                <td className="text-right p-2">{w.saAcceptedWorse.toLocaleString()}</td>
                <td className="text-right p-2">{w.saRejectedByProb.toLocaleString()}</td>
                <td className="text-right p-2">{w.candidates > 0 ? (w.saAcceptedWorse / w.candidates * 100).toFixed(1) : '—'}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      <TemperatureCurve
        initialTemperature={initialTemp}
        effectiveCoolingRate={effectiveRate}
        iterationsPerWorker={iterations}
        coolingMode={coolingMode}
        minTemperature={minTemp}
        branches={branches}
      />

      <PlateauDetection events={plateaus} />

      <WorkerLifecycles workers={workerLifecycles} />
    </div>
  );
}
