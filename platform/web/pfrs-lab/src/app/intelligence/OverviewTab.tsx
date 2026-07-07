'use client';

import Card from '@/components/Card';

export default function OverviewTab() {
  return (
    <div className="space-y-4">
      <Card title="Search Intelligence">
        <p className="text-xs text-gray-400 leading-relaxed mb-4">
          Search Intelligence is a universal advisory system that monitors search behaviour
          and recommends compute allocation decisions. It operates across all four domains
          (NRP, CVRP, JSS, VRPTW) through three integration styles.
        </p>

        {/* Architecture */}
        <div className="bg-gray-800 rounded-lg p-4 mb-4">
          <p className="text-[10px] text-gray-500 uppercase mb-3">Architecture</p>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <StyleCard
              name="WorkerAssist"
              target="NRP Beam Search"
              actions="Skip / reduce / increase workers"
            />
            <StyleCard
              name="SearchAssist"
              target="SA / LAHC / Tabu"
              actions="Early stop / budget extend / budget reduce"
            />
            <StyleCard
              name="PortfolioAssist"
              target="All Portfolio modes"
              actions="Learned budget allocation across strategies"
            />
          </div>
        </div>

        {/* Modes */}
        <div className="bg-gray-800 rounded-lg p-4 mb-4">
          <p className="text-[10px] text-gray-500 uppercase mb-3">Modes</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <ModeCard mode="off" desc="No intelligence. Zero overhead." colour="gray" />
            <ModeCard mode="shadow" desc="Records predictions. No behaviour change." colour="blue" />
            <ModeCard mode="assist" desc="Applies safe recommendations." colour="emerald" />
            <ModeCard mode="adaptive" desc="Live-updating decisions." colour="amber" />
          </div>
        </div>

        {/* Pipeline */}
        <div className="bg-gray-800 rounded-lg p-4 mb-4">
          <p className="text-[10px] text-gray-500 uppercase mb-3">Intelligence Pipeline</p>
          <div className="flex items-center gap-2 text-[10px] text-gray-400 flex-wrap justify-center">
            {['Observe', 'Learn', 'Predict', 'Explain', 'Simulate', 'Validate', 'Adaptive'].map((step, i) => (
              <span key={step} className="flex items-center gap-2">
                {i > 0 && <span className="text-gray-600">→</span>}
                <span className={`px-2 py-1 rounded ${i === 6 ? 'bg-emerald-900/30 border border-emerald-700 text-emerald-400' : 'bg-gray-700'}`}>
                  {step}
                </span>
              </span>
            ))}
          </div>
        </div>

        {/* Validation status */}
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-[10px] text-gray-500 uppercase mb-3">Validation Status</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <StatusCard domain="CVRP" status="safe" detail="60–73% compute saved" />
            <StatusCard domain="JSS" status="safe" detail="40% compute saved (Tabu)" />
            <StatusCard domain="VRPTW" status="better" detail="19% better quality" />
            <StatusCard domain="NRP" status="safe" detail="Within variance" />
          </div>
          <p className="text-[10px] text-gray-500 mt-3 text-center">
            320 validation runs · 10 seeds · Welch t-test at 95% confidence · Validated on tested configurations
          </p>
        </div>
      </Card>
    </div>
  );
}

function StyleCard({ name, target, actions }: { name: string; target: string; actions: string }) {
  return (
    <div className="bg-gray-750 border border-gray-700 rounded p-3">
      <p className="text-xs text-blue-400 font-semibold">{name}</p>
      <p className="text-[10px] text-gray-500 mt-1">{target}</p>
      <p className="text-[10px] text-gray-400 mt-1">{actions}</p>
    </div>
  );
}

function ModeCard({ mode, desc, colour }: { mode: string; desc: string; colour: string }) {
  const borders: Record<string, string> = { gray: 'border-gray-700', blue: 'border-blue-800', emerald: 'border-emerald-800', amber: 'border-amber-800' };
  const titles: Record<string, string> = { gray: 'text-gray-400', blue: 'text-blue-400', emerald: 'text-emerald-400', amber: 'text-amber-400' };
  return (
    <div className={`border ${borders[colour]} rounded p-2`}>
      <p className={`text-xs font-semibold ${titles[colour]}`}>{mode}</p>
      <p className="text-[10px] text-gray-500 mt-0.5">{desc}</p>
    </div>
  );
}

function StatusCard({ domain, status, detail }: { domain: string; status: 'safe' | 'better'; detail: string }) {
  const bg = status === 'better' ? 'border-emerald-700' : 'border-gray-700';
  const badge = status === 'better'
    ? <span className="text-[8px] bg-emerald-900/50 text-emerald-300 px-1 py-0.5 rounded">IMPROVED</span>
    : <span className="text-[8px] bg-gray-700 text-gray-300 px-1 py-0.5 rounded">SAFE</span>;
  return (
    <div className={`border ${bg} rounded p-2`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-gray-200">{domain}</span>
        {badge}
      </div>
      <p className="text-[10px] text-gray-500 mt-1">{detail}</p>
    </div>
  );
}
