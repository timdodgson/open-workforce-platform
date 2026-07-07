'use client';

/**
 * PFRS Lab Architecture Diagram — visually striking, landing-page quality.
 * 
 * Layered flow from Problem Domains through to Research Evidence,
 * with colour-coded layers, gradients, depth, and generous whitespace.
 */
export default function ArchitectureDiagram() {
  return (
    <svg
      viewBox="0 0 900 520"
      className="w-full"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label="PFRS Lab platform architecture diagram showing flow from problem domains through algorithms, search intelligence, telemetry, and dashboard to research evidence"
    >
      <defs>
        {/* Background gradient */}
        <linearGradient id="archBg" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#f8fafc" />
          <stop offset="50%" stopColor="#f1f5f9" />
          <stop offset="100%" stopColor="#e2e8f0" />
        </linearGradient>

        {/* Layer gradients */}
        <linearGradient id="domainGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#3b82f6" />
          <stop offset="100%" stopColor="#6366f1" />
        </linearGradient>
        <linearGradient id="interfaceGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#7c3aed" />
          <stop offset="100%" stopColor="#a855f7" />
        </linearGradient>
        <linearGradient id="algoGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#f59e0b" />
          <stop offset="100%" stopColor="#f97316" />
        </linearGradient>
        <linearGradient id="siGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#059669" />
          <stop offset="100%" stopColor="#10b981" />
        </linearGradient>
        <linearGradient id="telemetryGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#0ea5e9" />
          <stop offset="100%" stopColor="#38bdf8" />
        </linearGradient>
        <linearGradient id="dashGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#6366f1" />
          <stop offset="100%" stopColor="#818cf8" />
        </linearGradient>
        <linearGradient id="evidenceGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#0f172a" />
          <stop offset="100%" stopColor="#1e293b" />
        </linearGradient>

        {/* Drop shadow */}
        <filter id="layerShadow" x="-2%" y="-5%" width="104%" height="115%">
          <feDropShadow dx="0" dy="3" stdDeviation="4" floodColor="#000" floodOpacity="0.08" />
        </filter>
        <filter id="glowSI" x="-5%" y="-15%" width="110%" height="130%">
          <feDropShadow dx="0" dy="2" stdDeviation="6" floodColor="#10b981" floodOpacity="0.25" />
        </filter>

        {/* Arrow marker */}
        <marker id="archArrow" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">
          <path d="M1,1 L7,4 L1,7" fill="none" stroke="#94a3b8" strokeWidth="1.5" strokeLinecap="round" />
        </marker>
        <marker id="archArrowGreen" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto">
          <path d="M1,1 L7,4 L1,7" fill="none" stroke="#10b981" strokeWidth="1.5" strokeLinecap="round" />
        </marker>
      </defs>

      {/* Background */}
      <rect width="900" height="520" fill="url(#archBg)" rx="16" />

      {/* ─── Layer 1: Problem Domains ─── */}
      <g filter="url(#layerShadow)">
        <rect x="100" y="24" width="700" height="50" rx="10" fill="url(#domainGrad)" />
        <text x="450" y="42" textAnchor="middle" fontSize="9" fill="rgba(255,255,255,0.7)" fontWeight="600" letterSpacing="1">PROBLEM DOMAINS</text>
        <text x="200" y="62" textAnchor="middle" fontSize="13" fill="white" fontWeight="600">NRP</text>
        <text x="350" y="62" textAnchor="middle" fontSize="13" fill="white" fontWeight="600">CVRP</text>
        <text x="500" y="62" textAnchor="middle" fontSize="13" fill="white" fontWeight="600">JSS</text>
        <text x="650" y="62" textAnchor="middle" fontSize="13" fill="white" fontWeight="600">VRPTW</text>
      </g>

      {/* Arrow 1→2 */}
      <path d="M450 78 L450 100" stroke="#94a3b8" strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#archArrow)" />

      {/* ─── Layer 2: Generic Problem Interface ─── */}
      <g filter="url(#layerShadow)">
        <rect x="160" y="104" width="580" height="42" rx="8" fill="url(#interfaceGrad)" />
        <text x="450" y="122" textAnchor="middle" fontSize="9" fill="rgba(255,255,255,0.7)" fontWeight="600" letterSpacing="1">GENERIC PROBLEM INTERFACE</text>
        <text x="450" y="138" textAnchor="middle" fontSize="10" fill="rgba(255,255,255,0.9)">TryMove · Evaluate · Undo · Constraints · Serialize</text>
      </g>

      {/* Arrow 2→3 */}
      <path d="M450 150 L450 172" stroke="#94a3b8" strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#archArrow)" />

      {/* ─── Layer 3: Algorithms ─── */}
      <g filter="url(#layerShadow)">
        <rect x="100" y="176" width="700" height="50" rx="10" fill="url(#algoGrad)" />
        <text x="450" y="194" textAnchor="middle" fontSize="9" fill="rgba(255,255,255,0.7)" fontWeight="600" letterSpacing="1">SEARCH ALGORITHMS</text>
        <text x="170" y="215" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">SA</text>
        <text x="300" y="215" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">LAHC</text>
        <text x="450" y="215" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">Tabu</text>
        <text x="590" y="215" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">Portfolio</text>
        <text x="730" y="215" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">Adaptive</text>
      </g>

      {/* Arrow 3→4 */}
      <path d="M450 230 L450 252" stroke="#10b981" strokeWidth="2" markerEnd="url(#archArrowGreen)" />

      {/* ─── Layer 4: Search Intelligence (centrepiece) ─── */}
      <g filter="url(#glowSI)">
        <rect x="80" y="256" width="740" height="64" rx="12" fill="url(#siGrad)" />
        <text x="450" y="278" textAnchor="middle" fontSize="10" fill="rgba(255,255,255,0.8)" fontWeight="700" letterSpacing="1.5">SEARCH INTELLIGENCE</text>
        <text x="450" y="300" textAnchor="middle" fontSize="11" fill="white" fontWeight="400">Observe → Learn → Predict → Explain → Simulate → Validate → Guide</text>
      </g>

      {/* Arrow 4→5 */}
      <path d="M450 324 L450 346" stroke="#94a3b8" strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#archArrow)" />

      {/* ─── Layer 5: Telemetry + Learning ─── */}
      <g filter="url(#layerShadow)">
        <rect x="140" y="350" width="620" height="42" rx="8" fill="url(#telemetryGrad)" />
        <text x="450" y="368" textAnchor="middle" fontSize="9" fill="rgba(255,255,255,0.7)" fontWeight="600" letterSpacing="1">TELEMETRY &amp; LEARNING</text>
        <text x="450" y="383" textAnchor="middle" fontSize="10" fill="rgba(255,255,255,0.9)">Metrics · Search History · Models · S3 Storage</text>
      </g>

      {/* Arrow 5→6 */}
      <path d="M450 396 L450 418" stroke="#94a3b8" strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#archArrow)" />

      {/* ─── Layer 6: Dashboard ─── */}
      <g filter="url(#layerShadow)">
        <rect x="180" y="422" width="540" height="38" rx="8" fill="url(#dashGrad)" />
        <text x="450" y="446" textAnchor="middle" fontSize="11" fill="white" fontWeight="500">Dashboard · Statistics · Benchmarks · Visualisations</text>
      </g>

      {/* Arrow 6→7 */}
      <path d="M450 464 L450 480" stroke="#94a3b8" strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#archArrow)" />

      {/* ─── Layer 7: Research Evidence ─── */}
      <g filter="url(#layerShadow)">
        <rect x="240" y="484" width="420" height="28" rx="6" fill="url(#evidenceGrad)" />
        <text x="450" y="503" textAnchor="middle" fontSize="10" fill="white" fontWeight="600" letterSpacing="0.5">Research Evidence</text>
      </g>

      {/* Feedback loop arrow (SI learns from telemetry) */}
      <path d="M830 310 C860 310, 860 375, 830 375" stroke="#10b981" strokeWidth="1.5" fill="none" strokeDasharray="3 2" opacity="0.6" />
      <text x="870" y="345" fontSize="8" fill="#10b981" opacity="0.7">feedback</text>
    </svg>
  );
}
