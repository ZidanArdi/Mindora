import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import {
  SettingsIcon,
  ShieldIcon,
  ClockIcon,
  ExportIcon,
  LogoutIcon,
} from '../components/atoms/Icons';
import type { UserProfile } from '../types';
import { useAuth } from '../hooks';

export default function SettingsPage() {
  const { user, logout } = useAuth();

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
        workspaceName: user.role === 'admin' ? 'System Administration' : 'Academic Workspace',
      }
    : {
        id: 'usr_guest',
        name: 'Student',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'ST',
        workspaceName: 'Academic Workspace',
      };

  const detectedTimezone = typeof Intl !== 'undefined'
    ? Intl.DateTimeFormat().resolvedOptions().timeZone
    : 'UTC';

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-4xl mx-auto">
        {/* Header Section */}
        <div className="border-b border-border pb-5">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
              System Configuration
            </span>
            <span className="text-xs text-text-muted">Preferences & Security</span>
          </div>
          <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
            Application Settings
          </h1>
          <p className="text-sm text-text-secondary">
            Timezone calibration, session security, and data portability controls.
          </p>
        </div>

        {/* Section 1: System Timezone Alignment */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ClockIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Timezone Calibration</h2>
              <p className="text-xs text-text-muted">
                Governs date boundaries for Study Tracking, Pomodoro Focus Intervals, and Productivity Reports.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 text-xs">
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Detected Browser Timezone (IANA)
              </span>
              <span className="font-mono text-sm font-bold text-primary mt-1 block">
                {detectedTimezone || 'UTC'}
              </span>
              <p className="text-text-secondary text-[11px] mt-0.5">
                All daily metrics and 7-day trend calculations are aggregated relative to midnight in this location.
              </p>
            </div>
            <span className="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 self-start sm:self-center">
              Synchronized
            </span>
          </div>
        </div>

        {/* Section 2: Data Portability & Reports */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ExportIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Data Portability & Reports</h2>
              <p className="text-xs text-text-muted">
                Export your verified academic sessions, focus logs, and habit records.
              </p>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl bg-slate-50 border border-slate-200/80 text-xs">
            <div>
              <strong className="text-text-primary block">Comprehensive Productivity Dossier</strong>
              <p className="text-text-secondary mt-0.5">
                Generate structured CSV exports and print-friendly PDF summaries for academic milestone audits.
              </p>
            </div>
            <Link to="/reports" className="shrink-0">
              <Button variant="outline" size="sm" className="text-xs font-bold gap-2">
                <ExportIcon className="w-3.5 h-3.5" />
                Go to Reports Hub
              </Button>
            </Link>
          </div>
        </div>

        {/* Section 3: Authentication & Session Security */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ShieldIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Session Security & Access</h2>
              <p className="text-xs text-text-muted">
                Cryptographic authentication state and session termination controls.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Token Architecture
              </span>
              <span className="text-text-secondary mt-1 block">
                Dual-token JWT authentication with short-lived 15-minute access tokens and encrypted HttpOnly session rotation.
              </span>
            </div>

            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Active Session Status
              </span>
              <div className="flex items-center gap-2 mt-1">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                <span className="font-bold text-text-primary">Authenticated</span>
                <span className="text-[11px] text-text-muted">({displayUser.email})</span>
              </div>
            </div>
          </div>

          <div className="pt-2 flex justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={logout}
              className="text-xs font-bold text-rose-600 border-rose-200 hover:bg-rose-50 gap-2"
            >
              <LogoutIcon className="w-4 h-4 text-rose-500" />
              Sign Out of Current Session
            </Button>
          </div>
        </div>

        {/* Section 4: Security Roadmap (Transparently Deferred) */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <SettingsIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Planned Security Roadmap</h2>
              <p className="text-xs text-text-muted">
                Upcoming security enhancements planned for subsequent releases.
              </p>
            </div>
          </div>

          <div className="space-y-3 text-xs">
            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <strong className="text-text-primary block">Multi-Factor Authentication (TOTP 2FA)</strong>
                <p className="text-text-secondary text-[11px] mt-0.5">
                  Hardware and authenticator app-based two-factor verification.
                </p>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-slate-200 text-slate-700 self-start sm:self-center">
                Security Roadmap
              </span>
            </div>

            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <strong className="text-text-primary block">Automated Email Activity Summaries</strong>
                <p className="text-text-secondary text-[11px] mt-0.5">
                  Scheduled weekly productivity reports delivered via email.
                </p>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-slate-200 text-slate-700 self-start sm:self-center">
                Deferred
              </span>
            </div>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}
