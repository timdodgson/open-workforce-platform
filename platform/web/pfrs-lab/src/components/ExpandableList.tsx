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
 * If items <= defaultCount, shows everything with no controls.
 */
export default function ExpandableList<T>({ items, defaultCount = 10, renderItem, className = '' }: ExpandableListProps<T>) {
  const [expanded, setExpanded] = useState(false);
  const visible = expanded ? items : items.slice(0, defaultCount);
  const hasMore = items.length > defaultCount;

  return (
    <div className={className}>
      {visible.map((item, i) => renderItem(item, i))}
      {hasMore && (
        <div className="text-center mt-3 pt-2 border-t border-gray-800">
          <p className="text-[9px] text-gray-500 mb-1">
            Showing {expanded ? `all ${items.length}` : `${defaultCount} of ${items.length}`} rows
          </p>
          <button
            onClick={() => setExpanded(!expanded)}
            className="text-[10px] text-blue-400 hover:text-blue-300 bg-gray-800 hover:bg-gray-700 px-3 py-1 rounded transition-colors"
          >
            {expanded ? 'Show fewer' : `Show all ${items.length}`}
          </button>
        </div>
      )}
    </div>
  );
}

/**
 * ExpandableTable — wraps table rows with expand/collapse.
 * Renders <tbody> rows only; caller provides <table> and <thead>.
 */
interface ExpandableTableProps<T> {
  items: T[];
  defaultCount?: number;
  renderRow: (item: T, index: number) => React.ReactNode;
}

export function ExpandableTable<T>({ items, defaultCount = 10, renderRow }: ExpandableTableProps<T>) {
  const [expanded, setExpanded] = useState(false);
  const visible = expanded ? items : items.slice(0, defaultCount);
  const hasMore = items.length > defaultCount;

  return (
    <>
      <tbody>
        {visible.map((item, i) => renderRow(item, i))}
      </tbody>
      {hasMore && (
        <tfoot>
          <tr>
            <td colSpan={100} className="text-center py-2 border-t border-gray-800">
              <p className="text-[9px] text-gray-500 mb-1">
                Showing {expanded ? `all ${items.length}` : `${defaultCount} of ${items.length}`} rows
              </p>
              <button
                onClick={() => setExpanded(!expanded)}
                className="text-[10px] text-blue-400 hover:text-blue-300 bg-gray-800 hover:bg-gray-700 px-3 py-1 rounded transition-colors"
              >
                {expanded ? 'Show fewer' : `Show all ${items.length}`}
              </button>
            </td>
          </tr>
        </tfoot>
      )}
    </>
  );
}
