import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import {
  Button,
  BellIcon,
  ClockIcon,
  LogoutIcon,
  ExportIcon,
  PlusIcon,
  MenuIcon,
  Avatar,
  UserIcon,
  SettingsIcon,
  SparklesIcon,
  ArrowRightIcon,
} from '../atoms';
import { SearchInput } from '../molecules';
import type { UserProfile, TaskItem, RecommendationItem } from '../../types';
import { cn, translateCategory, translatePriority } from '../../utils';
import { useAuth } from '../../hooks';
import { productivityService } from '../../services/productivityService';

export interface TopNavbarProps {
  role: 'user' | 'admin';
  currentUser: UserProfile;
  onOpenMobileMenu?: () => void;
  className?: string;
}

interface InAppNotification {
  id: string;
  title: string;
  message: string;
  priority: string;
  read: boolean;
  actionUrl?: string;
  timestamp: string;
}

export const TopNavbar = ({
  role,
  currentUser,
  onOpenMobileMenu,
  className,
}: TopNavbarProps) => {
  const navigate = useNavigate();
  const { logout } = useAuth();

  // Search state
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<TaskItem[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [showSearchDropdown, setShowSearchDropdown] = useState(false);
  const searchContainerRef = useRef<HTMLDivElement>(null);
  const searchReqIdRef = useRef(0);

  // Notification state
  const [notifications, setNotifications] = useState<InAppNotification[]>([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const notificationContainerRef = useRef<HTMLDivElement>(null);

  // Profile menu state
  const [showProfileMenu, setShowProfileMenu] = useState(false);
  const profileContainerRef = useRef<HTMLDivElement>(null);

  // Fetch initial in-app alert recommendations (if student)
  useEffect(() => {
    let isMounted = true;
    if (role === 'user') {
      productivityService
        .getRecommendations({ evaluate: false })
        .then((res) => {
          if (!isMounted) return;
          const items = (res.items || []).slice(0, 5).map((rec: RecommendationItem) => ({
            id: rec.id,
            title: rec.title,
            message: rec.message,
            priority: rec.priority,
            read: false,
            actionUrl: '/recommendations',
            timestamp: 'Pemberitahuan aktif',
          }));
          setNotifications(items);
        })
        .catch(() => {
          // Gracefully omit if offline or unavailable
        });
    }
    return () => {
      isMounted = false;
    };
  }, [role]);

  // Handle outside clicks to close popovers
  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      const target = e.target as Node;
      if (
        searchContainerRef.current &&
        !searchContainerRef.current.contains(target)
      ) {
        setShowSearchDropdown(false);
      }
      if (
        notificationContainerRef.current &&
        !notificationContainerRef.current.contains(target)
      ) {
        setShowNotifications(false);
      }
      if (
        profileContainerRef.current &&
        !profileContainerRef.current.contains(target)
      ) {
        setShowProfileMenu(false);
      }
    };

    document.addEventListener('mousedown', handleOutsideClick);
    return () => {
      document.removeEventListener('mousedown', handleOutsideClick);
    };
  }, []);

  // Debounced search effect
  useEffect(() => {
    const trimmed = searchQuery.trim();
    if (!trimmed) {
      return;
    }

    const reqId = ++searchReqIdRef.current;
    const timer = setTimeout(async () => {
      setIsSearching(true);
      setHasSearched(false);
      try {
        const data = await productivityService.listTasks({
          search: trimmed,
          limit: 5,
        });
        if (searchReqIdRef.current === reqId) {
          setSearchResults(data.tasks || []);
          setHasSearched(true);
          setShowSearchDropdown(true);
        }
      } catch (err: unknown) {
        if (searchReqIdRef.current === reqId) {
          console.error('Header search error:', err);
          setSearchResults([]);
          setHasSearched(true);
        }
      } finally {
        if (searchReqIdRef.current === reqId) {
          setIsSearching(false);
        }
      }
    }, 300);

    return () => {
      clearTimeout(timer);
    };
  }, [searchQuery]);

  const handleSearchSubmit = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = searchQuery.trim();
    setShowSearchDropdown(false);
    if (trimmed) {
      navigate(`/tasks?search=${encodeURIComponent(trimmed)}`);
    } else {
      navigate('/tasks');
    }
  };

  const handleSelectSearchResult = (task: TaskItem) => {
    setShowSearchDropdown(false);
    setSearchQuery('');
    navigate(`/tasks?search=${encodeURIComponent(task.title)}`);
  };

  const handleClearSearch = () => {
    setSearchQuery('');
    setSearchResults([]);
    setShowSearchDropdown(false);
  };

  const handleMarkAllNotificationsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
    toast.success('Semua notifikasi ditandai telah dibaca.');
  };

  const handleClickNotification = (item: InAppNotification) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === item.id ? { ...n, read: true } : n))
    );
    setShowNotifications(false);
    if (item.actionUrl) {
      navigate(item.actionUrl);
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
      toast.success('Berhasil keluar dari akun.');
    } finally {
      navigate('/login');
    }
  };

  const unreadCount = notifications.filter((n) => !n.read).length;

  return (
    <header
      className={cn(
        'sticky top-0 z-30 bg-surface/90 backdrop-blur-md border-b border-border px-4 lg:px-8 py-3.5 flex items-center justify-between gap-4 transition-all',
        className
      )}
    >
      {/* Left: Mobile hamburger & search */}
      <div
        ref={searchContainerRef}
        className="relative flex items-center gap-3 flex-1 max-w-lg"
      >
        {onOpenMobileMenu && (
          <button
            type="button"
            onClick={onOpenMobileMenu}
            className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted lg:hidden cursor-pointer"
            aria-label="Buka navigasi bilah sisi"
          >
            <MenuIcon className="w-5 h-5" />
          </button>
        )}

        <form onSubmit={handleSearchSubmit} className="w-full relative">
          <SearchInput
            placeholder={
              role === 'user'
                ? 'Cari tugas...'
                : 'Cari anggota, tugas, ruang kerja...'
            }
            value={searchQuery}
            onChange={(e) => {
              const val = e.target.value;
              setSearchQuery(val);
              if (!val.trim()) {
                setSearchResults([]);
                setIsSearching(false);
                setHasSearched(false);
                setShowSearchDropdown(false);
              } else if (!showSearchDropdown) {
                setShowSearchDropdown(true);
              }
            }}
            onFocus={() => {
              if (searchQuery.trim()) {
                setShowSearchDropdown(true);
              }
            }}
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                setShowSearchDropdown(false);
              }
            }}
            onClear={handleClearSearch}
            isLoading={isSearching}
            className="w-full"
          />
        </form>

        {/* Global Search Results Dropdown */}
        {showSearchDropdown && (searchQuery.trim().length > 0) && (
          <div className="absolute left-0 right-0 top-full mt-2 bg-white rounded-2xl border border-border shadow-soft-xl z-50 overflow-hidden animate-fadeIn">
            <div className="p-3 border-b border-border-subtle flex items-center justify-between text-xs text-text-secondary bg-surface-muted/50">
              <span className="font-semibold text-text-primary">
                {isSearching ? 'Mencari...' : `Hasil Pencarian (${searchResults.length})`}
              </span>
              <button
                type="button"
                onClick={() => setShowSearchDropdown(false)}
                className="text-text-muted hover:text-text-primary text-[11px] cursor-pointer"
              >
                Tutup
              </button>
            </div>

            <div className="max-h-72 overflow-y-auto divide-y divide-border-subtle">
              {searchResults.map((task) => (
                <div
                  key={task.id}
                  onClick={() => handleSelectSearchResult(task)}
                  className="p-3 hover:bg-primary-50/40 cursor-pointer transition-colors flex items-center justify-between gap-3 text-xs"
                >
                  <div className="space-y-1 min-w-0 flex-1">
                    <p className="font-semibold text-text-primary truncate">
                      {task.title}
                    </p>
                    <div className="flex items-center gap-2 text-[11px] text-text-muted">
                      <span className="px-1.5 py-0.5 rounded-md bg-surface-muted border border-border text-text-secondary">
                        {translateCategory(task.category)}
                      </span>
                      {task.priority && (
                        <span
                          className={cn(
                            'px-1.5 py-0.5 rounded-md text-[10px] font-bold uppercase',
                            task.priority === 'urgent'
                              ? 'bg-rose-100 text-rose-800'
                              : task.priority === 'high'
                              ? 'bg-amber-100 text-amber-800'
                              : 'bg-emerald-100 text-emerald-800'
                          )}
                        >
                          {translatePriority(task.priority)}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="shrink-0 flex items-center gap-2">
                    <span
                      className={cn(
                        'text-[10px] font-semibold px-2 py-0.5 rounded-full',
                        task.status === 'completed'
                          ? 'bg-emerald-100 text-emerald-800'
                          : 'bg-slate-100 text-slate-700'
                      )}
                    >
                      {task.status === 'completed' ? 'Selesai' : 'Tertunda'}
                    </span>
                    <ArrowRightIcon className="w-3.5 h-3.5 text-text-muted" />
                  </div>
                </div>
              ))}

              {!isSearching && hasSearched && searchResults.length === 0 && (
                <div className="p-6 text-center text-xs text-text-muted space-y-2">
                  <p>Tidak ada tugas yang cocok dengan <strong className="text-text-primary">"{searchQuery}"</strong>.</p>
                  <button
                    type="button"
                    onClick={() => {
                      setShowSearchDropdown(false);
                      navigate('/tasks');
                    }}
                    className="text-xs font-semibold text-primary hover:underline cursor-pointer"
                  >
                    Buka Halaman Tugas
                  </button>
                </div>
              )}
            </div>

            <div className="p-2.5 bg-slate-50 border-t border-border-subtle text-center">
              <button
                type="button"
                onClick={() => handleSearchSubmit()}
                className="text-xs font-semibold text-primary hover:underline cursor-pointer flex items-center justify-center gap-1.5 w-full"
              >
                <span>Lihat semua hasil di Manajemen Tugas</span>
                <ArrowRightIcon className="w-3 h-3" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Right: Actions, Notifications, User */}
      <div className="flex items-center gap-3">
        {role === 'user' ? (
          <div className="hidden sm:flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate('/reports')}
              className="gap-2 text-xs font-semibold cursor-pointer"
              title="Ekspor ringkasan & laporan produktivitas"
            >
              <ExportIcon className="w-3.5 h-3.5 text-text-secondary" />
              <span>Ekspor ringkasan</span>
            </Button>
          </div>
        ) : (
          <Button
            variant="primary"
            size="sm"
            onClick={() => navigate('/admin')}
            className="gap-2 text-xs font-semibold cursor-pointer"
            title="Kelola & pantau anggota ruang kerja"
          >
            <PlusIcon className="w-3.5 h-3.5" />
            <span>Undang anggota</span>
          </Button>
        )}

        <div className="flex items-center gap-1.5 border-l border-border pl-3">
          {/* Activity / Time Tracking Icon */}
          {role === 'user' && (
            <button
              type="button"
              onClick={() => navigate('/focus')}
              className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors relative hidden sm:flex cursor-pointer"
              title="Pelacak waktu & riwayat fokus"
              aria-label="Pelacak waktu dan riwayat sesi fokus"
            >
              <ClockIcon className="w-4 h-4" />
            </button>
          )}

          {/* Notifications Bell & Dropdown */}
          <div ref={notificationContainerRef} className="relative">
            <button
              type="button"
              onClick={() => setShowNotifications((prev) => !prev)}
              className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors relative cursor-pointer"
              title="Notifikasi"
              aria-label="Buka notifikasi"
            >
              <BellIcon className="w-4 h-4" />
              {unreadCount > 0 && (
                <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-emerald-500 ring-2 ring-white" />
              )}
            </button>

            {showNotifications && (
              <div className="absolute right-0 top-full mt-2 w-80 sm:w-96 bg-white rounded-2xl border border-border shadow-soft-xl z-50 overflow-hidden animate-fadeIn">
                <div className="p-3.5 border-b border-border flex items-center justify-between bg-surface-muted/50">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-text-primary">
                      Notifikasi
                    </span>
                    {unreadCount > 0 && (
                      <span className="px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-primary text-white">
                        {unreadCount} baru
                      </span>
                    )}
                  </div>
                  {unreadCount > 0 && (
                    <button
                      type="button"
                      onClick={handleMarkAllNotificationsRead}
                      className="text-[11px] font-semibold text-primary hover:underline cursor-pointer"
                    >
                      Tandai dibaca
                    </button>
                  )}
                </div>

                <div className="max-h-80 overflow-y-auto divide-y divide-border-subtle">
                  {notifications.map((item) => (
                    <div
                      key={item.id}
                      onClick={() => handleClickNotification(item)}
                      className={cn(
                        'p-3.5 hover:bg-surface-muted/80 cursor-pointer transition-colors text-xs flex items-start gap-3',
                        !item.read && 'bg-primary-50/30'
                      )}
                    >
                      <div className="w-7 h-7 rounded-xl bg-primary-100 text-primary-800 flex items-center justify-center shrink-0 mt-0.5">
                        <SparklesIcon className="w-3.5 h-3.5" />
                      </div>
                      <div className="space-y-1 flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-2">
                          <p className="font-semibold text-text-primary truncate">
                            {item.title}
                          </p>
                          {!item.read && (
                            <span className="w-1.5 h-1.5 rounded-full bg-primary shrink-0" />
                          )}
                        </div>
                        <p className="text-[11px] text-text-secondary leading-relaxed line-clamp-2">
                          {item.message}
                        </p>
                        <p className="text-[10px] text-text-muted">
                          {item.timestamp}
                        </p>
                      </div>
                    </div>
                  ))}

                  {notifications.length === 0 && (
                    <div className="p-8 text-center text-xs text-text-muted space-y-2">
                      <div className="w-10 h-10 rounded-2xl bg-slate-100 text-text-muted flex items-center justify-center mx-auto">
                        <BellIcon className="w-5 h-5 text-text-muted" />
                      </div>
                      <p className="font-semibold text-text-primary">
                        Belum ada notifikasi baru
                      </p>
                      <p className="text-[11px] text-text-secondary leading-relaxed max-w-xs mx-auto">
                        Anda akan menerima pemberitahuan otomatis saat rekomendasi belajar atau pengingat sesi fokus tersedia.
                      </p>
                    </div>
                  )}
                </div>

                {notifications.length > 0 && (
                  <div className="p-2.5 bg-slate-50 border-t border-border-subtle text-center">
                    <button
                      type="button"
                      onClick={() => {
                        setShowNotifications(false);
                        navigate('/recommendations');
                      }}
                      className="text-xs font-semibold text-primary hover:underline cursor-pointer"
                    >
                      Buka Pusat Rekomendasi
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Quick Header Logout Button */}
          <button
            type="button"
            onClick={handleLogout}
            className="p-2 rounded-xl text-text-secondary hover:text-red-600 hover:bg-surface-muted transition-colors hidden sm:flex cursor-pointer"
            title="Keluar"
            aria-label="Keluar dari akun"
          >
            <LogoutIcon className="w-4 h-4" />
          </button>

          {/* Profile Avatar & Interactive Menu */}
          <div ref={profileContainerRef} className="relative">
            <button
              type="button"
              onClick={() => setShowProfileMenu((prev) => !prev)}
              className="flex items-center rounded-full focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
              aria-label="Menu profil pengguna"
            >
              <Avatar
                initials={currentUser.avatarText}
                size="md"
                className="ml-1 cursor-pointer"
              />
            </button>

            {showProfileMenu && (
              <div className="absolute right-0 top-full mt-2 w-56 bg-white rounded-2xl border border-border shadow-soft-xl z-50 overflow-hidden animate-fadeIn py-1.5 text-xs">
                <div className="px-4 py-2.5 border-b border-border-subtle space-y-0.5">
                  <p className="font-bold text-text-primary truncate">
                    {currentUser.name}
                  </p>
                  <p className="text-[11px] text-text-secondary truncate">
                    {currentUser.email}
                  </p>
                  <span className="inline-block mt-1 px-2 py-0.5 rounded-full text-[10px] font-bold uppercase bg-primary-100 text-primary-900">
                    {role === 'admin' ? 'Administrator' : 'Mahasiswa'}
                  </span>
                </div>

                <div className="py-1">
                  <button
                    type="button"
                    onClick={() => {
                      setShowProfileMenu(false);
                      navigate('/profile');
                    }}
                    className="w-full px-4 py-2 text-left text-text-primary hover:bg-surface-muted flex items-center gap-2.5 transition-colors cursor-pointer"
                  >
                    <UserIcon className="w-4 h-4 text-text-secondary" />
                    <span>Profil Saya</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setShowProfileMenu(false);
                      navigate('/settings');
                    }}
                    className="w-full px-4 py-2 text-left text-text-primary hover:bg-surface-muted flex items-center gap-2.5 transition-colors cursor-pointer"
                  >
                    <SettingsIcon className="w-4 h-4 text-text-secondary" />
                    <span>Pengaturan Akun</span>
                  </button>
                </div>

                <div className="pt-1 border-t border-border-subtle">
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="w-full px-4 py-2 text-left text-red-600 hover:bg-red-50 flex items-center gap-2.5 transition-colors cursor-pointer font-semibold"
                  >
                    <LogoutIcon className="w-4 h-4 text-red-500" />
                    <span>Keluar</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};
