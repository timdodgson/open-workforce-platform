import Card from '@/components/Card';

export default function DiversityPage() {
  return (
    <div>
      <Card title="Beam Spread">
        <Placeholder text="Beam spread analysis — how retained paths diverge over weeks." />
      </Card>
      <Card title="Hamming Distance">
        <Placeholder text="Parent→child Hamming distance analysis." detail="Requires tree-csv with roster fingerprints / Hamming distance." />
      </Card>
      <Card title="Structural Diversity">
        <Placeholder text="Roster structural diversity metrics across beam paths." />
      </Card>
      <Card title="Near-Duplicate Detection">
        <Placeholder text="Detection of near-duplicate rosters (Hamming distance < 5%) within the beam." />
      </Card>
    </div>
  );
}

function Placeholder({ text, detail }: { text: string; detail?: string }) {
  return (
    <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
      <p>{text}</p>
      {detail && <p className="text-xs mt-2 text-gray-600">{detail}</p>}
    </div>
  );
}
