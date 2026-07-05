import Link from 'next/link';
import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import DeleteRunButton from './DeleteRunButton';
import RunList from './RunList';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const runs = await listRunsAsync();

  return (
    <div className="space-y-6">
      {/* Platform introduction */}
      <Card title="PFRS Research Lab">
        <p className="text-sm text-gray-300 mb-3">
          A multi-domain optimisation research platform for solving NP-hard combinatorial problems
          using metaheuristic search algorithms.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          <DomainCard
            title="Nurse Rostering (NRP)"
            description="Assign shifts to nurses over an 8-week horizon. Minimise soft constraint penalties while satisfying coverage, skills, and contractual rules."
            objective="Minimise total penalty (weighted soft constraint violations)"
            benchmark="INRC-II Competition"
            link="http://mobiz.vives.be/inrc2/"
            colour="blue"
          />
          <DomainCard
            title="Vehicle Routing (CVRP)"
            description="Find minimum-distance routes for capacity-limited vehicles serving customers from a depot. Every customer visited exactly once."
            objective="Minimise total travel distance (Euclidean)"
            benchmark="CVRPLIB"
            link="http://vrp.atd-lab.inf.puc-rio.br/index.php/en/"
            colour="emerald"
          />
          <DomainCard
            title="Job Shop Scheduling (JSS)"
            description="Schedule operations across machines. Each job has ordered operations, each requiring a specific machine. No machine overlap allowed."
            objective="Minimise makespan (total completion time)"
            benchmark="Taillard / OR-Library"
            link="http://jobshop.jjvh.nl/"
            colour="amber"
          />
          <DomainCard
            title="Vehicle Routing + Time Windows (VRPTW)"
            description="Extend CVRP with time constraints. Each customer has a service time window — vehicles must arrive within the window or wait. Late arrivals are infeasible."
            objective="Minimise total travel distance (time-feasible routes)"
            benchmark="Solomon Benchmarks"
            link="https://www.sintef.no/projectweb/top/vrptw/solomon-benchmark/"
            colour="rose"
          />
        </div>
      </Card>

      {/* How it works */}
      <Card title="How It Works">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
          <div className="bg-gray-800 rounded p-3">
            <p className="text-blue-400 font-semibold mb-1">1. Define Problem</p>
            <p className="text-gray-400">Each domain implements a generic Problem interface: initial solution, move generation, evaluation, validation.</p>
          </div>
          <div className="bg-gray-800 rounded p-3">
            <p className="text-emerald-400 font-semibold mb-1">2. Run Algorithms</p>
            <p className="text-gray-400">SA, LAHC, Tabu Search, or Portfolio mode — all work on any problem through the same interface.</p>
          </div>
          <div className="bg-gray-800 rounded p-3">
            <p className="text-amber-400 font-semibold mb-1">3. Analyse Results</p>
            <p className="text-gray-400">Full telemetry: discovery timeline, search progress, diversity, and cross-run statistical comparison.</p>
          </div>
        </div>
      </Card>

      {/* Platform Features */}
      <Card title="Platform Features">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 text-xs">
          <FeatureCard
            title="Research Analytics"
            colour="blue"
            items={[
              'Statistical comparison (Welch\'s t-test, box plots, Cohen\'s d)',
              'Benchmark Ladder with algorithm leaderboard',
              'Gap-to-optimal tracking per instance',
              'Trend analysis with regression across experiments',
            ]}
          />
          <FeatureCard
            title="Search Telemetry"
            colour="emerald"
            items={[
              'Every global best improvement timestamped',
              'Discovery timeline and convergence curves',
              'Fitness landscape (Hamming distance vs penalty)',
              'Worker lifecycle and branch genealogy',
            ]}
          />
          <FeatureCard
            title="Multi-Algorithm Portfolio"
            colour="amber"
            items={[
              'SA, LAHC, Tabu, Portfolio through one interface',
              'Same algorithm code solves any problem domain',
              'Deterministic replay with seed control',
              'Best-move Tabu with configurable neighbourhood',
            ]}
          />
          <FeatureCard
            title="Beam Search (NRP)"
            colour="purple"
            items={[
              'Multi-week planning with diversity preservation',
              'Look-ahead (amortized global constraint projection)',
              'Lineage entropy and family health monitoring',
              'Final-window coupling for horizon problems',
            ]}
          />
          <FeatureCard
            title="Exact Benchmarks (ILP)"
            colour="rose"
            items={[
              'MILP formulations solved by HiGHS',
              'Proven optimality bounds for calibration',
              'Heuristic gap% tracked per instance',
              'NRP and CVRP models supported',
            ]}
          />
          <FeatureCard
            title="Infrastructure"
            colour="gray"
            items={[
              'S3 telemetry storage with versioned history',
              'ECS Fargate deployed dashboard (AWS)',
              'CI/CD via GitHub Actions + semantic release',
              'Cognito authentication for AI assistant',
            ]}
          />
        </div>
      </Card>

      {/* Algorithms reference */}
      <Card title="Algorithms">
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3 text-xs">
          <AlgoCard name="SA" full="Simulated Annealing" desc="Accepts worse moves probabilistically. Temperature cools over time." />
          <AlgoCard name="LAHC" full="Late Acceptance" desc="Accepts if better than current or better than L iterations ago." />
          <AlgoCard name="Tabu" full="Tabu Search" desc="Best-move neighbourhood. Forbids recent moves to force exploration." />
          <AlgoCard name="Portfolio" full="Portfolio Mode" desc="Runs all strategies, keeps the best result." />
          <AlgoCard name="Adaptive" full="Adaptive Hyper-Heuristic" desc="SA primary with LAHC escape bursts on stagnation. Learns when to switch." />
        </div>
      </Card>

      {/* Objective explanation */}
      <Card title="Understanding the Objective">
        <div className="text-xs text-gray-400 space-y-2">
          <p><span className="text-blue-400 font-semibold">NRP Penalty:</span> Sum of weighted soft constraint violations. Zero = perfect schedule. Higher = worse. Hard constraints (skills, coverage) must always be satisfied — solutions with hard violations are invalid.</p>
          <p><span className="text-emerald-400 font-semibold">CVRP Distance:</span> Total Euclidean travel distance across all vehicle routes. Lower = shorter routes = better. Vehicle capacity is a hard constraint — overloaded routes are invalid.</p>
          <p><span className="text-amber-400 font-semibold">ILP Reference:</span> Exact mathematical solver (HiGHS). Provides proven optimal or bounded solutions for small instances. Used to measure how close heuristics get to the true optimum.</p>
        </div>
      </Card>

      {/* Run list */}
      {runs.length > 0 ? (
        <RunList runs={runs} />
      ) : (
        <Card title="Getting Started">
          <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
            <p className="mb-3">No runs yet. Start by running an experiment:</p>
            <div className="text-left bg-gray-800 rounded p-3 text-xs font-mono text-gray-400 space-y-1">
              <p className="text-gray-500"># NRP (Nurse Rostering)</p>
              <p>go run ./cmd/owp tune-pfrs --pfrs-run-label my-nrp-run</p>
              <p className="text-gray-500 mt-2"># CVRP (Vehicle Routing)</p>
              <p>go run ./cmd/owp solve-cvrp --instance A-n32-k5.vrp --run-label my-cvrp-run</p>
            </div>
          </div>
        </Card>
      )}

      {/* References */}
      <Card title="References">
        <div className="text-xs text-gray-400 space-y-1">
          <p>• <a href="http://mobiz.vives.be/inrc2/" target="_blank" className="text-blue-400 hover:underline">INRC-II Competition</a> — International Nurse Rostering Competition (academic benchmark)</p>
          <p>• <a href="http://vrp.atd-lab.inf.puc-rio.br/index.php/en/" target="_blank" className="text-blue-400 hover:underline">CVRPLIB</a> — Capacitated Vehicle Routing Problem Library</p>
          <p>• <a href="https://github.com/ERGO-Code/HiGHS" target="_blank" className="text-blue-400 hover:underline">HiGHS</a> — Open-source linear programming solver (ILP benchmarks)</p>
          <p>• <a href="https://en.wikipedia.org/wiki/Simulated_annealing" target="_blank" className="text-blue-400 hover:underline">Simulated Annealing</a> — Metropolis acceptance criterion</p>
          <p>• <a href="https://en.wikipedia.org/wiki/Late_acceptance_hill_climbing" target="_blank" className="text-blue-400 hover:underline">Late Acceptance Hill Climbing</a> — Burke & Bykov, 2017</p>
        </div>
      </Card>
    </div>
  );
}

