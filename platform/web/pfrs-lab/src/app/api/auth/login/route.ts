import { NextResponse } from 'next/server';

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

    // Use Cognito InitiateAuth API directly via HTTP.
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
      }
    );

    const result = await authResponse.json();

    if (result.__type && result.__type.includes('Exception')) {
      const message = result.message || 'Authentication failed';
      return NextResponse.json({ error: message }, { status: 401 });
    }

    if (result.ChallengeName === 'NEW_PASSWORD_REQUIRED') {
      return NextResponse.json({ error: 'Password change required. Contact admin.' }, { status: 401 });
    }

    if (result.AuthenticationResult?.IdToken) {
      return NextResponse.json({
        idToken: result.AuthenticationResult.IdToken,
        expiresIn: result.AuthenticationResult.ExpiresIn,
      });
    }

    return NextResponse.json({ error: 'Unexpected auth response' }, { status: 500 });
  } catch (err) {
    console.error('Login error:', err);
    return NextResponse.json({ error: 'Login failed' }, { status: 500 });
  }
}
