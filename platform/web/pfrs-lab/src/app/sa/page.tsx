import { loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import SACharts from './SACharts';

export default async function SAPage() {
  const d = await loadRunSummary();

  const chartData = d.weeks.map(w => ({
    week: `W${w.week}`,
    better: w.saAcceptedBetter,
    worse: w.saAcceptedWorse,
    rejectedProb: w.saRejectedByProb,
    worsePct: w.candidates > 0 ? +(w.saAcceptedWorse / w.candidates * 100).toFixed(2) : 0,
  }));

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

      <Card title="Temperature Curve">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          Temperature decay curve will be available when per-iteration samples are exported.
        </div>
      </Card>

      <Card title="Plateau Detection">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          Plateau detection will identify stagnation regions in future iterations.
        </div>
      </Card>
    </div>
  );
}
