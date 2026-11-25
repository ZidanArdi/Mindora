import { apiFetch } from './api';
import type {
  ApiResponse,
  TaskItem,
  CreateTaskPayload,
  UpdateTaskPayload,
  TaskStatus,
  StudySession,
  StudyDailyStats,
  FocusSession,
  FocusDailyStats,
  FocusSessionType,
  LifestyleRecord,
  UpsertLifestylePayload,
  DashboardSummary,
  DailyAnalytics,
  WeeklyAnalytics,
  MonthlyAnalytics,
  AnalyticsSummary,
  ProductivityPrediction,
  PredictionInput,
  PredictionHistoryResponse,
  RecommendationItem,
  RecommendationsResponse,
} from '../types';

export const productivityService = {
  // =========================================================================
  // 1. Task Management APIs
  // =========================================================================
  async listTasks(params: {
    status?: string;
    priority?: string;
    category?: string;
    search?: string;
    sort_by?: string;
    sort_order?: string;
    page?: number;
    limit?: number;
  } = {}): Promise<{ tasks: TaskItem[]; total: number }> {
    const query = new URLSearchParams();
    if (params.status) query.set('status', params.status);
    if (params.priority) query.set('priority', params.priority);
    if (params.category) query.set('category', params.category);
    if (params.search) query.set('search', params.search);
    if (params.sort_by) query.set('sort_by', params.sort_by);
    if (params.sort_order) query.set('sort_order', params.sort_order);
    if (params.page) query.set('page', params.page.toString());
    if (params.limit) query.set('limit', params.limit.toString());

    const res = await apiFetch(`/tasks?${query.toString()}`);
    if (!res.ok) {
      throw new Error(`Failed to list tasks: ${res.statusText}`);
    }
    const data: ApiResponse<{ tasks: TaskItem[]; meta: { total: number } }> = await res.json();
    return {
      tasks: data.data?.tasks || [],
      total: data.data?.meta?.total || 0,
    };
  },

  async createTask(payload: CreateTaskPayload): Promise<TaskItem> {
    const res = await apiFetch('/tasks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to create task');
    }
    const data: ApiResponse<TaskItem> = await res.json();
    if (!data.data) throw new Error('No task returned');
    return data.data;
  },

  async getTask(id: string): Promise<TaskItem> {
    const res = await apiFetch(`/tasks/${id}`);
    if (!res.ok) throw new Error('Task not found');
    const data: ApiResponse<TaskItem> = await res.json();
    if (!data.data) throw new Error('No task returned');
    return data.data;
  },

  async updateTask(id: string, payload: UpdateTaskPayload): Promise<TaskItem> {
    const res = await apiFetch(`/tasks/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to update task');
    }
    const data: ApiResponse<TaskItem> = await res.json();
    if (!data.data) throw new Error('No task returned');
    return data.data;
  },

  async updateTaskStatus(id: string, status: TaskStatus): Promise<TaskItem> {
    const res = await apiFetch(`/tasks/${id}/status`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to update task status');
    }
    const data: ApiResponse<TaskItem> = await res.json();
    if (!data.data) throw new Error('No task returned');
    return data.data;
  },

  async deleteTask(id: string): Promise<void> {
    const res = await apiFetch(`/tasks/${id}`, {
      method: 'DELETE',
    });
    if (!res.ok) {
      throw new Error('Failed to delete task');
    }
  },

  // =========================================================================
  // 2. Study Tracking APIs
  // =========================================================================
  async startStudySession(payload: { subject: string; task_id?: string }): Promise<StudySession> {
    const res = await apiFetch('/study/sessions/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to start study session');
    }
    const data: ApiResponse<StudySession> = await res.json();
    if (!data.data) throw new Error('No session returned');
    return data.data;
  },

  async pauseStudySession(id: string): Promise<StudySession> {
    const res = await apiFetch(`/study/sessions/${id}/pause`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to pause study session');
    const data: ApiResponse<StudySession> = await res.json();
    if (!data.data) throw new Error('No session returned');
    return data.data;
  },

  async resumeStudySession(id: string): Promise<StudySession> {
    const res = await apiFetch(`/study/sessions/${id}/resume`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to resume study session');
    const data: ApiResponse<StudySession> = await res.json();
    if (!data.data) throw new Error('No session returned');
    return data.data;
  },

  async finishStudySession(id: string): Promise<StudySession> {
    const res = await apiFetch(`/study/sessions/${id}/finish`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to finish study session');
    const data: ApiResponse<StudySession> = await res.json();
    if (!data.data) throw new Error('No session returned');
    return data.data;
  },

  async cancelStudySession(id: string): Promise<StudySession> {
    const res = await apiFetch(`/study/sessions/${id}/cancel`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to cancel study session');
    const data: ApiResponse<StudySession> = await res.json();
    if (!data.data) throw new Error('No session returned');
    return data.data;
  },

  async getActiveStudySession(): Promise<StudySession | null> {
    const res = await apiFetch('/study/sessions/active');
    if (!res.ok) return null;
    const data: ApiResponse<StudySession | null> = await res.json();
    return data.data || null;
  },

  async listStudySessions(params: {
    start_date?: string;
    end_date?: string;
    page?: number;
    limit?: number;
  } = {}): Promise<{ sessions: StudySession[]; total: number }> {
    const query = new URLSearchParams();
    if (params.start_date) query.set('start_date', params.start_date);
    if (params.end_date) query.set('end_date', params.end_date);
    if (params.page) query.set('page', params.page.toString());
    if (params.limit) query.set('limit', params.limit.toString());

    const res = await apiFetch(`/study/sessions?${query.toString()}`);
    if (!res.ok) throw new Error('Failed to list study sessions');
    const data: ApiResponse<{ sessions: StudySession[]; meta: { total: number } }> = await res.json();
    return {
      sessions: data.data?.sessions || [],
      total: data.data?.meta?.total || 0,
    };
  },

  async getDailyStudyStats(date?: string, tz?: string): Promise<StudyDailyStats> {
    const params = new URLSearchParams();
    if (date) params.set('date', date);
    const resolvedTz = tz || Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (resolvedTz) params.set('tz', resolvedTz);
    const query = params.toString() ? `?${params.toString()}` : '';
    const res = await apiFetch(`/study/stats/daily${query}`);
    if (!res.ok) throw new Error('Failed to fetch study statistics');
    const data: ApiResponse<StudyDailyStats> = await res.json();
    return data.data || { total_completed_seconds: 0, completed_sessions_count: 0 };
  },

  // =========================================================================
  // 3. Focus Tracking APIs (Individual Interval Architecture)
  // =========================================================================
  async startFocusInterval(payload: {
    session_type: FocusSessionType;
    task_id?: string;
    target_minutes?: number;
  }): Promise<FocusSession> {
    const res = await apiFetch('/focus/sessions/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to start focus interval');
    }
    const data: ApiResponse<FocusSession> = await res.json();
    if (!data.data) throw new Error('No interval returned');
    return data.data;
  },

  async recordInterruption(id: string): Promise<FocusSession> {
    const res = await apiFetch(`/focus/sessions/${id}/interruption`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to record interruption');
    const data: ApiResponse<FocusSession> = await res.json();
    if (!data.data) throw new Error('No interval returned');
    return data.data;
  },

  async incrementInterruption(id: string): Promise<FocusSession> {
    return this.recordInterruption(id);
  },

  async finishFocusInterval(id: string): Promise<FocusSession> {
    const res = await apiFetch(`/focus/sessions/${id}/finish`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to finish interval');
    const data: ApiResponse<FocusSession> = await res.json();
    if (!data.data) throw new Error('No interval returned');
    return data.data;
  },

  async abandonFocusInterval(id: string): Promise<FocusSession> {
    const res = await apiFetch(`/focus/sessions/${id}/abandon`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to abandon interval');
    const data: ApiResponse<FocusSession> = await res.json();
    if (!data.data) throw new Error('No interval returned');
    return data.data;
  },

  async getActiveFocusInterval(): Promise<FocusSession | null> {
    const res = await apiFetch('/focus/sessions/active');
    if (!res.ok) return null;
    const data: ApiResponse<FocusSession | null> = await res.json();
    return data.data || null;
  },

  async listFocusHistory(params: {
    session_type?: FocusSessionType;
    page?: number;
    limit?: number;
  } = {}): Promise<{ intervals: FocusSession[]; total: number }> {
    const query = new URLSearchParams();
    if (params.session_type) query.set('session_type', params.session_type);
    if (params.page) query.set('page', params.page.toString());
    if (params.limit) query.set('limit', params.limit.toString());

    const res = await apiFetch(`/focus/sessions/history?${query.toString()}`);
    if (!res.ok) throw new Error('Failed to list focus history');
    const data: ApiResponse<{ intervals: FocusSession[]; meta: { total: number } }> = await res.json();
    return {
      intervals: data.data?.intervals || [],
      total: data.data?.meta?.total || 0,
    };
  },

  async listFocusIntervals(params: {
    session_type?: FocusSessionType;
    page?: number;
    limit?: number;
  } = {}): Promise<{ intervals: FocusSession[]; total: number }> {
    return this.listFocusHistory(params);
  },

  async getDailyFocusStats(date?: string): Promise<FocusDailyStats> {
    const query = date ? `?date=${date}` : '';
    const res = await apiFetch(`/focus/stats/daily${query}`);
    if (!res.ok) throw new Error('Failed to fetch focus statistics');
    const data: ApiResponse<FocusDailyStats> = await res.json();
    return data.data || {
      completed_focus_minutes: 0,
      completed_focus_count: 0,
      completed_breaks_count: 0,
      total_interruptions: 0,
    };
  },

  // =========================================================================
  // 4. Daily Lifestyle Tracking APIs (8 Variables)
  // =========================================================================
  async getLifestyleRecord(date?: string): Promise<LifestyleRecord | null> {
    const query = date ? `?date=${date}` : '';
    const res = await apiFetch(`/lifestyle/record${query}`);
    if (!res.ok) return null;
    const data: ApiResponse<LifestyleRecord | null> = await res.json();
    return data.data || null;
  },

  async upsertLifestyleRecord(payload: UpsertLifestylePayload): Promise<LifestyleRecord> {
    const res = await apiFetch('/lifestyle/record', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to save lifestyle log');
    }
    const data: ApiResponse<LifestyleRecord> = await res.json();
    if (!data.data) throw new Error('No record returned');
    return data.data;
  },

  async listLifestyleHistory(params: {
    page?: number;
    limit?: number;
  } = {}): Promise<{ records: LifestyleRecord[]; total: number }> {
    const query = new URLSearchParams();
    if (params.page) query.set('page', params.page.toString());
    if (params.limit) query.set('limit', params.limit.toString());

    const res = await apiFetch(`/lifestyle/history?${query.toString()}`);
    if (!res.ok) throw new Error('Failed to list lifestyle history');
    const data: ApiResponse<{ records: LifestyleRecord[]; meta: { total: number } }> = await res.json();
    return {
      records: data.data?.records || [],
      total: data.data?.meta?.total || 0,
    };
  },

  // =========================================================================
  // 5. Dashboard Summary API
  // =========================================================================
  async getDashboardSummary(): Promise<DashboardSummary> {
    const res = await apiFetch('/dashboard/summary');
    if (!res.ok) {
      throw new Error('Failed to fetch dashboard summary');
    }
    const data: ApiResponse<DashboardSummary> = await res.json();
    if (!data.data) throw new Error('No summary returned');
    return data.data;
  },

  // =========================================================================
  // 6. Productivity Analytics APIs (Phase 3)
  // =========================================================================
  async getDailyAnalytics(date?: string, tz?: string): Promise<DailyAnalytics> {
    const params = new URLSearchParams();
    if (date) params.set('date', date);
    if (tz) params.set('tz', tz);
    const query = params.toString() ? `?${params.toString()}` : '';

    const res = await apiFetch(`/analytics/daily${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch daily analytics');
    }
    const data: ApiResponse<DailyAnalytics> = await res.json();
    if (!data.data) throw new Error('No daily analytics returned');
    return data.data;
  },

  async getWeeklyAnalytics(endDate?: string, tz?: string): Promise<WeeklyAnalytics> {
    const params = new URLSearchParams();
    if (endDate) params.set('end_date', endDate);
    if (tz) params.set('tz', tz);
    const query = params.toString() ? `?${params.toString()}` : '';

    const res = await apiFetch(`/analytics/weekly${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch weekly analytics');
    }
    const data: ApiResponse<WeeklyAnalytics> = await res.json();
    if (!data.data) throw new Error('No weekly analytics returned');
    return data.data;
  },

  async getMonthlyAnalytics(endDate?: string, tz?: string): Promise<MonthlyAnalytics> {
    const params = new URLSearchParams();
    if (endDate) params.set('end_date', endDate);
    if (tz) params.set('tz', tz);
    const query = params.toString() ? `?${params.toString()}` : '';

    const res = await apiFetch(`/analytics/monthly${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch monthly analytics');
    }
    const data: ApiResponse<MonthlyAnalytics> = await res.json();
    if (!data.data) throw new Error('No monthly analytics returned');
    return data.data;
  },

  async getRangeAnalytics(startDate: string, endDate: string, tz?: string): Promise<WeeklyAnalytics> {
    const params = new URLSearchParams();
    params.set('start_date', startDate);
    params.set('end_date', endDate);
    if (tz) params.set('tz', tz);

    const res = await apiFetch(`/analytics/range?${params.toString()}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || (res.statusText ? `Failed to fetch range analytics (${res.status} ${res.statusText})` : 'Failed to fetch range analytics'));
    }
    const data: ApiResponse<WeeklyAnalytics> = await res.json();
    if (!data.data) throw new Error('No range analytics returned');
    return data.data;
  },

  async getAnalyticsSummary(tz?: string): Promise<AnalyticsSummary> {
    const query = tz ? `?tz=${encodeURIComponent(tz)}` : '';
    const res = await apiFetch(`/analytics/summary${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch analytics summary');
    }
    const data: ApiResponse<AnalyticsSummary> = await res.json();
    if (!data.data) throw new Error('No analytics summary returned');
    return data.data;
  },

  // =========================================================================
  // 7. ML Productivity Prediction APIs (Phase 3)
  // =========================================================================
  async createPrediction(input?: PredictionInput): Promise<ProductivityPrediction> {
    const res = await apiFetch('/predictions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input || {}),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to generate productivity prediction');
    }
    const data: ApiResponse<ProductivityPrediction> = await res.json();
    if (!data.data) throw new Error('No prediction returned');
    return data.data;
  },

  async listPredictions(params: { page?: number; limit?: number } = {}): Promise<PredictionHistoryResponse> {
    const q = new URLSearchParams();
    if (params.page) q.set('page', params.page.toString());
    if (params.limit) q.set('limit', params.limit.toString());
    const query = q.toString() ? `?${q.toString()}` : '';

    const res = await apiFetch(`/predictions${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch prediction history');
    }
    const data: ApiResponse<PredictionHistoryResponse> = await res.json();
    if (!data.data) throw new Error('No prediction history returned');
    return data.data;
  },

  async getPredictionById(id: string): Promise<ProductivityPrediction> {
    const res = await apiFetch(`/predictions/${id}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to fetch prediction details');
    }
    const data: ApiResponse<ProductivityPrediction> = await res.json();
    if (!data.data) throw new Error('No prediction found');
    return data.data;
  },

  // =========================================================================
  // 6. Recommendation Engine APIs (Phase 4.1)
  // =========================================================================
  async getRecommendations(params: { evaluate?: boolean } = {}): Promise<RecommendationsResponse> {
    const query = params.evaluate ? '?evaluate=true' : '';
    const res = await apiFetch(`/recommendations${query}`);
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to retrieve recommendations');
    }
    const data: ApiResponse<RecommendationsResponse> = await res.json();
    return (
      data.data || {
        items: [],
        total: 0,
        user_data_tier: 'cold_start',
        generated_at: new Date().toISOString(),
      }
    );
  },

  async updateRecommendationStatus(
    id: string,
    status: 'dismissed' | 'completed'
  ): Promise<RecommendationItem> {
    const res = await apiFetch(`/recommendations/${id}/status`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ status }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to update recommendation status');
    }
    const data: ApiResponse<RecommendationItem> = await res.json();
    if (!data.data) throw new Error('No updated recommendation data returned');
    return data.data;
  },

  // =========================================================================
  // 6. Groq LLM Academic Productivity Coach APIs
  // =========================================================================
  async askAI(prompt: string): Promise<string> {
    const res = await apiFetch('/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt: prompt.trim() }),
    });
    if (!res.ok) {
      throw new Error(await parseAIResponseError(res, 'Gagal menghubungi asisten AI.'));
    }
    const data: ApiResponse<{ reply: string }> = await res.json();
    return data.data?.reply || '';
  },

  async explainPrediction(predictionId?: string): Promise<string> {
    const res = await apiFetch('/ai/explain-prediction', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(predictionId ? { prediction_id: predictionId } : {}),
    });
    if (!res.ok) {
      throw new Error(
        await parseAIResponseError(res, 'Gagal menjelaskan prediksi.', 'Data prediksi belum tersedia untuk dijelaskan.')
      );
    }
    const data: ApiResponse<{ reply: string }> = await res.json();
    return data.data?.reply || '';
  },

  async generateStudyPlan(): Promise<string> {
    const res = await apiFetch('/ai/study-plan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    if (!res.ok) {
      throw new Error(await parseAIResponseError(res, 'Gagal menyusun rencana belajar.'));
    }
    const data: ApiResponse<{ reply: string }> = await res.json();
    return data.data?.reply || '';
  },
};

async function parseAIResponseError(res: Response, fallback: string, notFoundMsg?: string): Promise<string> {
  const err = await res.json().catch(() => ({} as { message?: string }));
  if (res.status === 401) {
    return 'Sesi login tidak valid. Silakan login kembali.';
  }
  if (res.status === 403) {
    return err.message || 'Anda tidak memiliki izin untuk mengakses asisten AI.';
  }
  if (res.status === 404) {
    return err.message || notFoundMsg || 'Layanan AI tidak ditemukan pada server backend.';
  }
  if (res.status === 429) {
    return err.message || 'Layanan AI sedang sibuk. Silakan coba beberapa saat lagi.';
  }
  if (res.status === 504) {
    return err.message || 'Koneksi ke layanan AI mengalami timeout.';
  }
  if (res.status === 502 || res.status === 503) {
    return err.message || 'Layanan AI sedang tidak tersedia.';
  }
  return err.message || fallback;
}
