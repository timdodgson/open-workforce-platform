interface StatusBadgeProps {
  label: string;
  tone?: 'neutral' | 'success' | 'warning' | 'danger' | 'info';
}

const TONE: Record<NonNullable<StatusBadgeProps['tone']>, string> = {
  neutral: 'border-gray-600 bg-gray-800/50 text-gray-400',
  success: 'border-emerald-500/40 bg-emerald-950/30 text-emerald-300',
  warning: 'border-amber-500/40 bg-amber-950/30 text-amber-300',
  danger: 'border-red-500/40 bg-red-950/30 text-red-300',
  info: 'border-blue-500/40 bg-blue-950/30 text-blue-300',
};

export default function StatusBadge({ label, tone = 'neutral' }: StatusBadgeProps) {
  return (
    <span className={`inline-block text-[10px] px-2 py-0.5 rounded border ${TONE[tone]}`}>
      {label}
    </span>
  );
}
