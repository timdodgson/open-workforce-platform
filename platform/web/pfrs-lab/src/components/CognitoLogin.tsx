'use client';

import { useState } from 'react';
import Card from '@/components/Card';

interface CognitoLoginProps {
  title: string;
  description: string;
  onSuccess?: () => void;
}

export default function CognitoLogin({ title, description, onSuccess }: CognitoLoginProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

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
        return;
      }

      const { idToken } = await res.json();
      localStorage.setItem('pfrs-auth-token', idToken);
      onSuccess?.();
      window.location.reload();
    } catch {
      setError('Network error');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex items-center justify-center py-16">
      <Card title={title}>
        <p className="text-xs text-gray-500 mb-6">{description}</p>
        <form onSubmit={handleLogin} className="space-y-4 max-w-sm">
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full bg-gray-950 border border-gray-700 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
              required
            />
          </div>
          <div>
            <label className="text-[10px] uppercase text-gray-500 block mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-gray-950 border border-gray-700 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
              required
            />
          </div>
          {error && <p className="text-xs text-red-400">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-gray-700 text-white py-2 rounded text-sm font-medium transition-colors"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </Card>
    </div>
  );
}
