import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { SparklesIcon, SendIcon, CheckIcon, CloseIcon, ArrowRightIcon } from '../../components/atoms';
import { QuickActionChips, MarkdownContent } from '../../components/molecules';
import { cn, translatePriority } from '../../utils';
import { productivityService } from '../../services/productivityService';
import type { RecommendationItem } from '../../types';

export interface AIAssistantWidgetProps {
  className?: string;
  onAskAI?: (prompt: string) => void;
}

export const AIAssistantWidget = ({
  className,
  onAskAI,
}: AIAssistantWidgetProps) => {
  const [prompt, setPrompt] = useState('');
  const [loadingAI, setLoadingAI] = useState<boolean>(false);
  const [aiResponse, setAiResponse] = useState<string | null>(null);
  const [aiError, setAiError] = useState<string | null>(null);
  const [lastFailedPrompt, setLastFailedPrompt] = useState<string | null>(null);
  const [urgentRecs, setUrgentRecs] = useState<RecommendationItem[]>([]);
  const [loadingRecs, setLoadingRecs] = useState<boolean>(true);

  // Fetch top 1-2 active recommendations (read-only evaluate=false)
  useEffect(() => {
    let isMounted = true;
    const fetchTopRecs = async () => {
      try {
        const res = await productivityService.getRecommendations({ evaluate: false });
        if (isMounted) {
          setUrgentRecs(res.items.slice(0, 2));
        }
      } catch (err: unknown) {
        console.error('Failed to load dashboard recommendations:', err);
      } finally {
        if (isMounted) setLoadingRecs(false);
      }
    };

    fetchTopRecs();
    return () => {
      isMounted = false;
    };
  }, []);

  const handleDismissRec = async (id: string) => {
    try {
      await productivityService.updateRecommendationStatus(id, 'dismissed');
      setUrgentRecs((prev) => prev.filter((r) => r.id !== id));
    } catch (err: unknown) {
      console.error('Failed to dismiss recommendation:', err);
    }
  };

  const handleCompleteRec = async (id: string) => {
    try {
      await productivityService.updateRecommendationStatus(id, 'completed');
      setUrgentRecs((prev) => prev.filter((r) => r.id !== id));
    } catch (err: unknown) {
      console.error('Failed to complete recommendation:', err);
    }
  };

  const quickPrompts = [
    'Rencanakan hari saya',
    'Jelaskan estimasi produktivitas',
    'Tips fokus belajar',
  ];

  const handleSend = async (textToSend?: string) => {
    const text = (textToSend || prompt).trim();
    if (!text || loadingAI) return;

    setLoadingAI(true);
    setAiError(null);
    setAiResponse(null);
    setLastFailedPrompt(null);

    try {
      let reply = '';
      if (text === 'Rencanakan hari saya') {
        reply = await productivityService.generateStudyPlan();
      } else if (text === 'Jelaskan estimasi produktivitas') {
        reply = await productivityService.explainPrediction();
      } else {
        reply = await productivityService.askAI(text);
      }
      setAiResponse(reply);
      onAskAI?.(text);
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : 'Gagal menghubungi asisten AI';
      setAiError(errMsg);
      setLastFailedPrompt(text);
    } finally {
      setLoadingAI(false);
      setPrompt('');
    }
  };

  return (
    <div
      className={cn(
        'bg-surface rounded-2xl border border-border p-6 shadow-soft-sm space-y-4 hover:border-primary/30 transition-all',
        className
      )}
    >
      {/* Header with Sparkle Badge */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary-100 text-primary-900 text-xs font-semibold">
          <SparklesIcon className="w-3.5 h-3.5 text-primary" />
          <span>ASISTEN PRODUKTIVITAS AI</span>
        </div>
        <span className="text-[11px] font-medium text-text-secondary">
          Didukung model cerdas Groq & pola belajar Anda
        </span>
      </div>

      {/* Top Urgent Recommendation Integration */}
      {!loadingRecs && urgentRecs.length > 0 && (
        <div className="space-y-2 pt-1 pb-1">
          <div className="flex items-center justify-between text-[11px] font-bold text-text-secondary">
            <span className="uppercase tracking-wider text-[#138F78] flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-[#138F78] animate-pulse" />
              Wawasan Produktivitas Mendesak
            </span>
            <Link
              to="/recommendations"
              className="text-[#138F78] hover:text-[#123C32] hover:underline transition-colors inline-flex items-center gap-1"
            >
              <span>Lihat Pusat Rekomendasi ({urgentRecs.length})</span>
              <ArrowRightIcon className="w-3 h-3" />
            </Link>
          </div>

          <div className="space-y-2">
            {urgentRecs.map((rec) => (
              <div
                key={rec.id}
                className="p-3.5 rounded-xl bg-white border border-emerald-200/80 shadow-soft-xs hover:border-emerald-400 transition-all space-y-2"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2">
                      <span
                        className={`text-[9px] font-black uppercase tracking-wider px-2 py-0.5 rounded ${
                          rec.priority === 'critical'
                            ? 'bg-rose-100 text-rose-800'
                            : rec.priority === 'high'
                            ? 'bg-amber-100 text-amber-900'
                            : 'bg-emerald-100 text-[#123C32]'
                        }`}
                      >
                        {translatePriority(rec.priority)}
                      </span>
                      <span className="text-xs font-bold text-[#123C32]">
                        {rec.title}
                      </span>
                    </div>
                    <p className="text-[11px] text-[#526B60] leading-snug pt-0.5">
                      {rec.message}
                    </p>
                  </div>

                  <button
                    type="button"
                    onClick={() => handleDismissRec(rec.id)}
                    className="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer shrink-0"
                    title="Abaikan (48 jam)"
                    aria-label="Abaikan rekomendasi"
                  >
                    <CloseIcon className="w-3 h-3" />
                  </button>
                </div>

                <div className="flex items-center justify-between pt-1 border-t border-slate-100 text-[11px]">
                  {rec.action_route ? (
                    <Link
                      to={rec.action_route}
                      className="font-bold text-[#138F78] hover:text-[#123C32] hover:underline inline-flex items-center gap-1"
                    >
                      <span>{rec.action_label || 'Ambil Tindakan'}</span>
                      <ArrowRightIcon className="w-3 h-3" />
                    </Link>
                  ) : (
                    <span />
                  )}

                  <button
                    type="button"
                    onClick={() => handleCompleteRec(rec.id)}
                    className="inline-flex items-center gap-1 font-semibold text-slate-500 hover:text-emerald-700 cursor-pointer"
                  >
                    <CheckIcon className="w-3 h-3 text-[#138F78]" />
                    <span>Selesai</span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Headline */}
      <h3 className="text-lg sm:text-xl font-bold tracking-tight text-text-primary">
        Ada yang ingin kamu diskusikan hari ini?
      </h3>

      {/* Quick Action Chips */}
      <QuickActionChips
        chips={quickPrompts}
        onSelect={(chip) => handleSend(chip)}
        size="sm"
      />

      {/* Input Prompt Box */}
      <div className="relative flex items-center pt-1">
        <input
          type="text"
          value={prompt}
          disabled={loadingAI}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          placeholder="Tanya tips belajar, saran alur hari ini, atau cara fokus..."
          className="w-full pl-4 pr-12 py-3 rounded-xl border border-border bg-surface-subtle text-sm text-text-primary placeholder:text-text-secondary/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary focus:bg-white transition-all shadow-soft-sm disabled:opacity-50"
        />
        <button
          type="button"
          disabled={loadingAI || !prompt.trim()}
          onClick={() => handleSend()}
          className="absolute right-2 p-2 rounded-lg bg-primary hover:bg-primary-hover text-white transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
          aria-label="Kirim ke Asisten AI"
        >
          <SendIcon className="w-4 h-4" />
        </button>
      </div>

      {/* Loading state */}
      {loadingAI && (
        <div className="p-4 rounded-xl bg-primary-50/70 border border-primary-200 text-xs text-primary-900 leading-relaxed font-medium animate-fadeIn flex items-center gap-3">
          <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin shrink-0" />
          <span>Mindora AI sedang menganalisis aktivitas dan menyusun respons terbaik untuk Anda...</span>
        </div>
      )}

      {/* Error state */}
      {aiError && (
        <div className="p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-xs text-rose-800 leading-relaxed font-medium animate-fadeIn flex items-center justify-between gap-2">
          <span>{aiError}</span>
          <div className="flex items-center gap-2 shrink-0">
            {lastFailedPrompt && (
              <button
                type="button"
                onClick={() => handleSend(lastFailedPrompt)}
                className="font-bold text-rose-700 underline hover:no-underline cursor-pointer"
              >
                Coba Lagi
              </button>
            )}
            <button
              type="button"
              onClick={() => {
                setAiError(null);
                setLastFailedPrompt(null);
              }}
              className="text-rose-500 hover:text-rose-700 p-0.5 cursor-pointer"
            >
              <CloseIcon className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}

      {/* AI Response Box */}
      {aiResponse && (
        <div className="p-4 rounded-xl bg-white border border-primary-200/80 shadow-soft-xs text-xs text-text-primary leading-relaxed space-y-3 animate-fadeIn">
          <div className="flex items-center justify-between pb-2 border-b border-border/60">
            <div className="flex items-center gap-2 text-primary font-bold">
              <SparklesIcon className="w-4 h-4" />
              <span>Rekomendasi Asisten Mindora AI</span>
            </div>
            <button
              type="button"
              onClick={() => setAiResponse(null)}
              className="text-text-muted hover:text-text-primary p-1 rounded hover:bg-slate-100 transition-colors cursor-pointer"
              title="Tutup tanggapan"
            >
              <CloseIcon className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="max-h-72 overflow-y-auto pr-1 text-slate-700 leading-relaxed">
            <MarkdownContent content={aiResponse} />
          </div>

          <div className="pt-2 border-t border-border/60 flex items-center justify-between text-[11px]">
            <span className="text-text-muted">Ingin eksplorasi lebih lanjut?</span>
            <Link
              to="/messages"
              className="text-primary font-bold hover:underline inline-flex items-center gap-1"
            >
              <span>Buka Ruang Percakapan Lengkap</span>
              <ArrowRightIcon className="w-3 h-3" />
            </Link>
          </div>
        </div>
      )}
    </div>
  );
};
