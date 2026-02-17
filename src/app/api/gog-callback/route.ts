import { NextRequest, NextResponse } from 'next/server';
import fs from 'fs';
import path from 'path';
import os from 'os';

const CLIENT_ID     = '***REDACTED_GOOGLE_CLIENT_ID***';
const CLIENT_SECRET = '***REDACTED_GOOGLE_SECRET***';
const REDIRECT_URI  = 'https://ocprd-sgp1-01.***REDACTED_HOST***:3000/api/gog-callback';
const KEYRING_DIR   = path.join(os.homedir(), '.config/gogcli/keyring');
const KEYRING_PATH  = path.join(KEYRING_DIR, 'sa-***REDACTED_EMAIL***.json');

export async function GET(req: NextRequest) {
  const code = req.nextUrl.searchParams.get('code');
  if (!code) return NextResponse.json({ error: 'no code' }, { status: 400 });

  const res = await fetch('https://oauth2.googleapis.com/token', {
    method: 'POST',
    headers: { 'content-type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({ code, client_id: CLIENT_ID, client_secret: CLIENT_SECRET, redirect_uri: REDIRECT_URI, grant_type: 'authorization_code' }),
  });

  if (!res.ok) {
    const err = await res.text();
    return new NextResponse(`Token exchange failed: ${err}`, { status: 500 });
  }

  const tokens = await res.json();
  // Save to a plain JSON file readable by scripts
  const tokenPath = path.join(os.homedir(), '.config/gogcli/google-tokens.json');
  fs.mkdirSync(path.dirname(tokenPath), { recursive: true });
  fs.writeFileSync(tokenPath, JSON.stringify({ ...tokens, saved_at: new Date().toISOString() }, null, 2), { mode: 0o600 });

  return new NextResponse(
    '<h2 style="font-family:sans-serif;color:green;padding:2rem">✅ Google authorized! You can close this tab.</h2>',
    { headers: { 'content-type': 'text/html' } }
  );
}
