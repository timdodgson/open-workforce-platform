'use client';
import Card from '@/components/Card';
import type { RoutingProblemType, RoutingSolution } from '@/lib/solution-types';

const ROUTE_COLOURS = [
  '#60a5fa', '#34d399', '#fbbf24', '#f87171', '#a78bfa',
  '#fb923c', '#2dd4bf', '#e879f9', '#84cc16', '#f472b6',
  '#06b6d4', '#eab308', '#8b5cf6', '#14b8a6', '#ef4444',
];

interface Props {
  solution: RoutingSolution;
  problemType: RoutingProblemType;
  capacity?: number;
}

export default function RouteViewer({ solution, problemType, capacity = 0 }: Props) {
  if (solution.routes.length === 0) return null;

  const isVRPTW = problemType === 'vrptw';

  return (
    <Card title="Route Visualisation">
      <p className="text-xs text-gray-500 mb-3">
        {isVRPTW
          ? 'VRPTW routes with capacity and per-route feasibility. Red route badge = capacity or time-window violation.'
          : 'CVRP routes by vehicle. Depot (D) is the start and end of every route.'}
      </p>
      <div className="space-y-2">
        {solution.routes.map((route, i) => {
          const colour = ROUTE_COLOURS[i % ROUTE_COLOURS.length];
          const utilisation = capacity > 0 ? (route.load / capacity) * 100 : 0;
          const overloaded = capacity > 0 && route.load > capacity;
          const routeBad = isVRPTW && route.feasible === false;
          return (
            <div key={i} className="flex items-center gap-2">
              <div
                className={`w-6 h-6 rounded flex items-center justify-center text-[10px] font-bold text-white ${routeBad || overloaded ? 'ring-2 ring-red-500' : ''}`}
                style={{ backgroundColor: colour }}
                title={routeBad ? 'Route infeasible (capacity or TW)' : undefined}
              >
                {i + 1}
              </div>
              <div className="flex-1 flex items-center gap-1 flex-wrap">
                <span className="text-[10px] text-gray-500 font-mono bg-gray-800 px-1 rounded">D</span>
                <span className="text-gray-600">→</span>
                {route.customers.map((c, j) => (
                  <span key={j} className="flex items-center gap-1">
                    <span className="text-[10px] font-mono px-1 rounded" style={{ backgroundColor: colour + '30', color: colour }}>
                      {c}
                    </span>
                    {j < route.customers.length - 1 && <span className="text-gray-700">→</span>}
                  </span>
                ))}
                <span className="text-gray-600">→</span>
                <span className="text-[10px] text-gray-500 font-mono bg-gray-800 px-1 rounded">D</span>
              </div>
              <div className="text-[10px] text-gray-500 whitespace-nowrap text-right">
                <div>{route.distance} dist · {route.load} load{capacity > 0 ? ` (${utilisation.toFixed(0)}%)` : ''}</div>
                {isVRPTW && (
                  <div className={routeBad ? 'text-red-400' : 'text-emerald-400'}>
                    {route.feasible === false ? 'infeasible' : 'feasible'}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
