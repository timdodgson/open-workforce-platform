'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const NAV_ITEMS = [
  { href: '/summary', label: 'Summary', icon: '📋' },
  { href: '/search', label: 'Search Progress', icon: '📈' },
  { href: '/tree', label: 'Search Tree', icon: '🌳' },
  { href: '/sa', label: 'Simulated Annealing', icon: '🔥' },
  { href: '/workers', label: 'Workers', icon: '👷' },
  { href: '/diversity', label: 'Diversity', icon: '🌍' },
  { href: '/compare', label: 'Compare Runs', icon: '⚖️' },
];

export default function Sidebar() {
  const pathname = usePathname();
  return (
    <nav className="w-56 bg-gray-900 border-r border-gray-700 fixed top-0 left-0 bottom-0 flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h1 className="text-sm font-bold text-blue-400">PFRS Research Lab</h1>
        <p className="text-xs text-gray-500 mt-1">Performance Analysis</p>
      </div>
      <ul className="flex-1 py-2">
        {NAV_ITEMS.map(({ href, label, icon }) => {
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
