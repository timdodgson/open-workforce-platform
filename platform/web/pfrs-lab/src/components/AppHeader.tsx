'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';

interface AuthUser {
  email: string;
  initials: string;
}

export default function AppHeader() {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/auth/me', { credentials: 'include' })
      .then((res) => (res.ok ? res.json() : { user: null }))
      .then((data) => setUser(data.user ?? null))
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  async function handleLogout() {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' });
    localStorage.removeItem('pfrs-auth-token');
    window.location.href = '/';
  }

  return (
    <header className="fixed top-0 left-0 right-0 z-50 h-14 bg-gray-950 border-b border-gray-800 flex items-center justify-between px-6 gap-4">
      <Link href="/lab" className="block min-w-0 shrink-0">
        <h1 className="text-sm font-bold text-blue-400 leading-tight">PFRS Lab</h1>
        <p className="text-[10px] text-gray-500 leading-tight">Adaptive Optimisation Research</p>
      </Link>

      <nav className="flex items-center gap-3 sm:gap-4 text-xs text-gray-400 min-w-0 overflow-x-auto" aria-label="Site links">
        <Link href="/" className="hover:text-gray-200 whitespace-nowrap shrink-0">
          Public site
        </Link>
        <Link href="/getting-started" className="hover:text-gray-200 whitespace-nowrap shrink-0">
          Get started
        </Link>
        <Link href="/about" className="hover:text-gray-200 whitespace-nowrap shrink-0">
          About
        </Link>
      </nav>

      <div className="flex items-center gap-3 shrink-0">
        {!loading && (
          user ? (
            <>
              <div
                className="w-8 h-8 rounded-full bg-blue-600 text-white text-xs font-semibold flex items-center justify-center"
                title={user.email}
              >
                {user.initials}
              </div>
              <button
                type="button"
                onClick={handleLogout}
                className="text-xs text-gray-400 hover:text-gray-200 transition-colors"
              >
                Sign out
              </button>
            </>
          ) : (
            <Link
              href="/admin"
              className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
            >
              Sign in
            </Link>
          )
        )}
      </div>
    </header>
  );
}
