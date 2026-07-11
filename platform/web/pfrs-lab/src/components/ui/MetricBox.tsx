import MetricCard from '@/components/MetricCard';

interface MetricBoxProps {
  label: string;
  value: string | number;
  colour?: 'default' | 'emerald' | 'amber' | 'blue' | 'red';
}

const COLOUR_MAP = {
  default: 'default',
  emerald: 'green',
  amber: 'amber',
  blue: 'blue',
  red: 'red',
} as const satisfies Record<NonNullable<MetricBoxProps['colour']>, 'default' | 'green' | 'amber' | 'blue' | 'red'>;

/** @deprecated Use MetricCard — thin alias for legacy colour prop spelling. */
export default function MetricBox({ label, value, colour = 'default' }: MetricBoxProps) {
  return (
    <MetricCard
      label={label}
      value={String(value)}
      color={COLOUR_MAP[colour]}
    />
  );
}
