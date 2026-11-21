import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import {
  MessageSquareIcon,
  SparklesIcon,
  ClockIcon,
  ShieldIcon,
  CheckIcon,
} from '../components/atoms/Icons';
import type { UserProfile } from '../types';
import { useAuth } from '../hooks';

export default function MessagesPage() {
  const { user } = useAuth();

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText:
          user.name
            .split(' ')
            .map((n) => n[0])
            .join('')
            .slice(0, 2)
            .toUpperCase() || 'ST',
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
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-amber-100 text-amber-800 border border-amber-200">
                Phase 4.3 · LLM Service Deferred
              </span>
              <span className="text-xs text-text-muted">Academic Communications & AI Guidance</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Messages & Conversations
            </h1>
            <p className="text-sm text-text-secondary">
              Productivity guidance, communication channels, and conversational assistance status.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Link to="/recommendations">
              <Button
                variant="primary"
                size="sm"
                className="gap-2 text-xs font-semibold bg-primary hover:bg-primary-hover text-white"
              >
                <SparklesIcon className="w-4 h-4 text-emerald-300" />
                View Active Recommendations
              </Button>
            </Link>
          </div>
        </div>

        {/* Technical Status & Architecture Transparency Card */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-2xl bg-amber-50 border border-amber-200 flex items-center justify-center shrink-0 text-amber-700">
              <ShieldIcon className="w-6 h-6" />
            </div>
            <div className="space-y-1.5 flex-1">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-bold text-primary-900">
                  Conversational AI Planning Engine — Deferred Status
                </h2>
                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-bold bg-amber-100 text-amber-900">
                  <span className="w-2 h-2 rounded-full bg-amber-600 animate-pulse" />
                  Service Not Connected
                </span>
              </div>
              <p className="text-xs text-text-secondary leading-relaxed">
                The conversational LLM assistant endpoint (originally planned as{' '}
                <code className="px-1.5 py-0.5 rounded bg-slate-100 font-mono text-[11px] text-text-primary">
                  POST /api/v1/ai/plan-day
                </code>
                ) is intentionally deferred in this release. Mindora follows an offline-first, privacy-preserving
                engineering policy: no external cloud LLM APIs (OpenAI, Gemini, Anthropic) or ungrounded generative
                chatbots are active, ensuring student activity data remains strictly confidential and protected against
                unverified hallucinated guidance.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Privacy Standard
              </span>
              <p className="text-text-secondary mt-0.5">
                Zero third-party LLM transmission; personal study logs remain on-premise.
              </p>
            </div>
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Active Alternative
              </span>
              <p className="text-text-secondary mt-0.5">
                Deterministic rule engine (REC-TASK, REC-SLEEP, REC-BREAK, REC-COLD, REC-STUDY).
              </p>
            </div>
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Simulated AI Policy
              </span>
              <p className="text-text-secondary mt-0.5">
                Zero fake responses or mock chatbots; all shown productivity data is 100% verified.
              </p>
            </div>
          </div>
        </div>

        {/* Two-Column Grid: Active Productivity Hub vs Future Peer Collaboration Roadmap */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Column 1: Active Guidance Tools (Working Today) */}
          <div className="p-6 rounded-2xl bg-gradient-to-br from-[#123C32] to-[#0A261F] text-white shadow-soft-md flex flex-col justify-between space-y-6">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-emerald-400/20 text-emerald-300 border border-emerald-400/30">
                  <SparklesIcon className="w-3 h-3 text-emerald-400" />
                  Production Ready (Phase 4.1)
                </span>
                <span className="text-[11px] font-mono text-white/60">Deterministic Engine</span>
              </div>

              <div>
                <h3 className="text-lg font-bold text-white">
                  Deterministic Recommendation Engine
                </h3>
                <p className="text-xs text-white/70 mt-1 leading-relaxed">
                  Instead of ungrounded probabilistic chatbots, Mindora evaluates your verified habits against
                  rigorous mathematical rules to generate personalized recommendations:
                </p>
              </div>

              <div className="space-y-2.5 text-xs text-white/90">
                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Overdue Task Interventions:</strong> Real-time alerts for
                    high-priority deadlines.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Habit Calibration Prompts:</strong> Cold-start guidance when daily
                    sleep or screen time logs are unrecorded.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Focus-to-Break Balancing:</strong> Pomodoro fatigue recovery alerts
                    when focus exceeds 100m with insufficient breaks.
                  </div>
                </div>
              </div>
            </div>

            <div className="pt-4 border-t border-white/10">
              <Link to="/recommendations" className="block">
                <Button
                  variant="primary"
                  size="md"
                  className="w-full justify-center bg-emerald-500 hover:bg-emerald-600 text-slate-900 font-bold border-none"
                >
                  Explore Recommendations Hub
                </Button>
              </Link>
            </div>
          </div>

          {/* Column 2: Peer Collaboration & Study Groups Roadmap */}
          <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between space-y-6">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-slate-100 text-text-muted">
                  <MessageSquareIcon className="w-3 h-3 text-text-muted" />
                  Future Milestone
                </span>
                <span className="text-[11px] font-mono text-text-muted">Peer Channels</span>
              </div>

              <div>
                <h3 className="text-lg font-bold text-primary-900">
                  Study Groups & Peer Collaboration
                </h3>
                <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                  The peer-to-peer messaging architecture is designated for a multi-user collaborative release.
                  Planned capabilities include:
                </p>
              </div>

              <div className="space-y-2.5 text-xs text-text-secondary">
                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    1
                  </span>
                  <div>
                    <strong className="text-text-primary block">Course-Based Study Channels</strong>
                    Discussions and resource sharing organized by academic subjects and courses.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    2
                  </span>
                  <div>
                    <strong className="text-text-primary block">Shared Pomodoro Focus Rooms</strong>
                    Synchronized focus intervals and accountability timers with study partners.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    3
                  </span>
                  <div>
                    <strong className="text-text-primary block">Direct Academic Inquiries</strong>
                    Private messaging between enrolled students and course teaching assistants.
                  </div>
                </div>
              </div>
            </div>

            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200 text-xs text-text-muted flex items-center justify-between">
              <span>Status: Planned for Collaborative Framework</span>
              <span className="font-semibold text-text-secondary">Offline</span>
            </div>
          </div>
        </div>

        {/* Empty / Deferred Channel State Notice */}
        <div className="p-10 rounded-2xl bg-white border border-border shadow-soft-xs text-center space-y-3">
          <div className="w-12 h-12 rounded-2xl bg-slate-100 text-text-muted flex items-center justify-center mx-auto">
            <MessageSquareIcon className="w-6 h-6" />
          </div>
          <h3 className="text-sm font-bold text-text-primary">
            No Active Messaging Conversations
          </h3>
          <p className="text-xs text-text-muted max-w-md mx-auto leading-relaxed">
            Messaging channels are offline while external conversational LLM services remain deferred. To plan your
            day and optimize your academic workload, use Mindora’s verified productivity modules.
          </p>
          <div className="pt-2 flex flex-wrap items-center justify-center gap-3">
            <Link to="/study">
              <Button variant="secondary" size="sm" className="text-xs font-bold gap-1.5">
                <ClockIcon className="w-3.5 h-3.5" />
                Study Tracking
              </Button>
            </Link>
            <Link to="/focus">
              <Button variant="outline" size="sm" className="text-xs font-bold">
                Focus Timer
              </Button>
            </Link>
            <Link to="/reports">
              <Button variant="outline" size="sm" className="text-xs font-bold">
                Productivity Reports
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}
