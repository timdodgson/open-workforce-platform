import Card from '@/components/Card';
import TabSpinner from '@/components/TabSpinner';

interface EmptyStateProps {
  title: string;
  message: string;
  loading?: boolean;
}

/** Shared empty/loading state for intelligence tabs and run pages. */
export default function EmptyState({ title, message, loading }: EmptyStateProps) {
  if (loading) {
    return (
      <Card title={title}>
        <TabSpinner />
      </Card>
    );
  }
  return (
    <Card title={title}>
      <p className="text-xs text-gray-500 text-center py-12">{message}</p>
    </Card>
  );
}
