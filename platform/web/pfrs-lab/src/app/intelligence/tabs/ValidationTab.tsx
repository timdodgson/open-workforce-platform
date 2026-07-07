'use client';

import Card from '@/components/Card';

export default function ValidationTab() {
  return (
    <div className="space-y-4">
      <Card title="Assist Validation — Can Search Intelligence Be Trusted?">
        <p className="text-xs text-gray-400 mb-4 leading-relaxed">
          Release-level validation evidence for Search Intelligence across all four domains.
          320 benchmark runs with statistical significance testing.
        </p>

        {/* Summary table */}
        <div className="overflow-x-auto mb-4">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase border-b border-gray-700">
                <th className="text-left p-2">Domain</th>
                <th className="text-left p-2">Algorithm</th>
                <th className="text-right p-2">Off Mean</th>
                <th className="text-right p-2">Adaptive Mean</th>
                <th className="text-right p-2">p-value</th>
                <th className="text-right p-2">Compute Saved</th>
                <th className="text-center p-2">Verdict</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              <ValidationRow domain="CVRP" alg="SA" off="802.8" adaptive="802.8" p="1.000" saved="60%" verdict="safe" />
              <ValidationRow domain="CVRP" alg="Portfolio" off="790.8" adaptive="790.8" p="1.000" saved="73%" verdict="safe" />
              <ValidationRow domain="JSS" alg="Tabu" off="666.0" adaptive="666.0" p="1.000" saved="41%" verdict="safe" />
              <ValidationRow domain="JSS" alg="Portfolio" off="682.3" adaptive="678.8" p="0.237" saved="23%" verdict="safe" />
              <ValidationRow domain="VRPTW" alg="SA" off="1137.6" adaptive="923.2" p="<0.001" saved="—" verdict="better" />
              <ValidationRow domain="VRPTW" alg="Portfolio" off="972.5" adaptive="861.3" p="<0.001" saved="—" verdict="better" />
              <ValidationRow domain="NRP" alg="SA" off="3943.5" adaptive="3922.0" p="0.183" saved="—" verdict="safe" />
              <ValidationRow domain="NRP" alg="Portfolio" off="3964.5" adaptive="3910.0" p="0.215" saved="—" verdict="safe" />
            </tbody>
          </table>
        </div>

        {/* Acceptance criteria */}
        <div className="bg-gray-800 rounded-lg p-4 mb-4">
          <p className="text-[10px] text-gray-500 uppercase mb-2">Acceptance Criteria</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-[10px]">
            <Criterion text="Adaptive not statistically worse than off" pass />
            <Criterion text="Assist/adaptive reduce compute (2+ domains)" pass />
            <Criterion text="Zero feasibility regressions" pass />
            <Criterion text="Zero missed best-known discoveries" pass />
            <Criterion text="Shadow behaviourally neutral" pass />
            <Criterion text="Results reproducible from documented commands" pass />
          </div>
        </div>

        {/* Known issues */}
        <div className="bg-gray-800 rounded-lg p-4 mb-4">
          <p className="text-[10px] text-gray-500 uppercase mb-2">Known Issues</p>
          <ul className="text-[10px] text-gray-400 space-y-1">
            <li>• JSS Portfolio: SA-bias heuristic causes 2/10 marginal degradations (fixed by learned model)</li>
            <li>• CVRP Portfolio a80k10: 1 seed marginally worse (+3, 0.2%) due to seed sensitivity</li>
            <li>• Not yet validated on instances larger than 100 customers / 30 nurses</li>
          </ul>
        </div>

        {/* Verdict */}
        <div className="bg-emerald-900/20 border border-emerald-700 rounded-lg p-4 text-center">
          <p className="text-emerald-400 font-bold text-sm">SAFE FOR RELEASE</p>
          <p className="text-[10px] text-gray-400 mt-1">Validated on tested configurations. Not claimed universal.</p>
        </div>
      </Card>

      <Card title="Full Reports">
        <p className="text-xs text-gray-500">
          Detailed validation reports available at{' '}
          <a href="/assist" className="text-blue-400 hover:underline">/assist</a>{' '}
          (per-decision telemetry analysis)
        </p>
      </Card>
    </div>
  );
}

function ValidationRow({ domain, alg, off, adaptive, p, saved, verdict }: {
  domain: string; alg: string; off: string; adaptive: string; p: string; saved: string; verdict: 'safe' | 'better';
}) {
  const verdictBadge = verdict === 'better'
    ? <span className="text-emerald-400">✅ BETTER</span>
    : <span className="text-gray-400">✅ SAFE</span>;
  return (
    <tr className="border-t border-gray-800">
      <td className="p-2 text-blue-400">{domain}</td>
      <td className="p-2">{alg}</td>
      <td className="text-right p-2">{off}</td>
      <td className="text-right p-2">{adaptive}</td>
      <td className="text-right p-2 text-amber-400">{p}</td>
      <td className="text-right p-2">{saved}</td>
      <td className="text-center p-2">{verdictBadge}</td>
    </tr>
  );
}

function Criterion({ text, pass }: { text: string; pass: boolean }) {
  return (
    <div className="flex items-center gap-2">
      <span className={pass ? 'text-emerald-400' : 'text-red-400'}>{pass ? '✓' : '✗'}</span>
      <span className="text-gray-300">{text}</span>
    </div>
  );
}
