'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';

interface PageItem {
  path: string;
  label: string;
  icon: string;
}

const PFRS_PAGES: PageItem[] = [
  { path: 'summary', label: 'Summary', icon: '📋' },
  { path: 'timeline', label: 'Timeline', icon: '⏱️' },
  { path: 'schedule', label: 'Schedule', icon: '📅' },
  { path: 'search', label: 'Search Progress', icon: '📈' },
  { path: 'replay', label: 'Replay', icon: '🎬' },
  { path: 'map', label: 'Search Map', icon: '🗺️' },
  { path: 'temperature', label: 'Temperature', icon: '🌡️' },
  { path: 'waterfall', label: 'Waterfall', icon: '🌊' },
  { path: 'plateaus', label: 'Plateaus', icon: '🏔️' },
  { path: 'causality', label: 'Causality', icon: '🔗' },
  { path: 'efficiency', label: 'Efficiency', icon: '⚡' },
  { path: 'constraints', label: 'Constraints', icon: '⚖️' },
  { path: 'workers', label: 'Workers', icon: '👷' },
  { path: 'archetypes', label: 'Archetypes', icon: '🎭' },
  { path: 'genealogy', label: 'Genealogy', icon: '🌲' },
  { path: 'families', label: 'Families', icon: '👨‍👩‍👧‍👦' },
  { path: 'dna', label: 'Search DNA', icon: '🧪' },
  { path: 'tree', label: 'Search Tree', icon: '🌳' },
  { path: 'pathdiff', label: 'Path Diff', icon: '🔃' },
  { path: 'inheritance', label: 'Inheritance', icon: '🧬' },
  { path: 'insights', label: 'Insights', icon: '💡' },
  { path: 'diversity', label: 'Diversity', icon: '🌍' },
  { path: 'landscape', label: 'Landscape', icon: '🏔️' },
  { path: 'report', label: 'Report', icon: '📄' },
  { path: 'export', label: 'Export', icon: '📤' },
  { path: 'explain', label: 'Explain', icon: '🧠' },
];

const ILP_PAGES: PageItem[] = [
  { path: 'summary', label: 'Summary', icon: '📋' },
  { path: 'baseline', label: 'Solve Progress', icon: '📐' },
  { path: 'schedule', label: 'Schedule', icon: '📅' },
  { path: 'export', label: 'Export', icon: '📤' },
];

const CVRP_PAGES: PageItem[] = [
  { path: 'summary', label: 'Summary', icon: '📋' },
  { path: 'routes', label: 'Route Viewer', icon: '🚛' },
  { path: 'search', label: 'Search Progress', icon: '📈' },
  { path: 'timeline', label: 'Timeline', icon: '⏱️' },
  { path: 'workers', label: 'Workers', icon: '👷' },
  { path: 'dna', label: 'Search DNA', icon: '🧪' },
  { path: 'map', label: 'Search Map', icon: '🗺️' },
  { path: 'export', label: 'Export', icon: '📤' },
];

const GLOBAL_ITEMS = [
  { href: '/compare', label: 'Compare', icon: '🔀' },
  { href: '/statistics', label: 'Statistics', icon: '📊' },
  { href: '/benchmarks', label: 'Benchmarks', icon: '🏆' },
  { href: '/trends', label: 'Trends', icon: '📈' },
  { href: '/experiments', label: 'Experiments', icon: '🧬' },
  { href: '/experiments/chat', label: 'Assistant', icon: '🤖' },
  { href: '/knowledge', label: 'Knowledge', icon: '📚' },
];

export default function Sidebar() {
  const pathname = usePathname();
  const [runMode, setRunMode] = useState<string | null>(null);

  // Detect if we're inside a run: /runs/<id>/...
  const runMatch = pathname.match(/^\/runs\/([^/]+)/);
  const runId = runMatch ? runMatch[1] : null;

  // Fetch run mode when run changes.
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

  // Select pages based on run mode.
  const pages = runMode === 'ilp' ? ILP_PAGES : runMode === 'cvrp' ? CVRP_PAGES : PFRS_PAGES;

  const navItems = pages.map(item => ({
    href: runId ? `/runs/${runId}/${item.path}` : `/${item.path}`,
    label: item.label,
    icon: item.icon,
  }));

  return (
    <nav className="w-56 bg-gray-900 border-r border-gray-700 fixed top-0 left-0 bottom-0 flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <Link href="/" className="block">
          <h1 className="text-sm font-bold text-blue-400">PFRS Research Lab</h1>
          <p className="text-xs text-gray-500 mt-1">Performance Analysis</p>
        </Link>
      </div>

      {/* Show current run label */}
      {runId && (
        <div className="px-4 py-2 border-b border-gray-800">
          <p className="text-[10px] uppercase text-gray-600 tracking-wider">Current Run</p>
          <p className="text-xs text-emerald-400 font-medium truncate">{runId}</p>
          {runMode && (
            <p className="text-[10px] text-gray-600 mt-0.5">{runMode.toUpperCase()}</p>
          )}
        </div>
      )}

      <ul className="flex-1 py-2 overflow-y-auto">
        <li>
          <Link href="/"
            className={`block px-4 py-2 text-sm border-l-2 transition-colors ${
              pathname === '/'
                ? 'text-blue-400 border-blue-400 bg-blue-400/10'
                : 'text-gray-400 border-transparent hover:text-white hover:bg-gray-800'
            }`}>
            <span className="mr-2">🏠</span>All Runs
          </Link>
        </li>
        {runId && navItems.map(({ href, label, icon }) => {
          const active = pathname === href;
          return (
            <li key={href}>
              <Link href={href}
                className={`block px-4 py-2 text-sm border-l-2 transition-colors ${
                  active
                    ? 'text-blue-400 border-blue-400 bg-blue-400/10'
                    : 'text-gray-400 border-transparent hover:text-white hover:bg-gray-800'
                }`}>
                <span className="mr-2">{icon}</span>{label}
              </Link>
            </li>
          );
        })}
        {GLOBAL_ITEMS.map(({ href, label, icon }) => {
          const active = pathname === href;
          return (
            <li key={href}>
              <Link href={href}
                className={`block px-4 py-2 text-sm border-l-2 transition-colors ${
                  active
                    ? 'text-blue-400 border-blue-400 bg-blue-400/10'
                    : 'text-gray-400 border-transparent hover:text-white hover:bg-gray-800'
                }`}>
                <span className="mr-2">{icon}</span>{label}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
