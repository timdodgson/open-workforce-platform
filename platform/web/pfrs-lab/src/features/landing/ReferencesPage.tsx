import Link from 'next/link';
import { REFERENCE_KIND_LABEL, REFERENCES, type ReferenceLink } from '@/lib/field-guide';

const KIND_ORDER: ReferenceLink['kind'][] = ['society', 'benchmark', 'archive', 'solver'];

export default function ReferencesPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">Field guide</p>
        <h1 className="site-title site-title--single">Standards and references for optimisation research</h1>
        <p className="site-lead">
          Societies, benchmark libraries, instance archives, and open solvers this work sits
          against — outbound links to the bodies that define operational research practice.
        </p>
        <div className="site-hero-actions">
          <Link href="/domains" className="site-btn-secondary">Domains →</Link>
          <Link href="/algorithms" className="site-btn-secondary">Algorithms →</Link>
          <Link href="/reproduce" className="site-btn-primary">Citation &amp; reproduce</Link>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">How to use this page</h2>
        <p className="site-body">
          These are not endorsements or partnerships. They are the public standards and datasets
          a serious OR portfolio is expected to name. If you are writing a paper, teaching a
          module, or reviewing engineering evidence, start here — then verify results in the lab.
        </p>
      </section>

      {KIND_ORDER.map((kind) => {
        const items = REFERENCES.filter((r) => r.kind === kind);
        return (
          <section key={kind} className="site-section site-section--panel">
            <h2 className="site-heading">{REFERENCE_KIND_LABEL[kind]}</h2>
            <ul className="site-ref-list">
              {items.map((r) => (
                <li key={r.id} className="site-ref-item">
                  <a href={r.url} className="site-ref-name" target="_blank" rel="noopener noreferrer">
                    {r.name}
                  </a>
                  <p className="site-ref-why">{r.why}</p>
                  <p className="site-ref-url">{r.url.replace(/^https?:\/\//, '')}</p>
                </li>
              ))}
            </ul>
          </section>
        );
      })}

      <section className="site-section">
        <h2 className="site-heading">Platform documentation</h2>
        <ul className="site-list">
          <li>
            <Link href="/reproduce" className="site-inline-link">Reproduce &amp; cite</Link>
            {' '}— academic citation path and student ladder
          </li>
          <li>
            <Link href="/research" className="site-inline-link">Research reference</Link>
            {' '}— algorithms, beam maths, Search Intelligence
          </li>
          <li>
            <a
              href="https://github.com/timdodgson/open-workforce-platform/blob/main/docs/BENCHMARKS.md"
              className="site-inline-link"
              target="_blank"
              rel="noopener noreferrer"
            >
              BENCHMARKS.md
            </a>
            {' '}— instance tables and gaps in the repository
          </li>
        </ul>
      </section>
    </div>
  );
}
