import type { Metadata } from 'next';
import { getStorageProvider } from '@/lib/storage/factory';
import CapabilitiesView, { type RegistrySummary } from './CapabilitiesView';
import type { DomainId } from '@/lib/capability-matrix';

export const metadata: Metadata = {
  title: 'Capabilities',
  description: 'Platform capability matrix across NRP, CVRP, VRPTW, and JSS — solvers, Search Intelligence, viewers, and gaps.',
};

export const dynamic = 'force-dynamic';

async function loadRegistrySummary(): Promise<RegistrySummary | null> {
  try {
    const store = getStorageProvider();
    const content = await store.readRootFile('policy_registry.json');
    if (!content) return null;
    const data = JSON.parse(content) as {
      versions?: Array<{ domain?: string; promotion_ready?: boolean }>;
    };
    const versions = data.versions ?? [];
    const domains = new Set<DomainId>();
    let promotionReady = 0;
    for (const v of versions) {
      if (v.promotion_ready) promotionReady++;
      const d = v.domain?.toLowerCase();
      if (d === 'nrp' || d === 'cvrp' || d === 'vrptw' || d === 'jss') domains.add(d);
    }
    return {
      total: versions.length,
      promotionReady,
      domains: [...domains],
    };
  } catch {
    return null;
  }
}

export default async function CapabilitiesPage() {
  const registry = await loadRegistrySummary();
  return <CapabilitiesView registry={registry} />;
}
