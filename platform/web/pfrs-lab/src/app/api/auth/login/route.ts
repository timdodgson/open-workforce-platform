import { NextResponse } from 'next/server';
import { COOKIE_NAME } from '@/lib/auth/session';

const REGION = process.env.AWS_REGION ?? 'eu-west-1';
const USER_POOL_ID = process.env.COGNITO_USER_POOL_ID ?? '';
const CLIENT_ID = process.env.COGNITO_CLIENT_ID ?? '';

export async function POST(request: Request) {
  try {
    const { email, password } = await request.json();

    if (!email || !password) {
      return NextResponse.json({ error: 'Email and password required' }, { status: 400 });
    }

    if (!USER_POOL_ID || !CLIENT_ID) {
      return NextResponse.json({ error: 'Authentication not configured' }, { status: 500 });
    }

    const authResponse = await fetch(
      `https://cognito-idp.${REGION}.amazonaws.com/`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-amz-json-1.1',
          'X-Amz-Target': 'AWSCognitoIdentityProviderService.InitiateAuth',
        },
        body: JSON.stringify({
          AuthFlow: 'USER_PASSWORD_AUTH',
          ClientId: CLIENT_ID,
          AuthParameters: {
            USERNAME: email,
            PASSWORD: password,
          },
        }),
      },
    );

    const result = await authResponse.json();

    if (result.__type && result.__type.includes('Exception')) {
      const message = result.message || 'Authentication failed';
      return NextResponse.json({ error: message }, { status: 401 });
    }

    if (result.ChallengeName === 'NEW_PASSWORD_REQUIRED') {
      return NextResponse.json({ error: 'Password change required. Contact admin.' }, { status: 401 });
    }

    const idToken = result.AuthenticationResult?.IdToken;
    const expiresIn = result.AuthenticationResult?.ExpiresIn ?? 3600;
    if (!idToken) {
      return NextResponse.json({ error: 'Unexpected auth response' }, { status: 500 });
    }

    const response = NextResponse.json({ idToken, expiresIn });
    response.cookies.set(COOKIE_NAME, idToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/',
      maxAge: expiresIn,
    });
    return response;
  } catch (err) {
    console.error('Login error:', err);
    return NextResponse.json({ error: 'Login failed' }, { status: 500 });
  }
}
