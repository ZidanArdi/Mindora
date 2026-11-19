import { useState, useEffect, useCallback } from 'react';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import { productivityService } from '../services/productivityService';
import type { LifestyleRecord, UpsertLifestylePayload, UserProfile } from '../types';
import { useAuth } from '../hooks';

interface DropdownOption {
  value: string;
  label: string;
}

// 1. Sleep Duration Options (midpoints: 2.0, 4.5, 5.5, 6.5, 7.5, 8.5; open-ended: 9.5)
const SLEEP_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '2.0', label: '0–4 hours' },
  { value: '4.5', label: '4–5 hours' },
  { value: '5.5', label: '5–6 hours' },
  { value: '6.5', label: '6–7 hours' },
  { value: '7.5', label: '7–8 hours' },
  { value: '8.5', label: '8–9 hours' },
  { value: '9.5', label: '9+ hours' },
];

// 2. Phone Screen Time Options (midpoints: 0.5, 1.5, 2.5, 3.5, 4.5, 5.5; open-ended: 6.5)
const PHONE_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0.5', label: '0–1 hours' },
  { value: '1.5', label: '1–2 hours' },
  { value: '2.5', label: '2–3 hours' },
  { value: '3.5', label: '3–4 hours' },
  { value: '4.5', label: '4–5 hours' },
  { value: '5.5', label: '5–6 hours' },
  { value: '6.5', label: '6+ hours' },
];

// 3. Social Media Options (midpoints: 0.5, 1.5, 2.5, 3.5; open-ended: 4.5)
const SOCIAL_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0.5', label: '0–1 hours' },
  { value: '1.5', label: '1–2 hours' },
  { value: '2.5', label: '2–3 hours' },
  { value: '3.5', label: '3–4 hours' },
  { value: '4.5', label: '4+ hours' },
];

// 4. YouTube Stream Options (midpoints: 0.5, 1.5, 2.5, 3.5; open-ended: 4.5)
const YOUTUBE_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0.5', label: '0–1 hours' },
  { value: '1.5', label: '1–2 hours' },
  { value: '2.5', label: '2–3 hours' },
  { value: '3.5', label: '3–4 hours' },
  { value: '4.5', label: '4+ hours' },
];

// 5. Gaming Time Options (exact 0; midpoints: 1.5, 2.5, 3.5; open-ended: 4.5)
const GAMING_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0', label: '0 hours' },
  { value: '1.5', label: '1–2 hours' },
  { value: '2.5', label: '2–3 hours' },
  { value: '3.5', label: '3–4 hours' },
  { value: '4.5', label: '4+ hours' },
];

// 6. Caffeine Intake Options (exact 0; integer midpoints in mg: 25, 75, 125, 175, 250; open-ended: 350)
const CAFFEINE_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0', label: '0 mg' },
  { value: '25', label: '1–50 mg' },
  { value: '75', label: '51–100 mg' },
  { value: '125', label: '101–150 mg' },
  { value: '175', label: '151–200 mg' },
  { value: '250', label: '201–300 mg' },
  { value: '350', label: '300+ mg' },
];

// 7. Exercise / Workout Options (exact 0; midpoints: 8, 23, 38, 53, 75; open-ended: 105)
const EXERCISE_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '0', label: '0 minutes' },
  { value: '8', label: '1–15 minutes' },
  { value: '23', label: '16–30 minutes' },
  { value: '38', label: '31–45 minutes' },
  { value: '53', label: '46–60 minutes' },
  { value: '75', label: '61–90 minutes' },
  { value: '105', label: '90+ minutes' },
];

// 8. Perceived Stress Options (1 through 10)
const STRESS_OPTIONS: DropdownOption[] = [
  { value: '', label: 'Not recorded' },
  { value: '1', label: '1 - Low / Calm' },
  { value: '2', label: '2' },
  { value: '3', label: '3' },
  { value: '4', label: '4' },
  { value: '5', label: '5 - Moderate' },
  { value: '6', label: '6' },
  { value: '7', label: '7' },
  { value: '8', label: '8' },
  { value: '9', label: '9' },
  { value: '10', label: '10 - High / Overwhelmed' },
];

/**
 * Safely preserves and renders existing numeric values that may not match predefined midpoints.
 */
function getOptionsWithCustom(options: DropdownOption[], currentValue: string, unit: string): DropdownOption[] {
  if (!currentValue || options.some((opt) => opt.value === currentValue)) {
    return options;
  }
  const customOption: DropdownOption = {
    value: currentValue,
    label: `Custom: ${currentValue} ${unit}`,
  };
  return [...options, customOption];
}

