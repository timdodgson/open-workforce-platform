'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';

interface PageItem {
  path: string;
  label: string;
  icon: string;
}

interface PageGroup {
  title: string;
  items: PageItem[];
}

// --- NRP beam search pages grouped by purpose ---
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
      { path: 'families', label: 'Families', icon: '👨‍👩‍👧‍👦' },
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

// --- CVRP pages grouped ---
const CVRP_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'routes', label: 'Route Viewer', icon: '🚛' },
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

// --- ILP pages ---
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

// --- JSS (Job Shop) pages ---
const JSS_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'gantt', label: 'Gantt Chart', icon: '📊' },
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

// --- VRPTW (Vehicle Routing with Time Windows) pages ---
const VRPTW_GROUPS: PageGroup[] = [
  {
    title: 'Overview',
    items: [
      { path: 'summary', label: 'Summary', icon: '📋' },
      { path: 'routes', label: 'Route Viewer', icon: '🚛' },
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

const GLOBAL_ITEMS = [
  { href: '/', label: 'Home', icon: '🏠' },
  { href: '/benchmarks', label: 'Benchmarks', icon: '🏆' },
  { href: '/statistics', label: 'Statistics', icon: '📊' },
  { href: '/compare', label: 'Compare', icon: '🔀' },
  { href: '/trends', label: 'Trends', icon: '📈' },
  { href: '/experiments', label: 'Experiments', icon: '🧬' },
  { href: '/experiments/chat', label: 'Assistant', icon: '🤖' },
  { href: '/knowledge', label: 'Knowledge', icon: '📚' },
  { href: '/learning', label: 'Learning', icon: '🧠' },
  { href: '/feature-importance', label: 'Model', icon: '🔬' },
  { href: '/predictions', label: 'Predictions', icon: '🧪' },
  { href: '/what-if', label: 'What-If Lab', icon: '⚗️' },
  { href: '/assist', label: 'Assist', icon: '🤝' },
  { href: '/decisions', label: 'Decisions', icon: '🎯' },
  { href: '/admin', label: 'Admin', icon: '⚙️' },
];

export default function Sidebar() {
  const pathname = usePathname();
  const [runMode, setRunMode] = useState<string | null>(null);

  const runMatch = pathname.match(/^\/runs\/([^/]+)/);
  const runId = runMatch ? runMatch[1] : null;

  useEffect(() => {
    if (!runId) {
      setRunMode(null);
      return;
    }
    fetch(`/api/runs/${runId}/meta`)
      .then(res => res.ok ? res.json() : null)
      .then(data => setRunMode(data?.mode ?? 'pfrs'))
      .catch(() => setRunMode('pfrs'));
  }, [runId]);

  // Select page groups based on run mode.
  const groups = runMode === 'ilp' ? ILP_GROUPS : runMode === 'cvrp' ? CVRP_GROUPS : runMode === 'jss' ? JSS_GROUPS : runMode === 'vrptw' ? VRPTW_GROUPS : PFRS_GROUPS;

  return (
    <nav className="w-56 bg-gray-900 border-r border-gray-700 fixed top-0 left-0 bottom-0 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <Link href="/" className="block">
          <h1 className="text-sm font-bold text-blue-400">PFRS Research Lab</h1>
          <p className="text-[10px] text-gray-500 mt-0.5">Optimisation Research Platform</p>
        </Link>
      </div>

      {/* Current run */}
      {runId && (
        <div className="px-4 py-2 border-b border-gray-800">
          <p className="text-[9px] uppercase text-gray-600 tracking-wider">Current Run</p>
          <p className="text-xs text-emerald-400 font-medium truncate">{runId}</p>
          {runMode && (
            <p className="text-[9px] text-gray-600 mt-0.5">{runMode.toUpperCase()}</p>
          )}
        </div>
      )}

      <div className="flex-1 overflow-y-auto py-2">
        {/* Global navigation */}
        <div className="mb-3">
          <p className="px-4 py-1 text-[9px] uppercase text-gray-600 tracking-wider font-semibold">Platform</p>
          {GLOBAL_ITEMS.map(({ href, label, icon }) => (
            <Link key={href} href={href}
              className={`block px-4 py-1.5 text-xs border-l-2 transition-colors ${
                pathname === href
                  ? 'text-blue-400 border-blue-400 bg-blue-400/10'
                  : 'text-gray-400 border-transparent hover:text-white hover:bg-gray-800'
              }`}>
              <span className="mr-2">{icon}</span>{label}
            </Link>
          ))}
        </div>

        {/* Run-specific grouped navigation */}
        {runId && groups.map(group => (
          <div key={group.title} className="mb-2">
            <p className="px-4 py-1 text-[9px] uppercase text-gray-600 tracking-wider font-semibold">{group.title}</p>
            {group.items.map(({ path, label, icon }) => {
              const href = `/runs/${runId}/${path}`;
              const active = pathname === href;
              return (
                <Link key={href} href={href}
                  className={`block px-4 py-1.5 text-xs border-l-2 transition-colors ${
                    active
                      ? 'text-blue-400 border-blue-400 bg-blue-400/10'
                      : 'text-gray-400 border-transparent hover:text-white hover:bg-gray-800'
                  }`}>
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
