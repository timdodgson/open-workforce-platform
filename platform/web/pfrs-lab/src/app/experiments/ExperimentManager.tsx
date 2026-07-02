'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { RunInfo } from './page';

interface Experiment {
  id: string;
  title: string;
  hypothesis: string;
  description: string;
  algorithm: string;
  parameters: string;
  expectedOutcome: string;
  conclusion: string;
  result: 'pending' | 'supported' | 'refuted' | 'inconclusive';
  confidence: 'low' | 'moderate' | 'high';
  runIds: string[];
  tags: string[];
  notes: string;
  createdAt: string;
}

interface Props {
  runs: RunInfo[];
}

const STORAGE_KEY = 'pfrs-experiments';

function loadExperiments(): Experiment[] {
  if (typeof window === 'undefined') return [];
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? JSON.parse(stored) : [];
  } catch { return []; }
}

function saveExperiments(experiments: Experiment[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(experiments));
}

export default function ExperimentManager({ runs }: Props) {
  const [experiments, setExperiments] = useState<Experiment[]>(() => loadExperiments());
  const [selected, setSelected] = useState<Experiment | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [search, setSearch] = useState('');
  const [filterTag, setFilterTag] = useState<string | null>(null);
  const [filterResult, setFilterResult] = useState<string | null>(null);

  // All tags across experiments.
  const allTags = useMemo(() => {
    const tags = new Set<string>();
    for (const exp of experiments) {
      for (const t of exp.tags) tags.add(t);
    }
    return Array.from(tags).sort();
  }, [experiments]);

  // Filtered experiments.
  const filtered = useMemo(() => {
    return experiments.filter(exp => {
      if (search && !exp.title.toLowerCase().includes(search.toLowerCase()) &&
          !exp.hypothesis.toLowerCase().includes(search.toLowerCase()) &&
          !exp.tags.some(t => t.toLowerCase().includes(search.toLowerCase()))) return false;
      if (filterTag && !exp.tags.includes(filterTag)) return false;
      if (filterResult && exp.result !== filterResult) return false;
      return true;
    });
  }, [experiments, search, filterTag, filterResult]);

  function createExperiment(exp: Experiment) {
    const updated = [...experiments, exp];
    setExperiments(updated);
    saveExperiments(updated);
    setShowCreate(false);
  }

  function updateExperiment(exp: Experiment) {
    const updated = experiments.map(e => e.id === exp.id ? exp : e);
    setExperiments(updated);
    saveExperiments(updated);
    setSelected(exp);
  }

  function deleteExperiment(id: string) {
    const updated = experiments.filter(e => e.id !== id);
    setExperiments(updated);
    saveExperiments(updated);
    setSelected(null);
  }

  function resultColor(r: string): string {
    switch (r) {
      case 'supported': return 'text-emerald-400';
      case 'refuted': return 'text-red-400';
      case 'inconclusive': return 'text-amber-400';
      default: return 'text-gray-500';
    }
  }

  function resultIcon(r: string): string {
    switch (r) {
      case 'supported': return '✓';
      case 'refuted': return '✗';
      case 'inconclusive': return '?';
      default: return '○';
    }
  }

  return (
    <div className="space-y-4">
      {/* Header + controls */}
      <Card title="Experiment Manager">
        <div className="flex gap-2 items-center mb-3">
          <input type="text" placeholder="Search experiments..." value={search}
            onChange={e => setSearch(e.target.value)}
            className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-gray-200 placeholder-gray-600" />
          <button onClick={() => setShowCreate(true)}
            className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-sm font-medium">
            + New Experiment
          </button>
        </div>
        <div className="flex gap-2 flex-wrap">
          {allTags.length > 0 && (
            <div className="flex items-center gap-1">
              <span className="text-[10px] text-gray-500">Tags:</span>
              <button onClick={() => setFilterTag(null)}
                className={`px-2 py-0.5 rounded text-[10px] ${!filterTag ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>All</button>
              {allTags.map(t => (
                <button key={t} onClick={() => setFilterTag(t)}
                  className={`px-2 py-0.5 rounded text-[10px] ${filterTag === t ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{t}</button>
              ))}
            </div>
          )}
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Result:</span>
            {['pending', 'supported', 'refuted', 'inconclusive'].map(r => (
              <button key={r} onClick={() => setFilterResult(filterResult === r ? null : r)}
                className={`px-2 py-0.5 rounded text-[10px] ${filterResult === r ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{r}</button>
            ))}
          </div>
        </div>
      </Card>

      {/* Experiment list */}
      {filtered.length === 0 && !showCreate && (
        <Card title="No Experiments">
          <p className="text-gray-500 text-sm">Create your first experiment to start tracking hypotheses.</p>
        </Card>
      )}

      {filtered.map(exp => (
        <div key={exp.id} onClick={() => setSelected(exp)}
          className={`bg-gray-850 border rounded-lg p-4 cursor-pointer transition hover:border-blue-500 ${
            selected?.id === exp.id ? 'border-blue-500' : 'border-gray-700'
          }`}>
          <div className="flex items-start justify-between">
            <div>
              <h3 className="text-sm font-medium text-gray-200">{exp.title}</h3>
              <p className="text-[10px] text-gray-500 mt-0.5">{exp.hypothesis}</p>
            </div>
            <div className="flex items-center gap-2">
              <span className={`text-xs font-medium ${resultColor(exp.result)}`}>
                {resultIcon(exp.result)} {exp.result}
              </span>
              <span className="text-[10px] text-gray-600">{exp.runIds.length} runs</span>
            </div>
          </div>
          <div className="flex gap-1 mt-2">
            {exp.tags.map(t => (
              <span key={t} className="px-1.5 py-0.5 bg-gray-800 rounded text-[9px] text-gray-400">{t}</span>
            ))}
          </div>
        </div>
      ))}

      {/* Selected experiment detail */}
      {selected && !showCreate && (
        <ExperimentDetail
          experiment={selected}
          runs={runs}
          onUpdate={updateExperiment}
          onDelete={deleteExperiment}
          onClose={() => setSelected(null)}
        />
      )}

      {/* Create form */}
      {showCreate && (
        <CreateExperimentForm
          runs={runs}
          onCreate={createExperiment}
          onCancel={() => setShowCreate(false)}
        />
      )}
    </div>
  );
}

function ExperimentDetail({ experiment: exp, runs, onUpdate, onDelete, onClose }: {
  experiment: Experiment; runs: RunInfo[];
  onUpdate: (e: Experiment) => void; onDelete: (id: string) => void; onClose: () => void;
}) {
  const [editing, setEditing] = useState(false);
  const [notes, setNotes] = useState(exp.notes);
  const [conclusion, setConclusion] = useState(exp.conclusion);
  const [result, setResult] = useState(exp.result);

  const expRuns = runs.filter(r => exp.runIds.includes(r.id));
  const penalties = expRuns.map(r => r.totalPenalty).filter(p => p > 0);
  const mean = penalties.length > 0 ? penalties.reduce((s, p) => s + p, 0) / penalties.length : 0;
  const best = penalties.length > 0 ? Math.min(...penalties) : 0;
  const worst = penalties.length > 0 ? Math.max(...penalties) : 0;

  function save() {
    onUpdate({ ...exp, notes, conclusion, result });
    setEditing(false);
  }

  return (
    <Card title={exp.title}>
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-4 text-xs">
          <div><p className="text-gray-500">Hypothesis</p><p className="text-gray-200">{exp.hypothesis}</p></div>
          <div><p className="text-gray-500">Expected Outcome</p><p className="text-gray-200">{exp.expectedOutcome}</p></div>
          <div><p className="text-gray-500">Algorithm</p><p className="text-blue-400">{exp.algorithm}</p></div>
          <div><p className="text-gray-500">Parameters</p><p className="font-mono text-gray-300">{exp.parameters}</p></div>
        </div>

        {exp.description && (
          <div><p className="text-[10px] text-gray-500">Description</p><p className="text-xs text-gray-300">{exp.description}</p></div>
        )}

        {/* Run comparison within experiment */}
        {expRuns.length > 0 && (
          <div>
            <p className="text-[10px] text-gray-500 uppercase mb-1">Runs ({expRuns.length})</p>
            <div className="grid grid-cols-3 gap-2 mb-2 text-center">
              <div className="bg-gray-800 rounded p-2"><p className="text-lg font-bold text-emerald-400">{best.toLocaleString()}</p><p className="text-[9px] text-gray-500">Best</p></div>
              <div className="bg-gray-800 rounded p-2"><p className="text-lg font-bold text-gray-300">{mean.toFixed(0)}</p><p className="text-[9px] text-gray-500">Mean</p></div>
              <div className="bg-gray-800 rounded p-2"><p className="text-lg font-bold text-red-400">{worst.toLocaleString()}</p><p className="text-[9px] text-gray-500">Worst</p></div>
            </div>
            <div className="space-y-0.5 max-h-32 overflow-y-auto">
              {expRuns.map(r => (
                <div key={r.id} className="flex items-center gap-2 text-[10px] px-2 py-1 bg-gray-800/50 rounded">
                  <span className="text-blue-400 font-mono truncate flex-1">{r.id}</span>
                  <span className="text-gray-400">{r.totalPenalty.toLocaleString()}</span>
                  <span className="text-gray-600">{(r.totalDurationMs / 1000).toFixed(1)}s</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Result + conclusion */}
        <div className="border-t border-gray-700 pt-3">
          {editing ? (
            <div className="space-y-2">
              <div>
                <label className="text-[10px] text-gray-500">Result</label>
                <select value={result} onChange={e => setResult(e.target.value as Experiment['result'])}
                  className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs">
                  <option value="pending">Pending</option>
                  <option value="supported">Supported</option>
                  <option value="refuted">Refuted</option>
                  <option value="inconclusive">Inconclusive</option>
                </select>
              </div>
              <div>
                <label className="text-[10px] text-gray-500">Conclusion</label>
                <textarea value={conclusion} onChange={e => setConclusion(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs h-16" />
              </div>
              <div>
                <label className="text-[10px] text-gray-500">Notes</label>
                <textarea value={notes} onChange={e => setNotes(e.target.value)}
                  className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs h-16" />
              </div>
              <div className="flex gap-2">
                <button onClick={save} className="px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-[10px]">Save</button>
                <button onClick={() => setEditing(false)} className="px-3 py-1 bg-gray-700 rounded text-[10px] text-gray-400">Cancel</button>
              </div>
            </div>
          ) : (
            <div>
              <div className="flex items-center gap-2 mb-1">
                <span className={`text-sm font-medium ${exp.result === 'supported' ? 'text-emerald-400' : exp.result === 'refuted' ? 'text-red-400' : 'text-gray-400'}`}>
                  Result: {exp.result}
                </span>
                <span className="text-[10px] text-gray-600">Confidence: {exp.confidence}</span>
              </div>
              {exp.conclusion && <p className="text-xs text-gray-300 mb-1">{exp.conclusion}</p>}
              {exp.notes && <p className="text-[10px] text-gray-500 italic">{exp.notes}</p>}
              <div className="flex gap-2 mt-2">
                <button onClick={() => setEditing(true)} className="px-3 py-1 bg-gray-700 rounded text-[10px] text-gray-300">Edit</button>
                <button onClick={() => { if (confirm('Delete?')) onDelete(exp.id); }} className="px-3 py-1 bg-red-900 hover:bg-red-800 rounded text-[10px] text-red-300">Delete</button>
                <button onClick={onClose} className="px-3 py-1 bg-gray-800 rounded text-[10px] text-gray-500">Close</button>
              </div>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}

function CreateExperimentForm({ runs, onCreate, onCancel }: {
  runs: RunInfo[]; onCreate: (e: Experiment) => void; onCancel: () => void;
}) {
  const [title, setTitle] = useState('');
  const [hypothesis, setHypothesis] = useState('');
  const [description, setDescription] = useState('');
  const [algorithm, setAlgorithm] = useState('');
  const [parameters, setParameters] = useState('');
  const [expectedOutcome, setExpectedOutcome] = useState('');
  const [tags, setTags] = useState('');
  const [selectedRuns, setSelectedRuns] = useState<Set<string>>(new Set());

  function toggleRun(id: string) {
    setSelectedRuns(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  function submit() {
    if (!title.trim()) return;
    const exp: Experiment = {
      id: `exp-${Date.now()}`,
      title: title.trim(),
      hypothesis: hypothesis.trim(),
      description: description.trim(),
      algorithm: algorithm.trim(),
      parameters: parameters.trim(),
      expectedOutcome: expectedOutcome.trim(),
      conclusion: '',
      result: 'pending',
      confidence: 'low',
      runIds: Array.from(selectedRuns),
      tags: tags.split(',').map(t => t.trim()).filter(Boolean),
      notes: '',
      createdAt: new Date().toISOString(),
    };
    onCreate(exp);
  }

  return (
    <Card title="New Experiment">
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Title *</label>
            <input value={title} onChange={e => setTitle(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
              placeholder="Budget Strategy Investigation" />
          </div>
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Algorithm</label>
            <input value={algorithm} onChange={e => setAlgorithm(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
              placeholder="SA / LAHC" />
          </div>
        </div>

        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Hypothesis</label>
          <input value={hypothesis} onChange={e => setHypothesis(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
            placeholder="Budget diversity delays entropy collapse." />
        </div>

        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Description</label>
          <textarea value={description} onChange={e => setDescription(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs h-14"
            placeholder="Detailed description..." />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Parameters</label>
            <input value={parameters} onChange={e => setParameters(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
              placeholder="beam=5, budget, weight=1.0" />
          </div>
          <div>
            <label className="text-[10px] text-gray-500 block mb-0.5">Expected Outcome</label>
            <input value={expectedOutcome} onChange={e => setExpectedOutcome(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
              placeholder="Lower entropy collapse rate" />
          </div>
        </div>

        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Tags (comma-separated)</label>
          <input value={tags} onChange={e => setTags(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs"
            placeholder="diversity, beam-strategy, budget" />
        </div>

        {/* Run selector */}
        <div>
          <label className="text-[10px] text-gray-500 block mb-0.5">Assign Runs ({selectedRuns.size} selected)</label>
          <div className="max-h-32 overflow-y-auto border border-gray-700 rounded p-1 space-y-0.5">
            {runs.map(r => (
              <label key={r.id} className="flex items-center gap-2 px-2 py-0.5 rounded hover:bg-gray-800 cursor-pointer text-[10px]">
                <input type="checkbox" checked={selectedRuns.has(r.id)}
                  onChange={() => toggleRun(r.id)} className="rounded" />
                <span className="text-blue-400 font-mono truncate">{r.id}</span>
                <span className="text-gray-500 ml-auto">{r.totalPenalty.toLocaleString()}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex gap-2">
          <button onClick={submit} disabled={!title.trim()}
            className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-700 disabled:text-gray-500 text-white rounded text-xs font-medium">
            Create Experiment
          </button>
          <button onClick={onCancel}
            className="px-4 py-1.5 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded text-xs">
            Cancel
          </button>
        </div>
      </div>
    </Card>
  );
}
