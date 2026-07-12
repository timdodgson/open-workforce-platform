/**
 * LLM backends for the research assistant (/api/chat).
 * Default: Anthropic Messages API. Bedrock remains available when AWS enables it.
 */

import type { TokenUsage } from './usage';

export type LlmProviderName = 'anthropic' | 'bedrock';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface LlmCompleteInput {
  system: string;
  messages: ChatMessage[];
  maxTokens?: number;
  temperature?: number;
}

export interface LlmCompleteResult {
  text: string;
  usage: TokenUsage;
  model: string;
  provider: LlmProviderName;
}

function resolveProvider(): LlmProviderName {
  const raw = (process.env.LLM_PROVIDER || 'anthropic').toLowerCase().trim();
  if (raw === 'bedrock') return 'bedrock';
  return 'anthropic';
}

export function getLlmProviderName(): LlmProviderName {
  return resolveProvider();
}

export function getLlmModelId(provider: LlmProviderName = resolveProvider()): string {
  if (provider === 'bedrock') {
    return process.env.BEDROCK_MODEL_ID ?? 'eu.anthropic.claude-3-haiku-20240307-v1:0';
  }
  return process.env.ANTHROPIC_MODEL_ID || 'claude-haiku-4-5-20251001';
}

function parseUsage(raw: unknown): TokenUsage {
  const u = raw as { input_tokens?: number; output_tokens?: number; inputTokens?: number; outputTokens?: number } | null;
  return {
    inputTokens: Number(u?.input_tokens ?? u?.inputTokens) || 0,
    outputTokens: Number(u?.output_tokens ?? u?.outputTokens) || 0,
  };
}

async function completeAnthropic(input: LlmCompleteInput): Promise<LlmCompleteResult> {
  const { getAnthropicApiKey } = await import('./secrets');
  const apiKey = await getAnthropicApiKey();
  const model = getLlmModelId('anthropic');

  const res = await fetch('https://api.anthropic.com/v1/messages', {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      'x-api-key': apiKey,
      'anthropic-version': '2023-06-01',
    },
    body: JSON.stringify({
      model,
      max_tokens: input.maxTokens ?? 2048,
      temperature: input.temperature ?? 0.3,
      system: input.system,
      messages: input.messages.map((m) => ({
        role: m.role,
        content: m.content,
      })),
    }),
  });

  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const detail =
      typeof body?.error?.message === 'string'
        ? body.error.message
        : typeof body?.message === 'string'
          ? body.message
          : res.statusText;
    throw new Error(`Anthropic API ${res.status}: ${detail}`);
  }

  const text = body?.content?.find((b: { type?: string }) => b.type === 'text')?.text;
  return {
    text: typeof text === 'string' && text.length > 0 ? text : 'No response from model.',
    usage: parseUsage(body?.usage),
    model,
    provider: 'anthropic',
  };
}

async function completeBedrock(input: LlmCompleteInput): Promise<LlmCompleteResult> {
  const model = getLlmModelId('bedrock');
  const region = process.env.AWS_REGION ?? 'eu-west-1';

  // Dynamic require — only loads when Bedrock is selected.
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { BedrockRuntimeClient, ConverseCommand } = require('@aws-sdk/client-bedrock-runtime');

  const client = new BedrockRuntimeClient({ region });
  const command = new ConverseCommand({
    modelId: model,
    system: [{ text: input.system }],
    messages: input.messages.map((m) => ({
      role: m.role,
      content: [{ text: m.content }],
    })),
    inferenceConfig: {
      maxTokens: input.maxTokens ?? 2048,
      temperature: input.temperature ?? 0.3,
    },
  });

  const response = await client.send(command);
  const text = response.output?.message?.content?.[0]?.text ?? 'No response from model.';
  const usage = response.usage
    ? {
        inputTokens: Number(response.usage.inputTokens) || 0,
        outputTokens: Number(response.usage.outputTokens) || 0,
      }
    : { inputTokens: 0, outputTokens: 0 };

  return {
    text,
    usage,
    model,
    provider: 'bedrock',
  };
}

export async function completeChat(input: LlmCompleteInput): Promise<LlmCompleteResult> {
  const provider = resolveProvider();
  if (provider === 'bedrock') {
    return completeBedrock(input);
  }
  return completeAnthropic(input);
}
