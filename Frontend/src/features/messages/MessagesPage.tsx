import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import { Button } from '../../components/atoms';
import { MarkdownContent } from '../../components/molecules';
import {
  SparklesIcon,
  SendIcon,
  ShieldIcon,
  ArrowRightIcon,
  ClockIcon,
} from '../../components/atoms/Icons';
import type { UserProfile, AIChatMessage } from '../../types';
import { useAuth } from '../../hooks';
import { productivityService } from '../../services/productivityService';

function createChatMessage(
  sender: 'user' | 'assistant',
  content: string,
  idOverride?: string
): AIChatMessage {
  const now = new Date();
  return {
    id: idOverride || `${sender}-${now.getTime()}`,
    sender,
    content,
    timestamp: now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
  };
}

export default function MessagesPage() {
  const { user } = useAuth();
  const [messages, setMessages] = useState<AIChatMessage[]>(() => [
    createChatMessage(
      'assistant',
      'Halo! Aku teman belajar Mindora-mu. Kamu bisa ngobrol santai denganku untuk bantu menyusun alur belajar yang fleksibel, memahami ritme produktivitasmu, atau cari tips agar tetap fokus tanpa cepat lelah.\n\nAda yang sedang ingin kamu tanyakan atau diskusikan hari ini?',
      'welcome'
    ),
  ]);
  const [inputPrompt, setInputPrompt] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [lastFailedPrompt, setLastFailedPrompt] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

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
            .toUpperCase() || 'MH',
        workspaceName: 'Ruang Kerja Akademik',
      }
    : {
        id: 'usr_guest',
        name: 'Mahasiswa',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'MH',
        workspaceName: 'Ruang Kerja Akademik',
      };

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  const handleSendMessage = async (customPrompt?: string) => {
    const promptToSend = (customPrompt || inputPrompt).trim();
    if (!promptToSend || isLoading) return;

    const userMsg = createChatMessage('user', promptToSend);

    setMessages((prev) => [...prev, userMsg]);
    setInputPrompt('');
    setIsLoading(true);
    setErrorMsg(null);
    setLastFailedPrompt(null);

    try {
      let replyText = '';
      if (promptToSend === 'Susun Jadwal Belajar Hari Ini' || promptToSend === 'Rencanakan hari saya') {
        replyText = await productivityService.generateStudyPlan();
      } else if (
        promptToSend === 'Jelaskan Estimasi Produktivitas Saya' ||
        promptToSend === 'Jelaskan estimasi produktivitas'
      ) {
        replyText = await productivityService.explainPrediction();
      } else {
        replyText = await productivityService.askAI(promptToSend);
      }

      const aiMsg = createChatMessage('assistant', replyText);

      setMessages((prev) => [...prev, aiMsg]);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menghubungi asisten AI';
      setErrorMsg(msg);
      setLastFailedPrompt(promptToSend);
    } finally {
      setIsLoading(false);
    }
  };

  const quickActionPrompts = [
    { label: '📅 Susun Alur Belajar Hari Ini', action: 'Rencanakan hari saya' },
    { label: '💡 Jelaskan Ritme Produktivitasku', action: 'Jelaskan estimasi produktivitas' },
    {
      label: '⚡ Tips Biar Nggak Tergoda Main HP',
      action: 'Bagaimana tips praktis agar aku tidak tergoda bermain HP saat sesi belajar?',
    },
    {
      label: '🎯 Cara Mulai Belajar Saat Lagi Mager',
      action: 'Aku lagi kurang semangat dan agak malas mulai belajar, enaknya gimana ya?',
    },
  ];

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-5xl mx-auto">
        {/* Header Section */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-emerald-100 text-emerald-900 border border-emerald-200">
                <span className="w-2 h-2 rounded-full bg-emerald-600 animate-pulse" />
                Teman Belajar AI Aktif
              </span>
              <span className="text-xs text-text-muted">Pendamping Belajar & Manajemen Waktu</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Ruang Teman Belajar & Konsultasi AI
            </h1>
            <p className="text-sm text-text-secondary">
              Konsultasikan alur belajar, strategi fokus, atau minta saran saat kamu merasa lelah.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Link to="/recommendations">
              <Button
                variant="outline"
                size="sm"
                className="gap-2 text-xs font-semibold"
              >
                <SparklesIcon className="w-4 h-4 text-emerald-600" />
                Pusat Rekomendasi
              </Button>
            </Link>
          </div>
        </div>

        {/* Quick Action Suggestion Bar */}
        <div className="space-y-2">
          <div className="text-[11px] font-bold uppercase tracking-wider text-text-secondary flex items-center gap-1.5">
            <SparklesIcon className="w-3.5 h-3.5 text-primary" />
            <span>Aksi Cepat Konsultasi:</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {quickActionPrompts.map((item, idx) => (
              <button
                key={idx}
                type="button"
                disabled={isLoading}
                onClick={() => handleSendMessage(item.action)}
                className="px-3 py-1.5 rounded-xl bg-white border border-border hover:border-primary/50 hover:bg-primary-50/50 text-xs font-medium text-text-primary transition-all cursor-pointer shadow-soft-xs disabled:opacity-50"
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* Main Conversation Container */}
        <div className="rounded-2xl border border-border bg-surface-subtle shadow-soft-sm flex flex-col h-[580px] overflow-hidden">
          {/* Chat Messages Stream */}
          <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-4">
            {messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex gap-3 ${
                  msg.sender === 'user' ? 'justify-end' : 'justify-start'
                }`}
              >
                {msg.sender === 'assistant' && (
                  <div className="w-8 h-8 rounded-xl bg-primary text-white flex items-center justify-center shrink-0 shadow-soft-xs">
                    <SparklesIcon className="w-4 h-4" />
                  </div>
                )}

                <div
                  className={`max-w-[85%] sm:max-w-[75%] rounded-2xl p-4 text-xs sm:text-sm leading-relaxed ${
                    msg.sender === 'user'
                      ? 'bg-primary text-white shadow-soft-xs rounded-br-xs'
                      : 'bg-white border border-border/80 text-text-primary shadow-soft-xs rounded-bl-xs'
                  }`}
                >
                  <MarkdownContent content={msg.content} />
                  <div
                    className={`text-[10px] mt-2 flex items-center gap-1 ${
                      msg.sender === 'user' ? 'text-primary-100 justify-end' : 'text-text-muted justify-end'
                    }`}
                  >
                    <ClockIcon className="w-3 h-3" />
                    <span>{msg.timestamp}</span>
                  </div>
                </div>

                {msg.sender === 'user' && (
                  <div className="w-8 h-8 rounded-xl bg-primary-100 text-primary-900 font-bold text-xs flex items-center justify-center shrink-0">
                    {displayUser.avatarText}
                  </div>
                )}
              </div>
            ))}

            {/* Loading Typing Indicator */}
            {isLoading && (
              <div className="flex gap-3 justify-start animate-fadeIn">
                <div className="w-8 h-8 rounded-xl bg-primary text-white flex items-center justify-center shrink-0 shadow-soft-xs">
                  <SparklesIcon className="w-4 h-4" />
                </div>
                <div className="bg-white border border-border/80 rounded-2xl rounded-bl-xs p-3.5 shadow-soft-xs flex items-center gap-2.5 text-xs text-text-secondary">
                  <div className="flex gap-1">
                    <span className="w-2 h-2 rounded-full bg-primary animate-bounce [animation-delay:-0.3s]" />
                    <span className="w-2 h-2 rounded-full bg-primary animate-bounce [animation-delay:-0.15s]" />
                    <span className="w-2 h-2 rounded-full bg-primary animate-bounce" />
                  </div>
                  <span>Mindora AI sedang menganalisis dan menyusun rekomendasi...</span>
                </div>
              </div>
            )}

            {/* Error Notification inside Stream */}
            {errorMsg && (
              <div className="p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-xs text-rose-800 flex items-center justify-between gap-2 animate-fadeIn">
                <span>{errorMsg}</span>
                <div className="flex items-center gap-3 shrink-0">
                  {lastFailedPrompt && (
                    <button
                      type="button"
                      onClick={() => handleSendMessage(lastFailedPrompt)}
                      className="font-bold text-rose-700 underline hover:no-underline cursor-pointer"
                    >
                      Coba Lagi
                    </button>
                  )}
                  <button
                    type="button"
                    onClick={() => {
                      setErrorMsg(null);
                      setLastFailedPrompt(null);
                    }}
                    className="font-bold underline hover:no-underline cursor-pointer"
                  >
                    Tutup
                  </button>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {/* Privacy & Safety Note */}
          <div className="px-4 py-2 bg-slate-50 border-t border-border/60 flex items-center justify-between text-[11px] text-text-muted">
            <span className="inline-flex items-center gap-1.5">
              <ShieldIcon className="w-3.5 h-3.5 text-emerald-600" />
              <span>Privasi Terjamin: Identitas dan kata sandi Anda tidak pernah ditransmisikan.</span>
            </span>
            <Link
              to="/analytics"
              className="font-semibold text-primary hover:underline hidden sm:inline-flex items-center gap-1"
            >
              <span>Lihat Analisis Belajar</span>
              <ArrowRightIcon className="w-3 h-3" />
            </Link>
          </div>

          {/* Chat Input Bar */}
          <div className="p-4 bg-white border-t border-border">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSendMessage();
              }}
              className="relative flex items-center"
            >
              <input
                type="text"
                value={inputPrompt}
                disabled={isLoading}
                onChange={(e) => setInputPrompt(e.target.value)}
                placeholder="Tulis pesan atau pertanyaan santai untuk teman belajar Mindora..."
                className="w-full pl-4 pr-12 py-3.5 rounded-xl border border-border bg-surface-subtle text-sm text-text-primary placeholder:text-text-secondary/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary focus:bg-white transition-all shadow-soft-sm disabled:opacity-50"
              />
              <button
                type="submit"
                disabled={isLoading || !inputPrompt.trim()}
                className="absolute right-2 p-2 rounded-lg bg-primary hover:bg-primary-hover text-white transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                aria-label="Kirim Pesan"
              >
                <SendIcon className="w-4 h-4" />
              </button>
            </form>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}
