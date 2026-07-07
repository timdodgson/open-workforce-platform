import { render, screen, fireEvent } from '@testing-library/react';
import ExpandableList, { ExpandableTable } from '@/components/ExpandableList';

describe('ExpandableList', () => {
  it('shows all items when <= defaultCount', () => {
    const items = ['a', 'b', 'c'];
    render(
      <ExpandableList items={items} renderItem={(item, i) => <div key={i}>{item}</div>} />
    );
    expect(screen.getByText('a')).toBeDefined();
    expect(screen.getByText('b')).toBeDefined();
    expect(screen.getByText('c')).toBeDefined();
    // No expand button.
    expect(screen.queryByText(/Show all/)).toBeNull();
  });

  it('shows expand button when > defaultCount', () => {
    const items = Array.from({ length: 15 }, (_, i) => `item-${i}`);
    render(
      <ExpandableList items={items} defaultCount={10} renderItem={(item, i) => <div key={i}>{item}</div>} />
    );
    // First 10 visible.
    expect(screen.getByText('item-0')).toBeDefined();
    expect(screen.getByText('item-9')).toBeDefined();
    // 11th not visible.
    expect(screen.queryByText('item-10')).toBeNull();
    // Button shows.
    expect(screen.getByText('Show all 15')).toBeDefined();
  });

  it('expands all rows on click', () => {
    const items = Array.from({ length: 15 }, (_, i) => `item-${i}`);
    render(
      <ExpandableList items={items} defaultCount={10} renderItem={(item, i) => <div key={i}>{item}</div>} />
    );
    fireEvent.click(screen.getByText('Show all 15'));
    // All items now visible.
    expect(screen.getByText('item-14')).toBeDefined();
    // Button changes.
    expect(screen.getByText('Show fewer')).toBeDefined();
  });

  it('collapses back to defaultCount', () => {
    const items = Array.from({ length: 15 }, (_, i) => `item-${i}`);
    render(
      <ExpandableList items={items} defaultCount={10} renderItem={(item, i) => <div key={i}>{item}</div>} />
    );
    fireEvent.click(screen.getByText('Show all 15'));
    fireEvent.click(screen.getByText('Show fewer'));
    // Back to 10.
    expect(screen.queryByText('item-10')).toBeNull();
    expect(screen.getByText('Show all 15')).toBeDefined();
  });
});
