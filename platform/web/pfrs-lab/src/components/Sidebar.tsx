'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import type { RunMode } from '@/features/runs/run-mode';

interface PageItem {
  path: string;
  label: string;
  icon: string;
}

interface PageGroup {
  title: string;
  items: PageItem[];
}

const PFRS_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'schedule', label: 'Schedule', icon: '📅' },
      { path: 'report', label: 'Report', icon: '📄' },
    ],
  },
  {
    title: 'Search Analysis',
    items: [
      { path: 'search', label: 'Search Progress', icon: '📈' },
      { path: 'timeline', label: 'Timeline', icon: '⏱️' },
      { path: 'map', label: 'Search Map', icon: '🗺️' },
      { path: 'replay', label: 'Replay', icon: '🎬' },
      { path: 'landscape', label: 'Landscape', icon: '🏔️' },
    ],
  },
  {
    title: 'Algorithm',
    items: [
      { path: 'temperature', label: 'Temperature', icon: '🌡️' },
      { path: 'plateaus', label: 'Plateaus', icon: '🏔️' },
      { path: 'workers', label: 'Workers', icon: '👷' },
      { path: 'efficiency', label: 'Efficiency', icon: '⚡' },
      { path: 'causality', label: 'Causality', icon: '🔗' },
    ],
  },
  {
    title: 'Beam Search',
    items: [
      { path: 'tree', label: 'Search Tree', icon: '🌳' },
      { path: 'genealogy', label: 'Genealogy', icon: '🌲' },
      { path: 'families', label: 'Families', icon: '👪' },
      { path: 'inheritance', label: 'Inheritance', icon: '🧬' },
      { path: 'diversity', label: 'Diversity', icon: '🌍' },
      { path: 'pathdiff', label: 'Path Diff', icon: '🔃' },
    ],
  },
  {
    title: 'Diagnostics',
    items: [
      { path: 'dna', label: 'Search DNA', icon: '🧪' },
      { path: 'archetypes', label: 'Archetypes', icon: '🎭' },
      { path: 'constraints', label: 'Constraints', icon: '⚖️' },
      { path: 'waterfall', label: 'Waterfall', icon: '🌊' },
      { path: 'insights', label: 'Insights', icon: '💡' },
      { path: 'explain', label: 'Explain', icon: '🧠' },
    ],
  },
  {
    title: 'Export',
    items: [
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const CVRP_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'routes', label: 'Route Viewer', icon: '🚛' },
      { path: 'constraints', label: 'Constraints', icon: '⚖️' },
    ],
  },
  {
    title: 'Search Analysis',
    items: [
      { path: 'search', label: 'Search Progress', icon: '📈' },
      { path: 'timeline', label: 'Timeline', icon: '⏱️' },
      { path: 'map', label: 'Search Map', icon: '🗺️' },
    ],
  },
  {
    title: 'Algorithm',
    items: [
      { path: 'workers', label: 'Workers', icon: '👷' },
      { path: 'dna', label: 'Search DNA', icon: '🧪' },
    ],
  },
  {
    title: 'Export',
    items: [
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const ILP_GROUPS: PageGroup[] = [
  {
    title: 'Results',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'baseline', label: 'Solve Progress', icon: '📐' },
      { path: 'schedule', label: 'Schedule', icon: '📅' },
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const JSS_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'gantt', label: 'Gantt Chart', icon: '📊' },
      { path: 'constraints', label: 'Constraints', icon: '⚖️' },
    ],
  },
  {
    title: 'Search Analysis',
    items: [
      { path: 'search', label: 'Search Progress', icon: '📈' },
      { path: 'timeline', label: 'Timeline', icon: '⏱️' },
    ],
  },
  {
    title: 'Export',
    items: [
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const VRPTW_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'routes', label: 'Route Viewer', icon: '🚛' },
      { path: 'constraints', label: 'Constraints', icon: '⚖️' },
    ],
  },
  {
    title: 'Search Analysis',
    items: [
      { path: 'search', label: 'Search Progress', icon: '📈' },
      { path: 'timeline', label: 'Timeline', icon: '⏱️' },
    ],
  },
  {
    title: 'Export',
    items: [
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const TSP_GROUPS: PageGroup[] = [
  {
    title: 'BYOD run',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'explain', label: 'Explain', icon: '💬' },
      { path: 'export', label: 'Export', icon: '📤' },
    ],
  },
];

