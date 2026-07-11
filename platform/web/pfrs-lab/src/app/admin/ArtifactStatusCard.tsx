import Card from '@/components/Card';
import type { ArtifactStatus } from '@/lib/intelligence/types';
import { formatArtifactAge } from '@/lib/intelligence/format-utils';

function StatusBadge({ stale }: { stale: boolean }) {
  return (
    <span
      className={`text-[10px] font-semibold uppercase px-2 py-0.5 rounded ${
        stale ? 'bg-amber-900/50 text-amber-300' : 'bg-emerald-900/50 text-emerald-300'
      }`}
    >
      {stale ? 'Stale' : 'Fresh'}
    </span>
  );
}

function ArtifactRow({
  label,
  file,
}: {
  label: string;
  file: ArtifactStatus['summary'];
}) {
  return (
    <tr className="border-t border-gray-800">
      <td className="p-2 text-gray-300">{label}</td>
      <td className="p-2">
        <span className={file.exists ? 'text-emerald-400' : 'text-red-400'}>
          {file.exists ? 'Present' : 'Missing'}
        </span>
      </td>
      <td className="p-2 text-gray-400">{formatArtifactAge(file.generatedAt)}</td>
      <td className="p-2 text-gray-400 text-right">{file.totalRuns ?? '—'}</td>
      <td className="p-2 text-gray-400 text-right">{file.runsScanned ?? '—'}</td>
    </tr>
  );
}

export default function ArtifactStatusCard({ status }: { status: ArtifactStatus }) {
  return (
    <Card title="Intelligence Artifact Status">
      <div className="flex items-center gap-3 mb-4">
        <StatusBadge stale={status.stale} />
        <span className="text-xs text-gray-400">
          {status.currentTotalRuns} runs in storage
        </span>
      </div>

      {status.staleReason && (
        <p className="text-xs text-amber-300 mb-4">{status.staleReason}</p>
      )}

      <table className="w-full text-xs mb-4">
        <thead>
          <tr className="text-gray-500 uppercase">
            <th className="text-left p-2">Artifact</th>
            <th className="text-left p-2">Status</th>
            <th className="text-left p-2">Generated</th>
            <th className="text-right p-2">Runs</th>
            <th className="text-right p-2">Scanned</th>
          </tr>
        </thead>
        <tbody>
          <ArtifactRow label="intelligence_summary.json" file={status.summary} />
          <ArtifactRow label="intelligence_learning.json" file={status.learning} />
          <ArtifactRow label="policy_dashboard.json" file={status.policy} />
        </tbody>
      </table>

      <p className="text-[10px] text-gray-500">
        Artifacts are stale when missing, out of sync with the current run count, or rebuilt at different times.
        Rebuild after uploading runs or retraining policies.
      </p>
    </Card>
  );
}
