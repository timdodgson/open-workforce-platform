import { loadRunSummary } from '@/lib/data-loader';
import MetricCard from '@/components/MetricCard';
import Card from '@/components/Card';

function fmt(n: number): string {
  return n.toLocaleString('en-US');
}

export default async function SummaryPage() {
  const d = await loadRunSummary();
  const meta = d.metadata;
  const prev = d.previousBest;

  // Week 8 / horizon pressure analysis
  const avgWeek1to7 = d.weeks.length > 1
    ? d.weeks.slice(0, -1).reduce((s, w) => s + w.finalPenalty, 0) / (d.weeks.length - 1)
    : 0;
  const lastWeek = d.weeks[d.weeks.length - 1];
  const lastWeekPct = d.totalPenalty > 0
    ? (d.maxWeekPenalty / d.totalPenalty * 100) : 0;
  const lastWeekMultiple = avgWeek1to7 > 0
    ? (d.maxWeekPenalty / avgWeek1to7) : 0;

  // Improvement over previous best
  const improvementAbs = prev ? prev.bestPenalty - d.totalPenalty : null;
  const improvementPct = prev && prev.bestPenalty > 0
    ? ((prev.bestPenalty - d.totalPenalty) / prev.bestPenalty * 100) : null;

  return (
    <div>
      {/* Experiment Configuration */}
      <Card title="Experiment">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
          <MetricCard label="Algorithm" value={meta?.algorithm || d.weeks[0]?.mode || '—'} />
          <MetricCard label="Cooling Mode" value={meta?.coolingMode || d.weeks[0]?.coolingMode || '—'} />
          <MetricCard label="Beam Width" value={meta ? String(meta.beamWidth) : '—'} />
          <MetricCard label="Beam Seeds" value={meta ? meta.beamSeeds.join(', ') : '—'} />
          <MetricCard label="Candidate Budget" value={meta ? fmt(meta.iterationsPerWorker) : fmt(d.weeks[0]?.iterationsPerWorker || 0)} />
          <MetricCard label="Initial Temp" value={meta ? String(meta.initialTemperature) : String(d.weeks[0]?.initialTemperature || 0)} />
          <MetricCard label="Effective Cooling" value={meta ? meta.effectiveCoolingRate.toFixed(10) : (d.weeks[0]?.effectiveCoolingRate || 0).toFixed(10)} />
          <MetricCard label="CPUs" value={meta ? String(meta.cpus) : String(d.weeks[0]?.maxConcurrent || '—')} />
          <MetricCard label="Instance" value={meta?.instance || d.weeks[0]?.instance || '—'} />
          <MetricCard label="Max Workers" value={meta ? String(meta.maxTotalWorkers) : String(d.weeks[0]?.maxTotalWorkers || 0)} />
        </div>
      </Card>

      {/* Current Result */}
      <Card title="Current Result">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
          <MetricCard label="Total Penalty" value={fmt(d.totalPenalty)} color="green" />
          <MetricCard label="Previous Best" value={prev ? fmt(prev.bestPenalty) : '—'} color="blue" />
          <MetricCard label="Improvement" value={improvementAbs !== null ? `${fmt(improvementAbs)} (${improvementPct!.toFixed(1)}%)` : '—'} color={improvementAbs !== null && improvementAbs > 0 ? 'green' : 'default'} />
          <MetricCard label="Runtime" value={`${(d.totalDurationMs / 1000).toFixed(1)}s`} color="blue" />
          <MetricCard label="Total Candidates" value={fmt(d.totalCandidates)} color="blue" />
          <MetricCard label="Workers Used" value={fmt(d.totalWorkers)} color="blue" />
          <MetricCard label="Branches" value={fmt(d.totalBranches)} color="blue" />
          <MetricCard label="Weeks" value={String(d.numWeeks)} color="blue" />
        </div>
      </Card>

      {/* Horizon Pressure */}
      <Card title="Horizon Pressure Analysis">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-3">
          <MetricCard label={`Week ${d.maxWeekNum} Penalty`} value={fmt(d.maxWeekPenalty)} color="amber" />
          <MetricCard label="% of Total" value={`${lastWeekPct.toFixed(1)}%`} color="amber" />
          <MetricCard label="vs Avg Weeks 1-N" value={`${lastWeekMultiple.toFixed(1)}x`} color="amber" />
          <MetricCard label="Rank" value={`1 of ${d.numWeeks}`} />
        </div>
        <p className="text-xs text-gray-400">
          Week {d.maxWeekNum} contributes {fmt(d.maxWeekPenalty)} / {fmt(d.totalPenalty)} = {lastWeekPct.toFixed(1)}%.
          This is {lastWeekMultiple.toFixed(1)}x the average of other weeks.
        </p>
      </Card>

      {/* Research Findings */}
      <Card title="Research Findings">
        <div className="space-y-2 text-sm">
          <Finding text={`Week ${d.maxWeekNum} contributes ${lastWeekPct.toFixed(0)}% of total penalty`} type="warn" />
          <Finding text={`Beam retained up to ${Math.max(...d.weeks.map(w => w.workersStarted))} histories per week`} />
          <Finding text={`Accepted worse moves ${d.acceptWorseRate.toFixed(1)}%`} />
          <Finding text={`Hard reject rate ${d.hardRejectRate.toFixed(0)}%`} type={d.hardRejectRate > 70 ? 'warn' : 'ok'} />
          {d.weeks[0]?.coolingMode === 'adaptive' && <Finding text="Adaptive cooling active" />}
          {d.weeks.every(w => w.hardViolations === 0)
            ? <Finding text="Beam search completed successfully — all weeks hard-feasible" />
            : <Finding text="Some weeks have hard violations" type="warn" />}
          {d.totalCandidates > 0 && (() => {
            const bestEff = d.weeks.reduce((best, w) => {
              const eff = w.candidates > 0 ? w.improvement / (w.candidates / 1_000_000) : 0;
              return eff > best.eff ? { eff, week: w.week } : best;
            }, { eff: 0, week: 0 });
            return bestEff.eff > 0
              ? <Finding text={`Candidate efficiency highest in week ${bestEff.week} (${bestEff.eff.toFixed(1)} penalty/M candidates)`} type="info" />
              : null;
          })()}
        </div>
      </Card>

      {/* Previous Best */}
      {!prev && (
        <Card title="Previous Best">
          <p className="text-sm text-gray-500">No previous best configured. Add <code className="text-xs bg-gray-800 px-1 rounded">data/best.json</code> to enable comparison.</p>
        </Card>
      )}
    </div>
  );
}

function Finding({ text, type = 'ok' }: { text: string; type?: 'ok' | 'warn' | 'info' }) {
  const styles = {
    ok: 'border-emerald-500 bg-emerald-500/10',
    warn: 'border-amber-500 bg-amber-500/10',
    info: 'border-blue-500 bg-blue-500/10',
  };
  return (
    <div className={`border-l-2 ${styles[type]} px-3 py-1.5 rounded-r`}>
      <span className="mr-2">{type === 'warn' ? '⚠' : '✓'}</span>{text}
    </div>
  );
}
