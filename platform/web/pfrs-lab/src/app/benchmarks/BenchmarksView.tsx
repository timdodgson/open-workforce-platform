'use client';

import { useState } from 'react';
import type { BenchmarkRun } from './page';
import type { ProblemType } from '@/lib/benchmark-suites';
import { BENCHMARK_SUITES } from '@/lib/benchmark-suites';
import SIComparison from './SIComparison';
import DomainBenchmarkCard from './DomainBenchmarkCard';

const DOMAINS: Array<{ id: 'all' | ProblemType; label: string }> = [
  { id: 'all', label: 'All domains' },
  { id: 'nrp', label: 'NRP' },
  { id: 'cvrp', label: 'CVRP' },
  { id: 'jss', label: 'JSS' },
  { id: 'vrptw', label: 'VRPTW' },
];

export default function BenchmarksView({
  runs,
  knownOptimal,
}: {
  runs: BenchmarkRun[];
  knownOptimal: Record<string, { value: number; source: string }>;
}) {
  const [domainFilter, setDomainFilter] = useState<'all' | ProblemType>('all');

  const suites =
    domainFilter === 'all'
      ? BENCHMARK_SUITES
      : BENCHMARK_SUITES.filter((s) => s.id === domainFilter);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-3 p-3 rounded-lg bg-gray-900/80 border border-gray-800">
        <span className="text-[10px] text-gray-500 uppercase">Domain</span>
        {DOMAINS.map((d) => (
          <button
            key={d.id}
            type="button"
            onClick={() => setDomainFilter(d.id)}
            className={`text-xs px-3 py-1 rounded border transition-colors ${
              domainFilter === d.id
                ? 'bg-blue-900/50 text-blue-300 border-blue-700'
                : 'bg-gray-800 text-gray-400 border-gray-700 hover:border-gray-600'
            }`}
          >
            {d.label}
          </button>
        ))}
      </div>

      <SIComparison runs={runs} domainFilter={domainFilter} />

      {suites.map((suite) => (
        <DomainBenchmarkCard key={suite.id} suite={suite} runs={runs} knownOptimal={knownOptimal} />
      ))}
    </div>
  );
}
