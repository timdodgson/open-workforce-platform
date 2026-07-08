'use client';

import Card from '@/components/Card';
import TabSpinner from '@/components/TabSpinner';

interface IntelligenceTabPanelProps {
  title: string;
  loading?: boolean;
  empty?: boolean;
  emptyMessage: string;
  children: React.ReactNode;
}

export default function IntelligenceTabPanel({
  title,
  loading,
  empty,
  emptyMessage,
  children,
}: IntelligenceTabPanelProps) {
  if (loading) {
    return (
      <Card title={title}>
        <TabSpinner />
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
