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

const HIGHLIGHTS = [
  {
    title: '£1M+ diagnostic platform',
    detail:
      'Built independently as a Java engineer — still in use — tracing aggregator-to-site errors and credited with saving the business over £1 million.',
  },
  {
    title: 'Three greenfield serverless products',
    detail:
      'Took AWS / IaC solutions from blank page to production as a Senior, then scaled that pattern as Principal.',
  },
  {
    title: 'Banking & broker mobile apps',
    detail: 'Shipped Android / Java production apps for major financial clients under real release pressure.',
  },
  {
    title: 'Principal technical leadership',
    detail:
      '“You build it, you run it” culture, reusable serverless blueprints, mentoring across squads — and a bias to embrace AI early.',
  },
] as const;

const CERTS_PRIMARY = [
  {
    name: 'AWS Certified Cloud Practitioner',
    issuer: 'Amazon Web Services',
    year: '2022',
    url: 'https://credly.com/badges/146f2309-b4f1-4f05-b8a0-e2a7c86eb760',
  },
  {
    name: 'Well-Architected Proficient',
    issuer: 'Amazon Web Services',
    year: '2024',
    url: 'https://credly.com/badges/4437a264-d9c3-484b-915e-ef3f03a39325',
  },
  {
    name: 'GitLab Certified CI/CD Associate',
    issuer: 'GitLab',
    year: '2024',
    url: 'https://credly.com/badges/6f8cbc53-cd9b-4684-b5c6-8143da7e52fb',
  },
  {
    name: 'GitLab Certified Security Specialist',
    issuer: 'GitLab',
    year: '2024',
    url: 'https://credly.com/badges/9ba1c837-7285-40e4-b7ba-c78c83aff011',
  },
  {
    name: 'Professional Scrum Master I (PSM I)',
    issuer: 'Scrum.org',
    year: '2017',
    url: 'https://credly.com/badges/653a4034-5855-4d65-aa63-ca07cadd97db',
  },
] as const;

const JOURNEY = [
  {
    era: 'Trades',
    title: 'Mechanic → white goods → TV & video',
    desc:
      'City & Guilds as a car mechanic, then white-goods repair, then television and video. At the peak of that craft I repaired to component level — find the one faulty transistor. Integrated circuits arrived and the board became a black box: know the inputs, expect the outputs, stop pretending you can see every gate.',
  },
  {
    era: 'Parallel track',
    title: 'BBC Micro → 286 → 386 → 486 → networking',
    desc:
      'Through all of that I was deep in computers — BBC Micro world first, then each generation of PCs and networking. Hands-on systems thinking long before it was a job title.',
  },
  {
    era: '~2000',
    title: 'Career change into IT',
    desc:
      'Made the deliberate move into technology as a profession. Network Administrator at Greenbank Preparatory School (2000–2012) — keep the estate running, solve real user problems, own the outcome.',
  },
  {
    era: '2012–2015',
    title: 'Manchester Metropolitan University',
    desc:
      'BSc First Class. Dissertation on nurse rostering optimisation — the seed problem that became PFRS Lab a decade later.',
  },
  {
    era: '2016–present',
    title: 'CDL Software — Junior → Principal',
    desc:
      'Junior (2016) → Intermediate → Senior → Principal (2021–). Java and Android production apps, the diagnostic platform, then serverless architecture and technical leadership across squads for major insurance and financial clients.',
  },
  {
    era: 'Now',
    title: 'PFRS Lab — AI-native optimisation',
    desc:
      'Reunites dissertation research with production engineering: multi-domain solvers, learned policies, statistical validation, BYOD SDK, and this public lab.',
  },
] as const;

