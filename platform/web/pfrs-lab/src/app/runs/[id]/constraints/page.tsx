import { loadRunSummary, loadRoster } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import {
  buildJSSFeasibility,
  buildRoutingFeasibility,
} from '@/lib/feasibility-summary';
import {
  buildMachineLoadRows,
  buildRouteConstraintRows,
  validateJSSSolution,
} from '@/lib/constraint-analysis';
import { isJSSSolution, isRoutingSolution } from '@/lib/solution-types';
import { resolveProblemType } from '@/lib/resolve-problem-type';
import RunPageShell from '@/features/runs/RunPageShell';
import ConstraintAnalysis from './ConstraintAnalysis';
import RoutingConstraints from './RoutingConstraints';
import JSSConstraints from './JSSConstraints';
import JSSILPConstraints from './JSSILPConstraints';

export const dynamic = 'force-dynamic';

export default async function ConstraintsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const storage = getStorageProvider();
  const rawMeta = await storage.readFile(id, 'run.json');
  const meta = rawMeta ? JSON.parse(rawMeta) : {};
  const problemType = resolveProblemType(id, meta);

  try {
    if (problemType === 'cvrp' || problemType === 'vrptw') {
      const solutionContent = await storage.readFile(id, 'solution.json');
      if (!solutionContent) {
        return (
          <RunPageShell
            title="Constraints"
            empty
            emptyMessage="No solution.json — run owp solve cvrp or owp solve vrptw with --run-label first."
          >
            {null}
          </RunPageShell>
        );
      }

      const parsed = JSON.parse(solutionContent);
      if (!isRoutingSolution(parsed)) {
        return (
          <RunPageShell title="Constraints" error="solution.json is not a routing solution">
            {null}
          </RunPageShell>
        );
      }

      const capacity = Number(meta.capacity) || 0;
      const summary = buildRoutingFeasibility(parsed, problemType, capacity);
      const rows = buildRouteConstraintRows(parsed, capacity, problemType === 'vrptw');

      return (
        <RunPageShell title={problemType === 'vrptw' ? 'VRPTW Constraints' : 'CVRP Constraints'}>
          <RoutingConstraints
            problemType={problemType}
            summary={summary}
            rows={rows}
            timeWindowViolations={parsed.timeWindowViolations ?? 0}
          />
        </RunPageShell>
      );
    }

    if (problemType === 'jss') {
      const solutionContent = await storage.readFile(id, 'solution.json');
      if (!solutionContent) {
        const benchRaw = await storage.readFile(id, 'ilp-benchmark.json');
        const benchmark = benchRaw ? JSON.parse(benchRaw) : null;
        const isILP = String(meta.mode || '').toLowerCase() === 'ilp' || id.toLowerCase().includes('ilp-jss');
        if (isILP) {
          return (
            <RunPageShell title="JSS Constraints">
              <JSSILPConstraints runMeta={meta} benchmark={benchmark} />
            </RunPageShell>
          );
        }
        return (
          <RunPageShell
            title="Constraints"
            empty
            emptyMessage="No solution.json — run owp solve jss with --run-label first."
          >
            {null}
          </RunPageShell>
        );
      }

      const parsed = JSON.parse(solutionContent);
      if (!isJSSSolution(parsed)) {
        return (
          <RunPageShell title="Constraints" error="solution.json is not a JSS schedule">
            {null}
          </RunPageShell>
        );
      }

      const summary = buildJSSFeasibility(parsed);
      const violations = validateJSSSolution(parsed);
      const machineLoads = buildMachineLoadRows(parsed);

      return (
        <RunPageShell title="JSS Constraints">
          <JSSConstraints summary={summary} violations={violations} machineLoads={machineLoads} />
        </RunPageShell>
      );
    }

    // NRP (default).
    const [summary, roster, breakdownRaw] = await Promise.all([
      loadRunSummary(id),
      loadRoster(id),
      storage.readFile(id, 'constraint-breakdown.json'),
    ]);

    let exportedBreakdown: {
      totalPenalty: number;
      numWeeks: number;
      hardViolations: number;
      constraints: { id: string; penalty: number; violations: number }[];
    } | null = null;

    if (breakdownRaw) {
      try {
        exportedBreakdown = JSON.parse(breakdownRaw);
      } catch {
        exportedBreakdown = null;
      }
    }

    const totalPenalty = exportedBreakdown?.totalPenalty
      ?? summary.totalPenalty
      ?? (Number(meta.objective) || 0);
    const numWeeks = exportedBreakdown?.numWeeks
      ?? summary.numWeeks
      ?? (Number(meta.weeks) || 0);

    return (
      <RunPageShell title="Constraints">
        <ConstraintAnalysis
          weeks={summary.weeks}
          totalPenalty={totalPenalty}
          roster={roster}
          numWeeks={numWeeks}
          exportedBreakdown={exportedBreakdown}
        />
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Constraints" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
