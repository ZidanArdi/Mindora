/**
 * Centralized Indonesian (id-ID) Localization Utilities for Mindora
 */

export const ID_LOCALE = 'id-ID';

/**
 * Formats a Date or ISO string into localized Indonesian date format.
 * Example: "21 November 2026"
 */
export function formatDate(
  date: string | Date | null | undefined,
  options: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }
): string {
  if (!date) return '—';
  try {
    const d = typeof date === 'string' ? new Date(date) : date;
    if (isNaN(d.getTime())) return '—';
    return new Intl.DateTimeFormat(ID_LOCALE, options).format(d);
  } catch {
    return '—';
  }
}

/**
 * Formats a Date or ISO string into localized Indonesian date and time format.
 * Example: "21 Nov 2026, 14.30"
 */
export function formatDateTime(
  date: string | Date | null | undefined,
  options: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }
): string {
  if (!date) return '—';
  try {
    const d = typeof date === 'string' ? new Date(date) : date;
    if (isNaN(d.getTime())) return '—';
    return new Intl.DateTimeFormat(ID_LOCALE, options).format(d);
  } catch {
    return '—';
  }
}

/**
 * Formats a Date or ISO string into localized Indonesian time format.
 * Example: "14.30"
 */
export function formatTime(
  date: string | Date | null | undefined,
  options: Intl.DateTimeFormatOptions = {
    hour: '2-digit',
    minute: '2-digit',
  }
): string {
  if (!date) return '—';
  try {
    const d = typeof date === 'string' ? new Date(date) : date;
    if (isNaN(d.getTime())) return '—';
    return new Intl.DateTimeFormat(ID_LOCALE, options).format(d);
  } catch {
    return '—';
  }
}

/**
 * Formats seconds into natural Indonesian duration string.
 * Example: "2 jam 15 mnt" or "45 mnt"
 */
export function formatDuration(seconds: number): string {
  if (isNaN(seconds) || seconds <= 0) return '0 mnt';
  const totalMinutes = Math.round(seconds / 60);
  if (totalMinutes < 60) {
    return `${totalMinutes} mnt`;
  }
  const hours = Math.floor(totalMinutes / 60);
  const remainingMins = totalMinutes % 60;
  if (remainingMins === 0) {
    return `${hours} jam`;
  }
  return `${hours} jam ${remainingMins} mnt`;
}

/**
 * Formats seconds into consistent, human-readable Indonesian study session duration.
 * Accurately differentiates short seconds from multi-hour sessions without misleading truncation.
 *
 * Examples:
 * - 6s     -> "6 detik"
 * - 45s    -> "45 detik"
 * - 300s   -> "5 menit"
 * - 330s   -> "5 menit 30 detik"
 * - 7200s  -> "2 jam"
 * - 15349s -> "4 jam 15 menit"
 * - 16815s -> "4 jam 40 menit"
 */
export function formatStudyDuration(totalSeconds: number): string {
  if (isNaN(totalSeconds) || totalSeconds <= 0) return '0 detik';

  const sec = Math.floor(totalSeconds);
  if (sec < 60) {
    return `${sec} detik`;
  }

  const hours = Math.floor(sec / 3600);
  const remainingMins = Math.floor((sec % 3600) / 60);
  const remainingSecs = sec % 60;

  if (hours > 0) {
    if (remainingMins === 0) {
      return remainingSecs > 0 ? `${hours} jam ${remainingSecs} detik` : `${hours} jam`;
    }
    return `${hours} jam ${remainingMins} menit`;
  }

  // Under 1 hour
  if (remainingSecs === 0) {
    return `${remainingMins} menit`;
  }
  return `${remainingMins} menit ${remainingSecs} detik`;
}

/**
 * Formats seconds into HH:MM:SS digital timer clock string.
 * Example: 15349 -> "04:15:49", 6 -> "00:00:06"
 */
