import Link from 'next/link';
import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import RunList from './RunList';
import ArchitectureDiagram from './ArchitectureDiagram';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const runs = await listRunsAsync();

  return (
    <div className="space-y-8">
      {/* 1. Hero */}
      <section className="bg-gradient-to-b from-slate-900 to-slate-800 border border-slate-200 rounded-xl p-10 text-center shadow-lg">
        <h1 className="text-4xl font-bold text-white mb-3 tracking-tight">PFRS Lab</h1>
        <p className="text-base text-blue-300 font-medium mb-5">A research platform for adaptive optimisation.</p>
        <p className="text-sm text-slate-300 max-w-2xl mx-auto leading-relaxed mb-6">
          Solves NP-hard combinatorial problems across multiple domains. Benchmarks algorithms with
          statistical rigour. Collects telemetry from every search. Learns from previous runs. Uses
          Search Intelligence to improve how search is performed — automatically, safely, measurably.
        </p>
        <div className="flex flex-wrap justify-center gap-2 mb-6">
          <Badge text="4 Domains" />
          <Badge text="5 Algorithms" />
          <Badge text="320+ Validation Runs" />
          <Badge text="Search Intelligence" />
          <Badge text="S3 Dashboard" />
          <Badge text="Statistical Validation" />
        </div>
        <div className="flex justify-center gap-3">
          <Link href="/benchmarks" className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded transition-colors">
            View Benchmarks
          </Link>
          <Link href="/intelligence" className="px-4 py-2 bg-white/10 hover:bg-white/20 text-white text-xs font-semibold rounded border border-white/30 transition-colors">
            Explore Search Intelligence
          </Link>
        </div>
      </section>

      {/* 2. Architecture + Why Different (side by side) */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div>
          <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Architecture</h2>
          <div className="border border-slate-200 rounded-xl p-5 h-full shadow-sm">
            <ArchitectureDiagram />
          </div>
        </div>
        <div>
          <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Why It Is Different</h2>
          <div className="space-y-4 h-full">
            <div className="rounded-xl p-5 border border-slate-200 shadow-sm">
              <p className="text-[10px] text-slate-500 uppercase mb-3 font-semibold">Traditional Optimiser</p>
              <div className="flex items-center gap-3 text-sm text-slate-600">
                <Pill>Problem</Pill><Arrow /><Pill>Algorithm</Pill><Arrow /><Pill>Answer</Pill>
              </div>
              <p className="text-[10px] text-slate-400 mt-3">Run once. Get result. No learning.</p>
            </div>
            <div className="rounded-xl p-5 border-2 border-blue-200 bg-blue-50/50 shadow-sm">
              <p className="text-[10px] text-blue-600 uppercase mb-3 font-semibold">PFRS Lab</p>
              <div className="flex items-center gap-2 text-sm text-slate-700 flex-wrap">
                <PillB>Problem</PillB><Arrow /><PillB>Algorithm</PillB><Arrow /><PillB>Telemetry</PillB><Arrow /><PillB>SI</PillB><Arrow /><PillB>Learning</PillB><Arrow /><PillG>Better Search</PillG>
              </div>
              <p className="text-[10px] text-slate-500 mt-3">Every run improves the next. Compute goes where it matters.</p>
            </div>
          </div>
        </div>
      </section>

      {/* 4. Algorithms */}
      <section>
        <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Algorithms</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <AlgoCard name="Simulated Annealing" abbr="SA"
            formula="P(accept) = exp(−Δ / T)"
            explanation="Worse moves accepted probabilistically. Temperature cools over time — explores broadly early, converges late."
            strength="General-purpose. Works on all domains without tuning." />
          <AlgoCard name="Late Acceptance" abbr="LAHC"
            formula="accept if f(x') ≤ f(x) or f(x') ≤ f(x_{t−L})"
            explanation="Compares against historical fitness from L steps ago. Allows controlled escape from local minima without a temperature parameter."
            strength="Excels on structured problems. Hit optimal on CVRP A-n32-k5." />
          <AlgoCard name="Tabu Search" abbr="Tabu"
            formula="select best x' ∉ TabuList (or aspiration)"
            explanation="Evaluates full neighbourhood. Forbids recently visited moves. Aspiration overrides tabu if new global best found."
            strength="Strong on large instances. Best on JSS (optimal on la01)." />
          <AlgoCard name="Portfolio" abbr="Port"
            formula="result = min(SA, LAHC, Tabu)"
            explanation="Runs all strategies in parallel. Returns the best result. Never worse than any individual algorithm."
            strength="Safe default. Exploits multi-core hardware." />
          <AlgoCard name="Search Intelligence" abbr="SI"
            formula="budget(s) ← budget(s) × π(signal)"
            explanation="Learned policies allocate compute based on observed search progress. Extends productive searches, stops stagnating ones."
            strength="40–73% compute saved. 19% quality improvement on VRPTW." />
        </div>
      </section>

      {/* 5. Search Intelligence */}
      <section>
        <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Search Intelligence</h2>
        <div className="border border-slate-200 rounded-xl p-6 shadow-sm">
          <div className="flex items-center gap-2 text-[10px] text-slate-500 flex-wrap justify-center mb-5">
            {['Observe', 'Learn', 'Predict', 'Explain', 'Simulate', 'Validate', 'Guide'].map((s, i) => (
              <span key={s} className="flex items-center gap-2">
                {i > 0 && <span className="text-slate-300">→</span>}
                <span className={`px-2.5 py-1 rounded ${i === 6 ? 'bg-emerald-100 border border-emerald-300 text-emerald-700' : 'bg-slate-100 border border-slate-200 text-slate-600'}`}>{s}</span>
              </span>
            ))}
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
            <ModeCard mode="off" desc="Zero overhead. Existing behaviour." c="gray" />
            <ModeCard mode="shadow" desc="Records predictions. No behaviour change." c="blue" />
            <ModeCard mode="assist" desc="Safe recommendations. Static checkpoints." c="emerald" />
            <ModeCard mode="adaptive" desc="Live decisions. Learned models." c="amber" />
          </div>
          <div className="flex justify-center gap-4 text-[9px] text-slate-400">
            <span>WorkerAssist (NRP)</span>
            <span>SearchAssist (SA/LAHC/Tabu)</span>
            <span>PortfolioAssist (all domains)</span>
          </div>
        </div>
      </section>

      {/* 6. Research Validation */}
      <section>
        <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Research Validation</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-3">
          <EvidenceCard domain="CVRP" result="Identical quality" detail="60–73% compute saved" verdict="safe" />
          <EvidenceCard domain="JSS" result="Optimal (la01)" detail="40% compute saved" verdict="safe" />
          <EvidenceCard domain="VRPTW" result="19% better" detail="p < 0.001 (adaptive)" verdict="better" />
          <EvidenceCard domain="NRP" result="Within variance" detail="No degradation" verdict="safe" />
        </div>
        <p className="text-[9px] text-slate-400 text-center">
          320 runs · 10 seeds · Welch t-test · Mann-Whitney U · Cohen&apos;s d · Validated on tested configurations
        </p>
      </section>

      {/* 7. Origin Story */}
      <section className="border border-slate-200 rounded-xl p-6 shadow-sm">
        <h2 className="text-sm font-semibold text-slate-800 mb-3">Origin</h2>
        <p className="text-sm text-slate-600 leading-relaxed mb-3">
          This project began as a university dissertation over a decade ago — a study of metaheuristic
          algorithms applied to nurse rostering. The original work explored simulated annealing and
          constraint handling on a single problem domain with limited tooling.
        </p>
        <p className="text-sm text-slate-600 leading-relaxed mb-3">
          PFRS Lab revisits that research with twenty years of professional software engineering,
          modern cloud infrastructure, and the insight that optimisation research is fragmented —
          every domain reinvents the wheel. The question became: what if one platform could solve
          multiple NP-hard domains, benchmark them rigorously, and learn from its own search history?
        </p>
        <p className="text-sm text-slate-600 leading-relaxed">
          The result is a platform where algorithms are tested once and work everywhere, every
          experiment is reproducible, and Search Intelligence uses the platform&apos;s own telemetry
          to improve future runs — something the original dissertation could only imagine.
        </p>
      </section>

      {/* 8. Domains */}
      <section>
        <h2 className="text-xs text-slate-500 uppercase tracking-wider mb-3 px-1">Domains</h2>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          <DomainPill name="NRP" full="Nurse Rostering" bench="INRC-II" c="blue" />
          <DomainPill name="CVRP" full="Vehicle Routing" bench="CVRPLIB" c="emerald" />
          <DomainPill name="JSS" full="Job Shop" bench="OR-Library" c="amber" />
          <DomainPill name="VRPTW" full="Routing + TW" bench="Solomon" c="rose" />
        </div>
      </section>

      {/* 9. Principles */}
      <section className="text-center py-4">
        <div className="flex flex-wrap justify-center gap-4">
          {['Everything measurable.', 'Everything reproducible.', 'Everything benchmarked.', 'Everything explainable.'].map(p => (
            <span key={p} className="text-xs text-slate-500 border border-slate-200 rounded-full px-4 py-1.5">{p}</span>
          ))}
        </div>
      </section>

      {/* 9. Recent Runs */}
      {runs.length > 0 && <RunList runs={runs} />}
    </div>
  );
}


