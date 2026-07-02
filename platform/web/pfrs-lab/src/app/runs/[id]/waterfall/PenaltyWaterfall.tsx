'use client';

import Card from '@/components/Card';
import { WeekRecord } from '@/lib/types';

interface Props {
  weeks: WeekRecord[];
  totalPenalty: number;
}

interface WaterfallBar {
  week: number;
  penalty: number;
  cumulative: number;
  percentage: number;
  isLargest: boolean;
  isSmallest: boolean;
}

function generateCommentary(bars: WaterfallBar[], totalPenalty: number): string[] {
  const comments: string[] = [];
  const largest = bars.find(b => b.isLargest);
  const smallest = bars.find(b => b.isSmallest);

  if (largest) {
    comments.push(
      `Week ${largest.week} contributed the most penalty: ${largest.penalty.toLocaleString()} (${largest.percentage.toFixed(1)}% of total).`
    );
  }
  if (smallest && smallest.week !== largest?.week) {
    comments.push(
      `Week ${smallest.week} was the lightest: ${smallest.penalty.toLocaleString()} (${smallest.percentage.toFixed(1)}%).`
    );
  }

  // Trend analysis.
  if (bars.length >= 4) {
    const firstHalf = bars.slice(0, Math.floor(bars.length / 2));
    const secondHalf = bars.slice(Math.floor(bars.length / 2));
    const firstAvg = firstHalf.reduce((s, b) => s + b.penalty, 0) / firstHalf.length;
    const secondAvg = secondHalf.reduce((s, b) => s + b.penalty, 0) / secondHalf.length;

    if (secondAvg > firstAvg * 1.3) {
      comments.push('Penalty increased significantly in later weeks — history accumulation is costly.');
    } else if (firstAvg > secondAvg * 1.3) {
      comments.push('Penalty decreased in later weeks — the solver adapted well to constraints.');
    } else {
      comments.push('Penalty remained relatively stable across weeks.');
    }
  }

  // Concentration.
  if (largest && largest.percentage > 30) {
    comments.push(`⚠️ High concentration: Week ${largest.week} alone accounts for ${largest.percentage.toFixed(0)}% of total penalty.`);
  }

  return comments;
}

export default function PenaltyWaterfall({ weeks, totalPenalty }: Props) {
  // Build waterfall data.
  const maxPenalty = Math.max(...weeks.map(w => w.finalPenalty), 1);
  const minPenalty = Math.min(...weeks.map(w => w.finalPenalty));

  let cumulative = 0;
  const bars: WaterfallBar[] = weeks.map(w => {
    cumulative += w.finalPenalty;
    return {
      week: w.week,
      penalty: w.finalPenalty,
      cumulative,
      percentage: totalPenalty > 0 ? (w.finalPenalty / totalPenalty) * 100 : 0,
      isLargest: w.finalPenalty === maxPenalty,
      isSmallest: w.finalPenalty === minPenalty,
    };
  });

  const commentary = generateCommentary(bars, totalPenalty);

  return (
    <div className="space-y-4">
      {/* Summary */}
      <Card title="Penalty Waterfall">
        <div className="grid grid-cols-3 gap-4 text-center mb-6">
          <div>
            <p className="text-2xl font-bold text-emerald-400">{totalPenalty.toLocaleString()}</p>
            <p className="text-[10px] text-gray-500">Total Penalty</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-blue-400">{weeks.length}</p>
            <p className="text-[10px] text-gray-500">Weeks</p>
          </div>
          <div>
            <p className="text-2xl font-bold text-gray-300">
              {Math.round(totalPenalty / Math.max(weeks.length, 1)).toLocaleString()}
            </p>
            <p className="text-[10px] text-gray-500">Avg per Week</p>
          </div>
        </div>

        {/* Waterfall chart */}
        <div className="relative h-64 flex items-end gap-2 px-4">
          {bars.map(bar => {
            // Waterfall: each bar starts where the previous ended.
            const barHeight = (bar.penalty / totalPenalty) * 100;
            const bottomOffset = ((bar.cumulative - bar.penalty) / totalPenalty) * 100;

            let barClass = 'bg-blue-600';
            if (bar.isLargest) barClass = 'bg-red-500';
            else if (bar.isSmallest) barClass = 'bg-emerald-500';

            return (
              <div key={bar.week} className="flex-1 relative h-full flex flex-col justify-end">
                {/* The floating bar */}
                <div className="relative" style={{ marginBottom: `${bottomOffset}%` }}>
                  <div
                    className={`${barClass} rounded-t relative group cursor-default transition-all hover:opacity-80`}
                    style={{ height: `${Math.max(barHeight * 2.5, 8)}px` }}
                    title={`Week ${bar.week}: ${bar.penalty.toLocaleString()} (${bar.percentage.toFixed(1)}%)`}
                  >
                    {/* Penalty label */}
                    <span className="absolute -top-5 left-1/2 -translate-x-1/2 text-[9px] text-gray-300 whitespace-nowrap">
                      {bar.penalty.toLocaleString()}
                    </span>
                  </div>
                </div>
                {/* Week label */}
                <span className="text-[9px] text-gray-500 text-center mt-1 absolute bottom-0 left-0 right-0">
                  W{bar.week}
                </span>
              </div>
            );
          })}

          {/* Total bar */}
          <div className="flex-1 relative h-full flex flex-col justify-end border-l border-gray-700 pl-2">
            <div
              className="bg-gray-500 rounded-t"
              style={{ height: '100%' }}
            >
              <span className="absolute -top-5 left-1/2 -translate-x-1/2 text-[9px] text-gray-200 font-bold whitespace-nowrap">
                {totalPenalty.toLocaleString()}
              </span>
            </div>
            <span className="text-[9px] text-gray-400 text-center mt-1 absolute bottom-0 left-0 right-0">
              Total
            </span>
          </div>
        </div>
      </Card>

      {/* Detail table */}
      <Card title="Week Breakdown">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Penalty</th>
              <th className="text-right p-2">Cumulative</th>
              <th className="text-right p-2">% of Total</th>
              <th className="p-2">Contribution</th>
              <th className="text-center p-2">Flag</th>
            </tr>
          </thead>
          <tbody>
            {bars.map(bar => (
              <tr key={bar.week} className={`border-t border-gray-800 ${
                bar.isLargest ? 'bg-red-900/10' : bar.isSmallest ? 'bg-emerald-900/10' : ''
              }`}>
                <td className="p-2 font-medium">Week {bar.week}</td>
                <td className="text-right p-2 font-mono">{bar.penalty.toLocaleString()}</td>
                <td className="text-right p-2 text-gray-400">{bar.cumulative.toLocaleString()}</td>
                <td className="text-right p-2">{bar.percentage.toFixed(1)}%</td>
                <td className="p-2">
                  <div className="h-3 bg-gray-800 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${
                        bar.isLargest ? 'bg-red-500' : bar.isSmallest ? 'bg-emerald-500' : 'bg-blue-600'
                      }`}
                      style={{ width: `${bar.percentage}%` }}
                    />
                  </div>
                </td>
                <td className="text-center p-2">
                  {bar.isLargest && <span className="text-red-400 text-[10px]">▲ highest</span>}
                  {bar.isSmallest && <span className="text-emerald-400 text-[10px]">▼ lowest</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Commentary */}
      <Card title="Analysis">
        <div className="space-y-2">
          {commentary.map((c, i) => (
            <p key={i} className="text-sm text-gray-300 leading-relaxed">
              {c}
            </p>
          ))}
        </div>
      </Card>
    </div>
  );
}
