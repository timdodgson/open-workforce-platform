import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import RunList from './RunList';
import ArchitectureDiagram from './ArchitectureDiagram';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const runs = await listRunsAsync();

  return (
    <div className="space-y-6">
      {/* Hero */}
      <div className="bg-gray-850 border border-gray-700 rounded-lg p-8 text-center">
        <h1 className="text-3xl font-bold text-gray-100 mb-2">PFRS Lab</h1>
        <p className="text-sm text-blue-400 mb-4">A research platform for adaptive optimisation.</p>
        <p className="text-xs text-gray-400 max-w-2xl mx-auto leading-relaxed">
          PFRS Lab solves and studies NP-hard optimisation problems across multiple domains.
          It provides a common Problem interface, shared metaheuristic algorithms, exact benchmarks
          where practical, unified telemetry, statistical analysis, and Search Intelligence modes
          that observe, explain, assist, and adapt the search process.
        </p>
      </div>

      {/* What is this? */}
      <Card title="What Is This?">
        <p className="text-xs text-gray-400 leading-relaxed">
          PFRS Lab is a research platform for studying how metaheuristic algorithms behave on
          hard combinatorial problems. It solves nurse rostering, vehicle routing, job shop scheduling,
          and routing with time windows — all through a single generic search engine. Every run produces
          structured telemetry. Every result is statistically compared. And the platform learns from its
          own history to make better search decisions over time.
        </p>
      </Card>

      {/* Why it exists */}
      <Card title="Why It Exists">
        <p className="text-xs text-gray-400 leading-relaxed">
          Optimisation research is fragmented — each problem domain gets its own bespoke solver,
          its own evaluation scripts, and its own reporting. Comparing algorithms across domains
          means rebuilding everything from scratch. PFRS Lab eliminates that. One interface, one
          telemetry system, one dashboard, one algorithm portfolio — applied to any NP-hard domain.
          The goal is not just to solve problems, but to understand search behaviour and improve it.
        </p>
      </Card>

      {/* What makes it different */}
      <Card title="What Makes It Different">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <p className="text-[10px] text-gray-500 uppercase mb-2">Traditional Optimiser</p>
            <div className="flex items-center gap-2 text-xs text-gray-400">
              <span className="bg-gray-700 px-2 py-1 rounded">Problem</span>
              <span className="text-gray-600">→</span>
              <span className="bg-gray-700 px-2 py-1 rounded">Algorithm</span>
              <span className="text-gray-600">→</span>
              <span className="bg-gray-700 px-2 py-1 rounded">Answer</span>
            </div>
          </div>
          <div className="bg-gray-800 rounded-lg p-4 border border-blue-800">
            <p className="text-[10px] text-blue-400 uppercase mb-2">PFRS Lab</p>
            <div className="flex items-center gap-2 text-xs text-gray-300 flex-wrap">
              <span className="bg-blue-900/30 border border-blue-700 px-2 py-1 rounded">Problem</span>
              <span className="text-gray-600">→</span>
              <span className="bg-blue-900/30 border border-blue-700 px-2 py-1 rounded">Algorithm</span>
              <span className="text-gray-600">→</span>
              <span className="bg-blue-900/30 border border-blue-700 px-2 py-1 rounded">Telemetry</span>
              <span className="text-gray-600">→</span>
              <span className="bg-blue-900/30 border border-blue-700 px-2 py-1 rounded">Search Intelligence</span>
              <span className="text-gray-600">→</span>
              <span className="bg-blue-900/30 border border-blue-700 px-2 py-1 rounded">Learning</span>
              <span className="text-gray-600">→</span>
              <span className="bg-emerald-900/30 border border-emerald-700 px-2 py-1 rounded text-emerald-300">Better Search</span>
            </div>
          </div>
        </div>
      </Card>

      {/* Architecture overview */}
      <Card title="Architecture Overview">
        <ArchitectureDiagram />
      </Card>

      {/* Platform stats */}
      <Card title="Platform at a Glance">
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
          <StatCard value="4" label="Domains" />
          <StatCard value="5" label="Algorithms" />
          <StatCard value="4" label="SI Modes" />
          <StatCard value="320+" label="Validation Runs" />
          <StatCard value="ILP" label="Exact Benchmarks" />
          <StatCard value="S3" label="Cloud Storage" />
          <StatCard value="p<0.05" label="Statistical Rigour" />
          <StatCard value="40+" label="Dashboard Pages" />
        </div>
      </Card>

      {/* Supported Domains */}
      <Card title="Supported Domains">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <DomainCard
            name="NRP"
            full="Nurse Rostering"
            description="Assign shifts to nurses over multi-week horizons. Satisfy coverage, skills, and contracts."
            objective="Minimise soft constraint penalties"
            benchmark="INRC-II"
            colour="blue"
          />
          <DomainCard
            name="CVRP"
            full="Capacitated Vehicle Routing"
            description="Find minimum-distance routes for vehicles with capacity limits. Every customer visited once."
            objective="Minimise total travel distance"
            benchmark="CVRPLIB"
            colour="emerald"
          />
          <DomainCard
            name="JSS"
            full="Job Shop Scheduling"
            description="Schedule operations on machines. Each job has ordered operations. No machine overlap."
            objective="Minimise makespan"
            benchmark="OR-Library"
            colour="amber"
          />
          <DomainCard
            name="VRPTW"
            full="Vehicle Routing + Time Windows"
            description="Route vehicles to customers within time windows. Late arrivals are infeasible."
            objective="Minimise distance (time-feasible)"
            benchmark="Solomon"
            colour="rose"
          />
        </div>
      </Card>

      {/* Algorithms */}
      <Card title="Algorithms">
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
          <AlgoCard name="SA" full="Simulated Annealing" desc="Probabilistic acceptance. Temperature cools over time." />
          <AlgoCard name="LAHC" full="Late Acceptance" desc="Accepts if better than current or L iterations ago." />
          <AlgoCard name="Tabu" full="Tabu Search" desc="Best-move neighbourhood. Forbids recent moves." />
          <AlgoCard name="Portfolio" full="Portfolio Mode" desc="Runs all strategies in parallel. Keeps the best." />
          <AlgoCard name="Adaptive" full="Adaptive Hyper-Heuristic" desc="SA + LAHC escape. Learns stagnation patterns." />
        </div>
      </Card>

      {/* Search Intelligence */}
      <Card title="Search Intelligence">
        <p className="text-xs text-gray-400 mb-4 leading-relaxed">
          A Search Intelligence advisory system that monitors search progress and makes safe
          compute allocation decisions. It observes, learns, predicts, explains, simulates,
          and guides — without ever compromising solution feasibility.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
          <ModeCard mode="off" description="No intelligence. Existing behaviour. Zero overhead." colour="gray" />
          <ModeCard mode="shadow" description="Records predictions. No behaviour change. Safe data collection." colour="blue" />
          <ModeCard mode="assist" description="Applies safe recommendations. Static checkpoints. Safety overrides." colour="emerald" />
          <ModeCard mode="adaptive" description="Live-updating decisions. Learns improvement curves. Extends or cuts budgets." colour="amber" />
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <p className="text-[10px] text-gray-500 uppercase mb-2">Intelligence Pipeline</p>
          <div className="flex items-center gap-2 text-[10px] text-gray-400 flex-wrap justify-center">
            <span className="bg-gray-700 px-2 py-1 rounded">Observe</span>
            <span className="text-gray-600">→</span>
            <span className="bg-gray-700 px-2 py-1 rounded">Learn</span>
            <span className="text-gray-600">→</span>
            <span className="bg-gray-700 px-2 py-1 rounded">Predict</span>
            <span className="text-gray-600">→</span>
            <span className="bg-gray-700 px-2 py-1 rounded">Explain</span>
            <span className="text-gray-600">→</span>
            <span className="bg-gray-700 px-2 py-1 rounded">Simulate</span>
            <span className="text-gray-600">→</span>
            <span className="bg-emerald-900/30 border border-emerald-700 px-2 py-1 rounded text-emerald-400">Guide</span>
          </div>
        </div>
      </Card>

      {/* Evidence */}
      <Card title="Validation Evidence">
        <p className="text-xs text-gray-400 mb-3">
          320 experiment runs. 10 seeds per configuration. Welch t-test at 95% confidence.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
          <EvidenceCard domain="CVRP" result="Identical quality" detail="60–73% compute saved" verdict="safe" />
          <EvidenceCard domain="JSS" result="Identical quality" detail="40% compute saved (Tabu)" verdict="safe" />
          <EvidenceCard domain="VRPTW" result="19% better quality" detail="p < 0.001 (adaptive)" verdict="better" />
          <EvidenceCard domain="NRP" result="Within variance" detail="No degradation" verdict="safe" />
        </div>
      </Card>

      {/* Principles */}
      <Card title="Principles">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <PrincipleCard text="Everything measurable." />
          <PrincipleCard text="Everything reproducible." />
          <PrincipleCard text="Everything benchmarked." />
          <PrincipleCard text="Everything explainable." />
        </div>
      </Card>

      {/* Run list */}
      {runs.length > 0 ? (
        <RunList runs={runs} />
      ) : (
        <Card title="Getting Started">
          <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
            <p className="mb-3">No runs yet. Start an experiment:</p>
            <div className="text-left bg-gray-800 rounded p-3 text-xs font-mono text-gray-400 space-y-1">
              <p className="text-gray-500"># CVRP with Search Intelligence</p>
              <p>owp solve-cvrp --instance A-n32-k5.vrp --mode portfolio --worker-decision-mode adaptive --run-label my-first-run</p>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}


// --- Components ---

function StatCard({ value, label }: { value: string; label: string }) {
  return (
    <div className="bg-gray-800 rounded-lg p-3 text-center">
      <div className="text-lg font-bold text-gray-100">{value}</div>
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
    </div>
  );
}

function DomainCard({ name, full, description, objective, benchmark, colour }: {
  name: string; full: string; description: string; objective: string; benchmark: string; colour: string;
}) {
  const borders: Record<string, string> = {
    blue: 'border-blue-800', emerald: 'border-emerald-800', amber: 'border-amber-800', rose: 'border-rose-800',
  };
  const titles: Record<string, string> = {
    blue: 'text-blue-400', emerald: 'text-emerald-400', amber: 'text-amber-400', rose: 'text-rose-400',
  };
  return (
    <div className={`bg-gray-800 border ${borders[colour]} rounded-lg p-4`}>
      <div className="flex items-baseline gap-2 mb-2">
        <span className={`text-sm font-bold ${titles[colour]}`}>{name}</span>
        <span className="text-[10px] text-gray-500">{full}</span>
      </div>
      <p className="text-[11px] text-gray-400 mb-2">{description}</p>
      <p className="text-[10px] text-gray-500"><span className="text-gray-400">Objective:</span> {objective}</p>
      <p className="text-[10px] text-gray-500"><span className="text-gray-400">Benchmark:</span> {benchmark}</p>
    </div>
  );
}

function AlgoCard({ name, full, desc }: { name: string; full: string; desc: string }) {
  return (
    <div className="bg-gray-800 rounded-lg p-3">
      <p className="text-amber-400 font-bold text-sm">{name}</p>
      <p className="text-[9px] text-gray-500 mb-1">{full}</p>
      <p className="text-[10px] text-gray-400">{desc}</p>
    </div>
  );
}

function ModeCard({ mode, description, colour }: { mode: string; description: string; colour: string }) {
  const borders: Record<string, string> = {
    gray: 'border-gray-700', blue: 'border-blue-800', emerald: 'border-emerald-800', amber: 'border-amber-800',
  };
  const titles: Record<string, string> = {
    gray: 'text-gray-400', blue: 'text-blue-400', emerald: 'text-emerald-400', amber: 'text-amber-400',
  };
  return (
    <div className={`bg-gray-800 border ${borders[colour]} rounded-lg p-3`}>
      <p className={`text-xs font-semibold ${titles[colour]} mb-1`}>{mode}</p>
      <p className="text-[10px] text-gray-400">{description}</p>
    </div>
  );
}

function EvidenceCard({ domain, result, detail, verdict }: { domain: string; result: string; detail: string; verdict: 'safe' | 'better' }) {
  const bg = verdict === 'better' ? 'border-emerald-700 bg-emerald-900/10' : 'border-gray-700';
  const badge = verdict === 'better'
    ? <span className="text-[9px] bg-emerald-900/50 text-emerald-300 px-1.5 py-0.5 rounded">IMPROVED</span>
    : <span className="text-[9px] bg-gray-700 text-gray-300 px-1.5 py-0.5 rounded">SAFE</span>;
  return (
    <div className={`bg-gray-800 border ${bg} rounded-lg p-3`}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-semibold text-gray-200">{domain}</span>
        {badge}
      </div>
      <p className="text-[11px] text-gray-300">{result}</p>
      <p className="text-[10px] text-gray-500">{detail}</p>
    </div>
  );
}

function PrincipleCard({ text }: { text: string }) {
  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 text-center">
      <p className="text-xs text-gray-300 font-medium">{text}</p>
    </div>
  );
}