const GLOBAL_ITEMS = [
  { href: '/lab', label: 'Lab Home', icon: '🏠' },
  { href: '/getting-started', label: 'Getting Started', icon: '🚀' },
  { href: '/about', label: 'About', icon: '👤' },
  { href: '/lab/byod', label: 'BYOD / BYOA', icon: '🔌' },
  { href: '/runs', label: 'All Runs', icon: '📂' },
  { href: '/benchmarks', label: 'Benchmarks', icon: '🏆' },
  { href: '/capabilities', label: 'Capabilities', icon: '✅' },
  { href: '/experiment-matrix', label: 'Experiment Matrix', icon: '🧪' },
  { href: '/statistics', label: 'Statistics', icon: '📊' },
  { href: '/compare', label: 'Compare', icon: '🔀' },
  { href: '/trends', label: 'Trends', icon: '📈' },
  { href: '/intelligence', label: 'Search Intelligence', icon: '🧠' },
  { href: '/knowledge', label: 'Knowledge Base', icon: '📚' },
  { href: '/experiments/chat', label: 'Assistant', icon: '🤖' },
  { href: '/admin', label: 'Admin', icon: '⚙️', adminOnly: true },
  { href: '/admin/ml-journey', label: 'ML Journey', icon: '📈', adminOnly: true },
] as const;

interface SidebarProps {
  runId?: string | null;
  runMode?: RunMode | null;
}

export default function Sidebar({ runId: runIdProp, runMode: runModeProp }: SidebarProps = {}) {
  const pathname = usePathname();
  const [navigating, setNavigating] = useState(false);
  const [lastPath, setLastPath] = useState(pathname);

  useEffect(() => {
    if (pathname !== lastPath) {
      setNavigating(false);
      setLastPath(pathname);
    }
  }, [pathname, lastPath]);

  const runMatch = pathname.match(/^\/runs\/([^/]+)/);
  const runId = runIdProp ?? (runMatch ? runMatch[1] : null);
  const runMode = runModeProp ?? null;
  const groups = runMode === 'ilp' ? ILP_GROUPS
    : runMode === 'cvrp' ? CVRP_GROUPS
      : runMode === 'jss' ? JSS_GROUPS
        : runMode === 'vrptw' ? VRPTW_GROUPS
          : runMode === 'tsp' ? TSP_GROUPS
            : PFRS_GROUPS;

  return (
    <nav className="w-56 bg-gray-950 border-r border-gray-800 shrink-0 fixed top-14 left-0 bottom-0 flex flex-col">
      {navigating && (
        <div className="px-4 py-2 border-b border-gray-800 flex items-center gap-2">
          <div className="w-3 h-3 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
          <span className="text-[9px] text-blue-400">Loading...</span>
        </div>
      )}

      {runId && (
        <div className="px-4 py-2 border-b border-gray-800">
          <p className="text-[9px] uppercase text-gray-600 tracking-wider">Current Run</p>
          <p className="text-xs text-blue-400 font-medium truncate">{runId}</p>
          {runMode && (
            <p className="text-[9px] text-gray-600 mt-0.5">{runMode.toUpperCase()}</p>
          )}
        </div>
      )}

      <div className="flex-1 overflow-y-auto py-2">
        <div className="mb-3">
          <p className="px-4 py-1 text-[9px] uppercase text-gray-600 tracking-wider font-semibold">Platform</p>
          {GLOBAL_ITEMS.filter((item) => {
            if ('adminOnly' in item && item.adminOnly) {
              const mode = typeof window !== 'undefined'
                ? (process.env.NEXT_PUBLIC_ADMIN_MODE || 'development')
                : 'development';
              return mode !== 'disabled';
            }
            return true;
          }).map(({ href, label, icon }) => (
            <Link
              key={href}
              href={href}
              onClick={() => { if (pathname !== href) setNavigating(true); }}
              className={`block px-4 py-1.5 text-xs border-l-2 transition-colors ${
                pathname === href
                  ? 'text-blue-400 border-blue-500 bg-gray-900'
                  : 'text-gray-400 border-transparent hover:text-gray-200 hover:bg-gray-900'
              }`}
            >
              <span className="mr-2">{icon}</span>{label}
            </Link>
          ))}
        </div>

        {runId && groups.map((group) => (
          <div key={group.title} className="mb-2">
            <p className="px-4 py-1 text-[9px] uppercase text-gray-600 tracking-wider font-semibold">{group.title}</p>
            {group.items.map(({ path, label, icon }) => {
              const href = `/runs/${runId}/${path}`;
              const active = pathname === href;
              return (
                <Link
                  key={href}
                  href={href}
                  onClick={() => { if (pathname !== href) setNavigating(true); }}
                  className={`block px-4 py-1.5 text-xs border-l-2 transition-colors ${
                    active
                      ? 'text-blue-400 border-blue-500 bg-gray-900'
                      : 'text-gray-400 border-transparent hover:text-gray-200 hover:bg-gray-900'
                  }`}
                >
                  <span className="mr-2">{icon}</span>{label}
                </Link>
              );
            })}
          </div>
        ))}
      </div>
    </nav>
  );
}
