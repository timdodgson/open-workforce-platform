import Link from 'next/link';
import RunList from '@/app/RunList';
import ByodExtensionSection from '@/features/landing/ByodExtensionSection';
import type { RunListEntry } from '@/lib/data-loader';
import { DOMAIN_GAPS, FLAGSHIP_CHALLENGE } from '@/lib/domain-challenge';

interface LandingPageProps {
  runs: RunListEntry[];
}

export default function LandingPage({ runs }: LandingPageProps) {
  return (
    <div className="landing landing--research">
      <div className="landing-research-banner">
        <p>
          Technical reference — algorithms, beam search maths, validation tables.
          {' '}<Link href="/">← Back to public site</Link>
          {' · '}<Link href="/lab">Open lab →</Link>
        </p>
      </div>

      <header className="landing-hero landing-hero--research">
        <p className="landing-eyebrow">Research reference</p>
        <h1 className="landing-title">Technical depth</h1>
        <p className="landing-subtitle">Algorithms, Search Intelligence, beam search, and validation evidence.</p>
      </header>

      <section className="landing-section landing-section-wide">
        <h2 className="landing-section-title">Platform Architecture</h2>
        <p className="landing-section-desc">
          Every domain flows through the same generic interface. Search Intelligence
          learns from telemetry and guides where compute goes.
        </p>
        <div className="arch-stack">
          <div className="arch-stack-row">
            <span className="arch-stack-label">Domains</span>
            <div className="arch-stack-chips">
              <span>NRP</span><span>CVRP</span><span>JSS</span><span>VRPTW</span>
            </div>
            <p className="arch-stack-note">INRC-II · CVRPLIB · OR-Library · Solomon</p>
          </div>
          <div className="arch-stack-connector" aria-hidden>↓</div>
          <div className="arch-stack-row">
            <span className="arch-stack-label">Generic interface</span>
            <p className="arch-stack-detail">TryMove · Evaluate · Undo · Constraints · Serialize</p>
          </div>
          <div className="arch-stack-connector" aria-hidden>↓</div>
          <div className="arch-stack-row">
            <span className="arch-stack-label">Search algorithms</span>
            <div className="arch-stack-chips">
              <span>SA</span><span>LAHC</span><span>Tabu</span><span>GA</span><span>Portfolio</span>
            </div>
          </div>
          <div className="arch-stack-connector arch-stack-connector--accent" aria-hidden>↓</div>
          <div className="arch-stack-row arch-stack-row--highlight">
            <span className="arch-stack-label">Search Intelligence</span>
            <p className="arch-stack-detail">Observe → Learn → Predict → Explain → Simulate → Validate → Guide</p>
            <p className="arch-stack-note">WorkerAssist · SearchAssist · PortfolioAssist</p>
          </div>
          <div className="arch-stack-connector" aria-hidden>↓</div>
          <div className="arch-stack-row">
            <span className="arch-stack-label">Telemetry &amp; storage</span>
            <p className="arch-stack-detail">Discoveries · worker lifecycle · learned models · S3</p>
          </div>
          <div className="arch-stack-connector" aria-hidden>↓</div>
          <div className="arch-stack-row">
            <span className="arch-stack-label">Lab dashboard</span>
            <div className="arch-stack-chips">
              <span>Benchmarks</span><span>Statistics</span><span>Experiment Matrix</span><span>Explain</span>
            </div>
          </div>
        </div>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Search Intelligence</h2>
        <p className="landing-section-desc">
          Two complementary layers — not a version upgrade. <strong>Assist</strong> handles
          rule-based compute safety; <strong>Policies</strong> adds learned stagnation,
          restart, and budget timing from distilled trees.
        </p>
        <div className="si-layers">
          <div className="si-layer">
            <span className="si-layer-name">Assist layer</span>
            <code className="si-layer-flag">--worker-decision-mode</code>
            <p className="si-layer-desc">
              WorkerAssist, SearchAssist, PortfolioAssist. Shadow, assist, or adaptive
              checkpoints. Validated on 320 runs — 40–73% compute saved on CVRP/JSS.
            </p>
          </div>
          <div className="si-layer">
            <span className="si-layer-name">Policy layer</span>
            <code className="si-layer-flag">--policy-mode</code>
            <p className="si-layer-desc">
              Learned JSON policies (rules / hybrid / learned). Stagnation, restart, budget,
              and worker policies executed in Go. 12 active lifecycle policies on val-* harness.
            </p>
          </div>
        </div>
        <div className="si-pipeline">
          {['Observe', 'Learn', 'Predict', 'Explain', 'Simulate', 'Validate', 'Guide'].map((step, i) => (
            <span key={step} className="si-pipeline-item">
              {i > 0 && <span className="si-pipeline-arrow">→</span>}
              <span className={`si-pipeline-step ${i === 6 ? 'si-pipeline-step--final' : ''}`}>{step}</span>
            </span>
          ))}
        </div>
        <p className="si-modes-label">Assist modes</p>
        <div className="si-modes">
          <div className="si-mode si-mode--off"><span className="si-mode-name">off</span><span className="si-mode-desc">Zero overhead. Existing behaviour.</span></div>
          <div className="si-mode si-mode--shadow"><span className="si-mode-name">shadow</span><span className="si-mode-desc">Records predictions. No behaviour change.</span></div>
          <div className="si-mode si-mode--assist"><span className="si-mode-name">assist</span><span className="si-mode-desc">Safe recommendations. Static checkpoints.</span></div>
          <div className="si-mode si-mode--adaptive"><span className="si-mode-name">adaptive</span><span className="si-mode-desc">Live decisions. Learned models.</span></div>
        </div>
        <p className="si-modes-label">Policy modes</p>
        <div className="si-modes">
          <div className="si-mode si-mode--off"><span className="si-mode-name">rules</span><span className="si-mode-desc">Rule checkpoints only.</span></div>
          <div className="si-mode si-mode--assist"><span className="si-mode-name">hybrid</span><span className="si-mode-desc">Learned when confident; rules fallback.</span></div>
          <div className="si-mode si-mode--shadow"><span className="si-mode-name">learned</span><span className="si-mode-desc">Learned stagnation + restart policies.</span></div>
        </div>
        <div className="si-assistants">
          <span>WorkerAssist <span className="si-assistants-domain">(NRP beam search)</span></span>
          <span>SearchAssist <span className="si-assistants-domain">(SA / LAHC / Tabu / GA)</span></span>
          <span>PortfolioAssist <span className="si-assistants-domain">(all domains)</span></span>
        </div>
      </section>

      <section className="landing-section landing-section-wide">
        <h2 className="landing-section-title">Flagship challenge</h2>
        <p className="landing-section-desc">
          Among published domains, <strong>{FLAGSHIP_CHALLENGE.label}</strong> is the one we have not
          yet cracked. CVRP, JSS, and VRPTW sit within ~0–4% of published optima; NRP remains
          +{FLAGSHIP_CHALLENGE.gapPct}% above the HiGHS ILP bound on {FLAGSHIP_CHALLENGE.instance}.
        </p>
        <div className="challenge-grid">
          <div className="challenge-card challenge-card--flagship">
            <span className="challenge-badge">Hardest to crack</span>
            <h3>{FLAGSHIP_CHALLENGE.instance}</h3>
            <p className="challenge-stat">
              Platform best <strong>{FLAGSHIP_CHALLENGE.platformBest.toLocaleString()}</strong>
              {' '}({FLAGSHIP_CHALLENGE.platformMode}) vs {FLAGSHIP_CHALLENGE.referenceLabel}{' '}
              <strong>{FLAGSHIP_CHALLENGE.referenceValue.toLocaleString()}</strong>
            </p>
            <ul className="challenge-why">
              {FLAGSHIP_CHALLENGE.whyHardest.map((line) => (
                <li key={line}>{line}</li>
              ))}
            </ul>
            <p className="challenge-cmd-label">Try the flagship path (cwd: {FLAGSHIP_CHALLENGE.cwd})</p>
            <code className="challenge-cmd">{FLAGSHIP_CHALLENGE.tryCommand}</code>
            <p className="challenge-cmd-label">Same beam with GA in the portfolio</p>
            <code className="challenge-cmd">{FLAGSHIP_CHALLENGE.gaCommand}</code>
          </div>
          <div className="challenge-gaps">
            <h3>Gap to reference by domain</h3>
            <table className="challenge-table">
              <thead>
                <tr>
                  <th>Domain</th>
                  <th>At optimal</th>
                  <th>Best gap</th>
                  <th>Worst gap</th>
                </tr>
              </thead>
              <tbody>
                {DOMAIN_GAPS.map((d) => (
                  <tr key={d.domain} className={d.domain === FLAGSHIP_CHALLENGE.domain ? 'challenge-row--flagship' : ''}>
                    <td>{d.label}</td>
                    <td>{d.atOptimal}/{d.instances}</td>
                    <td>{d.bestGapPct != null ? `+${d.bestGapPct}%` : '—'}</td>
                    <td>{d.worstGapPct != null ? `+${d.worstGapPct}%` : '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Algorithms</h2>
        <p className="landing-section-desc">
          Five search strategies plus Search Intelligence — one Problem interface.
          Portfolio Mode runs them in parallel. Search Intelligence learns which to fund.
        </p>
        <div className="algo-list">
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--sa">SA</span>
              <span className="algo-name">Simulated Annealing</span>
            </div>
            <code className="algo-formula">P(accept worse) = exp(−Δ / T)</code>
            <p className="algo-desc">Worse moves accepted probabilistically. Temperature cools geometrically — explores broadly early, converges late.</p>
          </article>
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--lahc">LAHC</span>
              <span className="algo-name">Late Acceptance Hill Climbing</span>
            </div>
            <code className="algo-formula">accept if f(new) ≤ f(current) or f(new) ≤ f(t−L)</code>
            <p className="algo-desc">Compares against historical fitness from L steps ago. Controlled escape from local minima without temperature.</p>
          </article>
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--tabu">Tabu</span>
              <span className="algo-name">Tabu Search</span>
            </div>
            <code className="algo-formula">move ∉ TabuList or aspiration improves best</code>
            <p className="algo-desc">Evaluates full neighbourhood. Forbids recently visited moves. Aspiration overrides tabu if new global best found.</p>
          </article>
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--ga">GA</span>
              <span className="algo-name">Genetic Algorithm</span>
            </div>
            <code className="algo-formula">pop ← elite ∪ crossover(parents) ∪ mutate</code>
            <p className="algo-desc">Population-based search with tournament selection and dual-parent crossover. New lever on the NRP flagship challenge.</p>
          </article>
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--port">Portfolio</span>
              <span className="algo-name">Portfolio Mode</span>
            </div>
            <code className="algo-formula">best = min(SA, LAHC, Tabu, GA)</code>
            <p className="algo-desc">Runs all strategies in parallel (default sa,lahc,tabu,ga). Returns the best result. Never worse than any individual algorithm.</p>
          </article>
          <article className="algo-item algo-item--highlight">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--si">SI</span>
              <span className="algo-name">Adaptive Hyper-Heuristic</span>
            </div>
            <code className="algo-formula">budget(strategy) ← budget(strategy) × policy(signal)</code>
            <p className="algo-desc">Learned policies allocate compute based on observed search progress. 40–73% compute saved, 19% quality improvement on VRPTW.</p>
          </article>
        </div>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Beam Search</h2>
        <p className="landing-section-desc">
          NRP uses parallel beam search across an 8-week planning horizon.
          Multiple candidate paths are maintained, expanded, and pruned each week —
          preventing commitment to bad early decisions.
        </p>

        <div className="beam-flow">
          <div className="beam-flow-steps">
            <span className="beam-flow-step">Expand paths × seeds</span>
            <span className="beam-flow-arrow">→</span>
            <span className="beam-flow-step">Run workers (SA/LAHC/Tabu/GA)</span>
            <span className="beam-flow-arrow">→</span>
            <span className="beam-flow-step">Rank by Φ(x)</span>
            <span className="beam-flow-arrow">→</span>
            <span className="beam-flow-step">Prune to top N</span>
          </div>
        </div>

        <div className="beam-maths">
          <div className="beam-math-block">
            <span className="beam-math-label">Look-Ahead Evaluation (Amortized Dynamic Boundary)</span>
            <code className="algo-formula">Φ(x) = f(x) + (ω · w/W) · Σ[ max(0, Â_n − A_max) · β₁ + max(0, Ŵ_n − W_max) · β₂ ]</code>
            <p className="beam-math-desc">
              Projects constraint trajectories forward. Time-scaled weight (weak early, strong late).
              Prevents invisible debt that explodes in week 8.
            </p>
          </div>

          <div className="beam-math-block">
            <span className="beam-math-label">Lineage Entropy</span>
            <code className="algo-formula">H(w) = −Σ p_f · log₂(p_f)</code>
            <p className="beam-math-desc">
              Shannon entropy over ancestor families per week.
              H = 1.0 means balanced diversity. H = 0.0 means total beam collapse.
            </p>
          </div>

          <div className="beam-math-block">
            <span className="beam-math-label">Beam Health Score</span>
            <code className="algo-formula">Score = H_norm·30 + min(1,r/10)·30 − (p_max/100)·20 − |collapse_weeks|·5</code>
            <p className="beam-math-desc">
              Composite 0–100 indicator combining diversity, innovation rate,
              monopoly penalty, and collapse count.
            </p>
          </div>
        </div>

        <div className="beam-features">
          <span className="beam-feature">Diversity Slots (prevent family monopoly)</span>
          <span className="beam-feature">Final Window Coupling (week 7+8 combined pruning)</span>
          <span className="beam-feature">Portfolio Branching (spawn per-strategy workers on global best)</span>
          <span className="beam-feature">Lineage Tracking (full ancestry for post-hoc analysis)</span>
        </div>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Explainability</h2>
        <p className="landing-section-desc">
          Every optimisation decision can be explained. The platform includes both
          automated narrative generation and an AI-powered research assistant.
        </p>

        <div className="explain-grid">
          <div className="explain-card">
            <span className="explain-card-title">Explain This Run</span>
            <p className="explain-card-desc">
              Rule-based narrative engine. Reads measured telemetry — discoveries,
              workers, beam ancestry, plateaus — and generates evidence-backed
              natural language explanations. Every statement links to supporting data.
            </p>
            <span className="explain-card-tag">No LLM required</span>
          </div>
          <div className="explain-card">
            <span className="explain-card-title">AI Research Assistant</span>
            <p className="explain-card-desc">
              LLM-powered experiment planner (Claude via AWS Bedrock). Designs experiments,
              generates CLI commands, interprets results, suggests next steps based on
              the platform&apos;s algorithms and parameters.
            </p>
            <span className="explain-card-tag">AWS Bedrock</span>
          </div>
          <div className="explain-card">
            <span className="explain-card-title">What-If Lab</span>
            <p className="explain-card-desc">
              Simulate alternative decisions and predict outcomes. Counterfactual
              analysis powered by historical telemetry and the learned model.
            </p>
            <span className="explain-card-tag">Simulation</span>
          </div>
          <div className="explain-card">
            <span className="explain-card-title">Feature Importance</span>
            <p className="explain-card-desc">
              Decision tree model visualisation. Shows which features drive worker
              value predictions and portfolio budget allocation decisions.
            </p>
            <span className="explain-card-tag">ML Transparency</span>
          </div>
        </div>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Research Validation</h2>
        <p className="landing-section-desc">
          320+ runs. 10 seeds per configuration. Welch t-test, Mann-Whitney U,
          Cohen&apos;s d. Validated on tested configurations.
        </p>
        <table className="evidence-table">
          <thead>
            <tr><th>Domain</th><th>Result</th><th>Detail</th><th>Verdict</th></tr>
          </thead>
          <tbody>
            <tr>
              <td className="evidence-domain">CVRP</td>
              <td>Identical quality</td>
              <td>60–73% compute saved</td>
              <td><span className="evidence-badge evidence-badge--safe">SAFE</span></td>
            </tr>
            <tr>
              <td className="evidence-domain">JSS</td>
              <td>Optimal (la01 = 666)</td>
              <td>40% compute saved</td>
              <td><span className="evidence-badge evidence-badge--safe">SAFE</span></td>
            </tr>
            <tr>
              <td className="evidence-domain">VRPTW</td>
              <td>19% better quality</td>
              <td>p &lt; 0.001, d = −2.91</td>
              <td><span className="evidence-badge evidence-badge--improved">IMPROVED</span></td>
            </tr>
            <tr>
              <td className="evidence-domain">NRP</td>
              <td>Within stochastic variance</td>
              <td>No degradation</td>
              <td><span className="evidence-badge evidence-badge--safe">SAFE</span></td>
            </tr>
          </tbody>
        </table>
        <p className="evidence-footnote">
          4 domains · SA/LAHC/Tabu/GA + Portfolio · Search Intelligence · 320+ validation runs · Zero feasibility regressions
        </p>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Optimality Gap</h2>
        <p className="landing-section-desc">
          HiGHS (ILP) gives two different numbers on n012w8 — do not conflate them. The dual/MIP
          lower bound is not a proven optimal roster; the best feasible ILP solution is the
          published reference used for the +14.7% gap.
        </p>
        <div className="ilp-comparison">
          <div className="ilp-row">
            <span className="ilp-label">ILP lower bound (dual / MIP gap, not a roster)</span>
            <span className="ilp-value">~{FLAGSHIP_CHALLENGE.ilpLowerBound.toLocaleString()}</span>
          </div>
          <div className="ilp-row">
            <span className="ilp-label">ILP best feasible (published reference)</span>
            <span className="ilp-value">{FLAGSHIP_CHALLENGE.ilpFeasible.toLocaleString()}</span>
          </div>
          <div className="ilp-row ilp-row--highlight">
            <span className="ilp-label">Best PFRS (portfolio beam)</span>
            <span className="ilp-value">{FLAGSHIP_CHALLENGE.platformBest.toLocaleString()}</span>
          </div>
          <div className="ilp-row">
            <span className="ilp-label">Gap to ILP feasible</span>
            <span className="ilp-value">~{FLAGSHIP_CHALLENGE.gapPct}%</span>
          </div>
        </div>
        <p className="evidence-footnote">
          Heuristics return feasible rosters in minutes. Closing the remaining gap to the ILP
          feasible reference is the open flagship challenge — not the dual bound alone.
        </p>
      </section>

      <section className="landing-section">
        <h2 className="landing-section-title">Domains</h2>
        <div className="domains-grid">
          <div className="domain-item domain-item--nrp">
            <span className="domain-abbr">NRP</span>
            <span className="domain-full">Nurse Rostering (NRP)</span>
            <span className="domain-bench">INRC-II benchmark</span>
          </div>
          <div className="domain-item domain-item--cvrp">
            <span className="domain-abbr">CVRP</span>
            <span className="domain-full">Vehicle Routing (CVRP)</span>
            <span className="domain-bench">CVRPLIB benchmark</span>
          </div>
          <div className="domain-item domain-item--jss">
            <span className="domain-abbr">JSS</span>
            <span className="domain-full">Job Shop Scheduling (JSS)</span>
            <span className="domain-bench">OR-Library benchmark</span>
          </div>
          <div className="domain-item domain-item--vrptw">
            <span className="domain-abbr">VRPTW</span>
            <span className="domain-full">Vehicle Routing + Time Windows (VRPTW)</span>
            <span className="domain-bench">Solomon benchmark</span>
          </div>
        </div>
      </section>

      <ByodExtensionSection />

      <section className="landing-section">
        <h2 className="landing-section-title">Related</h2>
        <p className="landing-section-desc">
          Field-guide pages and the cite path — same content, less density than this reference.
        </p>
        <div className="flex flex-wrap gap-4 text-sm">
          <Link href="/algorithms" className="text-blue-400 hover:underline">Algorithms →</Link>
          <Link href="/domains" className="text-blue-400 hover:underline">Domains →</Link>
          <Link href="/getting-started" className="text-blue-400 hover:underline">Getting started →</Link>
          <Link href="/reproduce" className="text-blue-400 hover:underline">Cite &amp; reproduce →</Link>
          <Link href="/lab/byod" className="text-blue-400 hover:underline">BYOD / BYOA →</Link>
        </div>
      </section>

      <section className="landing-section landing-principles">
        <span className="landing-principle">Everything measurable.</span>
        <span className="landing-principle">Everything reproducible.</span>
        <span className="landing-principle">Everything benchmarked.</span>
        <span className="landing-principle">Everything explainable.</span>
      </section>

      {runs.length > 0 && (
        <section className="landing-section landing-runs">
          <RunList runs={runs} />
        </section>
      )}
    </div>
  );
}
