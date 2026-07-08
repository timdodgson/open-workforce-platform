interface TabSpinnerProps {
  label?: string;
}

/** Consistent loading indicator for intelligence tabs and run pages. */
export default function TabSpinner({ label = 'Loading…' }: TabSpinnerProps) {
  return (
    <div className="flex items-center justify-center py-16">
      <div className="flex items-center gap-3">
        <div className="w-5 h-5 border-2 border-blue-400 border-t-transparent rounded-full animate-spin" />
        <span className="text-sm text-gray-400">{label}</span>
      </div>
    </div>
  );
}
