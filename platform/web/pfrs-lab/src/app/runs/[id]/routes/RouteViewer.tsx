'use client';
import Card from '@/components/Card';

interface CVRPRoute {
  customers: number[];
  load: number;
  distance: number;
}

interface CVRPSolution {
  routes: CVRPRoute[];
  totalCost: number;
  vehicles: number;
  feasible: boolean;
}

const ROUTE_COLOURS = [
  '#60a5fa', '#34d399', '#fbbf24', '#f87171', '#a78bfa',
  '#fb923c', '#2dd4bf', '#e879f9', '#84cc16', '#f472b6',
  '#06b6d4', '#eab308', '#8b5cf6', '#14b8a6', '#ef4444',
];

export default function RouteViewer({ solution }: { solution: CVRPSolution }) {
  if (solution.routes.length === 0) return null;

  // Build a simple SVG visualisation of routes.
  // We don't have coordinates here (they're in the dataset, not the solution JSON).
  // Show a schematic: depot in centre, routes radiating out.
  const numRoutes = solution.routes.length;

  return (
    <Card title="Route Visualisation">
      <p className="text-xs text-gray-500 mb-3">
        Schematic view of routes. Each colour represents one vehicle. Numbers are customer IDs in visit order.
        Depot (D) is the start and end of every route.
      </p>
      <div className="space-y-2">
        {solution.routes.map((route, i) => {
          const colour = ROUTE_COLOURS[i % ROUTE_COLOURS.length];
          const utilisation = route.load > 0 ? Math.min(100, (route.load / (solution.totalCost > 0 ? route.load : 1)) * 100) : 0;
          return (
            <div key={i} className="flex items-center gap-2">
              <div className="w-6 h-6 rounded flex items-center justify-center text-[10px] font-bold text-white"
                style={{ backgroundColor: colour }}>
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
              <div className="text-[10px] text-gray-500 whitespace-nowrap">
                {route.distance} dist · {route.load} load
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
