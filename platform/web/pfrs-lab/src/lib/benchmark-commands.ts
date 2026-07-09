import type { PolicyMode, ProblemType } from '@/lib/benchmark-suites';

const POLICY_DIR = '../ml/policies';

export type CommandTier = 'fast' | 'deep';

interface CommandSpec {
  tier: CommandTier;
  domain: ProblemType;
  mode: string;
  instanceSlug: string;
  labelPrefix: string;
  build: (policy: PolicyMode, seed: number) => string;
}

const COMMAND_SPECS: CommandSpec[] = [
  // --- fast (validate-si2.ps1) ---
  {
    tier: 'fast',
    domain: 'cvrp',
    mode: 'sa',
    instanceSlug: 'a32k5',
    labelPrefix: 'val-cvrp',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode sa --iterations 500000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-cvrp-a32k5-sa-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'cvrp',
    mode: 'portfolio',
    instanceSlug: 'a32k5',
    labelPrefix: 'val-cvrp',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --mode portfolio --iterations 500000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-cvrp-a32k5-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'jss',
    mode: 'tabu',
    instanceSlug: 'la01',
    labelPrefix: 'val-jss',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --mode tabu --iterations 100000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-jss-la01-tabu-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'jss',
    mode: 'portfolio',
    instanceSlug: 'la01',
    labelPrefix: 'val-jss',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --mode portfolio --iterations 100000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-jss-la01-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'vrptw',
    mode: 'sa',
    instanceSlug: 'c101',
    labelPrefix: 'val-vrptw',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode sa --iterations 100000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-vrptw-c101-sa-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'vrptw',
    mode: 'portfolio',
    instanceSlug: 'c101',
    labelPrefix: 'val-vrptw',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode portfolio --iterations 100000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-vrptw-c101-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'nrp',
    mode: 'sa',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-nrp',
    build: (policy, seed) =>
      `go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --seeds ${seed} --worker-decision-mode assist --policy-mode ${policy} --policy-dir ${POLICY_DIR} --pfrs-run-label val-nrp-n012w8-sa-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'fast',
    domain: 'nrp',
    mode: 'portfolio',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-nrp',
    build: (policy, seed) =>
      `go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 16 --seeds ${seed} --worker-decision-mode assist --policy-mode ${policy} --policy-dir ${POLICY_DIR} --pfrs-run-label val-nrp-n012w8-portfolio-${policy}-s${seed} --storage s3`,
  },
  // --- deep (validate-si2-deep.ps1) ---
  {
    tier: 'deep',
    domain: 'cvrp',
    mode: 'sa',
    instanceSlug: 'a80k10',
    labelPrefix: 'val-deep-cvrp',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --mode sa --iterations 5000000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-cvrp-a80k10-sa-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'cvrp',
    mode: 'portfolio',
    instanceSlug: 'a80k10',
    labelPrefix: 'val-deep-cvrp',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --mode portfolio --iterations 2000000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-cvrp-a80k10-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'jss',
    mode: 'tabu',
    instanceSlug: 'ft10',
    labelPrefix: 'val-deep-jss',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --mode tabu --iterations 1000000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-jss-ft10-tabu-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'jss',
    mode: 'portfolio',
    instanceSlug: 'ft10',
    labelPrefix: 'val-deep-jss',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --mode portfolio --iterations 500000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-jss-ft10-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'vrptw',
    mode: 'sa',
    instanceSlug: 'c101',
    labelPrefix: 'val-deep-vrptw',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode sa --iterations 2000000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-vrptw-c101-sa-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'vrptw',
    mode: 'portfolio',
    instanceSlug: 'c101',
    labelPrefix: 'val-deep-vrptw',
    build: (policy, seed) =>
      `go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --mode portfolio --iterations 1000000 --policy-mode ${policy} --policy-dir ${POLICY_DIR} --seed ${seed} --run-label val-deep-vrptw-c101-portfolio-${policy}-s${seed} --storage s3`,
  },
  {
    tier: 'deep',
    domain: 'nrp',
    mode: 'sa',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-deep-nrp',
    build: (policy, seed) =>
      `go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode sa --pfrs-iterations-per-worker 300000 --pfrs-max-total-workers 32 --seeds ${seed} --worker-decision-mode assist --policy-mode ${policy} --policy-dir ${POLICY_DIR} --pfrs-run-label val-deep-nrp-n012w8-sa-${policy}-s${seed} --pfrs-storage s3`,
  },
  {
    tier: 'deep',
    domain: 'nrp',
    mode: 'portfolio',
    instanceSlug: 'n012w8',
    labelPrefix: 'val-deep-nrp',
    build: (policy, seed) =>
      `go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-mode portfolio --pfrs-iterations-per-worker 300000 --pfrs-max-total-workers 32 --seeds ${seed} --worker-decision-mode assist --policy-mode ${policy} --policy-dir ${POLICY_DIR} --pfrs-run-label val-deep-nrp-n012w8-portfolio-${policy}-s${seed} --pfrs-storage s3`,
  },
];

export function buildSiGoCommand(
  domain: string,
  mode: string,
  seed: number,
  policy: PolicyMode,
  tier: CommandTier = 'fast',
): string | null {
  const spec = COMMAND_SPECS.find(
    (s) => s.tier === tier && s.domain === domain && s.mode === mode,
  );
  return spec ? spec.build(policy, seed) : null;
}

export function commandsForDomain(
  domain: ProblemType,
  tier: CommandTier,
): Array<{ mode: string; example: string }> {
  return COMMAND_SPECS.filter((s) => s.tier === tier && s.domain === domain).map((s) => ({
    mode: s.mode,
    example: s.build('hybrid', 42),
  }));
}

export const GO_CWD = 'cd platform/go';

export const BATCH_SCRIPTS = {
  fast: 'powershell -ExecutionPolicy Bypass -File .\\scripts\\validate-si2.ps1',
  deep: 'powershell -ExecutionPolicy Bypass -File .\\scripts\\validate-si2-deep.ps1',
  retrain: 'powershell -ExecutionPolicy Bypass -File .\\scripts\\retrain-si2-policies.ps1',
} as const;
