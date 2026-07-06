'use client';

import Card from '@/components/Card';

interface Props {
  tp: number;
  fp: number;
  tn: number;
  fn: number;
}

export default function ConfusionMatrix({ tp, fp, tn, fn }: Props) {
  const total = tp + fp + tn + fn;

  return (
    <Card title="Confusion Matrix">
      <p className="text-xs text-gray-500 mb-4">
        TP = correctly recommended skip (worker was useless). FP = recommended skip but worker improved (missed opportunity).
        TN = correctly recommended run (worker improved). FN = recommended run but worker was useless (wasted CPU).
      </p>
      <div className="flex justify-center">
        <div className="inline-block">
          {/* Header row */}
          <div className="grid grid-cols-[120px_120px_120px] gap-1 mb-1">
            <div />
            <div className="text-center text-[10px] text-gray-500 uppercase p-2">Actually Useless</div>
            <div className="text-center text-[10px] text-gray-500 uppercase p-2">Actually Useful</div>
          </div>
          {/* Predicted Skip row */}
          <div className="grid grid-cols-[120px_120px_120px] gap-1 mb-1">
            <div className="flex items-center justify-end pr-3 text-[10px] text-gray-500 uppercase">Predicted Skip</div>
            <Cell value={tp} label="TP" total={total} colour="emerald" />
            <Cell value={fp} label="FP" total={total} colour="red" />
          </div>
          {/* Predicted Run row */}
          <div className="grid grid-cols-[120px_120px_120px] gap-1">
            <div className="flex items-center justify-end pr-3 text-[10px] text-gray-500 uppercase">Predicted Run</div>
            <Cell value={fn} label="FN" total={total} colour="amber" />
            <Cell value={tn} label="TN" total={total} colour="emerald" />
          </div>
        </div>
      </div>
    </Card>
  );
}

function Cell({ value, label, total, colour }: { value: number; label: string; total: number; colour: string }) {
  const pct = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';
  const colourMap: Record<string, string> = {
    emerald: 'bg-emerald-900/30 border-emerald-700',
    red: 'bg-red-900/30 border-red-700',
    amber: 'bg-amber-900/30 border-amber-700',
  };
  const textMap: Record<string, string> = {
    emerald: 'text-emerald-400',
    red: 'text-red-400',
    amber: 'text-amber-400',
  };

  return (
    <div className={`border rounded p-3 text-center ${colourMap[colour]}`}>
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-xl font-bold ${textMap[colour]}`}>{value}</div>
      <div className="text-[10px] text-gray-500">{pct}%</div>
    </div>
  );
}
