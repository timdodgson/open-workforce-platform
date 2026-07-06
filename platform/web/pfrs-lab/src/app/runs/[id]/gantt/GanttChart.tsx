'use client';

interface ScheduledOp {
  JobID: number;
  OpIndex: number;
  Machine: number;
  Start: number;
  End: number;
  Duration: number;
}

interface Props {
  operations: ScheduledOp[];
  jobs: number;
  machines: number;
  makespan: number;
}

// Colour palette for jobs (distinct, colourblind-friendly).
const JOB_COLOURS = [
  '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
  '#06b6d4', '#ec4899', '#84cc16', '#f97316', '#6366f1',
  '#14b8a6', '#e11d48', '#a3e635', '#0ea5e9', '#d946ef',
  '#fbbf24', '#2dd4bf', '#f43f5e', '#4ade80', '#a78bfa',
];

export default function GanttChart({ operations, jobs, machines, makespan }: Props) {
  const rowHeight = 36;
  const headerWidth = 80;
  const chartWidth = 800;
  const chartHeight = machines * rowHeight;
  const totalHeight = chartHeight + 40; // extra for time axis
  const totalWidth = headerWidth + chartWidth;

  // Group by machine.
  const byMachine: ScheduledOp[][] = Array.from({ length: machines }, () => []);
  for (const op of operations) {
    if (op.Machine >= 0 && op.Machine < machines) {
      byMachine[op.Machine].push(op);
    }
  }

  // Time scale.
  const timeScale = (t: number) => (t / makespan) * chartWidth;

  // Generate time axis ticks.
  const tickCount = Math.min(10, makespan);
  const tickInterval = Math.ceil(makespan / tickCount);
  const ticks: number[] = [];
  for (let t = 0; t <= makespan; t += tickInterval) {
    ticks.push(t);
  }
  if (ticks[ticks.length - 1] !== makespan) ticks.push(makespan);

  return (
    <div className="overflow-x-auto">
      <svg
        viewBox={`0 0 ${totalWidth} ${totalHeight}`}
        className="w-full"
        style={{ minHeight: `${Math.max(totalHeight * 0.6, 200)}px`, fontFamily: 'Inter, system-ui, sans-serif' }}
      >
        {/* Background */}
        <rect width={totalWidth} height={totalHeight} fill="#111827" rx={4} />

        {/* Machine labels */}
        {Array.from({ length: machines }, (_, m) => (
          <text
            key={`label-${m}`}
            x={headerWidth - 8}
            y={m * rowHeight + rowHeight / 2 + 4}
            textAnchor="end"
            fontSize="10"
            fill="#9ca3af"
          >
            M{m}
          </text>
        ))}

        {/* Grid lines */}
        {Array.from({ length: machines + 1 }, (_, m) => (
          <line
            key={`grid-${m}`}
            x1={headerWidth}
            y1={m * rowHeight}
            x2={totalWidth}
            y2={m * rowHeight}
            stroke="#374151"
            strokeWidth={0.5}
          />
        ))}

        {/* Time axis ticks */}
        {ticks.map(t => (
          <g key={`tick-${t}`}>
            <line
              x1={headerWidth + timeScale(t)}
              y1={0}
              x2={headerWidth + timeScale(t)}
              y2={chartHeight}
              stroke="#374151"
              strokeWidth={0.5}
              strokeDasharray="2,2"
            />
            <text
              x={headerWidth + timeScale(t)}
              y={chartHeight + 14}
              textAnchor="middle"
              fontSize="8"
              fill="#6b7280"
            >
              {t}
            </text>
          </g>
        ))}

        {/* Operations */}
        {byMachine.map((machineOps, m) =>
          machineOps.map((op, i) => {
            const x = headerWidth + timeScale(op.Start);
            const w = timeScale(op.Duration);
            const y = m * rowHeight + 4;
            const h = rowHeight - 8;
            const colour = JOB_COLOURS[op.JobID % JOB_COLOURS.length];

            return (
              <g key={`op-${m}-${i}`}>
                <rect
                  x={x}
                  y={y}
                  width={Math.max(w, 1)}
                  height={h}
                  fill={colour}
                  rx={3}
                  opacity={0.85}
                >
                  <title>
                    Job {op.JobID} Op {op.OpIndex} | Machine {op.Machine} | [{op.Start}–{op.End}] ({op.Duration})
                  </title>
                </rect>
                {w > 20 && (
                  <text
                    x={x + w / 2}
                    y={y + h / 2 + 3}
                    textAnchor="middle"
                    fontSize="8"
                    fill="white"
                    fontWeight="bold"
                  >
                    J{op.JobID}
                  </text>
                )}
              </g>
            );
          })
        )}

        {/* Time axis label */}
        <text
          x={headerWidth + chartWidth / 2}
          y={chartHeight + 30}
          textAnchor="middle"
          fontSize="9"
          fill="#6b7280"
        >
          Time → (makespan: {makespan})
        </text>
      </svg>

      {/* Legend */}
      <div className="flex flex-wrap gap-2 mt-3">
        {Array.from({ length: jobs }, (_, j) => (
          <span key={j} className="flex items-center gap-1 text-[9px] text-gray-400">
            <span
              className="w-3 h-3 rounded-sm"
              style={{ background: JOB_COLOURS[j % JOB_COLOURS.length] }}
            />
            Job {j}
          </span>
        ))}
      </div>
    </div>
  );
}
