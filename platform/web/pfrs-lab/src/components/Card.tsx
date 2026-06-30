interface CardProps {
  title: string;
  children: React.ReactNode;
}

export default function Card({ title, children }: CardProps) {
  return (
    <div className="bg-gray-850 border border-gray-700 rounded-lg p-5 mb-4">
      <h3 className="text-sm font-semibold text-blue-400 mb-3">{title}</h3>
      {children}
    </div>
  );
}
