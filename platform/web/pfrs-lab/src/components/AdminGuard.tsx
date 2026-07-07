import { getAuthProvider, getAdminMode } from '@/lib/auth';
import Card from './Card';

/**
 * Server-side admin access guard.
 * Wraps admin-only content. Renders children only if the user is authorised.
 */
export default async function AdminGuard({ children }: { children: React.ReactNode }) {
  const mode = getAdminMode();

  // Disabled: admin section is hidden entirely.
  if (mode === 'disabled') {
    return (
      <Card title="Admin">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Admin section is disabled.</p>
          <p className="text-xs mt-1">Set <code className="text-blue-400">PFRS_ADMIN_MODE=development</code> to enable.</p>
        </div>
      </Card>
    );
  }

  // Development: bypass auth with visible indicator.
  if (mode === 'development') {
    return (
      <div>
        <div className="mb-4 bg-amber-900/20 border border-amber-700 rounded-lg px-4 py-2 text-center">
          <p className="text-[10px] text-amber-400 font-semibold">Development Mode — Authentication Disabled</p>
        </div>
        {children}
      </div>
    );
  }

  // Authenticated: check the auth provider.
  const auth = getAuthProvider();
  const isAdmin = await auth.isAdmin();

  if (!isAdmin) {
    return (
      <Card title="Access Denied">
        <div className="border-2 border-dashed border-red-800 rounded-lg p-8 text-center text-gray-400">
          <p className="text-red-400 font-semibold mb-2">Administrator access required.</p>
          <p className="text-xs">Sign in with an administrator account to access this page.</p>
        </div>
      </Card>
    );
  }

  return <>{children}</>;
}
