'use client';

/**
 * PFRS Lab Architecture Diagram — production landing-page quality.
 *
 * Inspired by the Open Workforce Platform diagram: colour-coded layers with
 * icons, depth, generous whitespace, and proper visual hierarchy.
 * Uses HTML/CSS instead of SVG for better rendering and responsiveness.
 */
export default function ArchitectureDiagram() {
  return (
    <div className="arch-diagram" role="img" aria-label="PFRS Lab platform architecture">
      {/* Layer 1: Problem Domains */}
      <div className="arch-layer arch-layer--domains">
        <span className="arch-layer-label">PROBLEM DOMAINS</span>
        <div className="arch-layer-items">
          <span className="arch-chip arch-chip--nrp">NRP</span>
          <span className="arch-chip arch-chip--cvrp">CVRP</span>
          <span className="arch-chip arch-chip--jss">JSS</span>
          <span className="arch-chip arch-chip--vrptw">VRPTW</span>
        </div>
        <span className="arch-layer-aside">Nurse Rostering · Vehicle Routing · Job Shop · Time Windows</span>
      </div>

      <div className="arch-arrow">↓</div>

      {/* Layer 2: Generic Problem Interface */}
      <div className="arch-layer arch-layer--interface">
        <span className="arch-layer-label">GENERIC PROBLEM INTERFACE</span>
        <span className="arch-layer-detail">TryMove · Evaluate · Undo · Constraints · Serialize</span>
      </div>

      <div className="arch-arrow">↓</div>

      {/* Layer 3: Algorithms */}
      <div className="arch-layer arch-layer--algorithms">
        <span className="arch-layer-label">SEARCH ALGORITHMS</span>
        <div className="arch-layer-items">
          <span className="arch-chip arch-chip--algo">SA</span>
          <span className="arch-chip arch-chip--algo">LAHC</span>
          <span className="arch-chip arch-chip--algo">Tabu</span>
          <span className="arch-chip arch-chip--algo">Portfolio</span>
          <span className="arch-chip arch-chip--algo">Adaptive</span>
        </div>
      </div>

      <div className="arch-arrow arch-arrow--highlight">↓</div>

      {/* Layer 4: Search Intelligence (centrepiece) */}
      <div className="arch-layer arch-layer--si">
        <span className="arch-layer-label">SEARCH INTELLIGENCE</span>
        <span className="arch-layer-detail">Observe → Learn → Predict → Explain → Simulate → Validate → Guide</span>
        <div className="arch-layer-items arch-layer-items--small">
          <span className="arch-chip arch-chip--si-mode">WorkerAssist</span>
          <span className="arch-chip arch-chip--si-mode">SearchAssist</span>
          <span className="arch-chip arch-chip--si-mode">PortfolioAssist</span>
        </div>
      </div>

      <div className="arch-arrow">↓</div>

      {/* Layer 5: Telemetry & Learning */}
      <div className="arch-layer arch-layer--telemetry">
        <span className="arch-layer-label">TELEMETRY &amp; LEARNING</span>
        <span className="arch-layer-detail">Discoveries · Worker Lifecycle · Learned Models · S3 Storage</span>
      </div>

      <div className="arch-arrow">↓</div>

      {/* Layer 6: Dashboard */}
      <div className="arch-layer arch-layer--dashboard">
        <span className="arch-layer-label">DASHBOARD</span>
        <div className="arch-layer-items arch-layer-items--small">
          <span className="arch-chip arch-chip--dash">Benchmarks</span>
          <span className="arch-chip arch-chip--dash">Statistics</span>
          <span className="arch-chip arch-chip--dash">Visualisations</span>
          <span className="arch-chip arch-chip--dash">Explain</span>
        </div>
      </div>

      <div className="arch-arrow">↓</div>

      {/* Layer 7: Research Evidence */}
      <div className="arch-layer arch-layer--evidence">
        <span className="arch-layer-label">RESEARCH EVIDENCE</span>
      </div>

      {/* Feedback loop annotation */}
      <div className="arch-feedback">
        <span className="arch-feedback-line" />
        <span className="arch-feedback-text">feedback loop</span>
      </div>
    </div>
  );
}
