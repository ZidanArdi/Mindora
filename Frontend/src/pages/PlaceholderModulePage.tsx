import { useLocation, Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import type { UserProfile } from '../types';
import { useAuth } from '../hooks';

interface ModuleConfig {
  title: string;
  subtitle: string;
  phase: string;
  badge: string;
  description: string;
  expectedFeatures: string[];
  icon: string;
}

const moduleConfigs: Record<string, ModuleConfig> = {
  '/analytics': {
    title: 'Cognitive & Habit Analytics',
    subtitle: 'Correlations between study duration, lifestyle variables, and distraction density.',
    phase: 'Phase 3',
    badge: 'Coming in Phase 3',
    description:
      'The Analytics module will provide deep statistical charts and longitudinal correlations across sleep, screen time, caffeine, and study consistency.',
    expectedFeatures: [
      'Circadian energy curve vs study performance',
      'Caffeine intake vs focus duration correlation',
      'Distraction frequency trends by time-of-day',
      'Weekly habit consistency radar chart',
    ],
    icon: '📊',
  },
  '/reports': {
    title: 'Productivity Reports & Export',
    subtitle: 'Comprehensive PDF summaries and CSV data portability.',
    phase: 'Phase 4',
    badge: 'Coming in Phase 4',
    description:
      'Generate academic milestone reports, verify study hours for scholarship requirements, and export your personal data safely.',
    expectedFeatures: [
      'Downloadable PDF monthly productivity dossiers',
      'Raw CSV export for personal research analysis',
      'Automated weekly email summaries',
      'Verified study proof certifications',
    ],
    icon: '📑',
  },
  '/recommendations': {
    title: 'AI Productivity Recommendations',
    subtitle: 'Machine learning suggestions based on your personal habit calibration.',
    phase: 'Phase 3',
    badge: 'Coming in Phase 3',
    description:
      'Once sufficient habit data is collected, the Machine Learning recommendation engine will suggest optimal study hours and sleep adjustments.',
    expectedFeatures: [
      'Tailored study session length recommendations',
      'Target bedtime notifications based on next-day study plan',
      'Screen-time intervention alerts prior to deep focus blocks',
      'Context-aware study break suggestions',
    ],
    icon: '✨',
  },
  '/profile': {
    title: 'Student Profile & Academic Context',
    subtitle: 'Manage your major, study goals, and academic semester schedule.',
    phase: 'Phase 4',
    badge: 'Coming in Phase 4',
    description:
      'Customize your student details, academic curriculum, exam periods, and notification preferences.',
    expectedFeatures: [
      'Academic semester & exam schedule manager',
      'Subject syllabus and credit weight mapping',
      'Profile avatar and student identity customization',
      'Session notification preferences',
    ],
    icon: '🎓',
  },
  '/settings': {
    title: 'Settings & Security',
    subtitle: 'Account authentication, session management, and privacy preferences.',
    phase: 'Phase 4',
    badge: 'Coming in Phase 4',
    description:
      'Configure password changes, 2FA security, active device session review, and data privacy policies.',
    expectedFeatures: [
      'Multi-factor authentication (MFA / 2FA)',
      'Active session inspection and remote logout',
      'Theme customization (Botanical Sage / Forest Night)',
      'Account export and privacy controls',
    ],
    icon: '⚙️',
  },
};

export default function PlaceholderModulePage() {
  const { user } = useAuth();
  const location = useLocation();

  const currentPath = location.pathname;
  const config = moduleConfigs[currentPath] || {
    title: 'Productivity Module',
    subtitle: 'This feature is currently under active development.',
    phase: 'Phase 3 / 4',
    badge: 'In Development',
    description: 'This module is scheduled for development in subsequent Mindora releases.',
    expectedFeatures: ['Deep feature integration', 'Responsive interface', 'Export capabilities'],
    icon: '🌱',
  };

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name.slice(0, 2).toUpperCase(),
        workspaceName: 'Academic Workspace',
      }
    : {
        id: 'usr_guest',
        name: 'Student',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'ST',
        workspaceName: 'Academic Workspace',
      };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="max-w-4xl mx-auto py-8 space-y-6">
        {/* Module Header Card */}
        <div className="relative overflow-hidden rounded-3xl bg-linear-to-br from-emerald-50/80 via-white to-teal-50/70 border border-emerald-200/70 p-8 shadow-soft-sm">
          {/* Subtle background decoration */}
          <div className="absolute right-0 bottom-0 pointer-events-none opacity-15">
            <svg width="240" height="200" viewBox="0 0 240 200" fill="none">
              <circle cx="180" cy="120" r="80" fill="#138F78" />
              <path d="M100 20 L150 150 L50 150 Z" fill="#123C32" />
            </svg>
          </div>

          <div className="relative z-10 space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-300">
              <span className="w-2 h-2 rounded-full bg-[#138F78] animate-ping" />
              <span>{config.badge}</span>
            </div>

            <div className="flex items-start gap-4">
              <div className="w-14 h-14 rounded-2xl bg-white border border-emerald-200/80 shadow-soft-xs flex items-center justify-center text-3xl shrink-0">
                {config.icon}
              </div>
              <div>
                <h1 className="text-2xl sm:text-3xl font-extrabold text-[#123C32] tracking-tight">
                  {config.title}
                </h1>
                <p className="text-sm text-[#526B60] mt-1">{config.subtitle}</p>
              </div>
            </div>

            <p className="text-sm text-text-secondary leading-relaxed pt-2 max-w-2xl">
              {config.description}
            </p>
          </div>
        </div>

        {/* Planned Capabilities Section */}
        <div className="bg-white rounded-3xl border border-border p-6 sm:p-8 shadow-soft-xs space-y-4">
          <h2 className="text-base font-bold text-[#123C32]">Planned Capabilities in {config.phase}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
            {config.expectedFeatures.map((feat, idx) => (
              <div
                key={idx}
                className="flex items-center gap-3 p-3.5 rounded-2xl bg-slate-50/80 border border-slate-200/60"
              >
                <span className="w-6 h-6 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold flex items-center justify-center shrink-0">
                  ✓
                </span>
                <span className="text-xs font-semibold text-text-primary">{feat}</span>
              </div>
            ))}
          </div>

          <div className="pt-4 border-t border-border/60 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <span className="text-xs text-text-secondary">
              Core productivity tracking (Tasks, Study, Focus, Lifestyle) is fully active in Phase 2.
            </span>
            <div className="flex items-center gap-3">
              <Link
                to="/dashboard"
                className="px-4 py-2 rounded-xl text-xs font-bold text-[#123C32] bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 transition-colors"
              >
                ← Return to Dashboard
              </Link>
              <Link
                to="/tasks"
                className="px-4 py-2 rounded-xl text-xs font-bold text-white bg-[#138F78] hover:bg-[#123C32] transition-colors shadow-soft-xs"
              >
                Open Tasks Manager →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}
