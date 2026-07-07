import Link from 'next/link';
import { listRunsAsync } from '@/lib/data-loader';
import RunList from './RunList';
import ArchitectureDiagram from './ArchitectureDiagram';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const runs = await listRunsAsync();

  return (
    <div className="landing">
      {/* ═══════════════════════════════════════════════════════════════
          Section 1: Hero — open, typographic, no box
      ═══════════════════════════════════════════════════════════════ */}
      <header className="landing-hero">
        <p className="landing-eyebrow">PFRS Research Lab</p>
        <h1 className="landing-title">PFRS Lab</h1>
        <p className="landing-subtitle">A research platform for adaptive optimisation.</p>
        <p className="landing-intro">
          PFRS Lab studies how metaheuristic algorithms behave across hard optimisation
          problems, captures telemetry from every run, and uses Search Intelligence to
          improve how compute is allocated — automatically, safely, measurably.
        </p>
        <div className="landing-actions">
          <Link href="/benchmarks" className="landing-btn-primary">
            View Benchmarks
          </Link>
          <Link href="/intelligence" className="landing-btn-secondary">
            Explore Search Intelligence
          </Link>
        </div>
      </header>

      {/* ═══════════════════════════════════════════════════════════════
          Section 2: Origin Story
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section">
        <p className="landing-prose">
          This project began as a university dissertation on nurse rostering optimisation
          over a decade ago. PFRS Lab revisits that research with twenty years of professional
          software engineering experience and a question: what if one platform could solve
          multiple NP-hard domains, benchmark them with statistical rigour, and learn from
          its own search history?
        </p>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 3: Architecture Diagram
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section landing-section-wide">
        <h2 className="landing-section-title">Platform Architecture</h2>
        <p className="landing-section-desc">
          Every domain flows through the same generic interface. Search Intelligence
          sits at the centre, learning from telemetry to allocate compute where it matters.
        </p>
        <div className="landing-diagram">
          <ArchitectureDiagram />
        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 4: Search Intelligence — the centrepiece
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section">
        <h2 className="landing-section-title">Search Intelligence</h2>
        <p className="landing-section-desc">
          This is not just another optimiser. Search Intelligence observes how algorithms
          behave, builds learned models, and reallocates compute in real time — extending
          productive searches and stopping stagnating ones.
        </p>

        {/* Pipeline */}
        <div className="si-pipeline">
          {['Observe', 'Learn', 'Predict', 'Explain', 'Simulate', 'Validate', 'Guide'].map((step, i) => (
            <span key={step} className="si-pipeline-item">
              {i > 0 && <span className="si-pipeline-arrow">→</span>}
              <span className={`si-pipeline-step ${i === 6 ? 'si-pipeline-step--final' : ''}`}>
                {step}
              </span>
            </span>
          ))}
        </div>

        {/* Modes */}
        <div className="si-modes">
          <div className="si-mode si-mode--off">
            <span className="si-mode-name">off</span>
            <span className="si-mode-desc">Zero overhead. Existing behaviour.</span>
          </div>
          <div className="si-mode si-mode--shadow">
            <span className="si-mode-name">shadow</span>
            <span className="si-mode-desc">Records predictions. No behaviour change.</span>
          </div>
          <div className="si-mode si-mode--assist">
            <span className="si-mode-name">assist</span>
            <span className="si-mode-desc">Safe recommendations. Static checkpoints.</span>
          </div>
          <div className="si-mode si-mode--adaptive">
            <span className="si-mode-name">adaptive</span>
            <span className="si-mode-desc">Live decisions. Learned models.</span>
          </div>
        </div>

        {/* Assistants */}
        <div className="si-assistants">
          <span>WorkerAssist <span className="si-assistants-domain">(NRP)</span></span>
          <span>SearchAssist <span className="si-assistants-domain">(SA / LAHC / Tabu)</span></span>
          <span>PortfolioAssist <span className="si-assistants-domain">(all domains)</span></span>
        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 5: Algorithms with Maths
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section">
        <h2 className="landing-section-title">Algorithms</h2>
        <p className="landing-section-desc">
          Five strategies, one interface. Each algorithm explores the search space differently.
          Portfolio Mode runs them in parallel. Search Intelligence learns which to fund.
        </p>

        <div className="algo-list">
          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--sa">SA</span>
              <span className="algo-name">Simulated Annealing</span>
            </div>
            <code className="algo-formula">P(accept worse) = exp(−Δ / T)</code>
            <p className="algo-desc">
              Worse moves accepted probabilistically. Temperature cools over time —
              explores broadly early, converges late. General-purpose across all domains.
            </p>
          </article>

          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--lahc">LAHC</span>
              <span className="algo-name">Late Acceptance Hill Climbing</span>
            </div>
            <code className="algo-formula">accept if f(new) ≤ f(current) or f(new) ≤ f(t−L)</code>
            <p className="algo-desc">
              Compares against historical fitness from L steps ago. Controlled escape from
              local minima without a temperature parameter. Hit optimal on CVRP A-n32-k5.
            </p>
          </article>

          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--tabu">Tabu</span>
              <span className="algo-name">Tabu Search</span>
            </div>
            <code className="algo-formula">move ∉ TabuList or aspiration improves best</code>
            <p className="algo-desc">
              Evaluates full neighbourhood. Forbids recently visited moves. Aspiration
              overrides tabu if new global best found. Best on JSS (optimal on la01).
            </p>
          </article>

          <article className="algo-item">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--port">Portfolio</span>
              <span className="algo-name">Portfolio Mode</span>
            </div>
            <code className="algo-formula">best = min(SA, LAHC, Tabu)</code>
            <p className="algo-desc">
              Runs all strategies in parallel. Returns the best result. Never worse than
              any individual algorithm. Safe default exploiting multi-core hardware.
            </p>
          </article>

          <article className="algo-item algo-item--highlight">
            <div className="algo-header">
              <span className="algo-abbr algo-abbr--si">SI</span>
              <span className="algo-name">Adaptive Hyper-Heuristic</span>
            </div>
            <code className="algo-formula">budget(strategy) ← budget(strategy) × policy(signal)</code>
            <p className="algo-desc">
              Learned policies allocate compute based on observed search progress. Extends
              productive searches, stops stagnating ones. 40–73% compute saved, 19% quality
              improvement on VRPTW.
            </p>
          </article>
        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 6: Validation Evidence
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section">
        <h2 className="landing-section-title">Research Validation</h2>
        <p className="landing-section-desc">
          320+ runs. 10 seeds per configuration. Welch t-test, Mann-Whitney U,
          Cohen&apos;s d. Validated on tested configurations.
        </p>

        <table className="evidence-table">
          <thead>
            <tr>
              <th>Domain</th>
              <th>Result</th>
              <th>Detail</th>
              <th>Verdict</th>
            </tr>
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
              <td>Optimal (la01)</td>
              <td>40% compute saved</td>
              <td><span className="evidence-badge evidence-badge--safe">SAFE</span></td>
            </tr>
            <tr>
              <td className="evidence-domain">VRPTW</td>
              <td>19% better quality</td>
              <td>p &lt; 0.001 (adaptive)</td>
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
          4 domains · 5 algorithms · 320+ validation runs
        </p>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 7: Domains
      ═══════════════════════════════════════════════════════════════ */}
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

      {/* ═══════════════════════════════════════════════════════════════
          Section 8: Principles
      ═══════════════════════════════════════════════════════════════ */}
      <section className="landing-section landing-principles">
        <span className="landing-principle">Everything measurable.</span>
        <span className="landing-principle">Everything reproducible.</span>
        <span className="landing-principle">Everything benchmarked.</span>
        <span className="landing-principle">Everything explainable.</span>
      </section>

      {/* ═══════════════════════════════════════════════════════════════
          Section 9: Recent Runs (compact)
      ═══════════════════════════════════════════════════════════════ */}
      {runs.length > 0 && (
        <section className="landing-section landing-runs">
          <RunList runs={runs} />
        </section>
      )}
    </div>
  );
}