export default function AboutPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">About · Tim Dodgson</p>
        <h1 className="site-title site-title--single">Principal engineer. Measurable systems.</h1>
        <p className="site-lead">
          From component-level repair to Principal Software Engineer — a journey of systems,
          leadership, and knowing when the black box is the point. PFRS Lab is that discipline
          applied to multi-domain optimisation and applied ML.
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
              Serverless · AWS · CI/CD · IaC · System design at scale · Stockport, UK
            </p>
            <p className="site-profile-bio">
              10+ years at CDL, junior developer to technical leader. Before that: trades,
              networking, and a First-class degree. I lead by getting people to want to achieve
              with me — and I look at the big picture when the industry shifts.
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
              <li>NRP: flagship challenge — honest gaps, not only wins</li>
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

      <section id="journey" className="site-section site-section--panel">
        <h2 className="site-heading">The journey</h2>
        <p className="site-body" style={{ marginBottom: '1.5rem' }}>
          Not a straight line into software. A long apprenticeship in how systems fail — engines,
          appliances, circuit boards, school networks — then deliberate career change into IT,
          university, and principal-level engineering.
        </p>
        <div className="site-timeline">
          {JOURNEY.map((item) => (
            <div key={item.era} className="site-timeline-item">
              <span className="site-timeline-era">{item.era}</span>
              <div>
                <p className="site-timeline-title">{item.title}</p>
                <p className="site-timeline-desc">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section id="highlights" className="site-section">
        <h2 className="site-heading">Career high points</h2>
        <div className="site-verify-grid">
          {HIGHLIGHTS.map((h) => (
            <div key={h.title} className="site-verify-item">
              <p className="site-verify-label">{h.title}</p>
              <p className="site-verify-desc">{h.detail}</p>
            </div>
          ))}
        </div>
      </section>

      <section id="leadership" className="site-section site-section--panel">
        <h2 className="site-heading">Leadership and the black box</h2>
        <p className="site-body">
          I am a natural leader — not by title first, but by getting people to <em>want</em> to
          achieve with me. I also sit with the big picture. Right now many developers push back
          on AI in the workflow. I see the opposite: the tools will only get better. Embrace early,
          become the next expert — or risk being left behind. In five years, contract-first work
          with AI may be normal; today it still needs a careful frame so it is not mistaken for
          skipping engineering responsibility.
        </p>
        <blockquote className="site-quote">
          <p>
            When I repaired TV and video, I could work to component level — find the one faulty
            transistor. Then integrated circuits arrived. What happened inside was out of reach;
            the board became a black box. All that mattered was what goes in and what you should
            expect out. The IC was not “untested” — you verified the pinout and the behaviour.
          </p>
          <p>
            That is where software development is moving. The skill is getting AI to understand
            the contract — inputs and expected outputs. Default to the interface. Prove behaviour
            at the boundary with tests, types, reviews on risky paths, security, and performance.
            Open the box when something is wrong, the risk is high, or you are learning — not by
            staring at every gate when the pinout already tells the truth.
          </p>
        </blockquote>
        <p className="site-body">
          PFRS Lab is built in that spirit: clear interfaces, measured outcomes, and Search
          Intelligence that treats search itself as something you observe and improve — not only
          something you hand-tune forever.
        </p>
      </section>

      <section id="certifications" className="site-section">
        <h2 className="site-heading">Certifications</h2>
        <p className="site-body" style={{ marginBottom: '1.25rem' }}>
          Signal over inventory — Credly-verified where available. Full list on{' '}
          <a href={LINKEDIN} className="site-inline-link" target="_blank" rel="noopener noreferrer">
            LinkedIn
          </a>
          .
        </p>
        <ul className="site-cert-list">
          {CERTS_PRIMARY.map((c) => (
            <li key={c.name} className="site-cert-item">
              <a href={c.url} className="site-cert-name" target="_blank" rel="noopener noreferrer">
                {c.name}
              </a>
              <span className="site-cert-meta">
                {c.issuer} · {c.year}
              </span>
            </li>
          ))}
        </ul>
        <p className="site-body" style={{ marginTop: '1rem' }}>
          Also: AWS Knowledge badges (Compute, Cloud Essentials, Networking, Architecting, Serverless).
          Education: BSc (First), Manchester Metropolitan University.
        </p>
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">What you can verify</h2>
        <div className="site-verify-grid">
          <div className="site-verify-item">
            <p className="site-verify-label">Solvers</p>
            <p className="site-verify-desc">Go CLI — SA, LAHC, Tabu, GA, portfolio, beam search, ILP benchmarks</p>
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

      <section className="site-section">
        <h2 className="site-heading">Outbound kit</h2>
        <p className="site-body">
          Copy-paste lines for LinkedIn, CV, or a short intro email. Full versions in the repo at{' '}
          <code className="site-code">docs/portfolio/outbound-kit.md</code>.
        </p>

        <div className="site-copy-block">
          <p className="site-copy-label">LinkedIn one-liner</p>
          <p className="site-copy-text">
            PFRS Lab (pfrs-lab.com) — multi-domain optimisation + Search Intelligence I built end-to-end:
            Go solvers, ML on real telemetry, 320+ validated runs, public lab. Principal SWE · serverless/AWS.
            Journey: trades → networking → First-class BSc → CDL Principal.
          </p>
        </div>

        <div className="site-copy-block">
          <p className="site-copy-label">Email opener</p>
          <p className="site-copy-text">
            I built PFRS Lab (pfrs-lab.com) — a live optimisation platform across four NP-hard domains with
            Search Intelligence validated on 320+ runs. I&apos;m a Principal Software Engineer (CDL)
            with a non-linear path: City &amp; Guilds mechanic through electronics repair into IT,
            then First-class CS and principal-level serverless leadership. Start at /lab for metrics,
            /about for the journey. Source: github.com/timdodgson/open-workforce-platform
          </p>
        </div>

        <p className="site-body" style={{ marginTop: '1rem' }}>
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
