interface MetricBoxProps {
  label: string;
  value: string | number;
  colour?: 'default' | 'emerald' | 'amber' | 'blue' | 'red';
}

const COLOUR: Record<NonNullable<MetricBoxProps['colour']>, string> = {
  default: 'text-gray-200',
  emerald: 'text-emerald-400',
  amber: 'text-amber-400',
  blue: 'text-blue-400',
  red: 'text-red-400',
};

/** Compact metric tile used across intelligence and run dashboards. */
export default function MetricBox({ label, value, colour = 'default' }: MetricBoxProps) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <p className="text-[9px] text-gray-500 uppercase tracking-wide">{label}</p>
      <p className={`text-lg font-bold mt-1 ${COLOUR[colour]}`}>{value}</p>
    </div>
  );
}
