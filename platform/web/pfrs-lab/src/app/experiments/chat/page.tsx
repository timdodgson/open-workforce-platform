'use client';
import { useState, useEffect } from 'react';
import ChatWidget from './ChatWidget';

export default function ChatPage() {
  const [token, setToken] = useState<string | null>(null);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Check for saved token on mount.
  useEffect(() => {
    const saved = localStorage.getItem('pfrs-auth-token');
    if (saved) {
      // Check if expired.
      try {
        const payload = JSON.parse(atob(saved.split('.')[1]));
        if (payload.exp && payload.exp > Date.now() / 1000) {
          setToken(saved);
        } else {
          localStorage.removeItem('pfrs-auth-token');
        }
      } catch {
        localStorage.removeItem('pfrs-auth-token');
      }
    }
  }, []);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: 'Login failed' }));
        setError(body.error || 'Login failed');
        setLoading(false);
        return;
      }

      const { idToken } = await res.json();
      localStorage.setItem('pfrs-auth-token', idToken);
      setToken(idToken);
    } catch {
      setError('Network error');
    }

    setLoading(false);
  }

  function handleLogout() {
    localStorage.removeItem('pfrs-auth-token');
    setToken(null);
  }

  if (token) {
    return (
      <div className="relative h-full">
        <button onClick={handleLogout}
          className="absolute top-2 right-2 text-[10px] text-gray-500 hover:text-gray-300 z-10">
          Sign out
        </button>
        <ChatWidget token={token} />
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-8 w-full max-w-sm">
        <h2 className="text-lg font-bold text-blue-400 mb-1">🤖 Optimisation Assistant</h2>
        <p className="text-xs text-gray-500 mb-6">Sign in to access the experiment planner.</p>

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
              required
            />
          </div>
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
              required
            />
          </div>
          {error && <p className="text-xs text-red-400">{error}</p>}
          <button type="submit" disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-gray-700 text-white py-2 rounded text-sm font-medium transition-colors">
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>

        <p className="text-[9px] text-gray-600 mt-4 text-center">
          Dashboard is public. Authentication is only required for the AI assistant.
        </p>
      </div>
    </div>
  );
}
