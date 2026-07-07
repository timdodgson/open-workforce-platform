/**
 * Authentication abstraction for PFRS Lab.
 * 
 * Designed to be backed by Cognito in production.
 * For v3: supports disabled, development, and authenticated modes.
 */

export interface AuthUser {
  id: string;
  email: string;
  isAdmin: boolean;
}

export interface AuthProvider {
  /** Check if the current request is authenticated. */
  isAuthenticated(): Promise<boolean>;

  /** Check if the current user is an administrator. */
  isAdmin(): Promise<boolean>;

  /** Get the current user, or null if unauthenticated. */
  getCurrentUser(): Promise<AuthUser | null>;
}

/** Admin access mode, controlled by PFRS_ADMIN_MODE env var. */
export type AdminMode = 'disabled' | 'development' | 'authenticated';

export function getAdminMode(): AdminMode {
  const mode = process.env.PFRS_ADMIN_MODE || 'development';
  if (mode === 'disabled' || mode === 'authenticated') return mode;
  return 'development';
}
