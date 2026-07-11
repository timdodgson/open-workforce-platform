import Link from 'next/link';

const GITHUB = 'https://github.com/timdodgson/open-workforce-platform';
const CITATION_MD = `${GITHUB}/blob/main/docs/CITATION.md`;
const DISSERTATION = `${GITHUB}/blob/main/inspiration/Final_Project_Tim_Dodgson.pdf`;

const LEARNING_PATH = [
  {
    level: 'Level 1',
    title: 'Browse the live lab',
    time: '~30 min',
    steps: [
      'Open /lab for domain metrics and validation snapshot',
      'Compare algorithm gaps on /benchmarks (INRC-II, CVRPLIB, Solomon, OR-Library)',
      'Read Welch t-tests on /statistics',
      'Inspect one run end-to-end (summary → search → explain)',
    ],
    deliverable: 'One-page gap comparison with lab screenshots',
    href: '/lab',
    cta: 'Open lab',
  },
  {
    level: 'Level 2',
    title: 'Reproduce one run locally',
    time: '~2 hours',
    steps: [
      'Clone repo, install Go, run go test ./... -short in platform/go',
      'Run EXP-008 command (CVRP A-n32-k5, hybrid policy, seed 42)',
      'Start dashboard with npm run dev, inspect /runs/<your-label>/summary',
      'Compare objective to val-cvrp-a32k5-sa-hybrid-s42 on the public lab',
    ],
    deliverable: 'Reproduction report: command, hardware, objective, runtime',
    href: '#reproduce',
    cta: 'See command',
  },
  {
    level: 'Level 3',
    title: 'Algorithm ladder (one domain)',
    time: 'Half day',
    steps: [
      'Run SA, LAHC, Tabu, Portfolio on one instance (--worker-decision-mode off)',
      'Record gap-to-optimal for each mode',
      'Use /benchmarks reproduce hints or docs/06-engineering/benchmark-suite.md',
    ],
    deliverable: 'Modes × objective table with brief interpretation',
    href: '/benchmarks',
    cta: 'Benchmark ladder',
  },
  {
    level: 'Level 4',
    title: 'Statistical replication',
    time: 'Multi-day',
    steps: [
      'Pick EXP-007 (assist modes, 320 runs) or SI2 matrix subset (val-*)',
      'Smoke test: validate-si2-quick.ps1 (6 runs)',
      'Replicate one domain seed sweep; recompute Welch comparison',
    ],
    deliverable: 'Domain-specific replication study with p-values',
    href: '/experiment-matrix',
    cta: 'Experiment matrix',
  },
] as const;

const BIBTEX_SOFTWARE = `@software{dodgson2026pfrs,
  author       = {Dodgson, Tim},
  title        = {PFRS Lab: Multi-Domain Optimisation Research Platform},
  year         = {2026},
  url          = {https://pfrs-lab.com},
  repository   = {https://github.com/timdodgson/open-workforce-platform},
  note         = {NRP, CVRP, JSS, VRPTW with Search Intelligence validation}
}`;

const REPRO_COMMAND = `cd platform/go
go run ./cmd/owp solve cvrp \\
  --instance ../../examples/cvrp/A-n32-k5.vrp \\
  --mode sa --iterations 500000 \\
  --policy-mode hybrid --policy-dir ../ml/policies \\
  --seed 42 --run-label my-cvrp-sa-hybrid-s42 \\
  --storage local`;

