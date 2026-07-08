'use client';

import { useCallback, useEffect, useState } from 'react';
import type { PredictionsData } from './page.types';

interface FetchState {
  data: PredictionsData | null;
  loading: boolean;
  error: string | null;
  hasMore: boolean;
}

export function usePredictions(initialLimit = 2000) {
  const [state, setState] = useState<FetchState>({
    data: null,
    loading: true,
    error: null,
    hasMore: false,
  });

  const load = useCallback(async (limit: number) => {
    setState((s) => ({ ...s, loading: true, error: null }));
    try {
      const res = await fetch(`/api/predictions?limit=${limit}&offset=0`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      setState({
        data: {
          version: json.version,
          total_predictions: json.total_predictions,
          predictions: json.predictions,
        },
        loading: false,
        error: null,
        hasMore: Boolean(json.has_more),
      });
    } catch (e) {
      setState({
        data: null,
        loading: false,
        error: e instanceof Error ? e.message : 'Failed to load predictions',
        hasMore: false,
      });
    }
  }, []);

  useEffect(() => {
    load(initialLimit);
  }, [initialLimit, load]);

  return { ...state, reload: () => load(initialLimit) };
}
