import { AuthProvider, AuthUser, getAdminMode } from './types';

/**
 * Development auth provider — always grants admin access.
 * Used when PFRS_ADMIN_MODE=development.
 */
class DevelopmentAuthProvider implements AuthProvider {
  async isAuthenticated(): Promise<boolean> { return true; }
  async isAdmin(): Promise<boolean> { return true; }
  async getCurrentUser(): Promise<AuthUser> {
    return { id: 'dev', email: 'dev@localhost', isAdmin: true };
  }
}

/**
 * Disabled auth provider — never grants access.
 * Used when PFRS_ADMIN_MODE=disabled.
 */
class DisabledAuthProvider implements AuthProvider {
  async isAuthenticated(): Promise<boolean> { return false; }
  async isAdmin(): Promise<boolean> { return false; }
  async getCurrentUser(): Promise<AuthUser | null> { return null; }
}

/**
 * Placeholder for future Cognito integration.
 * When PFRS_ADMIN_MODE=authenticated, this would validate tokens.
 * For v3, it falls back to disabled (requires explicit auth setup).
 */
class AuthenticatedProvider implements AuthProvider {
  async isAuthenticated(): Promise<boolean> {
    // TODO: Implement Cognito token validation.
    return false;
  }
  async isAdmin(): Promise<boolean> { return false; }
  async getCurrentUser(): Promise<AuthUser | null> { return null; }
}

let instance: AuthProvider | null = null;

export function getAuthProvider(): AuthProvider {
  if (instance) return instance;

  const mode = getAdminMode();
  switch (mode) {
    case 'development':
      instance = new DevelopmentAuthProvider();
      break;
    case 'authenticated':
      instance = new AuthenticatedProvider();
      break;
    case 'disabled':
    default:
      instance = new DisabledAuthProvider();
      break;
  }

  return instance;
}