export default function ReproducePage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">Academic · Cite &amp; reproduce</p>
        <h1 className="site-title site-title--single">Evidence you can verify.</h1>
        <p className="site-lead">
          PFRS Lab publishes every benchmark command, seed, and statistical test. Use the live lab
          for exploration, clone the repo to reproduce a single run, or replicate full validation
          suites for coursework and papers.
        </p>
        <div className="site-profile-links">
          <Link href="/lab" className="site-btn-primary">Browse live results</Link>
          <a href={CITATION_MD} className="site-btn-secondary" target="_blank" rel="noopener noreferrer">
            Full citation guide (GitHub)
          </a>
          <a href={DISSERTATION} className="site-btn-secondary" target="_blank" rel="noopener noreferrer">
            Original dissertation
          </a>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">Student learning path</h2>
        <p className="site-body site-body--center">
          Four levels from browser-only to statistical replication. No AWS required for Levels 1–2.
        </p>
        <div className="site-learn-path">
          {LEARNING_PATH.map((item) => (
            <article key={item.level} className="site-learn-card">
              <div className="site-learn-card-head">
                <span className="site-learn-level">{item.level}</span>
                <span className="site-learn-time">{item.time}</span>
              </div>
              <h3 className="site-learn-title">{item.title}</h3>
              <ol className="site-learn-steps">
                {item.steps.map((step) => (
                  <li key={step}>{step}</li>
                ))}
              </ol>
              <p className="site-learn-deliverable">
                <strong>Deliverable:</strong> {item.deliverable}
              </p>
              <Link href={item.href} className="site-inline-link">{item.cta} →</Link>
            </article>
          ))}
        </div>
      </section>

      <section id="reproduce" className="site-section site-section--panel">
        <h2 className="site-heading">Reproduce one run (EXP-008)</h2>
        <p className="site-body">
          Canonical SI2 cell: CVRP A-n32-k5, SA, hybrid policy, seed 42. Same seed → same objective.
          Documented in <a href={`${GITHUB}/blob/main/docs/EXPERIMENTS.md`} className="site-inline-link" target="_blank" rel="noopener noreferrer">EXPERIMENTS.md § EXP-008</a>.
        </p>
        <div className="site-copy-block">
          <p className="site-copy-label">PowerShell (from repo root)</p>
          <pre className="site-copy-text site-copy-text--pre">{REPRO_COMMAND}</pre>
        </div>
        <p className="site-body site-body--compact">
          Then <code className="site-code">cd platform/web/pfrs-lab && npm run dev</code> and open{' '}
          <code className="site-code">/runs/my-cvrp-sa-hybrid-s42/summary</code>.
          Compare to the published run{' '}
          <Link href="/runs/val-cvrp-a32k5-sa-hybrid-s42/summary" className="site-inline-link">
            val-cvrp-a32k5-sa-hybrid-s42
          </Link>.
        </p>
      </section>

      <section id="cite" className="site-section">
        <h2 className="site-heading">How to cite</h2>
        <p className="site-body site-body--center">
          Use in papers, theses, and course materials. Instance-level citations (INRC-II, CVRPLIB, Solomon, OR-Library) should reference the original benchmarks.
        </p>
        <div className="site-copy-block">
          <p className="site-copy-label">BibTeX — software / platform</p>
          <pre className="site-copy-text site-copy-text--pre">{BIBTEX_SOFTWARE}</pre>
        </div>
        <div className="site-executive-grid site-executive-grid--cite">
          <div className="site-executive-block">
            <h3>320-run SI validation</h3>
            <p>
              Assist/adaptive modes across four domains. Live tables on{' '}
              <Link href="/statistics" className="site-inline-link">/statistics</Link>.
              Commands in the statistical validation report on GitHub.
            </p>
          </div>
          <div className="site-executive-block">
            <h3>288-run SI2 matrix</h3>
            <p>
              rules / hybrid / learned policies. Catalog on{' '}
              <Link href="/experiment-matrix" className="site-inline-link">/experiment-matrix</Link>.
              Coverage audit in docs/reports/val-gap-audit.md.
            </p>
          </div>
          <div className="site-executive-block">
            <h3>Dissertation lineage</h3>
            <p>
              Nurse rostering research from 2014 informs the NRP domain and beam-search design.
              <a href={DISSERTATION} className="site-inline-link" target="_blank" rel="noopener noreferrer"> PDF on GitHub</a>.
            </p>
          </div>
          <div className="site-executive-block">
            <h3>R&amp;D runbook</h3>
            <p>
              Full validation scripts, policy retrain, and production sync for researchers extending the platform.
              <a href={`${GITHUB}/blob/main/docs/RUNBOOK.md`} className="site-inline-link" target="_blank" rel="noopener noreferrer"> RUNBOOK.md</a>.
            </p>
          </div>
        </div>
      </section>

      <section className="site-section site-cta-band">
        <h2 className="site-cta-title">Ready to dig in?</h2>
        <p className="site-cta-desc">
          Start in the browser, clone when you need your own numbers, cite when you publish.
        </p>
        <div className="site-profile-links">
          <Link href="/benchmarks" className="site-btn-secondary">Benchmark ladder</Link>
          <Link href="/research" className="site-btn-secondary">Technical depth</Link>
          <Link href="/lab" className="site-btn-primary">Open research lab</Link>
        </div>
      </section>
    </div>
  );
}
