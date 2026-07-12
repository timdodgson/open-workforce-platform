/**
 * Build / deploy release metadata injected at deploy time.
 * Prefer CI-set NEXT_PUBLIC_* vars; fall back to "unknown" when running locally.
 */

export interface ReleaseInfo {
  version: string;
  gitSha: string;
  gitShaShort: string;
  deployedAt: string;
  llmProvider: string;
  llmModel: string;
}

export function getReleaseInfo(): ReleaseInfo {
  const version = process.env.NEXT_PUBLIC_APP_VERSION?.trim() || 'unknown';
  const gitSha = process.env.NEXT_PUBLIC_GIT_SHA?.trim() || 'unknown';
  const deployedAt = process.env.NEXT_PUBLIC_DEPLOYED_AT?.trim() || 'unknown';
  const llmProvider = (process.env.LLM_PROVIDER || 'anthropic').toLowerCase();
  const llmModel =
    llmProvider === 'bedrock'
      ? process.env.BEDROCK_MODEL_ID || 'eu.anthropic.claude-3-haiku-20240307-v1:0'
      : process.env.ANTHROPIC_MODEL_ID || 'claude-haiku-4-5-20251001';

  return {
    version,
    gitSha,
    gitShaShort: gitSha === 'unknown' ? 'unknown' : gitSha.slice(0, 7),
    deployedAt,
    llmProvider,
    llmModel,
  };
}
