import { NextRequest, NextResponse } from 'next/server';

// Routes that require authentication (server-side guard).
// NOTE: This middleware only runs when Next.js is deployed in server mode
// (i.e. NOT output: 'export'). For static exports, nginx auth_request
// handles /admin protection using the access_token cookie.
const PROTECTED_ROUTES = [
  '/dashboard',
  '/settings',
  '/analytics',
  '/ab-tests',
  '/audience',
  '/custom-domains',
  '/teams',
  '/billing',
  '/admin',
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isProtected = PROTECTED_ROUTES.some(
    (r) => pathname === r || pathname.startsWith(r + '/')
  );

  if (!isProtected) {
    return NextResponse.next();
  }

  // Token is stored in cookie (set alongside localStorage at login)
  const token = request.cookies.get('access_token')?.value;

  if (!token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/settings/:path*',
    '/analytics/:path*',
    '/ab-tests/:path*',
    '/audience/:path*',
    '/custom-domains/:path*',
    '/teams/:path*',
    '/billing/:path*',
    '/admin/:path*',
  ],
};
