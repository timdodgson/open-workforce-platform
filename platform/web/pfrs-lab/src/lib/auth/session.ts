import { cookies } from 'next/headers';
import { verifyCognitoIdToken } from './cognito';
import type { AuthUser } from './types';

const COOKIE_NAME = 'pfrs-auth-token';

/** Read the Cognito session from the httpOnly auth cookie (server-side). */
export async function getSessionUser(): Promise<AuthUser | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get(COOKIE_NAME)?.value;
  if (!token) return null;
  return verifyCognitoIdToken(token);
}

export { COOKIE_NAME };
