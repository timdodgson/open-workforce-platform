/**
 * Assistant / chat API configuration surfaced on Admin and used by /api/chat.
 * Prompt lives in assistant-prompt.md (repo); edit there and redeploy to change behaviour.
 */

import { readFileSync } from 'fs';
import { join } from 'path';
import { getLlmModelId, getLlmProviderName } from './complete';

/** Generation defaults sent on every Anthropic/Bedrock Messages call. */
export const ASSISTANT_MAX_TOKENS = 2048;
export const ASSISTANT_TEMPERATURE = 0.3;
export const ASSISTANT_RATE_LIMIT_PER_MINUTE = 20;
/** How many recent manifest runs are appended to the system prompt. */
export const ASSISTANT_MANIFEST_RUN_CONTEXT = 20;
export const ASSISTANT_PROMPT_PATH = 'src/lib/assistant-prompt.md';
export const ASSISTANT_ANTHROPIC_API_VERSION = '2023-06-01';

const FALLBACK_PROMPT =
  'You are a PFRS optimisation experiment planner. Help users design nurse rostering experiments.';

export function loadAssistantSystemPrompt(): string {
  try {
    return readFileSync(join(process.cwd(), ASSISTANT_PROMPT_PATH), 'utf-8');
  } catch {
    return FALLBACK_PROMPT;
  }
}

export interface AssistantConfigSnapshot {
  provider: string;
  model: string;
  maxTokens: number;
  temperature: number;
  rateLimitPerMinute: number;
  manifestRunContext: number;
  anthropicApiVersion: string;
  apiKeySource: string;
  promptPath: string;
  promptChars: number;
  promptLines: number;
  systemPrompt: string;
}

export function getAssistantConfigSnapshot(): AssistantConfigSnapshot {
  const provider = getLlmProviderName();
  const model = getLlmModelId(provider);
  const systemPrompt = loadAssistantSystemPrompt();
  const ssmPath = process.env.ANTHROPIC_API_KEY_SSM?.trim();
  const hasEnvKey = Boolean(process.env.ANTHROPIC_API_KEY?.trim());
  let apiKeySource = 'unset';
  if (provider === 'bedrock') {
    apiKeySource = 'AWS credentials (Bedrock IAM)';
  } else if (ssmPath) {
    apiKeySource = `SSM SecureString ${ssmPath}`;
  } else if (hasEnvKey) {
    apiKeySource = 'ANTHROPIC_API_KEY env';
  }

  return {
    provider,
    model,
    maxTokens: ASSISTANT_MAX_TOKENS,
    temperature: ASSISTANT_TEMPERATURE,
    rateLimitPerMinute: ASSISTANT_RATE_LIMIT_PER_MINUTE,
    manifestRunContext: ASSISTANT_MANIFEST_RUN_CONTEXT,
    anthropicApiVersion: ASSISTANT_ANTHROPIC_API_VERSION,
    apiKeySource,
    promptPath: ASSISTANT_PROMPT_PATH,
    promptChars: systemPrompt.length,
    promptLines: systemPrompt.split('\n').length,
    systemPrompt,
  };
}
