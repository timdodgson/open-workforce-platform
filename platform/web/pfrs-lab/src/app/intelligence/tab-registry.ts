export type TabId =
  | 'overview'
  | 'learning'
  | 'model'
  | 'predictions'
  | 'decisions'
  | 'what-if'
  | 'validation'
  | 'policies'
  | 'si-validation'
  | 'continuous-learning'
  | 'promotion'
  | 'counterfactual';

export type IntelligenceSection =
  | 'summary'
  | 'learning'
  | 'decisions'
  | 'model'
  | 'assist'
  | 'policies'
  | 'continuous-learning'
  | 'promotion'
  | 'counterfactual';

export interface IntelligenceTab {
  id: TabId;
  label: string;
  icon: string;
  section?: IntelligenceSection;
  paginated?: boolean;
}

export const INTELLIGENCE_TABS: IntelligenceTab[] = [
  { id: 'overview', label: 'Overview', icon: '🧠' },
  { id: 'learning', label: 'Worker Learning', icon: '📊', section: 'learning', paginated: true },
  { id: 'continuous-learning', label: 'Policy Learning', icon: '🔄', section: 'continuous-learning' },
  { id: 'model', label: 'Model', icon: '🔬', section: 'model' },
  { id: 'predictions', label: 'Predictions', icon: '🧪' },
  { id: 'decisions', label: 'Decisions', icon: '🎯', section: 'decisions', paginated: true },
  { id: 'counterfactual', label: 'Counterfactual', icon: '🔀', section: 'counterfactual' },
  { id: 'what-if', label: 'What-If', icon: '⚗️' },
  { id: 'validation', label: 'Assist Val.', icon: '✅', section: 'assist' },
  { id: 'policies', label: 'Policies', icon: '📋', section: 'policies' },
  { id: 'promotion', label: 'Promotion', icon: '🚀', section: 'promotion' },
  { id: 'si-validation', label: 'SI Val.', icon: '🧪' },
];

export const VALID_TAB_IDS: TabId[] = INTELLIGENCE_TABS.map((t) => t.id);

export const TAB_SECTION: Partial<Record<TabId, IntelligenceSection>> = Object.fromEntries(
  INTELLIGENCE_TABS.filter((t) => t.section).map((t) => [t.id, t.section!]),
) as Partial<Record<TabId, IntelligenceSection>>;

export const PAGINATED_SECTIONS = new Set<IntelligenceSection>(
  INTELLIGENCE_TABS.filter((t) => t.paginated && t.section).map((t) => t.section!),
);

export function isValidTabId(tab: string | null): tab is TabId {
  return tab !== null && VALID_TAB_IDS.includes(tab as TabId);
}
