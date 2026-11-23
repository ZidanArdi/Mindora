import { apiFetch } from './api';
import type { ApiResponse, AIResponseData } from '../types';

export const aiService = {
  /**
   * Send a free-form message or question to the Mindora AI Academic Productivity Coach.
   */
  async chat(prompt: string): Promise<AIResponseData> {
    const trimmed = prompt.trim();
    if (!trimmed) {
      throw new Error('Pertanyaan atau pesan tidak boleh kosong.');
    }
    const res = await apiFetch('/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt: trimmed }),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => null);
      throw new Error(errData?.message || `Gagal menghubungi asisten AI (${res.status})`);
    }

    const json: ApiResponse<AIResponseData> = await res.json();
    if (!json.data) {
      throw new Error('Respons asisten AI tidak valid.');
    }
    return json.data;
  },

  /**
   * Request a detailed explanation of a productivity prediction score.
   */
  async explainPrediction(predictionId?: string): Promise<AIResponseData> {
    const body: Record<string, string> = {};
    if (predictionId) {
      body.prediction_id = predictionId;
    }

    const res = await apiFetch('/ai/explain-prediction', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => null);
      throw new Error(errData?.message || `Gagal menjelaskan prediksi (${res.status})`);
    }

    const json: ApiResponse<AIResponseData> = await res.json();
    if (!json.data) {
      throw new Error('Respons penjelasan prediksi tidak valid.');
    }
    return json.data;
  },

  /**
   * Request an optimized daily study plan based on tasks and focus stamina.
   */
  async generateStudyPlan(): Promise<AIResponseData> {
    const res = await apiFetch('/ai/study-plan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => null);
      throw new Error(errData?.message || `Gagal menyusun jadwal belajar (${res.status})`);
    }

    const json: ApiResponse<AIResponseData> = await res.json();
    if (!json.data) {
      throw new Error('Respons jadwal belajar tidak valid.');
    }
    return json.data;
  },
};
