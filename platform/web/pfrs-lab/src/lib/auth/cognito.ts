import type { AuthUser } from './types';

const REGION = process.env.AWS_REGION ?? 'eu-west-1';

/** Decode and validate a Cognito ID token (issuer + expiry). */
export async function verifyCognitoIdToken(token: string): Promise<AuthUser | null> {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;

    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString()) as {
      sub?: string;
      email?: string;
      exp?: number;
      iss?: string;
    };

    const now = Math.floor(Date.now() / 1000);
    if (payload.exp && payload.exp < now) return null;

    const poolId = process.env.COGNITO_USER_POOL_ID;
    if (poolId) {
      const expectedIssuer = `https://cognito-idp.${REGION}.amazonaws.com/${poolId}`;
      if (payload.iss !== expectedIssuer) return null;
    }

    if (!payload.sub) return null;

    return {
      id: payload.sub,
      email: payload.email || payload.sub,
      isAdmin: true,
    };
  } catch {
    return null;
  }
}
