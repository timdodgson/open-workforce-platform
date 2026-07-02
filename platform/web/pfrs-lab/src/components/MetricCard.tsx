interface MetricCardProps {
  label: string;
  value: string;
  color?: 'green' | 'blue' | 'amber' | 'red' | 'default';
}

const COLOR_MAP = {
  green: 'text-emerald-400',
  blue: 'text-blue-400',
  amber: 'text-amber-400',
  red: 'text-red-400',
  default: 'text-white',
};

export default function MetricCard({ label, value, color = 'default' }: MetricCardProps) {
  return (
    <div className="bg-gray-800/80 rounded-md p-3">
      <div className="text-[10px] uppercase tracking-wider text-gray-500">{label}</div>
      <div className={`text-lg font-bold mt-1 ${COLOR_MAP[color]}`}>{value}</div>
    </div>
  );
}
