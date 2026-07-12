import Link from 'next/link';
import { ALGORITHMS } from '@/lib/field-guide';

export default function AlgorithmsPage() {
  return (
    <div className="site">
      <section className="site-hero site-hero--compact">
        <p className="site-eyebrow">Field guide</p>
        <h1 className="site-title site-title--single">Metaheuristic algorithms compared</h1>
        <p className="site-lead">
          Why Simulated Annealing, LAHC, Tabu, Genetic Algorithms, and portfolio search behave
          differently — and when this platform uses each one on real benchmarks.
        </p>
        <div className="site-hero-actions">
          <Link href="/domains" className="site-btn-secondary">Problem domains →</Link>
          <Link href="/references" className="site-btn-secondary">Standards &amp; references →</Link>
          <Link href="/getting-started" className="site-btn-primary">Run an example</Link>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">How to read this</h2>
        <p className="site-body">
          Every algorithm below plugs into the same move / evaluate / undo interface. Changing
          <code className="site-code"> --mode</code> does not change the domain model — only the
          search policy. That is why we can compare them fairly on CVRPLIB, Solomon, OR-Library,
          and INRC-II instances.
        </p>
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">At a glance</h2>
        <div className="site-table-wrap">
          <table className="site-table">
            <thead>
              <tr>
                <th scope="col">Algorithm</th>
                <th scope="col">Strength</th>
                <th scope="col">Trade-off</th>
                <th scope="col">Typical use here</th>
              </tr>
            </thead>
            <tbody>
              {ALGORITHMS.map((a) => (
                <tr key={a.id}>
                  <th scope="row">{a.abbr}</th>
                  <td>{a.strength}</td>
                  <td>{a.weakness}</td>
                  <td>{a.when}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="site-section">
        <h2 className="site-heading">Each algorithm in context</h2>
        <div className="site-guide-stack">
          {ALGORITHMS.map((a) => (
            <article key={a.id} className="site-guide-card" id={a.id}>
              <p className="site-guide-kicker">{a.family}</p>
              <h3 className="site-guide-title">
                <span className="site-guide-abbr">{a.abbr}</span> {a.name}
              </h3>
              <dl className="site-guide-dl">
                <div>
                  <dt>Strength</dt>
                  <dd>{a.strength}</dd>
                </div>
                <div>
                  <dt>Trade-off</dt>
                  <dd>{a.weakness}</dd>
                </div>
                <div>
                  <dt>When we use it</dt>
                  <dd>{a.when}</dd>
                </div>
                <div>
                  <dt>Real-world intuition</dt>
                  <dd>{a.realWorld}</dd>
                </div>
              </dl>
            </article>
          ))}
        </div>
      </section>

      <section className="site-section site-section--panel">
        <h2 className="site-heading">Not a beauty contest</h2>
        <p className="site-body">
          No single metaheuristic wins every instance. Portfolio mode exists because that is also
          how serious OR practice hedges. Search Intelligence (Assist and Policies) sits on top —
          it does not replace these algorithms; it decides where to spend their budgets.
        </p>
        <p className="site-body">
          Deeper maths and beam-search design live on the{' '}
          <Link href="/research" className="site-inline-link">research reference</Link>
          . Worked CLI commands are on{' '}
          <Link href="/getting-started" className="site-inline-link">Getting started</Link>.
        </p>
      </section>
    </div>
  );
}
