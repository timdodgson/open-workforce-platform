import { listRunsAsync } from '@/lib/data-loader';
import LandingPage from '@/features/landing/LandingPage';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  const runs = await listRunsAsync();
  return <LandingPage runs={runs} />;
}
