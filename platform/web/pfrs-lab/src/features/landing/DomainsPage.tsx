import Link from 'next/link';
import { DOMAINS } from '@/lib/field-guide';

export default function DomainsPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">Field guide</p>
        <h1 className="site-title site-title--single">Optimisation domains that map to real operations</h1>
        <p className="site-lead">
          Nurse rostering, vehicle routing, time windows, and job shops — why these NP-hard
          problems matter outside the lab, and which published benchmarks we measure against.
        </p>
        <div className="site-hero-actions">
          <Link href="/algorithms" className="site-btn-secondary">Algorithms →</Link>
          <Link href="/references" className="site-btn-secondary">Standards &amp; references →</Link>
          <Link href="/benchmarks" className="site-btn-primary">Live benchmark gaps</Link>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">One interface, four industries</h2>
        <p className="site-body">
          The platform is not four separate toys. Each domain implements the same search contract,
          so an algorithm improvement or Search Intelligence policy can transfer — while still
          respecting domain-specific hard constraints (coverage, capacity, time windows, machine
          conflicts).
        </p>
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">Domains</h2>
        <div className="site-guide-stack">
          {DOMAINS.map((d) => (
            <article key={d.id} className="site-guide-card" id={d.id}>
              <p className="site-guide-kicker">{d.shortName}</p>
              <h3 className="site-guide-title">{d.name}</h3>
              <dl className="site-guide-dl">
                <div>
                  <dt>Real-world problem</dt>
                  <dd>{d.realWorld}</dd>
                </div>
                <div>
                  <dt>Why it is hard</dt>
                  <dd>{d.whyHard}</dd>
                </div>
                <div>
                  <dt>Published benchmark</dt>
                  <dd>
                    <a href={d.benchmarkUrl} className="site-inline-link" target="_blank" rel="noopener noreferrer">
                      {d.benchmark}
                    </a>
                  </dd>
                </div>
                <div>
                  <dt>On this platform</dt>
                  <dd>{d.platformNote}</dd>
                </div>
                <div>
                  <dt>Algorithms we typically try</dt>
                  <dd>{d.relatedAlgos.join(' · ')}</dd>
                </div>
              </dl>
            </article>
          ))}
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">Flagship challenge</h2>
        <p className="site-body">
          Among these four, <strong>NRP</strong> remains the hardest to crack here: largest gap to
          a published ILP feasible reference (a yardstick from HiGHS — exact methods do not scale
          to the larger instances we care about), and the only domain where learned policies still
          sometimes regress versus rules on the validation harness. That is intentional honesty —
          the site shows the open problem, not only the wins.
        </p>
        <p className="site-body">
          See the{' '}
          <Link href="/research" className="site-inline-link">research flagship section</Link>
          {' '}and{' '}
          <Link href="/getting-started" className="site-inline-link">NRP beam examples</Link>.
        </p>
      </section>
    </div>
  );
}