// --- Components ---

function Badge({ text }: { text: string }) {
  return <span className="text-[10px] bg-white/10 border border-white/20 text-slate-200 px-2.5 py-1 rounded-full">{text}</span>;
}

function Pill({ children }: { children: React.ReactNode }) {
  return <span className="bg-slate-100 border border-slate-200 px-2.5 py-1 rounded text-xs text-slate-600">{children}</span>;
}
function PillB({ children }: { children: React.ReactNode }) {
  return <span className="bg-blue-50 border border-blue-200 px-2 py-0.5 rounded text-xs text-blue-700">{children}</span>;
}
function PillG({ children }: { children: React.ReactNode }) {
  return <span className="bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded text-xs text-emerald-700 font-medium">{children}</span>;
}
function Arrow() {
  return <span className="text-slate-300 text-xs">→</span>;
}

function AlgoCard({ name, abbr, formula, explanation, strength }: {
  name: string; abbr: string; formula: string; explanation: string; strength: string;
}) {
  return (
    <div className="border border-slate-200 rounded-xl p-4 shadow-sm">
      <div className="flex items-baseline gap-2 mb-2">
        <span className="text-sm font-bold text-amber-600">{abbr}</span>
        <span className="text-[10px] text-slate-500">{name}</span>
      </div>
      <code className="block text-[10px] text-blue-700 bg-blue-50 border border-blue-100 rounded px-2 py-1 mb-2 font-mono">{formula}</code>
      <p className="text-[10px] text-slate-600 mb-2 leading-relaxed">{explanation}</p>
      <p className="text-[9px] text-emerald-600 font-medium">{strength}</p>
    </div>
  );
}

