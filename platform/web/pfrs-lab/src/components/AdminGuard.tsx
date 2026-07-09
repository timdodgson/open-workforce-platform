import { getAuthProvider, getAdminMode } from '@/lib/auth';
import { getSessionUser } from '@/lib/auth/session';
import CognitoLogin from './CognitoLogin';
import Card from './Card';

/**
 * Server-side admin access guard.
 * Wraps admin-only content. Renders children only if the user is authorised.
 */
export default async function AdminGuard({ children }: { children: React.ReactNode }) {
  const mode = getAdminMode();

  if (mode === 'disabled') {
    return (
      <Card title="Admin">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Admin section is disabled.</p>
        </div>
      </Card>
    );
  }

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

  const user = await getSessionUser();
  if (!user) {
    return (
      <CognitoLogin
        title="Administrator Sign In"
        description="Sign in with your Cognito account to access platform administration."
      />
    );
  }

  return <>{children}</>;
}
