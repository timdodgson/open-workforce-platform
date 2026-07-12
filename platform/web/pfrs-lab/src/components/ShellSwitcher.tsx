'use client';

import { usePathname } from 'next/navigation';
import AppShell from '@/components/AppShell';
import SiteShell from '@/components/SiteShell';

/** Routes that use the public marketing shell (no sidebar). */
const SITE_ROUTES = new Set([
  '/',
  '/about',
  '/research',
  '/reproduce',
  '/algorithms',
  '/domains',
  '/references',
  '/getting-started',
]);

export default function ShellSwitcher({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isSite = SITE_ROUTES.has(pathname);

  if (isSite) {
    return <SiteShell>{children}</SiteShell>;
  }

  return <AppShell>{children}</AppShell>;
}