function ModeCard({ mode, desc, c }: { mode: string; desc: string; c: string }) {
  const borders: Record<string, string> = { gray: 'border-slate-200', blue: 'border-blue-200', emerald: 'border-emerald-200', amber: 'border-amber-200' };
  const titles: Record<string, string> = { gray: 'text-slate-600', blue: 'text-blue-600', emerald: 'text-emerald-600', amber: 'text-amber-600' };
  const bgs: Record<string, string> = { gray: 'bg-slate-50', blue: 'bg-blue-50', emerald: 'bg-emerald-50', amber: 'bg-amber-50' };
  return (
    <div className={`border ${borders[c]} ${bgs[c]} rounded-lg p-3`}>
      <p className={`text-xs font-semibold ${titles[c]} mb-1`}>{mode}</p>
      <p className="text-[10px] text-slate-500">{desc}</p>
    </div>
  );
}

function EvidenceCard({ domain, result, detail, verdict }: { domain: string; result: string; detail: string; verdict: 'safe' | 'better' }) {
  const ring = verdict === 'better' ? 'border-emerald-300 bg-emerald-50' : 'border-slate-200 bg-white';
  const badge = verdict === 'better'
    ? <span className="text-[8px] bg-emerald-100 text-emerald-700 px-1.5 py-0.5 rounded font-semibold">IMPROVED</span>
    : <span className="text-[8px] bg-slate-100 text-slate-600 px-1.5 py-0.5 rounded font-semibold">SAFE</span>;
  return (
    <div className={`border ${ring} rounded-xl p-4 shadow-sm`}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-semibold text-slate-800">{domain}</span>
        {badge}
      </div>
      <p className="text-xs text-slate-700">{result}</p>
      <p className="text-[10px] text-slate-500">{detail}</p>
    </div>
  );
}

function DomainPill({ name, full, bench, c }: { name: string; full: string; bench: string; c: string }) {
  const colours: Record<string, string> = { blue: 'text-blue-600 border-blue-200 bg-blue-50', emerald: 'text-emerald-600 border-emerald-200 bg-emerald-50', amber: 'text-amber-600 border-amber-200 bg-amber-50', rose: 'text-rose-600 border-rose-200 bg-rose-50' };
  const cls = colours[c] || colours.blue;
  return (
    <div className={`border rounded-lg p-3 ${cls}`}>
      <span className="text-sm font-bold">{name}</span>
      <span className="text-[10px] text-slate-500 ml-2">{full}</span>
      <p className="text-[9px] text-slate-400 mt-1">{bench}</p>
    </div>
  );
}
