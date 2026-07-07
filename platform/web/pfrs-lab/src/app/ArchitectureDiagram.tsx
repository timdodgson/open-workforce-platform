'use client';

/**
 * PFRS Lab v3 Architecture — compact CSS-based diagram.
 * No SVG. Uses flexbox for a clean, responsive layout.
 */
export default function ArchitectureDiagram() {
  return (
    <div className="space-y-2 py-2">
      <Layer label="Problems" colour="blue" items={['NRP', 'CVRP', 'JSS', 'VRPTW']} />
      <FlowArrow />
      <Layer label="Interface" colour="purple" items={['TryMove · Evaluate · Undo · Serialize']} wide />
      <FlowArrow />
      <Layer label="Algorithms" colour="amber" items={['SA', 'LAHC', 'Tabu', 'Portfolio', 'Adaptive']} />
      <FlowArrow />
      <Layer label="Search Intelligence" colour="emerald" items={['Off', 'Shadow', 'Assist', 'Adaptive']} />
      <FlowArrow />
      <Layer label="Telemetry → Learning → Storage" colour="gray" items={['run.json · models · S3']} wide />
      <FlowArrow />
      <Layer label="Dashboard → Research Evidence" colour="blue" items={['Benchmarks · Statistics · Validation']} wide />
    </div>
  );
}

function Layer({ label, colour, items, wide }: { label: string; colour: string; items: string[]; wide?: boolean }) {
  const colours: Record<string, { border: string; bg: string; text: string }> = {
    blue: { border: 'border-blue-700', bg: 'bg-blue-900/30', text: 'text-blue-400' },
    purple: { border: 'border-purple-700', bg: 'bg-purple-900/30', text: 'text-purple-400' },
    amber: { border: 'border-amber-700', bg: 'bg-amber-900/30', text: 'text-amber-400' },
    emerald: { border: 'border-emerald-700', bg: 'bg-emerald-900/30', text: 'text-emerald-400' },
    gray: { border: 'border-gray-700', bg: 'bg-gray-800/50', text: 'text-gray-400' },
  };
  const c = colours[colour] || colours.gray;

  return (
    <div className="text-center">
      <p className={`text-[9px] uppercase tracking-wider font-semibold ${c.text} mb-1`}>{label}</p>
      <div className={`flex flex-wrap justify-center gap-1.5 ${wide ? '' : ''}`}>
        {items.map(item => (
          <span key={item} className={`text-[10px] px-2.5 py-1 rounded border ${c.border} ${c.bg} text-gray-200`}>
            {item}
          </span>
        ))}
      </div>
    </div>
  );
}

function FlowArrow() {
  return <div className="text-center text-gray-600 text-[10px]">↓</div>;
}
