import Link from 'next/link';
import ByodExtensionSection from '@/features/landing/ByodExtensionSection';

export default function MarketingHome() {
  return (
    <div className="site">
      <section className="site-hero">
        <p className="site-eyebrow">Open optimisation research · Production-grade engineering</p>
        <h1 className="site-title">
          Hard problems.<br />
          <span className="site-title-accent">Measured answers.</span>
        </h1>
        <p className="site-lead">
          PFRS Lab is a live research platform for NP-hard optimisation — nurse rostering,
          vehicle routing, job shop scheduling, and time windows — with metaheuristic search,
          statistical validation, and Search Intelligence that learns where compute actually helps.
        </p>
        <div className="site-hero-actions">
          <Link href="/lab" className="site-btn-primary">Explore the live lab</Link>
          <Link href="/research" className="site-btn-secondary">Read the research</Link>
        </div>
      </section>

      <section className="site-metrics">
        <div className="site-metric">
          <span className="site-metric-value">4</span>
          <span className="site-metric-label">benchmark domains</span>
        </div>
        <div className="site-metric">
          <span className="site-metric-value">320+</span>
          <span className="site-metric-label">validated runs</span>
        </div>
        <div className="site-metric">
          <span className="site-metric-value">40–73%</span>
          <span className="site-metric-label">compute saved (SI)</span>
        </div>
        <div className="site-metric">
          <span className="site-metric-value">0</span>
          <span className="site-metric-label">feasibility regressions</span>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">Built for people who need evidence</h2>
        <p className="site-body site-body--center">
          Whether you are writing a paper, teaching a course, or evaluating whether adaptive
          search is real — everything here is reproducible, benchmarked, and explorable in the browser.
        </p>
        <div className="site-audience">
          <article className="site-audience-card">
            <h3>Researchers &amp; academics</h3>
            <p>
              Multi-domain platform with Welch t-tests, effect sizes, and a full experiment matrix
              documenting every run variation and which options are on or off — and why.
            </p>
            <Link href="/experiment-matrix" className="site-inline-link">Experiment matrix →</Link>
          </article>
          <article className="site-audience-card">
            <h3>University students</h3>
            <p>
              See how SA, LAHC, and Tabu behave on real INRC-II, CVRPLIB, Solomon, and OR-Library
              instances. Compare gaps to known optima. Follow the SI pipeline from telemetry to policy.
            </p>
            <Link href="/benchmarks" className="site-inline-link">Benchmark ladder →</Link>
          </article>
          <article className="site-audience-card">
            <h3>Engineers &amp; hiring managers</h3>
            <p>
              End-to-end system: Go solvers, serverless storage, ML policy training, Next.js lab UI,
              explainability, and statistical rigour — not a slide deck, a working product.
            </p>
            <Link href="/about" className="site-inline-link">About the builder →</Link>
          </article>
        </div>
      </section>

      <section className="site-section site-section--panel">
        <div className="site-split">
          <div>
            <h2 className="site-heading">What makes this different</h2>
            <ul className="site-list">
              <li>
                <strong>One generic interface</strong> — same move/evaluate/undo contract across
                four unrelated NP-hard problem classes.
              </li>
              <li>
                <strong>Search Intelligence</strong> — observes search telemetry, trains checkpoint
                policies, and reallocates compute with measured safety bounds.
              </li>
              <li>
                <strong>Explainability by default</strong> — rule-based run narratives, feature
                importance, what-if simulation, and an AI research assistant.
              </li>
            </ul>
          </div>
          <div className="site-proof-card">
            <p className="site-proof-title">Validation snapshot</p>
            <ul className="site-proof-rows">
              <li><span>CVRP</span><span>Identical quality, 60–73% less compute</span></li>
              <li><span>JSS</span><span>Optimal on la01, 40% compute saved</span></li>
              <li><span>VRPTW</span><span>19% better quality, p &lt; 0.001</span></li>
              <li><span>NRP</span><span>No degradation vs baseline</span></li>
            </ul>
            <Link href="/statistics" className="site-inline-link">Full statistical analysis →</Link>
          </div>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">From dissertation to production AI</h2>
        <p className="site-body">
          This work started as a{' '}
          <a
            href="https://github.com/timdodgson/open-workforce-platform/blob/main/inspiration/Final_Project_Tim_Dodgson.pdf"
            className="site-inline-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            university dissertation on nurse rostering
          </a>{' '}
          over a decade ago. It has been rebuilt with twenty years of principal-level engineering —
          full stack, serverless at scale, and now applied ML on real search telemetry — to answer
          one question: can a single platform solve multiple hard domains and learn from its own history?
        </p>
        <p className="site-body">
          <Link href="/about" className="site-inline-link">Read the full story →</Link>
        </p>
      </section>

      <ByodExtensionSection compact />

      <section className="site-cta-band">
        <h2 className="site-cta-title">Ready to dig into the data?</h2>
        <p className="site-cta-desc">
          The research lab is a separate workspace — benchmarks, runs, statistics, intelligence
          dashboards, and admin tools. No marketing fluff, just metrics and reproducibility.
        </p>
        <Link href="/lab" className="site-btn-primary site-btn-primary--large">Enter the research lab</Link>
      </section>
    </div>
  );
}
