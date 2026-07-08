'use client';

import { useCallback, useEffect, useState } from 'react';
import type { PredictionsData, WorkerPrediction } from './page.types';

interface FetchState {
  data: PredictionsData | null;
  loading: boolean;
  loadingMore: boolean;
  error: string | null;
  hasMore: boolean;
  offset: number;
}

const DEFAULT_PAGE = 500;

export function usePredictions(initialLimit = DEFAULT_PAGE) {
  const [state, setState] = useState<FetchState>({
    data: null,
    loading: true,
    loadingMore: false,
    error: null,
    hasMore: false,
    offset: 0,
  });

  const fetchPage = useCallback(async (limit: number, offset: number, append: boolean) => {
    setState((s) => ({
      ...s,
      loading: !append,
      loadingMore: append,
      error: null,
    }));
    try {
      const res = await fetch(`/api/predictions?limit=${limit}&offset=${offset}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      const predictions = json.predictions as WorkerPrediction[];
      setState((s) => ({
        data: append && s.data
          ? {
              version: json.version,
              total_predictions: json.total_predictions,
              predictions: [...s.data.predictions, ...predictions],
            }
          : {
              version: json.version,
              total_predictions: json.total_predictions,
              predictions,
            },
        loading: false,
        loadingMore: false,
        error: null,
        hasMore: Boolean(json.has_more),
        offset: offset + predictions.length,
      }));
    } catch (e) {
      setState((s) => ({
        ...s,
        loading: false,
        loadingMore: false,
        error: e instanceof Error ? e.message : 'Failed to load predictions',
      }));
    }
  }, []);

  useEffect(() => {
    fetchPage(initialLimit, 0, false);
  }, [initialLimit, fetchPage]);

  const loadMore = useCallback(() => {
    if (state.loadingMore || !state.hasMore) return;
    fetchPage(initialLimit, state.offset, true);
  }, [fetchPage, initialLimit, state.hasMore, state.loadingMore, state.offset]);

  return {
    data: state.data,
    loading: state.loading,
    loadingMore: state.loadingMore,
    error: state.error,
    hasMore: state.hasMore,
    reload: () => fetchPage(initialLimit, 0, false),
    loadMore,
  };
}
