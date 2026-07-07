'use client';

/**
 * PFRS Lab Platform Diagram — visual, article-quality.
 * Shows the platform flow with depth and colour.
 */
export default function ArchitectureDiagram() {
  return (
    <svg viewBox="0 0 700 320" className="w-full max-w-2xl mx-auto" xmlns="http://www.w3.org/2000/svg">
      {/* Background gradient */}
      <defs>
        <linearGradient id="flowGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#eff6ff" />
          <stop offset="100%" stopColor="#f8fafc" />
        </linearGradient>
        <filter id="shadow" x="-4%" y="-4%" width="108%" height="108%">
          <feDropShadow dx="0" dy="2" stdDeviation="3" floodOpacity="0.1" />
        </filter>
      </defs>
      <rect width="700" height="320" fill="url(#flowGrad)" rx="12" />

      {/* Row 1: Problems (NP-hard context) */}
      <text x="350" y="28" textAnchor="middle" fontSize="9" fill="#64748b" fontWeight="600">NP-HARD PROBLEMS</text>
      <g filter="url(#shadow)">
        <rect x="80" y="38" width="95" height="32" rx="6" fill="#3b82f6" />
        <text x="127" y="58" textAnchor="middle" fontSize="11" fill="white" fontWeight="600">NRP</text>
      </g>
      <g filter="url(#shadow)">
        <rect x="190" y="38" width="95" height="32" rx="6" fill="#10b981" />
        <text x="237" y="58" textAnchor="middle" fontSize="11" fill="white" fontWeight="600">CVRP</text>
      </g>
      <g filter="url(#shadow)">
        <rect x="300" y="38" width="95" height="32" rx="6" fill="#f59e0b" />
        <text x="347" y="58" textAnchor="middle" fontSize="11" fill="white" fontWeight="600">JSS</text>
      </g>
      <g filter="url(#shadow)">
        <rect x="410" y="38" width="95" height="32" rx="6" fill="#ef4444" />
        <text x="457" y="58" textAnchor="middle" fontSize="11" fill="white" fontWeight="600">VRPTW</text>
      </g>
      {/* Callout */}
      <text x="570" y="50" fontSize="8" fill="#94a3b8">Nurse Rostering</text>
      <text x="570" y="62" fontSize="8" fill="#94a3b8">Vehicle Routing</text>

      {/* Arrow */}
      <path d="M350 74 L350 88" stroke="#cbd5e1" strokeWidth="2" markerEnd="url(#arrowhead)" />
      <defs><marker id="arrowhead" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto"><path d="M0,0 L6,3 L0,6 Z" fill="#cbd5e1" /></marker></defs>

      {/* Row 2: Generic Interface */}
      <g filter="url(#shadow)">
        <rect x="140" y="92" width="420" height="30" rx="6" fill="#7c3aed" />
        <text x="350" y="111" textAnchor="middle" fontSize="10" fill="white">Generic Problem Interface — TryMove · Evaluate · Undo · Serialize</text>
      </g>

      {/* Arrow */}
      <path d="M350 126 L350 140" stroke="#cbd5e1" strokeWidth="2" markerEnd="url(#arrowhead)" />

      {/* Row 3: Algorithms */}
      <text x="350" y="152" textAnchor="middle" fontSize="9" fill="#64748b" fontWeight="600">SEARCH ALGORITHMS</text>
      <g>
        <rect x="95" y="158" width="70" height="26" rx="5" fill="#fef3c7" stroke="#f59e0b" strokeWidth="1" />
        <text x="130" y="175" textAnchor="middle" fontSize="10" fill="#92400e" fontWeight="500">SA</text>
      </g>
      <g>
        <rect x="175" y="158" width="70" height="26" rx="5" fill="#fef3c7" stroke="#f59e0b" strokeWidth="1" />
        <text x="210" y="175" textAnchor="middle" fontSize="10" fill="#92400e" fontWeight="500">LAHC</text>
      </g>
      <g>
        <rect x="255" y="158" width="70" height="26" rx="5" fill="#fef3c7" stroke="#f59e0b" strokeWidth="1" />
        <text x="290" y="175" textAnchor="middle" fontSize="10" fill="#92400e" fontWeight="500">Tabu</text>
      </g>
      <g>
        <rect x="335" y="158" width="80" height="26" rx="5" fill="#fef3c7" stroke="#f59e0b" strokeWidth="1" />
        <text x="375" y="175" textAnchor="middle" fontSize="10" fill="#92400e" fontWeight="500">Portfolio</text>
      </g>
      <g>
        <rect x="425" y="158" width="80" height="26" rx="5" fill="#fef3c7" stroke="#f59e0b" strokeWidth="1" />
        <text x="465" y="175" textAnchor="middle" fontSize="10" fill="#92400e" fontWeight="500">Adaptive</text>
      </g>

      {/* Arrow */}
      <path d="M350 188 L350 202" stroke="#cbd5e1" strokeWidth="2" markerEnd="url(#arrowhead)" />

      {/* Row 4: Search Intelligence (larger, prominent) */}
      <g filter="url(#shadow)">
        <rect x="100" y="206" width="500" height="44" rx="8" fill="#ecfdf5" stroke="#10b981" strokeWidth="1.5" />
        <text x="350" y="224" textAnchor="middle" fontSize="10" fill="#065f46" fontWeight="600">SEARCH INTELLIGENCE</text>
        <text x="350" y="240" textAnchor="middle" fontSize="9" fill="#047857">Observe → Learn → Predict → Explain → Simulate → Guide</text>
      </g>

      {/* Arrow */}
      <path d="M350 254 L350 268" stroke="#cbd5e1" strokeWidth="2" markerEnd="url(#arrowhead)" />

      {/* Row 5: Outputs */}
      <g filter="url(#shadow)">
        <rect x="100" y="272" width="230" height="30" rx="6" fill="#f1f5f9" stroke="#e2e8f0" strokeWidth="1" />
        <text x="215" y="291" textAnchor="middle" fontSize="9" fill="#475569">Telemetry → Learning → S3</text>
      </g>
      <g filter="url(#shadow)">
        <rect x="345" y="272" width="255" height="30" rx="6" fill="#dbeafe" stroke="#93c5fd" strokeWidth="1" />
        <text x="472" y="291" textAnchor="middle" fontSize="9" fill="#1e40af">Dashboard · Statistics · Validation</text>
      </g>
    </svg>
  );
}