function DomainCard({ title, description, objective, benchmark, link, colour }: {
  title: string; description: string; objective: string; benchmark: string; link: string; colour: string;
}) {
  const borderClass = colour === 'blue' ? 'border-blue-800' : colour === 'emerald' ? 'border-emerald-800' : colour === 'rose' ? 'border-rose-800' : 'border-amber-800';
  const titleClass = colour === 'blue' ? 'text-blue-400' : colour === 'emerald' ? 'text-emerald-400' : colour === 'rose' ? 'text-rose-400' : 'text-amber-400';
  return (
    <div className={`bg-gray-800 border ${borderClass} rounded-lg p-4`}>
      <h3 className={`text-sm font-semibold ${titleClass} mb-2`}>{title}</h3>
      <p className="text-xs text-gray-400 mb-2">{description}</p>
      <p className="text-[10px] text-gray-500"><span className="text-gray-400">Objective:</span> {objective}</p>
      <p className="text-[10px] text-gray-500 mt-1">
        <span className="text-gray-400">Benchmark:</span>{' '}
        <a href={link} target="_blank" className="text-blue-400 hover:underline">{benchmark}</a>
      </p>
    </div>
  );
}

function AlgoCard({ name, full, desc }: { name: string; full: string; desc: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <p className="text-amber-400 font-bold text-sm">{name}</p>
      <p className="text-[9px] text-gray-500 mb-1">{full}</p>
      <p className="text-[10px] text-gray-400">{desc}</p>
    </div>
  );
}

function FeatureCard({ title, colour, items }: { title: string; colour: string; items: string[] }) {
  const colourMap: Record<string, string> = {
    blue: 'text-blue-400 border-blue-800',
    emerald: 'text-emerald-400 border-emerald-800',
    amber: 'text-amber-400 border-amber-800',
    purple: 'text-purple-400 border-purple-800',
    rose: 'text-rose-400 border-rose-800',
    gray: 'text-gray-400 border-gray-700',
  };
  const classes = colourMap[colour] || colourMap.gray;
  const [textClass, borderClass] = classes.split(' ');

  return (
    <div className={`bg-gray-800 border ${borderClass} rounded-lg p-3`}>
      <p className={`${textClass} font-semibold text-xs mb-2`}>{title}</p>
      <ul className="space-y-1">
        {items.map((item, i) => (
          <li key={i} className="text-[10px] text-gray-400 flex items-start gap-1.5">
            <span className="text-gray-600 mt-0.5">•</span>
            <span>{item}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
