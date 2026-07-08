import PredictionsPageClient from './PredictionsPageClient';

export const dynamic = 'force-dynamic';

export { type WorkerPrediction, type PredictionsData } from './page.types';

export default function PredictionsPage() {
  return <PredictionsPageClient />;
}
