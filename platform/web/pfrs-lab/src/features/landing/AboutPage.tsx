import Link from 'next/link';
import ByodExtensionSection from '@/features/landing/ByodExtensionSection';

const GITHUB = 'https://github.com/timdodgson/open-workforce-platform';
const DISSERTATION = `${GITHUB}/blob/main/inspiration/Final_Project_Tim_Dodgson.pdf`;
const LINKEDIN = 'https://www.linkedin.com/in/tim-dodgson/';

const PROOF_POINTS = [
  {
    metric: '4 domains',
    detail: 'NRP, CVRP, JSS, VRPTW — one generic search interface, shared SI layer',
  },
  {
    metric: '320+ runs',
    detail: 'Multi-seed validation with statistical tests in the public lab',
  },
  {
    metric: 'owp-sdk',
    detail: 'Published BYOD module — plug in new problem types without forking the engine',
  },
] as const;

export default function AboutPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">About · Tim Dodgson</p>
        <h1 className="site-title site-title--single">Principal engineer. Measurable systems.</h1>
        <p className="site-lead">
          PFRS Lab is a portfolio-grade research platform — built end-to-end by a Principal
          Software Engineer with 10+ years shipping serverless production systems, now applying
          that discipline to multi-domain optimisation and applied ML.
        </p>
        <div className="site-profile-links">
          <a href={LINKEDIN} className="site-btn-secondary" target="_blank" rel="noopener noreferrer">
            LinkedIn
          </a>
          <a href={GITHUB} className="site-btn-secondary" target="_blank" rel="noopener noreferrer">
            GitHub
          </a>
          <a href={DISSERTATION} className="site-btn-secondary" target="_blank" rel="noopener noreferrer">
            Dissertation PDF
          </a>
          <Link href="/lab" className="site-btn-primary">Open research lab</Link>
        </div>
      </section>

      <section className="site-section">
        <div className="site-profile-card">
          <div>
            <p className="site-profile-role">Principal Software Engineer · CDL Software</p>
            <p className="site-profile-sub">
              Serverless · AWS · CI/CD · IaC · System design at scale
            </p>
            <p className="site-profile-bio">
              10+ years at CDL, junior developer to technical leader. Delivered greenfield
              serverless products, reusable AWS blueprints, and production mobile apps for major
              financial clients. Earlier career: built a diagnostic tracing platform still in use,
              credited with saving the business over £1 million. BSc (First), Manchester Metropolitan
              University. AWS &amp; GitLab certified.
            </p>
          </div>
          <ul className="site-proof-list">
            {PROOF_POINTS.map((p) => (
              <li key={p.metric}>
                <span className="site-proof-metric">{p.metric}</span>
                <span className="site-proof-detail">{p.detail}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <section id="summary" className="site-section site-section--panel site-executive">
        <h2 className="site-heading">Executive summary</h2>
        <p className="site-body site-executive-lead">
          For hiring managers and technical leaders — problem, approach, evidence, and audit path in one page.
        </p>

        <div className="site-executive-grid">
          <div className="site-executive-block">
            <h3>Problem</h3>
            <p>
              NP-hard optimisation is usually isolated solvers or demos. PFRS Lab is a single
              platform across four benchmark domains with statistical validation and a public audit trail.
            </p>
          </div>
          <div className="site-executive-block">
            <h3>Approach</h3>
            <p>
              Go metaheuristics + beam search, Search Intelligence policies on real telemetry,
              Python training pipeline, Next.js lab UI, S3 artifact storage — designed and built by one engineer.
            </p>
          </div>
          <div className="site-executive-block">
            <h3>Evidence</h3>
            <ul>
              <li>CVRP: identical quality, 60–73% less compute</li>
              <li>JSS: optimal on la01 reference, ~40% compute saved</li>
              <li>VRPTW: ~19% better quality (p &lt; 0.001)</li>
              <li>NRP: no degradation; zero feasibility regressions in validation</li>
            </ul>
          </div>
          <div className="site-executive-block">
            <h3>Audit in 5 minutes</h3>
            <ol>
              <li><Link href="/lab" className="site-inline-link">Open Lab</Link> — live run counts</li>
              <li><Link href="/benchmarks" className="site-inline-link">Benchmarks</Link> — gaps to known optima</li>
              <li><Link href="/statistics" className="site-inline-link">Statistics</Link> — SI significance tests</li>
              <li><Link href="/experiment-matrix" className="site-inline-link">Experiment matrix</Link> — every run variation documented</li>
              <li><Link href="/runs" className="site-inline-link">Runs</Link> — drill into artifacts</li>
            </ol>
          </div>
        </div>

        <p className="site-body mt-4">
          <a href={`${GITHUB}/blob/main/docs/portfolio/executive-summary.md`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
            Markdown version (repo) →
          </a>
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
                Nurse rostering optimisation (First-class BSc project) — the seed problem behind PFRS Lab.
              </p>
            </div>
          </div>
          <div className="site-timeline-item">
            <span className="site-timeline-era">2016–2025</span>
            <div>
              <p className="site-timeline-title">CDL Software — Junior to Principal</p>
              <p className="site-timeline-desc">
                Java and Android production apps, diagnostic platforms, then serverless architecture
                and technical leadership — three greenfield solutions to production, mentoring squads,
                AWS and GitLab certifications.
              </p>
            </div>
          </div>
          <div className="site-timeline-item">
            <span className="site-timeline-era">Now</span>
            <div>
              <p className="site-timeline-title">PFRS Lab — AI-native optimisation</p>
              <p className="site-timeline-desc">
                Reunites dissertation research with production engineering: multi-domain solvers,
                learned policies, statistical validation, BYOD SDK, and this public lab.
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
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">Outbound kit</h2>
        <p className="site-body">
          Copy-paste lines for LinkedIn, CV, or a short intro email. Full versions in the repo at{' '}
          <code className="text-xs text-gray-500">docs/portfolio/outbound-kit.md</code>.
        </p>

        <div className="site-copy-block">
          <p className="site-copy-label">LinkedIn one-liner</p>
          <p className="site-copy-text">
            PFRS Lab (pfrs-lab.com) — multi-domain optimisation + Search Intelligence I built end-to-end:
            Go solvers, ML on real telemetry, 320+ validated runs, public lab. Principal SWE · serverless/AWS.
          </p>
        </div>

        <div className="site-copy-block">
          <p className="site-copy-label">Email opener</p>
          <p className="site-copy-text">
            I built PFRS Lab (pfrs-lab.com) — a live optimisation platform across four NP-hard domains with
            Search Intelligence validated on 320+ runs. I&apos;m a Principal Software Engineer (CDL, 10+ years)
            applying production discipline to applied ML. Start at /lab for metrics, /about#summary for context.
            Source: github.com/timdodgson/open-workforce-platform
          </p>
        </div>

        <p className="site-body mt-4">
          <a href={`${GITHUB}/blob/main/docs/portfolio/outbound-kit.md`} className="site-inline-link" target="_blank" rel="noopener noreferrer">
            Full outbound kit →
          </a>
        </p>
      </section>

      <ByodExtensionSection />

      <section className="site-cta-band">
        <h2 className="site-cta-title">Judge the work on data</h2>
        <p className="site-cta-desc">
          Skip the marketing — open the lab, pick a benchmark, read the statistics.
        </p>
        <div className="site-hero-actions site-hero-actions--center">
          <Link href="/lab" className="site-btn-primary">Open research lab</Link>
          <Link href="/research" className="site-btn-secondary">Technical deep dive</Link>
        </div>
      </section>
    </div>
  );
}
