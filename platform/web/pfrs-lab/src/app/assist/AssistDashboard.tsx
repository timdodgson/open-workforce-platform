'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { UnifiedAssistRecord, WorkerAssistRecord, SearchAssistRecord, PortfolioAssistRecord } from './page';

interface Props {
  records: UnifiedAssistRecord[];
}

export default function AssistDashboard({ records }: Props) {
  // Split by architecture.
  const workerRecords = records.filter((r): r is WorkerAssistRecord => r.architecture === 'worker');
  const searchRecords = records.filter((r): r is SearchAssistRecord => r.architecture === 'search');
  const portfolioRecords = records.filter((r): r is PortfolioAssistRecord => r.architecture === 'portfolio');

  // Domains present.
  const domains = useMemo(() => [...new Set(records.map(r => r.domain))].sort(), [records]);

  // Overall stats.
  const stats = useMemo(() => {
    const total = records.length;

    // Accepted/rejected across all types.
    let accepted = 0;
    let rejected = 0;
    let safetyOverrides = 0;

    for (const r of workerRecords) {
      if (r.outcome === 'accepted') accepted++;
      else rejected++;
      if (r.safetyTriggered) safetyOverrides++;
    }
    for (const r of searchRecords) {
      if (r.accepted) accepted++;
      else rejected++;
      if (r.safetyTriggered) safetyOverrides++;
    }
    for (const r of portfolioRecords) {
      if (r.accepted) accepted++;
      else rejected++;
      if (r.safetyRejected) safetyOverrides++;
    }

    // CPU/objective impact.
    const workersSkipped = workerRecords.filter(r => r.finalAction === 'skip').length;
    const globalBestsMissed = workerRecords.filter(r => r.finalAction === 'skip' && r.producedGlobalBest).length;
    const earlyStops = searchRecords.filter(r => r.finalAction === 'early_stop' && r.accepted).length;
    const budgetAdjusted = portfolioRecords.filter(r => r.accepted && r.finalBudget !== r.originalBudget).length;

    return { total, accepted, rejected, safetyOverrides, workersSkipped, globalBestsMissed, earlyStops, budgetAdjusted, domains };
  }, [records, workerRecords, searchRecords, portfolioRecords, domains]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title="Search Intelligence — Assist Analysis">
        <p className="text-xs text-gray-500 mb-4">
          AI advisory decisions across all solver architectures. Each assist type integrates
          differently depending on the solver&apos;s architecture.
        </p>

        {/* Architecture breakdown */}
        <div className="grid grid-cols-3 gap-3 mb-4">
          <ArchCard label="Worker Assist" icon="👷" count={workerRecords.length} domain="NRP (beam search)" active={workerRecords.length > 0} />
          <ArchCard label="Search Assist" icon="🔍" count={searchRecords.length} domain="SA / LAHC / Tabu" active={searchRecords.length > 0} />
          <ArchCard label="Portfolio Assist" icon="📊" count={portfolioRecords.length} domain="CVRP / JSS / VRPTW" active={portfolioRecords.length > 0} />
        </div>

        {/* Key metrics */}
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2">
          <Stat label="Total Decisions" value={stats.total} colour="blue" />
          <Stat label="Accepted" value={stats.accepted} colour="emerald" />
          <Stat label="Rejected" value={stats.rejected} colour="amber" />
          <Stat label="Safety Overrides" value={stats.safetyOverrides} colour="red" />
          <Stat label="Workers Skipped" value={stats.workersSkipped} colour="blue" />
          <Stat label="GB Missed" value={stats.globalBestsMissed} colour={stats.globalBestsMissed > 0 ? 'red' : 'emerald'} />
          <Stat label="Early Stops" value={stats.earlyStops} colour="amber" />
          <Stat label="Budget Adjusted" value={stats.budgetAdjusted} colour="blue" />
        </div>

        {/* Domains */}
        <div className="mt-3 flex gap-2">
          {domains.map(d => (
            <span key={d} className="text-[10px] px-2 py-0.5 rounded bg-gray-800 text-blue-400">{d}</span>
          ))}
        </div>
      </Card>

      {/* Worker Assist Section (NRP) */}
      {workerRecords.length > 0 && <WorkerAssistSection records={workerRecords} />}

      {/* Portfolio Assist Section (CVRP/JSS/VRPTW) */}
      {portfolioRecords.length > 0 && <PortfolioAssistSection records={portfolioRecords} />}

      {/* Search Assist Section */}
      {searchRecords.length > 0 && <SearchAssistSection records={searchRecords} />}
    </div>
  );
}