export function formatTimer(totalSec: number): string {
  if (isNaN(totalSec) || totalSec <= 0) return '00:00:00';
  const hrs = Math.floor(totalSec / 3600);
  const mins = Math.floor((totalSec % 3600) / 60);
  const secs = Math.floor(totalSec % 60);
  return `${hrs.toString().padStart(2, '0')}:${mins
    .toString()
    .padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

/**
 * Formats seconds into hours with 1 decimal place using Indonesian decimal comma.
 * Example: "2,5 jam"
 */
export function formatHours(seconds: number): string {
  if (isNaN(seconds) || seconds <= 0) return '0 jam';
  const hrs = seconds / 3600;
  return `${new Intl.NumberFormat(ID_LOCALE, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  }).format(hrs)} jam`;
}

/**
 * Formats minutes into natural Indonesian string.
 * Example: "25 mnt" or "1 jam 30 mnt"
 */
export function formatMinutes(minutes: number): string {
  if (isNaN(minutes) || minutes <= 0) return '0 mnt';
  if (minutes < 60) return `${minutes} mnt`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m > 0 ? `${h} jam ${m} mnt` : `${h} jam`;
}

/**
 * Formats numbers into Indonesian locale representation.
 */
export function formatNumber(
  num: number,
  options?: Intl.NumberFormatOptions
): string {
  if (isNaN(num)) return '0';
  return new Intl.NumberFormat(ID_LOCALE, options).format(num);
}

/**
 * Formats percentages into localized string.
 */
export function formatPercent(num: number, decimals: number = 0): string {
  if (isNaN(num)) return '0%';
  return `${new Intl.NumberFormat(ID_LOCALE, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num)}%`;
}

/**
 * Formats a prediction target date (YYYY-MM-DD or ISO timestamp) into localized date string
 * without accidental timezone date shift.
 * Example: "2026-11-22T00:00:00Z" -> "22 November 2026"
 */
export function formatTargetDate(
  dateStr: string | null | undefined,
  locale: string = ID_LOCALE
): string {
  if (!dateStr) return '—';
  const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (match) {
    const [, yStr, mStr, dStr] = match;
    const year = parseInt(yStr, 10);
    const monthIndex = parseInt(mStr, 10) - 1;
    const day = parseInt(dStr, 10);
    const localDate = new Date(year, monthIndex, day);
    return new Intl.DateTimeFormat(locale, {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    }).format(localDate);
  }
  return formatDate(dateStr, { day: 'numeric', month: 'long', year: 'numeric' });
}

/**
 * Translates prediction state labels consistently.
 */
export function translatePredictionState(
  state?: string,
  fallbackLabel?: string,
  locale: string = ID_LOCALE
): string {
  const isEn = locale.startsWith('en');
  if (state === 'baseline' || fallbackLabel === 'Baseline Estimate') {
    return isEn ? 'Baseline Estimate' : 'Estimasi Awal';
  }
  if (state === 'personalized' || fallbackLabel === 'Personalized Estimate') {
    return isEn ? 'Personalized Estimate' : 'Estimasi Terpersonalisasi';
  }
  return fallbackLabel || (isEn ? 'Personalized Estimate' : 'Estimasi Terpersonalisasi');
}

/**
 * Translates prediction description strings to avoid raw English sentences in Indonesian locale.
 */
export function translatePredictionDescription(
  desc?: string,
  state?: string,
  locale: string = ID_LOCALE
): string {
  const isEn = locale.startsWith('en');
  if (isEn) {
    if (state === 'baseline' || desc?.includes('reference values')) {
      return 'This estimate uses reference values for one or more unrecorded habits. Record your daily habits to obtain a more personalized estimate.';
    }
    return desc || 'Estimated from habit patterns (study, sleep, phone, stress). Not an objective academic evaluation.';
  }

  // Indonesian
  if (state === 'baseline' || desc?.includes('reference values')) {
    return 'Estimasi ini menggunakan nilai referensi untuk kebiasaan yang belum dicatat. Catat kebiasaan harian Anda untuk memperoleh estimasi yang lebih personal.';
  }
  if (state === 'personalized' || desc?.includes('recorded study duration') || desc?.includes('calculated using')) {
    return 'Estimasi ini dihitung berdasarkan durasi belajar, tidur, penggunaan ponsel, dan tingkat stres yang Anda catat.';
  }
  return desc || 'Diestimasi dari pola kebiasaan (belajar, tidur, ponsel, stres). Bukan merupakan nilai akademik objektif.';
}

/**
 * Generates an appropriate time-aware greeting in user's locale.
 */
export function getGreeting(userName?: string, locale: string = ID_LOCALE): string {
  const hour = new Date().getHours();
  const isEn = locale.startsWith('en');
  let greet: string;

  if (isEn) {
    if (hour >= 4 && hour < 12) {
      greet = 'Good morning';
    } else if (hour >= 12 && hour < 17) {
      greet = 'Good afternoon';
    } else if (hour >= 17 && hour < 21) {
      greet = 'Good evening';
    } else {
      greet = 'Good night';
    }
  } else {
    if (hour >= 4 && hour < 11) {
      greet = 'Selamat Pagi';
    } else if (hour >= 11 && hour < 15) {
      greet = 'Selamat Siang';
    } else if (hour >= 15 && hour < 18) {
      greet = 'Selamat Sore';
    } else {
      greet = 'Selamat Malam';
    }
  }

  return userName ? `${greet}, ${userName}` : greet;
}

/**
 * Translates task and recommendation priority identifiers into Bahasa Indonesia.
 */
export function translatePriority(priority?: string): string {
  if (!priority) return 'Sedang';
  switch (priority.toLowerCase()) {
    case 'critical':
      return 'Kritis';
    case 'urgent':
      return 'Mendesak';
    case 'high':
      return 'Tinggi';
    case 'medium':
      return 'Sedang';
    case 'low':
      return 'Rendah';
    default:
      return priority;
  }
}

/**
 * Translates status identifiers into Bahasa Indonesia.
 */
export function translateStatus(status?: string): string {
  if (!status) return 'Menunggu';
  switch (status.toLowerCase()) {
    case 'completed':
      return 'Selesai';
    case 'pending':
      return 'Menunggu';
    case 'in_progress':
      return 'Sedang Dikerjakan';
    case 'active':
      return 'Aktif';
    case 'paused':
      return 'Dijeda';
    case 'cancelled':
      return 'Dibatalkan';
    case 'abandoned':
      return 'Dibatalkan';
    case 'dismissed':
      return 'Diabaikan';
    case 'expired':
      return 'Kedaluwarsa';
    case 'invalidated':
      return 'Tidak Berlaku';
    default:
      return status;
  }
}

/**
 * Translates category names into Bahasa Indonesia.
 */
export function translateCategory(category?: string): string {
  if (!category) return 'Lainnya';
  switch (category.toLowerCase()) {
    case 'study':
      return 'Belajar';
    case 'academic':
      return 'Akademik';
    case 'lab':
    case 'lab / practical':
      return 'Praktikum / Lab';
    case 'project':
      return 'Proyek';
    case 'personal':
      return 'Pribadi';
    case 'inbox':
      return 'Kotak Masuk';
    case 'work':
      return 'Pekerjaan';
    case 'health':
      return 'Kesehatan';
    case 'holiday trip':
      return 'Liburan';
    default:
      return category;
  }
}

/**
 * Known API error translation dictionary for user-friendly notifications.
 */
const API_ERROR_MAP: Record<string, string> = {
  'Authentication required': 'Autentikasi diperlukan. Silakan masuk kembali.',
  'Invalid credentials': 'Email atau kata sandi tidak valid.',
  'Invalid email or password': 'Email atau kata sandi tidak valid.',
  'Email already registered': 'Alamat email ini sudah terdaftar.',
  'Account is locked or inactive': 'Akun terkunci atau nonaktif. Hubungi administrator.',
  'Invalid or expired token': 'Token tidak valid atau telah kedaluwarsa.',
  'Forbidden: insufficient permissions': 'Akses ditolak: izin tidak memadai.',
  'start_date and end_date query parameters are required (YYYY-MM-DD)': 'Parameter tanggal mulai dan tanggal akhir wajib diisi (YYYY-MM-DD).',
  'both start_date and end_date are required (YYYY-MM-DD)': 'Tanggal mulai dan tanggal akhir wajib diisi (YYYY-MM-DD).',
  'invalid start_date format, expected YYYY-MM-DD': 'Format tanggal mulai tidak valid, gunakan YYYY-MM-DD.',
  'invalid end_date format, expected YYYY-MM-DD': 'Format tanggal akhir tidak valid, gunakan YYYY-MM-DD.',
  'start_date cannot be after end_date': 'Tanggal mulai tidak boleh melebihi tanggal akhir.',
  'date range cannot exceed 90 days': 'Rentang tanggal tidak boleh melebihi 90 hari.',
  'Failed to fetch range analytics': 'Gagal mengambil data analitik rentang tanggal.',
  'Failed to load tasks': 'Gagal memuat daftar tugas.',
  'Failed to create task': 'Gagal membuat tugas baru.',
  'Failed to update task': 'Gagal memperbarui tugas.',
  'Failed to delete task': 'Gagal menghapus tugas.',
  'Failed to start study session': 'Gagal memulai sesi belajar.',
  'Failed to pause study session': 'Gagal menjeda sesi belajar.',
  'Failed to resume study session': 'Gagal melanjutkan sesi belajar.',
  'Failed to finish study session': 'Gagal menyelesaikan sesi belajar.',
  'Failed to cancel study session': 'Gagal membatalkan sesi belajar.',
  'Failed to start focus interval': 'Gagal memulai interval fokus.',
  'Failed to finish interval': 'Gagal menyelesaikan interval.',
  'Failed to abandon interval': 'Gagal membatalkan interval.',
  'Failed to save lifestyle log': 'Gagal menyimpan catatan gaya hidup.',
  'Failed to fetch dashboard summary': 'Gagal memuat ringkasan dasbor.',
  'Failed to generate productivity prediction': 'Gagal menghasilkan prediksi produktivitas.',
  'Failed to retrieve recommendations': 'Gagal memuat rekomendasi.',
  'Failed to update recommendation status': 'Gagal memperbarui status rekomendasi.',
  'Failed to fetch user profile': 'Gagal memuat profil pengguna.',
  'Failed to update user profile': 'Gagal memperbarui profil pengguna.',
};

/**
 * Translates backend error messages to Bahasa Indonesia, falling back to original or custom message.
 */
export function translateErrorMessage(err: unknown, fallback?: string): string {
  if (err instanceof Error) {
    if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
      return 'Server backend tidak dapat dihubungi. Pastikan koneksi atau backend sedang berjalan.';
    }
    if (API_ERROR_MAP[err.message]) {
      return API_ERROR_MAP[err.message];
    }
    // Partial matching for known prefixes
    for (const [enKey, idVal] of Object.entries(API_ERROR_MAP)) {
      if (err.message.includes(enKey)) {
        return idVal;
      }
    }
    return err.message;
  }
  return fallback || 'Terjadi kesalahan sistem. Silakan coba beberapa saat lagi.';
}
