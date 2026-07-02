'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const PAGE_ITEMS = [
  { path: 'summary', label: 'Summary', icon: '📋' },
  { path: 'schedule', label: 'Schedule', icon: '📅' },
  { path: 'search', label: 'Search Progress', icon: '📈' },
  { path: 'tree', label: 'Search Tree', icon: '🌳' },
  { path: 'inheritance', label: 'Inheritance', icon: '🧬' },
  { path: 'insights', label: 'Insights', icon: '💡' },
  { path: 'diversity', label: 'Diversity', icon: '🌍' },
];

const GLOBAL_ITEMS = [
  { href: '/compare', label: 'Compare', icon: '⚖️' },
];

export default function Sidebar() {
  const pathname = usePathname();

  // Detect if we're inside a run: /runs/<id>/...
  const runMatch = pathname.match(/^\/runs\/([^/]+)/);
  const runId = runMatch ? runMatch[1] : null;

  // Build nav links based on whether we're in a run or not.
  const navItems = PAGE_ITEMS.map(item => ({
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
        </div>
      )}

      <ul className="flex-1 py-2">
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
