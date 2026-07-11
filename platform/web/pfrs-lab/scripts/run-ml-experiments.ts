/**
 * Step 8 — execute approved ML research experiments from queue.json.
 * Default is dry-run; pass --approve-ids=id1,id2 or --execute to run commands.
 *
 * Run from platform/web/pfrs-lab:
 *   npm run propose-ml-experiments
 *   npm run run-ml-experiments
 *   npm run run-ml-experiments -- --approve-ids=ml-exp-nrp-n012w8-sa-hybrid-s99
 */
import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';

type Proposal = {
  id: string;
  type: string;
  rationale: string;
  requires_approval: boolean;
  status: string;
  command: string;
  cwd: string;
  run_label?: string;
};

type ResearchQueue = {
  generated_at: string;
  human_approval_required: boolean;
  proposals: Proposal[];
  step8_loop_ok?: boolean;
  step8_promotion_ready?: boolean;
};

function repoRoot(): string {
  return resolve(process.cwd(), '..', '..', '..');
}

function queuePaths(): string[] {
  const root = repoRoot();
  return [
    join(root, 'docs', 'reports', 'ml-research', 'queue.json'),
    join(root, 'platform', 'ml', 'policies', 'research_queue.json'),
  ];
}

function loadQueue(): ResearchQueue | null {
  for (const p of queuePaths()) {
    if (!existsSync(p)) continue;
    try {
      return JSON.parse(readFileSync(p, 'utf-8')) as ResearchQueue;
    } catch {
      /* try next */
    }
  }
  return null;
}

function parseArgs() {
  const approveArg = process.argv.find((a) => a.startsWith('--approve-ids='));
  const approveIds = approveArg
    ? approveArg.split('=')[1].split(',').map((s) => s.trim()).filter(Boolean)
    : [];
  const executeAll = process.argv.includes('--execute');
  const dryRun = !executeAll && approveIds.length === 0;
  return { approveIds, executeAll, dryRun };
}

function runCommand(command: string, cwdRel: string): { ok: boolean; output: string } {
  const cwd = resolve(repoRoot(), cwdRel);
  const isWin = process.platform === 'win32';
  const result = spawnSync(command, {
    cwd,
    shell: true,
    encoding: 'utf-8',
    stdio: ['ignore', 'pipe', 'pipe'],
    env: process.env,
  });
  const output = `${result.stdout ?? ''}${result.stderr ?? ''}`.trim();
  return { ok: result.status === 0, output };
}

function saveQueue(queue: ResearchQueue) {
  const json = JSON.stringify(queue, null, 2);
  for (const p of queuePaths()) {
    try {
      writeFileSync(p, json);
      console.log(`Updated ${p}`);
    } catch {
      /* skip */
    }
  }
}

function main() {
  const { approveIds, executeAll, dryRun } = parseArgs();
  const queue = loadQueue();
  if (!queue) {
    console.error('No research queue found. Run: npm run propose-ml-experiments');
    process.exit(1);
  }

  const approved = new Set(approveIds);
  const toRun = queue.proposals.filter((p) => {
    if (p.type === 'retrain_policies') return false;
    if (executeAll) return p.requires_approval;
    return approved.has(p.id);
  });

  console.log(`ML research queue (${queue.proposals.length} proposals, dry-run=${dryRun})`);
  console.log(`  human_approval_required: ${queue.human_approval_required}`);
  console.log(`  step8_loop_ok: ${queue.step8_loop_ok ?? '—'}`);

  if (dryRun) {
    for (const p of queue.proposals) {
      console.log(`\n[${p.status}] ${p.id} (${p.type})`);
      console.log(`  ${p.rationale}`);
      console.log(`  cwd: ${p.cwd}`);
      console.log(`  ${p.command}`);
    }
    console.log('\nDry-run only. Approve with --approve-ids=<id> or --execute');
    return;
  }

  if (toRun.length === 0) {
    console.error('No matching proposals to execute.');
    process.exit(1);
  }

  for (const p of toRun) {
    console.log(`\nRunning ${p.id}...`);
    console.log(`  ${p.command}`);
    const { ok, output } = runCommand(p.command, p.cwd);
    p.status = ok ? 'completed' : 'failed';
    if (output) {
      const lines = output.split('\n').slice(-8);
      for (const line of lines) console.log(`  | ${line}`);
    }
    console.log(ok ? '  OK' : '  FAILED');
  }

  queue.generated_at = new Date().toISOString();
  saveQueue(queue);
}

main();
