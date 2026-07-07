'use client';

/**
 * PFRS Lab v3 Architecture Diagram.
 * Clean vertical flow. One page. Professional.
 */
export default function ArchitectureDiagram() {
  return (
    <div className="overflow-x-auto">
      <svg
        viewBox="0 0 600 680"
        className="w-full max-w-xl mx-auto"
        style={{ minWidth: '400px' }}
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect width="600" height="680" fill="transparent" />

        {/* Layer 1: Problems */}
        <Layer y={0} label="Problems" colour="#60a5fa" />
        <Row y={24} items={['NRP', 'CVRP', 'JSS', 'VRPTW']} colour="#3b82f6" bg="#1e3a5f" />
        <Arrow y={68} />

        {/* Layer 2: Problem Interface */}
        <Layer y={82} label="Generic Problem Interface" colour="#a78bfa" />
        <WideBox y={106} text="CreateInitialSolution · TryMove · Evaluate · Undo · Serialize" colour="#7c3aed" bg="#2d2248" />
        <Arrow y={148} />

        {/* Layer 3: Algorithms */}
        <Layer y={162} label="Algorithms" colour="#fbbf24" />
        <Row y={186} items={['SA', 'LAHC', 'Tabu', 'Portfolio', 'Adaptive']} colour="#f59e0b" bg="#3d2e1a" />
        <Arrow y={230} />

        {/* Layer 4: Search Intelligence */}
        <Layer y={244} label="Search Intelligence" colour="#34d399" />
        <Row y={268} items={['Off', 'Shadow', 'Assist', 'Adaptive']} colour="#10b981" bg="#1a3d2e" />
        <SubRow y={304} items={['WorkerAssist', 'SearchAssist', 'PortfolioAssist', 'Learned Model']} />
        <Arrow y={336} />

        {/* Layer 5: Telemetry */}
        <Layer y={350} label="Telemetry" colour="#60a5fa" />
        <WideBox y={374} text="run.json · discoveries.csv · worker_learning.csv · portfolio_assist.csv" colour="#334155" bg="#1e293b" />
        <Arrow y={416} />

        {/* Layer 6: Learning */}
        <Layer y={430} label="Learning" colour="#a78bfa" />
        <WideBox y={454} text="worker_model.json · portfolio_budget_model.json · Feature Importance" colour="#7c3aed" bg="#2d2248" />
        <Arrow y={496} />

        {/* Layer 7: Storage */}
        <Layer y={510} label="Storage" colour="#60a5fa" />
        <WideBox y={534} text="Local Filesystem · S3 (versioned) · Manifest Index" colour="#334155" bg="#1e293b" />
        <Arrow y={576} />

        {/* Layer 8: Dashboard */}
        <Layer y={590} label="Dashboard" colour="#60a5fa" />
        <WideBox y={614} text="Benchmarks · Statistics · Search Intelligence · Route Viewer · Gantt" colour="#3b82f6" bg="#1e3a5f" />
        <Arrow y={656} />

        {/* Layer 9: Research Outputs */}
        <text x={300} y={676} textAnchor="middle" fill="#34d399" fontSize={11} fontWeight="bold" fontFamily="system-ui, sans-serif">
          Research Outputs — Validation · Gap Analysis · Statistical Evidence
        </text>
      </svg>
    </div>
  );
}

function Layer({ y, label, colour }: { y: number; label: string; colour: string }) {
  return (
    <text x={300} y={y + 12} textAnchor="middle" fill={colour} fontSize={10} fontWeight="bold" fontFamily="system-ui, sans-serif">
      {label.toUpperCase()}
    </text>
  );
}

function Row({ y, items, colour, bg }: { y: number; items: string[]; colour: string; bg: string }) {
  const count = items.length;
  const boxW = 100;
  const gap = 12;
  const totalW = count * boxW + (count - 1) * gap;
  const startX = (600 - totalW) / 2;

  return (
    <g>
      {items.map((item, i) => {
        const x = startX + i * (boxW + gap);
        return (
          <g key={item}>
            <rect x={x} y={y} width={boxW} height={32} rx={4} fill={bg} stroke={colour} strokeWidth={1} />
            <text x={x + boxW / 2} y={y + 20} textAnchor="middle" fill="#e5e7eb" fontSize={11} fontFamily="system-ui, sans-serif">
              {item}
            </text>
          </g>
        );
      })}
    </g>
  );
}

function SubRow({ y, items }: { y: number; items: string[] }) {
  const count = items.length;
  const boxW = 110;
  const gap = 8;
  const totalW = count * boxW + (count - 1) * gap;
  const startX = (600 - totalW) / 2;

  return (
    <g>
      {items.map((item, i) => {
        const x = startX + i * (boxW + gap);
        return (
          <g key={item}>
            <rect x={x} y={y} width={boxW} height={22} rx={3} fill="#1f2937" stroke="#4b5563" strokeWidth={0.5} />
            <text x={x + boxW / 2} y={y + 14} textAnchor="middle" fill="#9ca3af" fontSize={9} fontFamily="system-ui, sans-serif">
              {item}
            </text>
          </g>
        );
      })}
    </g>
  );
}

function WideBox({ y, text, colour, bg }: { y: number; text: string; colour: string; bg: string }) {
  return (
    <g>
      <rect x={50} y={y} width={500} height={32} rx={4} fill={bg} stroke={colour} strokeWidth={1} />
      <text x={300} y={y + 20} textAnchor="middle" fill="#d1d5db" fontSize={9.5} fontFamily="system-ui, sans-serif">
        {text}
      </text>
    </g>
  );
}

function Arrow({ y }: { y: number }) {
  return (
    <g>
      <line x1={300} y1={y} x2={300} y2={y + 10} stroke="#4b5563" strokeWidth={1.5} />
      <polygon points={`296,${y + 8} 304,${y + 8} 300,${y + 14}`} fill="#4b5563" />
    </g>
  );
}
