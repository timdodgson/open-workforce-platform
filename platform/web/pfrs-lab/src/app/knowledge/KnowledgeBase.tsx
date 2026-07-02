'use client';

import { useState, useMemo, useEffect } from 'react';
import Card from '@/components/Card';

interface Finding {
  id: string;
  statement: string;
  evidence: string;
  confidence: 'high' | 'moderate' | 'low';
  runCount: number;
  supportingExperiments: string[];
  contradictingExperiments: string[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

interface Experiment {
  id: string;
  title: string;
  hypothesis: string;
  result: string;
  confidence: string;
  runIds: string[];
  tags: string[];
}

const FINDINGS_KEY = 'pfrs-knowledge-findings';
const EXPERIMENTS_KEY = 'pfrs-experiments';

function loadFindings(): Finding[] {
  if (typeof window === 'undefined') return [];
  try { return JSON.parse(localStorage.getItem(FINDINGS_KEY) || '[]'); } catch { return []; }
}

function saveFindings(findings: Finding[]) {
  localStorage.setItem(FINDINGS_KEY, JSON.stringify(findings));
}

function loadExperiments(): Experiment[] {
  if (typeof window === 'undefined') return [];
  try { return JSON.parse(localStorage.getItem(EXPERIMENTS_KEY) || '[]'); } catch { return []; }
}

export default function KnowledgeBase() {
  const [findings, setFindings] = useState<Finding[]>([]);
  const [experiments, setExperiments] = useState<Experiment[]>([]);
  const [search, setSearch] = useState('');
  const [filterConfidence, setFilterConfidence] = useState<string | null>(null);
  const [filterTag, setFilterTag] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [selected, setSelected] = useState<Finding | null>(null);

  useEffect(() => {
    setFindings(loadFindings());
    setExperiments(loadExperiments());
  }, []);

  const allTags = useMemo(() => {
    const tags = new Set<string>();
    for (const f of findings) for (const t of f.tags) tags.add(t);
    return Array.from(tags).sort();
  }, [findings]);

  const filtered = useMemo(() => {
    return findings.filter(f => {
      if (search && !f.statement.toLowerCase().includes(search.toLowerCase()) &&
          !f.tags.some(t => t.toLowerCase().includes(search.toLowerCase()))) return false;
      if (filterConfidence && f.confidence !== filterConfidence) return false;
      if (filterTag && !f.tags.includes(filterTag)) return false;
      return true;
    });
  }, [findings, search, filterConfidence, filterTag]);

  // Auto-derive findings from experiments.
  const derivedFindings = useMemo(() => {
    const derived: { statement: string; supporting: string[]; contradicting: string[]; runs: number; tags: string[] }[] = [];
    // Group experiments by similar hypothesis.
    const supported = experiments.filter(e => e.result === 'supported');
    for (const exp of supported) {
      const existing = derived.find(d => d.statement === exp.hypothesis);
      if (existing) {
        existing.supporting.push(exp.id);
        existing.runs += exp.runIds.length;
      } else {
        derived.push({
          statement: exp.hypothesis,
          supporting: [exp.id],
          contradicting: [],
          runs: exp.runIds.length,
          tags: exp.tags,
        });
      }
    }
    // Check for contradictions.
    const refuted = experiments.filter(e => e.result === 'refuted');
    for (const exp of refuted) {
      const match = derived.find(d => d.statement === exp.hypothesis);
      if (match) {
        match.contradicting.push(exp.id);
      }
    }
    return derived;
  }, [experiments]);

  function addFinding(f: Finding) {
    const updated = [...findings, f];
    setFindings(updated);
    saveFindings(updated);
    setShowCreate(false);
  }

  function deleteFinding(id: string) {
    const updated = findings.filter(f => f.id !== id);
    setFindings(updated);
    saveFindings(updated);
    setSelected(null);
  }

  function confidenceColor(c: string): string {
    switch (c) { case 'high': return 'text-emerald-400'; case 'moderate': return 'text-amber-400'; default: return 'text-gray-500'; }
  }

  function confidenceBg(c: string): string {
    switch (c) { case 'high': return 'bg-emerald-900/20 border-emerald-700'; case 'moderate': return 'bg-amber-900/20 border-amber-700'; default: return 'bg-gray-800 border-gray-700'; }
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title="Research Knowledge Base">
        <div className="flex gap-2 items-center mb-3">
          <input type="text" placeholder="Search findings..." value={search}
            onChange={e => setSearch(e.target.value)}
            className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-gray-200 placeholder-gray-600" />
          <button onClick={() => setShowCreate(true)}
            className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-sm font-medium">
            + Add Finding
          </button>
        </div>
        <div className="flex gap-2 flex-wrap">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Confidence:</span>
            {['high', 'moderate', 'low'].map(c => (
              <button key={c} onClick={() => setFilterConfidence(filterConfidence === c ? null : c)}
                className={`px-2 py-0.5 rounded text-[10px] ${filterConfidence === c ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{c}</button>
            ))}
          </div>
          {allTags.length > 0 && (
            <div className="flex items-center gap-1">
              <span className="text-[10px] text-gray-500">Tag:</span>
              {allTags.slice(0, 8).map(t => (
                <button key={t} onClick={() => setFilterTag(filterTag === t ? null : t)}
                  className={`px-2 py-0.5 rounded text-[10px] ${filterTag === t ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{t}</button>
              ))}
            </div>
          )}
        </div>
        <p className="text-[9px] text-gray-600 mt-2">{findings.length} findings, {experiments.length} experiments tracked</p>
      </Card>

      {/* Auto-derived findings from experiments */}
      {derivedFindings.length > 0 && (
        <Card title="Auto-Derived from Experiments">
          <div className="space-y-2">
            {derivedFindings.map((d, i) => (
              <div key={i} className={`border rounded p-3 ${d.contradicting.length > 0 ? 'border-amber-700 bg-amber-900/10' : 'border-emerald-700 bg-emerald-900/10'}`}>
                <p className="text-sm text-gray-200">{d.statement}</p>
                <div className="flex gap-3 mt-1 text-[10px]">
                  <span className="text-emerald-400">{d.supporting.length} supporting</span>
                  {d.contradicting.length > 0 && <span className="text-red-400">{d.contradicting.length} contradicting</span>}
                  <span className="text-gray-500">{d.runs} runs</span>
                  {d.tags.map(t => <span key={t} className="px-1 bg-gray-800 rounded text-gray-400">{t}</span>)}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Manual findings */}
      {filtered.length > 0 && (
        <Card title="Findings">
          <div className="space-y-2">
            {filtered.map(f => (
              <div key={f.id} onClick={() => setSelected(f)}
                className={`border rounded p-3 cursor-pointer transition hover:border-blue-500 ${confidenceBg(f.confidence)}`}>
                <div className="flex items-start justify-between">
                  <p className="text-sm text-gray-200">{f.statement}</p>
                  <span className={`text-[10px] px-2 py-0.5 rounded bg-gray-800 ${confidenceColor(f.confidence)}`}>
                    {f.confidence}
                  </span>
                </div>
                <div className="flex gap-3 mt-1 text-[10px] text-gray-500">
                  <span>{f.runCount} runs</span>
                  <span>{f.supportingExperiments.length} supporting</span>
                  {f.contradictingExperiments.length > 0 && (
                    <span className="text-red-400">{f.contradictingExperiments.length} contradicting</span>
                  )}
                  {f.tags.map(t => <span key={t} className="px-1 bg-gray-800 rounded">{t}</span>)}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Selected detail */}
      {selected && (
        <Card title="Finding Detail">
          <div className="space-y-2 text-xs">
            <div><p className="text-gray-500">Statement</p><p className="text-gray-200 font-medium">{selected.statement}</p></div>
            <div><p className="text-gray-500">Evidence</p><p className="text-gray-300">{selected.evidence}</p></div>
            <div className="grid grid-cols-3 gap-3">
              <div><p className="text-gray-500">Confidence</p><p className={confidenceColor(selected.confidence)}>{selected.confidence}</p></div>
              <div><p className="text-gray-500">Run Count</p><p>{selected.runCount}</p></div>
              <div><p className="text-gray-500">Last Updated</p><p className="text-gray-400">{new Date(selected.updatedAt).toLocaleDateString()}</p></div>
            </div>
            <div><p className="text-gray-500">Supporting</p><p className="text-emerald-400">{selected.supportingExperiments.join(', ') || 'None'}</p></div>
            <div><p className="text-gray-500">Contradicting</p><p className="text-red-400">{selected.contradictingExperiments.join(', ') || 'None'}</p></div>
          </div>
          <div className="flex gap-2 mt-3">
            <button onClick={() => deleteFinding(selected.id)} className="px-3 py-1 bg-red-900 hover:bg-red-800 rounded text-[10px] text-red-300">Delete</button>
            <button onClick={() => setSelected(null)} className="px-3 py-1 bg-gray-700 rounded text-[10px] text-gray-400">Close</button>
          </div>
        </Card>
      )}

      {/* Create form */}
      {showCreate && (
        <CreateFindingForm onSave={addFinding} onCancel={() => setShowCreate(false)} />
      )}

      {/* Empty state */}
      {findings.length === 0 && derivedFindings.length === 0 && !showCreate && (
        <Card title="Getting Started">
          <p className="text-gray-500 text-sm">No findings yet. Create experiments, run them, then document findings here.</p>
        </Card>
      )}
    </div>
  );
}

function CreateFindingForm({ onSave, onCancel }: { onSave: (f: Finding) => void; onCancel: () => void }) {
  const [statement, setStatement] = useState('');
  const [evidence, setEvidence] = useState('');
  const [confidence, setConfidence] = useState<'high' | 'moderate' | 'low'>('moderate');
  const [runCount, setRunCount] = useState('');
  const [supporting, setSupporting] = useState('');
  const [contradicting, setContradicting] = useState('');
  const [tags, setTags] = useState('');

  function submit() {
    if (!statement.trim()) return;
    const now = new Date().toISOString();
    onSave({
      id: `finding-${Date.now()}`,
      statement: statement.trim(),
      evidence: evidence.trim(),
      confidence,
      runCount: parseInt(runCount) || 0,
      supportingExperiments: supporting.split(',').map(s => s.trim()).filter(Boolean),
      contradictingExperiments: contradicting.split(',').map(s => s.trim()).filter(Boolean),
      tags: tags.split(',').map(t => t.trim()).filter(Boolean),
      createdAt: now,
      updatedAt: now,
    });
  }

  return (
    <Card title="Add Finding">
      <div className="space-y-2">
        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Finding Statement *</label>
          <input value={statement} onChange={e => setStatement(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
            placeholder="Budget Strategy delays entropy collapse." />
        </div>
        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Evidence</label>
          <textarea value={evidence} onChange={e => setEvidence(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs h-16"
            placeholder="Measured across 14 runs. Entropy remained above 1.5 until week 6 with budget, vs week 3 without." />
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Confidence</label>
            <select value={confidence} onChange={e => setConfidence(e.target.value as 'high'|'moderate'|'low')}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs">
              <option value="high">High</option>
              <option value="moderate">Moderate</option>
              <option value="low">Low</option>
            </select>
          </div>
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Run Count</label>
            <input value={runCount} onChange={e => setRunCount(e.target.value)} type="number"
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs" placeholder="14" />
          </div>
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Tags (comma-sep)</label>
            <input value={tags} onChange={e => setTags(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs" placeholder="diversity, budget" />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Supporting Experiments (IDs)</label>
            <input value={supporting} onChange={e => setSupporting(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs" placeholder="exp-1, exp-2" />
          </div>
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Contradicting Experiments</label>
            <input value={contradicting} onChange={e => setContradicting(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs" placeholder="None" />
          </div>
        </div>
        <div className="flex gap-2">
          <button onClick={submit} disabled={!statement.trim()}
            className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-700 text-white rounded text-xs">Add Finding</button>
          <button onClick={onCancel} className="px-4 py-1.5 bg-gray-700 text-gray-300 rounded text-xs">Cancel</button>
        </div>
      </div>
    </Card>
  );
}
