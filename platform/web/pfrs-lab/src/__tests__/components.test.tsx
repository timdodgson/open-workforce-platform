/**
 * Component rendering tests.
 * Verify pages render without crashing given valid data.
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';

describe('Card', () => {
  it('renders title and children', () => {
    render(<Card title="Test Card"><p>Content</p></Card>);
    expect(screen.getByText('Test Card')).toBeInTheDocument();
    expect(screen.getByText('Content')).toBeInTheDocument();
  });
});

describe('MetricCard', () => {
  it('renders label and value', () => {
    render(<MetricCard label="Penalty" value="3430" color="green" />);
    expect(screen.getByText('Penalty')).toBeInTheDocument();
    expect(screen.getByText('3430')).toBeInTheDocument();
  });
});
