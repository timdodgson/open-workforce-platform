import Card from '@/components/Card';

interface RunPageShellProps {
  title: string;
  error?: string | null;
  empty?: boolean;
  emptyMessage?: string;
  children: React.ReactNode;
}

/** Shared error/empty wrapper for /runs/[id]/* pages. */
export default function RunPageShell({
  title,
  error,
  empty,
  emptyMessage = 'No data available for this run.',
  children,
}: RunPageShellProps) {
  if (error) {
    return (
      <Card title="Error">
        <p className="text-red-400 text-sm">Failed to load data: {error}</p>
      </Card>
    );
  }
  if (empty) {
    return (
      <Card title={title}>
        <p className="text-xs text-gray-500 text-center py-12">{emptyMessage}</p>
      </Card>
    );
  }
  return <>{children}</>;
}