// --- Worker Assist Section ---

function WorkerAssistSection({ records }: { records: WorkerAssistRecord[] }) {
  const accepted = records.filter(r => r.outcome === 'accepted').length;
  const skipped = records.filter(r => r.finalAction === 'skip').length;
  const gbMissed = records.filter(r => r.finalAction === 'skip' && r.producedGlobalBest).length;
  const safetyCount = records.filter(r => r.safetyTriggered).length;

  return (
    <Card title="👷 Worker Assist — NRP Beam Search">
      <p className="text-xs text-gray-500 mb-3">
        Per-worker spawn decisions in the beam search. Workers can be run, skipped, or have budget adjusted.
      </p>
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 mb-3">
        <Stat label="Decisions" value={records.length} colour="blue" />
        <Stat label="Accepted" value={accepted} colour="emerald" />
        <Stat label="Workers Skipped" value={skipped} colour="amber" />
        <Stat label="GB Missed" value={gbMissed} colour={gbMissed > 0 ? 'red' : 'emerald'} />
        <Stat label="Safety Overrides" value={safetyCount} colour="red" />
      </div>
      <div className="overflow-x-auto max-h-[250px] overflow-y-auto">
        <table className="w-full text-[10px]">
          <thead className="sticky top-0 bg-gray-850">
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-1">Worker</th>
              <th className="text-left p-1">Algo</th>
              <th className="text-left p-1">Recommendation</th>
              <th className="text-right p-1">Confidence</th>
              <th className="text-center p-1">Safety</th>
              <th className="text-left p-1">Outcome</th>
              <th className="text-center p-1">Improved?</th>
              <th className="text-center p-1">GB?</th>
            </tr>
          </thead>
          <tbody>
            {records.slice(-30).reverse().map((r, i) => (
              <tr key={i} className={`border-t border-gray-800 ${r.safetyTriggered ? 'bg-red-900/10' : ''}`}>
                <td className="p-1">{r.workerId}</td>
                <td className="p-1 text-emerald-400">{r.algorithm}</td>
                <td className="p-1 text-blue-400">{r.recommendation}</td>
                <td className="text-right p-1 text-amber-400">{r.confidence.toFixed(2)}</td>
                <td className="text-center p-1">{r.safetyTriggered ? <span className="text-red-400">⚠</span> : '—'}</td>
                <td className="p-1"><OutcomeBadge outcome={r.outcome} /></td>
                <td className="text-center p-1">{r.improved ? '✓' : '✗'}</td>
                <td className="text-center p-1">{r.producedGlobalBest ? '⭐' : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

// --- Portfolio Assist Section ---

function PortfolioAssistSection({ records }: { records: PortfolioAssistRecord[] }) {
  const accepted = records.filter(r => r.accepted).length;
  const safetyCount = records.filter(r => r.safetyRejected).length;
  const budgetChanged = records.filter(r => r.finalBudget !== r.originalBudget).length;
  const winners = records.filter(r => r.strategyWon);

  // Group by domain.
  const domains = [...new Set(records.map(r => r.domain))];

  return (
    <Card title="📊 Portfolio Assist — Budget Allocation">
      <p className="text-xs text-gray-500 mb-3">
        Per-strategy budget decisions in portfolio runs. The AI advises how to distribute
        iterations across SA, LAHC, and Tabu.
      </p>
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 mb-3">
        <Stat label="Strategy Decisions" value={records.length} colour="blue" />
        <Stat label="Accepted" value={accepted} colour="emerald" />
        <Stat label="Budget Changed" value={budgetChanged} colour="amber" />
        <Stat label="Safety Rejected" value={safetyCount} colour="red" />
        <Stat label="Winners Advised" value={winners.length} colour="emerald" />
      </div>

      {/* Per-domain breakdown */}
      {domains.map(domain => {
        const domainRecs = records.filter(r => r.domain === domain);
        return (
          <div key={domain} className="mb-3">
            <h4 className="text-xs text-blue-400 uppercase font-semibold mb-2">{domain}</h4>
            <div className="overflow-x-auto">
              <table className="w-full text-[10px]">
                <thead>
                  <tr className="text-gray-500 uppercase">
                    <th className="text-left p-1">Instance</th>
                    <th className="text-left p-1">Strategy</th>
                    <th className="text-right p-1">Original</th>
                    <th className="text-right p-1">Recommended</th>
                    <th className="text-right p-1">Final</th>
                    <th className="text-left p-1">Action</th>
                    <th className="text-right p-1">Conf</th>
                    <th className="text-center p-1">Accepted</th>
                    <th className="text-right p-1">Result</th>
                    <th className="text-center p-1">Won?</th>
                  </tr>
                </thead>
                <tbody>
                  {domainRecs.slice(-20).map((r, i) => (
                    <tr key={i} className={`border-t border-gray-800 ${r.safetyRejected ? 'bg-red-900/10' : ''}`}>
                      <td className="p-1 text-gray-400 truncate max-w-[80px]">{r.instance}</td>
                      <td className="p-1 text-emerald-400">{r.strategy}</td>
                      <td className="text-right p-1">{(r.originalBudget / 1000).toFixed(0)}K</td>
                      <td className="text-right p-1 text-blue-400">{(r.recommendedBudget / 1000).toFixed(0)}K</td>
                      <td className="text-right p-1 font-semibold">{(r.finalBudget / 1000).toFixed(0)}K</td>
                      <td className="p-1"><ActionBadge action={r.recommendation} /></td>
                      <td className="text-right p-1 text-amber-400">{r.confidence.toFixed(2)}</td>
                      <td className="text-center p-1">{r.accepted ? <span className="text-emerald-400">✓</span> : <span className="text-gray-600">✗</span>}</td>
                      <td className="text-right p-1">{r.resultObjective > 0 ? r.resultObjective.toLocaleString() : '—'}</td>
                      <td className="text-center p-1">{r.strategyWon ? '★' : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        );
      })}
    </Card>
  );
}

// --- Search Assist Section ---

function SearchAssistSection({ records }: { records: SearchAssistRecord[] }) {
  const actioned = records.filter(r => r.recommendedAction !== 'continue');
  const accepted = actioned.filter(r => r.accepted).length;
  const earlyStops = records.filter(r => r.finalAction === 'early_stop' && r.accepted).length;
  const safetyCount = records.filter(r => r.safetyTriggered).length;

  return (
    <Card title="🔍 Search Assist — Single-Algorithm Hooks">
      <p className="text-xs text-gray-500 mb-3">
        Periodic checkpoints during search execution. The AI monitors progress and may recommend
        early stop, budget adjustment, or continuation.
      </p>
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 mb-3">
        <Stat label="Checkpoints" value={records.length} colour="blue" />
        <Stat label="Recommendations" value={actioned.length} colour="amber" />
        <Stat label="Accepted" value={accepted} colour="emerald" />
        <Stat label="Early Stops" value={earlyStops} colour="amber" />
        <Stat label="Safety Blocks" value={safetyCount} colour="red" />
      </div>
      <div className="overflow-x-auto max-h-[250px] overflow-y-auto">
        <table className="w-full text-[10px]">
          <thead className="sticky top-0 bg-gray-850">
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-1">Domain</th>
              <th className="text-left p-1">Algo</th>
              <th className="text-right p-1">Cands</th>
              <th className="text-right p-1">Best</th>
              <th className="text-right p-1">Plateau</th>
              <th className="text-left p-1">Rec</th>
              <th className="text-right p-1">Conf</th>
              <th className="text-center p-1">Safety</th>
              <th className="text-left p-1">Final</th>
            </tr>
          </thead>
          <tbody>
            {actioned.slice(-30).reverse().map((r, i) => (
              <tr key={i} className={`border-t border-gray-800 ${r.safetyTriggered ? 'bg-red-900/10' : ''}`}>
                <td className="p-1 text-blue-400">{r.domain}</td>
                <td className="p-1 text-emerald-400">{r.algorithm}</td>
                <td className="text-right p-1">{(r.candidates / 1000).toFixed(0)}K</td>
                <td className="text-right p-1">{r.bestPenalty.toLocaleString()}</td>
                <td className="text-right p-1 text-amber-400">{(r.plateauLength / 1000).toFixed(0)}K</td>
                <td className="p-1"><ActionBadge action={r.recommendedAction} /></td>
                <td className="text-right p-1 text-amber-400">{r.confidence.toFixed(2)}</td>
                <td className="text-center p-1">{r.safetyTriggered ? <span className="text-red-400">⚠</span> : '—'}</td>
                <td className="p-1"><ActionBadge action={r.finalAction} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}

// --- Shared Components ---

function Stat({ label, value, colour }: { label: string; value: string | number; colour: string }) {
  const colourMap: Record<string, string> = {
    blue: 'text-blue-400', emerald: 'text-emerald-400', amber: 'text-amber-400', red: 'text-red-400',
  };
  return (
    <div className="bg-gray-800 rounded p-2">
      <div className="text-[8px] text-gray-500 uppercase">{label}</div>
      <div className={`text-sm font-bold ${colourMap[colour] || 'text-gray-300'}`}>{value}</div>
    </div>
  );
}

function ArchCard({ label, icon, count, domain, active }: { label: string; icon: string; count: number; domain: string; active: boolean }) {
  return (
    <div className={`rounded-lg p-3 border ${active ? 'border-blue-700 bg-blue-900/10' : 'border-gray-800 bg-gray-800/50'}`}>
      <div className="flex items-center gap-2 mb-1">
        <span className="text-lg">{icon}</span>
        <span className={`text-xs font-semibold ${active ? 'text-blue-400' : 'text-gray-600'}`}>{label}</span>
      </div>
      <div className={`text-xl font-bold ${active ? 'text-gray-200' : 'text-gray-700'}`}>{count}</div>
      <div className="text-[9px] text-gray-500">{domain}</div>
    </div>
  );
}

function OutcomeBadge({ outcome }: { outcome: string }) {
  const styles: Record<string, string> = {
    accepted: 'text-emerald-400',
    rejected: 'text-red-400',
    overridden: 'text-amber-400',
  };
  return <span className={styles[outcome] || 'text-gray-400'}>{outcome}</span>;
}

function ActionBadge({ action }: { action: string }) {
  const styles: Record<string, string> = {
    run: 'bg-gray-700 text-gray-300',
    skip: 'bg-red-900/50 text-red-300',
    reduce_budget: 'bg-amber-900/50 text-amber-300',
    boost_budget: 'bg-emerald-900/50 text-emerald-300',
    early_stop: 'bg-red-900/50 text-red-300',
    adjust_budget: 'bg-amber-900/50 text-amber-300',
    continue: 'bg-gray-700 text-gray-500',
  };
  return (
    <span className={`text-[9px] px-1.5 py-0.5 rounded ${styles[action] || 'bg-gray-700 text-gray-400'}`}>
      {action}
    </span>
  );
}
