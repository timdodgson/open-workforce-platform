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
    <header className="fixed top-0 left-0 right-0 z-50 h-14 bg-gray-950 border-b border-gray-800 flex items-center justify-between px-6">
      <Link href="/" className="block min-w-0">
        <h1 className="text-sm font-bold text-blue-400 leading-tight">PFRS Lab</h1>
        <p className="text-[10px] text-gray-500 leading-tight">Adaptive Optimisation Research</p>
      </Link>

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
