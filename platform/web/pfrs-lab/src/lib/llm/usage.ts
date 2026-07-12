/**
 * Persist assistant token usage for Admin metering.
 * Stored at storage root as chat_usage.json (S3 or local data/).
 */

import type { StorageProvider } from '@/lib/storage';

export const CHAT_USAGE_FILENAME = 'chat_usage.json';
const MAX_RECENT = 40;

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
}

export interface ChatUsageDay {
  requests: number;
  inputTokens: number;
  outputTokens: number;
}

export interface ChatUsageRecent {
  at: string;
  inputTokens: number;
  outputTokens: number;
  model: string;
  provider: string;
  user?: string;
}

export interface ChatUsageStore {
  updatedAt: string;
  totals: ChatUsageDay;
  byDay: Record<string, ChatUsageDay>;
  recent: ChatUsageRecent[];
}

function emptyTotals(): ChatUsageDay {
  return { requests: 0, inputTokens: 0, outputTokens: 0 };
}

export function emptyChatUsage(): ChatUsageStore {
  return {
    updatedAt: new Date(0).toISOString(),
    totals: emptyTotals(),
    byDay: {},
    recent: [],
  };
}

export async function loadChatUsage(storage: StorageProvider): Promise<ChatUsageStore> {
  try {
    const raw = await storage.readRootFile(CHAT_USAGE_FILENAME);
    if (!raw) return emptyChatUsage();
    const parsed = JSON.parse(raw) as ChatUsageStore;
    return {
      updatedAt: parsed.updatedAt || emptyChatUsage().updatedAt,
      totals: {
        requests: Number(parsed.totals?.requests) || 0,
        inputTokens: Number(parsed.totals?.inputTokens) || 0,
        outputTokens: Number(parsed.totals?.outputTokens) || 0,
      },
      byDay: parsed.byDay && typeof parsed.byDay === 'object' ? parsed.byDay : {},
      recent: Array.isArray(parsed.recent) ? parsed.recent : [],
    };
  } catch {
    return emptyChatUsage();
  }
}

export async function recordChatUsage(
  storage: StorageProvider,
  entry: {
    usage: TokenUsage;
    model: string;
    provider: string;
    user?: string;
  },
): Promise<void> {
  const store = await loadChatUsage(storage);
  const now = new Date();
  const day = now.toISOString().slice(0, 10);
  const input = Math.max(0, Math.floor(entry.usage.inputTokens));
  const output = Math.max(0, Math.floor(entry.usage.outputTokens));

  store.totals.requests += 1;
  store.totals.inputTokens += input;
  store.totals.outputTokens += output;

  const dayBucket = store.byDay[day] ?? emptyTotals();
  dayBucket.requests += 1;
  dayBucket.inputTokens += input;
  dayBucket.outputTokens += output;
  store.byDay[day] = dayBucket;

  store.recent.unshift({
    at: now.toISOString(),
    inputTokens: input,
    outputTokens: output,
    model: entry.model,
    provider: entry.provider,
    user: entry.user,
  });
  store.recent = store.recent.slice(0, MAX_RECENT);
  store.updatedAt = now.toISOString();

  await storage.writeRootFile(CHAT_USAGE_FILENAME, JSON.stringify(store, null, 2));
}

/** Rough Anthropic list-price estimate for Haiku-class (USD). Override via env if needed. */
export function estimateCostUsd(inputTokens: number, outputTokens: number): number {
  const inPerM = Number(process.env.ANTHROPIC_INPUT_USD_PER_M) || 1;
  const outPerM = Number(process.env.ANTHROPIC_OUTPUT_USD_PER_M) || 5;
  return (inputTokens / 1_000_000) * inPerM + (outputTokens / 1_000_000) * outPerM;
}