export default function LifestyleTrackingPage() {
  const { user } = useAuth();

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

  // Selected date in YYYY-MM-DD format
  const [todayString] = useState(() => new Date().toISOString().split('T')[0]);
  const [selectedDate, setSelectedDate] = useState<string>(todayString);

  // 8 Canonical Lifestyle Variables (stored as string in form to preserve empty/NULL state)
  const [sleepHours, setSleepHours] = useState<string>('');
  const [phoneUsageHours, setPhoneUsageHours] = useState<string>('');
  const [socialMediaHours, setSocialMediaHours] = useState<string>('');
  const [youtubeHours, setYoutubeHours] = useState<string>('');
  const [gamingHours, setGamingHours] = useState<string>('');
  const [coffeeIntakeMg, setCoffeeIntakeMg] = useState<string>('');
  const [exerciseMinutes, setExerciseMinutes] = useState<string>('');
  const [stressLevel, setStressLevel] = useState<string>('');
  const [notes, setNotes] = useState<string>('');

  const [loadingRecord, setLoadingRecord] = useState(false);
  const [saving, setSaving] = useState(false);
  const [feedbackMsg, setFeedbackMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [history, setHistory] = useState<LifestyleRecord[]>([]);

  const loadHistory = useCallback(async () => {
    try {
      const data = await productivityService.listLifestyleHistory({ limit: 14 });
      setHistory(data.records);
    } catch (err: unknown) {
      console.error('Failed to load lifestyle history:', err);
    }
  }, []);

  const loadRecordForDate = useCallback(async (date: string) => {
    try {
      setLoadingRecord(true);
      setFeedbackMsg(null);
      const record = await productivityService.getLifestyleRecord(date);
      if (record) {
        setSleepHours(record.sleep_hours !== null && record.sleep_hours !== undefined ? String(record.sleep_hours) : '');
        setPhoneUsageHours(record.phone_usage_hours !== null && record.phone_usage_hours !== undefined ? String(record.phone_usage_hours) : '');
        setSocialMediaHours(record.social_media_hours !== null && record.social_media_hours !== undefined ? String(record.social_media_hours) : '');
        setYoutubeHours(record.youtube_hours !== null && record.youtube_hours !== undefined ? String(record.youtube_hours) : '');
        setGamingHours(record.gaming_hours !== null && record.gaming_hours !== undefined ? String(record.gaming_hours) : '');
        setCoffeeIntakeMg(record.coffee_intake_mg !== null && record.coffee_intake_mg !== undefined ? String(record.coffee_intake_mg) : '');
        setExerciseMinutes(record.exercise_minutes !== null && record.exercise_minutes !== undefined ? String(record.exercise_minutes) : '');
        setStressLevel(record.stress_level !== null && record.stress_level !== undefined ? String(record.stress_level) : '');
        setNotes(record.notes || '');
      } else {
        // Reset form to blank (preserving NULL/unrecorded)
        setSleepHours('');
        setPhoneUsageHours('');
        setSocialMediaHours('');
        setYoutubeHours('');
        setGamingHours('');
        setCoffeeIntakeMg('');
        setExerciseMinutes('');
        setStressLevel('');
        setNotes('');
      }
    } catch (err: unknown) {
      console.error('Failed to load lifestyle record:', err);
    } finally {
      setLoadingRecord(false);
    }
  }, []);

  useEffect(() => {
    loadRecordForDate(selectedDate);
    loadHistory();
  }, [selectedDate, loadRecordForDate, loadHistory]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setFeedbackMsg(null);

    // Helper to parse numeric or undefined (STRICTLY preserve NULL instead of forced 0.0)
    const parseNum = (val: string): number | undefined => {
      const trimmed = val.trim();
      if (trimmed === '') return undefined;
      const parsed = parseFloat(trimmed);
      return isNaN(parsed) ? undefined : parsed;
    };

    const parseIntVal = (val: string): number | undefined => {
      const trimmed = val.trim();
      if (trimmed === '') return undefined;
      const parsed = parseInt(trimmed, 10);
      return isNaN(parsed) ? undefined : parsed;
    };

    const payload: UpsertLifestylePayload = {
      record_date: selectedDate,
      sleep_hours: parseNum(sleepHours),
      phone_usage_hours: parseNum(phoneUsageHours),
      social_media_hours: parseNum(socialMediaHours),
      youtube_hours: parseNum(youtubeHours),
      gaming_hours: parseNum(gamingHours),
      coffee_intake_mg: parseIntVal(coffeeIntakeMg),
      exercise_minutes: parseIntVal(exerciseMinutes),
      stress_level: parseIntVal(stressLevel),
      notes: notes.trim() || undefined,
    };

    try {
      await productivityService.upsertLifestyleRecord(payload);
      setFeedbackMsg({
        type: 'success',
        text: `Lifestyle habits for ${selectedDate} saved successfully!`,
      });
      await loadHistory();
    } catch (err: unknown) {
      console.error('Failed to save lifestyle record:', err);
      setFeedbackMsg({
        type: 'error',
        text: 'Failed to save lifestyle habits. Please verify input ranges.',
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border/60">
          <div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-[#123C32] tracking-tight">
              Daily Lifestyle & Habit Calibration
            </h1>
            <p className="text-sm text-[#526B60] mt-1">
              Track 8 key physiological & digital habits that influence cognitive energy and focus.
            </p>
          </div>

          {/* Date Selector */}
          <div className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-2xl border border-border shadow-soft-xs">
            <span className="text-xs font-semibold text-text-secondary">Date:</span>
            <input
              type="date"
              value={selectedDate}
              max={todayString}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="text-xs font-bold text-[#123C32] bg-transparent outline-none cursor-pointer"
            />
          </div>
        </div>

        {/* Feedback Alert */}
        {feedbackMsg && (
          <div
            className={`p-4 rounded-2xl border text-xs font-semibold transition-all ${
              feedbackMsg.type === 'success'
                ? 'bg-emerald-50 text-emerald-800 border-emerald-200'
                : 'bg-rose-50 text-rose-800 border-rose-200'
            }`}
          >
            {feedbackMsg.text}
          </div>
        )}

        {/* Main Logging Card */}
        <div className="bg-white rounded-3xl border border-border p-6 sm:p-8 shadow-soft-sm relative overflow-hidden">
          {/* Subtle nature accent */}
          <div className="absolute right-0 top-0 pointer-events-none opacity-5">
            <svg width="260" height="260" viewBox="0 0 200 200" fill="none">
              <path
                d="M100 20 C120 70 170 80 170 100 C170 140 130 180 100 180 C70 180 30 140 30 100 C30 80 80 70 100 20 Z"
                fill="#138F78"
              />
            </svg>
          </div>

          <form onSubmit={handleSave} className="relative z-10 space-y-6">
            <div className="border-b border-border/60 pb-3 flex items-center justify-between">
              <div>
                <h2 className="text-base font-bold text-[#123C32]">
                  Habit Variables for {new Date(selectedDate).toLocaleDateString([], { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                </h2>
                <span className="text-xs text-[#526B60]">
                  Unrecorded metrics will be safely preserved as unmeasured (never forced to 0).
                </span>
              </div>
              {loadingRecord && (
                <span className="text-xs font-semibold text-teal-600 animate-pulse">
                  Loading day records...
                </span>
              )}
            </div>

            {/* 8 Canonical Variables in a responsive 2/4-column grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
              {/* 1. Sleep Duration */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-sleep" className="text-xs font-bold text-[#123C32]">
                    Sleep Duration
                  </label>
                  <span className="text-[11px] font-semibold text-teal-700 bg-teal-50 px-2 py-0.5 rounded-full border border-teal-200">
                    Hours
                  </span>
                </div>
                <select
                  id="select-sleep"
                  value={sleepHours}
                  onChange={(e) => setSleepHours(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(SLEEP_OPTIONS, sleepHours, 'hrs').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Total sleep including deep sleep & naps.
                  {sleepHours !== '' && (
                    <span className="text-[#138F78] font-bold block mt-0.5">
                      Calibrates as {sleepHours}h
                    </span>
                  )}
                </p>
              </div>

              {/* 2. Phone Screen Time */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-phone" className="text-xs font-bold text-[#123C32]">
                    Phone Screen Time
                  </label>
                  <span className="text-[11px] font-semibold text-teal-700 bg-teal-50 px-2 py-0.5 rounded-full border border-teal-200">
                    Hours
                  </span>
                </div>
                <select
                  id="select-phone"
                  value={phoneUsageHours}
                  onChange={(e) => setPhoneUsageHours(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(PHONE_OPTIONS, phoneUsageHours, 'hrs').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Smartphone active screen-on time.
                  {phoneUsageHours !== '' && (
                    <span className="text-[#138F78] font-bold block mt-0.5">
                      Calibrates as {phoneUsageHours}h
                    </span>
                  )}
                </p>
              </div>

              {/* 3. Social Media */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-social" className="text-xs font-bold text-[#123C32]">
                    Social Media
                  </label>
                  <span className="text-[11px] font-semibold text-teal-700 bg-teal-50 px-2 py-0.5 rounded-full border border-teal-200">
                    Hours
                  </span>
                </div>
                <select
                  id="select-social"
                  value={socialMediaHours}
                  onChange={(e) => setSocialMediaHours(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(SOCIAL_OPTIONS, socialMediaHours, 'hrs').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Instagram, TikTok, Twitter/X, and feeds.
                  {socialMediaHours !== '' && (
                    <span className="text-[#138F78] font-bold block mt-0.5">
                      Calibrates as {socialMediaHours}h
                    </span>
                  )}
                </p>
              </div>

              {/* 4. YouTube Stream */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-youtube" className="text-xs font-bold text-[#123C32]">
                    YouTube Stream
                  </label>
                  <span className="text-[11px] font-semibold text-teal-700 bg-teal-50 px-2 py-0.5 rounded-full border border-teal-200">
                    Hours
                  </span>
                </div>
                <select
                  id="select-youtube"
                  value={youtubeHours}
                  onChange={(e) => setYoutubeHours(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(YOUTUBE_OPTIONS, youtubeHours, 'hrs').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Video streaming duration on desktop or mobile.
                  {youtubeHours !== '' && (
                    <span className="text-[#138F78] font-bold block mt-0.5">
                      Calibrates as {youtubeHours}h
                    </span>
                  )}
                </p>
              </div>

              {/* 5. Gaming Time */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-gaming" className="text-xs font-bold text-[#123C32]">
                    Gaming Time
                  </label>
                  <span className="text-[11px] font-semibold text-teal-700 bg-teal-50 px-2 py-0.5 rounded-full border border-teal-200">
                    Hours
                  </span>
                </div>
                <select
                  id="select-gaming"
                  value={gamingHours}
                  onChange={(e) => setGamingHours(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(GAMING_OPTIONS, gamingHours, 'hrs').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Console, PC, or mobile gameplay duration.
                  {gamingHours !== '' && (
                    <span className="text-[#138F78] font-bold block mt-0.5">
                      Calibrates as {gamingHours}h
                    </span>
                  )}
                </p>
              </div>

              {/* 6. Coffee Intake (mg) */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-caffeine" className="text-xs font-bold text-[#123C32]">
                    Caffeine Intake
                  </label>
                  <span className="text-[11px] font-semibold text-amber-700 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-200">
                    Milligrams (mg)
                  </span>
                </div>
                <select
                  id="select-caffeine"
                  value={coffeeIntakeMg}
                  onChange={(e) => setCoffeeIntakeMg(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(CAFFEINE_OPTIONS, coffeeIntakeMg, 'mg').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  1 brewed cup ≈ 80–100mg; 1 espresso ≈ 64mg.
                  {coffeeIntakeMg !== '' && (
                    <span className="text-amber-700 font-bold block mt-0.5">
                      Calibrates as {coffeeIntakeMg}mg
                    </span>
                  )}
                </p>
              </div>

              {/* 7. Exercise Minutes */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-exercise" className="text-xs font-bold text-[#123C32]">
                    Exercise / Workout
                  </label>
                  <span className="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                    Minutes
                  </span>
                </div>
                <select
                  id="select-exercise"
                  value={exerciseMinutes}
                  onChange={(e) => setExerciseMinutes(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(EXERCISE_OPTIONS, exerciseMinutes, 'mins').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Running, gym, cycling, yoga, or brisk walking.
                  {exerciseMinutes !== '' && (
                    <span className="text-emerald-700 font-bold block mt-0.5">
                      Calibrates as {exerciseMinutes}m
                    </span>
                  )}
                </p>
              </div>

              {/* 8. Perceived Stress (1-10) */}
              <div className="p-4 rounded-2xl bg-slate-50/70 border border-slate-200/60 space-y-2">
                <div className="flex items-center justify-between">
                  <label htmlFor="select-stress" className="text-xs font-bold text-[#123C32]">
                    Perceived Stress
                  </label>
                  <span className="text-[11px] font-semibold text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded-full border border-indigo-200">
                    Scale 1–10
                  </span>
                </div>
                <select
                  id="select-stress"
                  value={stressLevel}
                  onChange={(e) => setStressLevel(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
                >
                  {getOptionsWithCustom(STRESS_OPTIONS, stressLevel, 'scale').map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <p className="text-[10px] text-text-secondary leading-tight">
                  Subjective stress level (1 = Calm, 10 = Overwhelmed).
                  {stressLevel !== '' && (
                    <span className="text-indigo-700 font-bold block mt-0.5">
                      Calibrates as Level {stressLevel}/10
                    </span>
                  )}
                </p>
              </div>
            </div>

            {/* Optional Notes */}
            <div className="space-y-2">
              <label className="text-xs font-bold text-[#123C32]">
                Reflective Notes (Optional)
              </label>
              <textarea
                rows={2}
                placeholder="Any special circumstances today? (e.g. midterms week, poor sleep due to noise, energy drinks)"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full text-xs p-3 rounded-2xl border border-border focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all"
              />
            </div>

            {/* Action Buttons */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <Button
                type="submit"
                variant="primary"
                size="md"
                disabled={saving}
                className="px-8 font-bold text-xs bg-[#138F78] hover:bg-[#123C32] text-white rounded-xl shadow-soft-sm"
              >
                {saving ? 'Saving Habits...' : 'Save & Calibrate Habits'}
              </Button>
            </div>
          </form>
        </div>

        {/* History Table of Recent Days */}
        <div className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-bold text-[#123C32]">Recent Lifestyle Logs</h2>
            <span className="text-xs font-medium text-text-secondary">
              Past 14 Days History
            </span>
          </div>

          {history.length === 0 ? (
            <div className="py-12 text-center space-y-2">
              <span className="text-3xl block">🌿</span>
              <p className="text-sm font-semibold text-text-primary">No lifestyle records saved yet</p>
              <p className="text-xs text-text-secondary">
                Calibrate today's metrics above to build your habit baseline.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-border/80 text-text-secondary uppercase tracking-wider font-semibold">
                    <th className="py-3 px-3">Date</th>
                    <th className="py-3 px-3">Sleep</th>
                    <th className="py-3 px-3">Phone Screen</th>
                    <th className="py-3 px-3">Social Media</th>
                    <th className="py-3 px-3">YouTube</th>
                    <th className="py-3 px-3">Gaming</th>
                    <th className="py-3 px-3">Caffeine</th>
                    <th className="py-3 px-3">Exercise</th>
                    <th className="py-3 px-3">Stress</th>
                    <th className="py-3 px-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {history.map((record) => (
                    <tr key={record.id} className="hover:bg-surface-muted/40 transition-colors">
                      <td className="py-3.5 px-3 font-bold text-[#123C32]">
                        {record.record_date}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.sleep_hours !== null && record.sleep_hours !== undefined
                          ? `${record.sleep_hours}h`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.phone_usage_hours !== null && record.phone_usage_hours !== undefined
                          ? `${record.phone_usage_hours}h`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.social_media_hours !== null && record.social_media_hours !== undefined
                          ? `${record.social_media_hours}h`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.youtube_hours !== null && record.youtube_hours !== undefined
                          ? `${record.youtube_hours}h`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.gaming_hours !== null && record.gaming_hours !== undefined
                          ? `${record.gaming_hours}h`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.coffee_intake_mg !== null && record.coffee_intake_mg !== undefined
                          ? `${record.coffee_intake_mg}mg`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.exercise_minutes !== null && record.exercise_minutes !== undefined
                          ? `${record.exercise_minutes}m`
                          : <span className="text-text-muted">—</span>}
                      </td>
                      <td className="py-3.5 px-3">
                        {record.stress_level !== null && record.stress_level !== undefined ? (
                          <span
                            className={`inline-flex px-2 py-0.5 rounded text-[11px] font-bold ${
                              record.stress_level <= 3
                                ? 'bg-emerald-100 text-emerald-800'
                                : record.stress_level <= 7
                                ? 'bg-amber-100 text-amber-800'
                                : 'bg-rose-100 text-rose-800'
                            }`}
                          >
                            {record.stress_level}/10
                          </span>
                        ) : (
                          <span className="text-text-muted">—</span>
                        )}
                      </td>
                      <td className="py-3.5 px-3 text-right">
                        <button
                          type="button"
                          onClick={() => setSelectedDate(record.record_date)}
                          className="text-xs font-semibold text-primary hover:text-primary-800 underline"
                        >
                          Edit
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </DashboardTemplate>
  );
}
