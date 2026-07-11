import { loadRunSummary } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import { buildJSSFeasibility, buildRoutingFeasibility } from '@/lib/feasibility-summary';
import { isJSSSolution, isRoutingSolution } from '@/lib/solution-types';
import type { FeasibilitySummary } from '@/lib/feasibility-summary';
import {
  CVRPSummaryView,
  ILPSummaryView,
  JSSSummaryView,
  NRPSummaryView,
  VRPTWSummaryView,
} from './SummaryViews';

async function loadFeasibilitySummary(
  runId: string,
  problemType: string,
  capacity: number,
): Promise<FeasibilitySummary | null> {
  const content = await getStorageProvider().readFile(runId, 'solution.json');
  if (!content) return null;
  try {
    const parsed = JSON.parse(content);
    if (problemType === 'jss' && isJSSSolution(parsed)) {
      return buildJSSFeasibility(parsed);
    }
    if ((problemType === 'cvrp' || problemType === 'vrptw') && isRoutingSolution(parsed)) {
      return buildRoutingFeasibility(parsed, problemType, capacity);
    }
  } catch {
    return null;
  }
  return null;
}

export async function renderRunSummary(id: string) {
  const [d, rawMeta] = await Promise.all([
    loadRunSummary(id),
    getStorageProvider().readFile(id, 'run.json'),
  ]);

  const meta = rawMeta ? JSON.parse(rawMeta) : {};
  const mode = meta.mode || d.metadata?.mode || '—';
  const problemType = meta.problemType || 'nrp';
  const isILP = mode === 'ilp';

  if (isILP) {
    return <ILPSummaryView id={id} meta={meta} />;
  }

  if (problemType === 'cvrp') {
    const feasibility = await loadFeasibilitySummary(id, 'cvrp', Number(meta.capacity) || 0);
    return (
      <CVRPSummaryView
        id={id}
        meta={meta}
        totalPenalty={d.totalPenalty}
        totalDurationMs={d.totalDurationMs}
        totalCandidates={d.totalCandidates}
        feasibility={feasibility}
      />
    );
  }

  if (problemType === 'jss') {
    const feasibility = await loadFeasibilitySummary(id, 'jss', 0);
    return (
      <JSSSummaryView
        id={id}
        meta={meta}
        totalPenalty={d.totalPenalty}
        totalDurationMs={d.totalDurationMs}
        feasibility={feasibility}
      />
    );
  }

  if (problemType === 'vrptw') {
    const feasibility = await loadFeasibilitySummary(id, 'vrptw', Number(meta.capacity) || 0);
    return <VRPTWSummaryView id={id} meta={meta} feasibility={feasibility} />;
  }

  return <NRPSummaryView id={id} d={d} />;
}
