'use client';

/**
 * PFRS Lab v3.0 Architecture Diagram.
 * Renders as inline SVG with dark theme styling.
 * Shows the full platform flow in one glance.
 */
export default function ArchitectureDiagram() {
  return (
    <div className="overflow-x-auto">
      <svg
        viewBox="0 0 900 720"
        className="w-full max-w-4xl mx-auto"
        style={{ minWidth: '600px' }}
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Background */}
        <rect width="900" height="720" fill="#111827" rx="8" />

        {/* --- Layer 1: Problem Domains --- */}
        <Layer y={20} label="Problem Domains" labelColour="#60a5fa" />
        <Box x={100} y={50} w={150} h={36} text="NRP / INRC-II" fill="#1e3a5f" stroke="#3b82f6" />
        <Box x={270} y={50} w={150} h={36} text="CVRP / CVRPLIB" fill="#1a3d2e" stroke="#10b981" />
        <Box x={440} y={50} w={150} h={36} text="JSS / OR-Library" fill="#3d2e1a" stroke="#f59e0b" />
        <Box x={610} y={50} w={170} h={36} text="VRPTW / Solomon" fill="#3d1a2e" stroke="#f43f5e" />

        {/* Arrow down */}
        <Arrow x={450} y1={90} y2={110} />

        {/* --- Layer 2: Problem Interface --- */}
        <Layer y={110} label="Generic Problem Interface" labelColour="#a78bfa" />
        <Box x={200} y={140} w={500} h={36} text="Initial Solution · Move Generation · Evaluation · Validation · Serialisation" fill="#2d2248" stroke="#7c3aed" />

        {/* Arrow down */}
        <Arrow x={450} y1={180} y2={200} />

        {/* --- Layer 3: Search Algorithms --- */}
        <Layer y={200} label="Search Algorithms" labelColour="#fbbf24" />
        <Box x={150} y={230} w={100} h={36} text="SA" fill="#3d2e1a" stroke="#f59e0b" />
        <Box x={270} y={230} w={100} h={36} text="LAHC" fill="#3d2e1a" stroke="#f59e0b" />
        <Box x={390} y={230} w={100} h={36} text="Tabu" fill="#3d2e1a" stroke="#f59e0b" />
        <Box x={510} y={230} w={120} h={36} text="Portfolio" fill="#3d2e1a" stroke="#f59e0b" />
        <Box x={650} y={230} w={120} h={36} text="Adaptive" fill="#3d2e1a" stroke="#f59e0b" />

        {/* Arrow down */}
        <Arrow x={450} y1={270} y2={290} />

        {/* --- Layer 4: Search Intelligence --- */}
        <Layer y={290} label="Search Intelligence" labelColour="#34d399" />
        <Box x={100} y={320} w={90} h={32} text="Off" fill="#1f2937" stroke="#4b5563" textSize={10} />
        <Box x={205} y={320} w={100} h={32} text="Shadow" fill="#1e3a5f" stroke="#3b82f6" textSize={10} />
        <Box x={320} y={320} w={100} h={32} text="Assist" fill="#1a3d2e" stroke="#10b981" textSize={10} />
        <Box x={435} y={320} w={110} h={32} text="Adaptive" fill="#3d2e1a" stroke="#f59e0b" textSize={10} />
        {/* Subtypes */}
        <Box x={580} y={316} w={120} h={18} text="WorkerAssist" fill="#1f2937" stroke="#6b7280" textSize={9} />
        <Box x={580} y={338} w={120} h={18} text="SearchAssist" fill="#1f2937" stroke="#6b7280" textSize={9} />
        <Box x={715} y={316} w={130} h={18} text="PortfolioAssist" fill="#1f2937" stroke="#6b7280" textSize={9} />
        <Box x={715} y={338} w={130} h={18} text="Learned Model" fill="#1f2937" stroke="#6b7280" textSize={9} />

        {/* Arrow down */}
        <Arrow x={450} y1={358} y2={378} />

        {/* --- Layer 5: Telemetry --- */}
        <Layer y={378} label="Telemetry Layer" labelColour="#60a5fa" />
        <Box x={100} y={405} w={700} h={32} text="results.csv · discoveries.csv · worker_learning.csv · portfolio_assist.csv · adaptive_assist.csv · run.json" fill="#1e293b" stroke="#334155" textSize={10} />

        {/* Arrow down */}
        <Arrow x={450} y1={442} y2={460} />

        {/* --- Layer 6: Learning --- */}
        <Layer y={460} label="Learning Layer" labelColour="#a78bfa" />
        <Box x={130} y={488} w={200} h={32} text="worker_model.json" fill="#2d2248" stroke="#7c3aed" textSize={10} />
        <Box x={350} y={488} w={220} h={32} text="portfolio_budget_model.json" fill="#2d2248" stroke="#7c3aed" textSize={10} />
        <Box x={590} y={488} w={200} h={32} text="Feature Importance · What-If" fill="#2d2248" stroke="#7c3aed" textSize={10} />

        {/* Arrow down */}
        <Arrow x={450} y1={525} y2={543} />

        {/* --- Layer 7: Storage + Dashboard --- */}
        <Layer y={543} label="Storage & Dashboard" labelColour="#60a5fa" />
        <Box x={100} y={570} w={180} h={32} text="Local / S3 / Versioned" fill="#1e293b" stroke="#334155" textSize={10} />
        <Box x={300} y={570} w={500} h={32} text="Benchmark Ladder · Statistics · Assist · Learning · Predictions · What-If · Route Viewer · Gantt" fill="#1e3a5f" stroke="#3b82f6" textSize={9} />

        {/* Arrow down */}
        <Arrow x={450} y1={607} y2={625} />

        {/* --- Layer 8: Research Outputs --- */}
        <Layer y={625} label="Research Outputs" labelColour="#34d399" />
        <Box x={120} y={655} w={660} h={36} text="Validation Reports · Optimality Gap · Statistical Significance · Failure Analysis · Release Evidence" fill="#1a3d2e" stroke="#10b981" textSize={10} />
      </svg>
    </div>
  );
}

// --- Primitives ---

function Layer({ y, label, labelColour }: { y: number; label: string; labelColour: string }) {
  return (
    <text x={450} y={y + 14} textAnchor="middle" fill={labelColour} fontSize={11} fontWeight="bold" fontFamily="system-ui, sans-serif">
      {label}
    </text>
  );
}

function Box({ x, y, w, h, text, fill, stroke, textSize = 11 }: {
  x: number; y: number; w: number; h: number; text: string; fill: string; stroke: string; textSize?: number;
}) {
  return (
    <g>
      <rect x={x} y={y} width={w} height={h} rx={4} fill={fill} stroke={stroke} strokeWidth={1} />
      <text x={x + w / 2} y={y + h / 2 + 4} textAnchor="middle" fill="#e5e7eb" fontSize={textSize} fontFamily="system-ui, sans-serif">
        {text}
      </text>
    </g>
  );
}

function Arrow({ x, y1, y2 }: { x: number; y1: number; y2: number }) {
  return (
    <g>
      <line x1={x} y1={y1} x2={x} y2={y2 - 4} stroke="#4b5563" strokeWidth={1.5} />
      <polygon points={`${x - 4},${y2 - 6} ${x + 4},${y2 - 6} ${x},${y2}`} fill="#4b5563" />
    </g>
  );
}
