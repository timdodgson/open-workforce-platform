import type { ReactNode } from 'react';

type StatGridCols = 2 | 3 | 4;

const COL_CLASSES: Record<StatGridCols, string> = {
  2: 'grid grid-cols-2 gap-3',
  3: 'grid grid-cols-2 md:grid-cols-3 gap-3',
  4: 'grid grid-cols-2 sm:grid-cols-4 gap-3',
};

interface StatGridProps {
  cols?: StatGridCols;
  className?: string;
  children: ReactNode;
}

/** Standard metric tile grid used across run summaries and dashboards. */
export default function StatGrid({ cols = 4, className, children }: StatGridProps) {
  return (
    <div className={className ? `${COL_CLASSES[cols]} ${className}` : COL_CLASSES[cols]}>
      {children}
    </div>
  );
}
