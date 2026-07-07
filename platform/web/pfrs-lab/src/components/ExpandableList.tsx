'use client';

import { useState } from 'react';

interface ExpandableListProps<T> {
  items: T[];
  defaultCount?: number;
  renderItem: (item: T, index: number) => React.ReactNode;
  className?: string;
}

/**
 * Renders a list with "Show all" / "Show fewer" toggle.
 * Defaults to showing 10 items.
 */
export default function ExpandableList<T>({ items, defaultCount = 10, renderItem, className = '' }: ExpandableListProps<T>) {
  const [expanded, setExpanded] = useState(false);
  const visible = expanded ? items : items.slice(0, defaultCount);
  const hasMore = items.length > defaultCount;

  return (
    <div className={className}>
      {visible.map((item, i) => renderItem(item, i))}
      {hasMore && (
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-[10px] text-blue-400 hover:text-blue-300 mt-2 px-1"
        >
          {expanded ? `Show fewer (${defaultCount})` : `Show all (${items.length})`}
        </button>
      )}
    </div>
  );
}
