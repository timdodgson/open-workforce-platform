import Link from 'next/link';
import ByodExtensionSection from '@/features/landing/ByodExtensionSection';

export default function AboutPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">About</p>
        <h1 className="site-title site-title--single">Built to be inspected</h1>
        <p className="site-lead">
          PFRS Lab is both a serious optimisation research platform and a portfolio piece —
          designed so an academic can cite the methodology and an engineering leader can audit
          the system end to end.
        </p>
      </section>

      <section className="site-section">
        <h2 className="site-heading">Who built this</h2>
        <p className="site-body">
          Tim Dodgson — Principal Developer with a trajectory from full-stack product engineering
          through serverless architecture at scale, into applied AI on production systems.
          This project is the convergence of that path: hard computer science, shipped software,
          and measurable ML — not a toy demo.
        </p>
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">The arc</h2>
        <div className="site-timeline">
          <div className="site-timeline-item">
            <span className="site-timeline-era">2014</span>
            <div>
              <p className="site-timeline-title">University dissertation</p>
              <p className="site-timeline-desc">
                Nurse rostering optimisation — the seed problem that started everything.
              </p>
            </div>
          </div>
          <div className="site-timeline-item">
            <span className="site-timeline-era">Career</span>
            <div>
              <p className="site-timeline-title">Principal engineer</p>
              <p className="site-timeline-desc">
                Full-stack products, distributed systems, serverless platforms — the engineering
                discipline that makes research reproducible at scale.
              </p>
            </div>
          </div>
          <div className="site-timeline-item">
            <span className="site-timeline-era">Now</span>
            <div>
              <p className="site-timeline-title">AI-native optimisation lab</p>
              <p className="site-timeline-desc">
                Multi-domain solvers, telemetry pipelines, learned checkpoint policies, statistical
                validation, and a public lab UI — one cohesive system you can run, measure, and explain.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">What you can verify</h2>
        <div className="site-verify-grid">
          <div className="site-verify-item">
            <p className="site-verify-label">Solvers</p>
            <p className="site-verify-desc">Go CLI — SA, LAHC, Tabu, portfolio, beam search, ILP benchmarks</p>
          </div>
          <div className="site-verify-item">
            <p className="site-verify-label">Intelligence</p>
            <p className="site-verify-desc">Python policy training on real checkpoint telemetry</p>
          </div>
          <div className="site-verify-item">
            <p className="site-verify-label">Storage</p>
            <p className="site-verify-desc">S3-backed run artifacts, manifest-first dashboard loading</p>
          </div>
          <div className="site-verify-item">
            <p className="site-verify-label">Lab UI</p>
            <p className="site-verify-desc">Next.js — benchmarks, statistics, experiment matrix, explainability</p>
          </div>
        </div>
        <p className="site-body mt-6">
          <a
            href="https://github.com/timdodgson/open-workforce-platform"
            className="site-inline-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            View source on GitHub →
          </a>
          {' · '}
          <a
            href="https://github.com/timdodgson/open-workforce-platform/blob/main/inspiration/Final_Project_Tim_Dodgson.pdf"
            className="site-inline-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            Original dissertation PDF →
          </a>
        </p>
      </section>

      <ByodExtensionSection />

      <section className="site-cta-band">
        <h2 className="site-cta-title">See the evidence yourself</h2>
        <p className="site-cta-desc">
          The lab is where benchmarks, runs, and statistical tests live. Open it and judge the work on data.
        </p>
        <div className="site-hero-actions site-hero-actions--center">
          <Link href="/lab" className="site-btn-primary">Open research lab</Link>
          <Link href="/research" className="site-btn-secondary">Technical deep dive</Link>
        </div>
      </section>
    </div>
  );
}
