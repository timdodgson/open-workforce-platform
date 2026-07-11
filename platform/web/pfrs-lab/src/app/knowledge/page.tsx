import type { Metadata } from 'next';
import KnowledgeBase from './KnowledgeBase';

export const metadata: Metadata = {
  title: 'Knowledge Base',
  description: 'Research findings and experiment notes stored locally in your browser.',
};

export const dynamic = 'force-dynamic';

export default function KnowledgePage() {
  return <KnowledgeBase />;
}
