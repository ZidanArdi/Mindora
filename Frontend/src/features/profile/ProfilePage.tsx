import { useState, useEffect, useCallback } from 'react';
import { DashboardTemplate } from '../../components/templates';
import { Button } from '../../components/atoms';
import {
  UserIcon,
  ShieldIcon,
  CheckIcon,
} from '../../components/atoms/Icons';
import { authService } from '../../services/authService';
import type { UserProfile, UserResponse } from '../../types';
import { useAuth } from '../../hooks';
import { formatDate } from '../../utils/i18n';

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();
  const [profile, setProfile] = useState<UserResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Edit Name state
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [nameInput, setNameInput] = useState<string>('');
  const [saving, setSaving] = useState<boolean>(false);

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
        workspaceName: user.role === 'admin' ? 'Administrasi Sistem' : 'Ruang Kerja Akademik',
      }
    : {
        id: 'usr_guest',
        name: 'Mahasiswa',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'MH',
        workspaceName: 'Ruang Kerja Akademik',
      };

  const fetchProfile = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    try {
      const data = await authService.getProfile();
      setProfile(data);
      setNameInput(data.name);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memuat profil pengguna';
      setErrorMessage(msg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  const handleSaveName = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nameInput.trim()) {
      setErrorMessage('Nama tidak boleh kosong');
      return;
    }

    setSaving(true);
    setErrorMessage(null);
    setSuccessMessage(null);
    try {
      const updated = await authService.updateProfile(nameInput.trim());
      setProfile(updated);
      setIsEditing(false);
      setSuccessMessage('Nama profil berhasil diperbarui.');
      // Refresh global session to propagate updated name across UI
      await refreshUser();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui profil';
      setErrorMessage(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    if (profile) {
      setNameInput(profile.name);
    }
    setErrorMessage(null);
  };

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-4xl mx-auto">
        {/* Header Section */}
        <div className="border-b border-border pb-5">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
              Identitas Akun
            </span>
            <span className="text-xs text-text-muted">Profil Akademik Pribadi</span>
          </div>
          <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
            Profil Mahasiswa
          </h1>
          <p className="text-sm text-text-secondary">
            Kredensial akademik terverifikasi, rincian autentikasi, dan hak akses sistem.
          </p>
        </div>

        {/* Feedback Messages */}
        {successMessage && (
          <div className="p-4 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckIcon className="w-4 h-4 text-emerald-600" />
              <span>{successMessage}</span>
            </div>
            <button
              type="button"
              onClick={() => setSuccessMessage(null)}
              className="text-emerald-700 hover:text-emerald-900 font-bold ml-2"
            >
              Tutup
            </button>
          </div>
        )}

        {errorMessage && (
          <div className="p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs flex items-center justify-between">
            <span>{errorMessage}</span>
            <button
              type="button"
              onClick={() => setErrorMessage(null)}
              className="text-red-500 hover:text-red-700 font-bold ml-2"
            >
              Tutup
            </button>
          </div>
        )}

        {/* Loading Skeleton */}
        {loading && (
          <div className="p-12 text-center space-y-3">
            <div className="inline-block w-8 h-8 border-3 border-primary border-t-transparent rounded-full animate-spin" />
            <p className="text-xs font-semibold text-text-muted">Memuat profil akun...</p>
          </div>
        )}

        {/* Profile Content */}
        {!loading && profile && (
          <div className="space-y-6">
            {/* Primary Profile Card */}
            <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-6">
              <div className="flex flex-col sm:flex-row sm:items-center gap-5 border-b border-border pb-6">
                <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-[#123C32] to-[#138F78] text-white flex items-center justify-center text-2xl font-black shadow-soft-sm shrink-0">
                  {displayUser.avatarText}
                </div>
                <div className="space-y-1 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-xl font-bold text-primary-900">{profile.name}</h2>
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${
                        profile.role === 'admin'
                          ? 'bg-purple-100 text-purple-800'
                          : 'bg-nature-mint text-primary-900'
                      }`}
                    >
                      {profile.role}
                    </span>
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-emerald-100 text-emerald-800">
                      {profile.status}
                    </span>
                  </div>
                  <p className="text-xs text-text-secondary">{profile.email}</p>
                  <p className="text-[11px] text-text-muted font-mono">User ID: {profile.id}</p>
                </div>
              </div>

              {/* Editable Name Field */}
              <div className="space-y-4">
                <h3 className="text-sm font-bold text-primary-900 uppercase tracking-wider text-[11px]">
                  Informasi Akun
                </h3>

                {isEditing ? (
                  <form onSubmit={handleSaveName} className="space-y-3 max-w-md">
                    <div>
                      <label htmlFor="profile-name-input" className="block text-xs font-semibold text-text-secondary mb-1">
                        Nama Lengkap
                      </label>
                      <input
                        id="profile-name-input"
                        type="text"
                        value={nameInput}
                        onChange={(e) => setNameInput(e.target.value)}
                        className="w-full px-3 py-2 text-sm rounded-xl border border-border bg-white text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
                        placeholder="Masukkan nama Anda"
                        disabled={saving}
                      />
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        type="submit"
                        variant="primary"
                        size="sm"
                        disabled={saving}
                        className="text-xs font-bold"
                      >
                        {saving ? 'Menyimpan...' : 'Simpan Perubahan'}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={handleCancelEdit}
                        disabled={saving}
                        className="text-xs"
                      >
                        Batal
                      </Button>
                    </div>
                  </form>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center justify-between">
                      <div>
                        <span className="text-[11px] font-bold text-text-muted uppercase tracking-wider block">
                          Nama Tampilan
                        </span>
                        <span className="text-sm font-bold text-text-primary mt-0.5 block">
                          {profile.name}
                        </span>
                      </div>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => setIsEditing(true)}
                        className="text-xs font-semibold px-2.5 py-1"
                      >
                        Ubah Nama
                      </Button>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
                      <span className="text-[11px] font-bold text-text-muted uppercase tracking-wider block">
                        Email Akun
                      </span>
                      <span className="text-sm font-bold text-text-primary mt-0.5 block">
                        {profile.email}
                      </span>
                      <span className="text-[10px] text-text-muted mt-0.5 block">
                        Pengenal autentikasi utama
                      </span>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
                      <span className="text-[11px] font-bold text-text-muted uppercase tracking-wider block">
                        Peran Ditetapkan
                      </span>
                      <span className="text-sm font-bold text-text-primary mt-0.5 block capitalize">
                        {profile.role}
                      </span>
                      <span className="text-[10px] text-text-muted mt-0.5 block">
                        {profile.role === 'admin'
                          ? 'Hak istimewa pemantauan administratif dan pengelolaan pengguna penuh'
                          : 'Akses ruang kerja produktivitas pribadi mahasiswa'}
                      </span>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
                      <span className="text-[11px] font-bold text-text-muted uppercase tracking-wider block">
                        Anggota Sejak
                      </span>
                      <span className="text-sm font-bold text-text-primary mt-0.5 block">
                        {profile.created_at ? formatDate(profile.created_at) : 'Aktif'}
                      </span>
                      <span className="text-[10px] text-text-muted mt-0.5 block">
                        Pendaftaran akun mahasiswa terverifikasi
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Permissions & Security Capabilities Card */}
            <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
              <div className="flex items-center gap-2 border-b border-border pb-3">
                <ShieldIcon className="w-5 h-5 text-primary" />
                <div>
                  <h3 className="text-base font-bold text-primary-900">
                    Hak Akses Sistem yang Diberikan (RBAC)
                  </h3>
                  <p className="text-xs text-text-muted">
                    Kemampuan yang diotorisasi untuk peran Anda sesuai dengan kebijakan keamanan Mindora.
                  </p>
                </div>
              </div>

              {profile.permissions && profile.permissions.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5 pt-1">
                  {profile.permissions.map((perm) => (
                    <div
                      key={perm}
                      className="p-3 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center gap-2 text-xs"
                    >
                      <CheckIcon className="w-4 h-4 text-emerald-600 shrink-0" />
                      <span className="font-mono font-semibold text-text-primary text-[11px]">{perm}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-text-muted">Tidak ada hak akses peran eksplisit yang ditetapkan.</p>
              )}
            </div>

            {/* Data Protection Notice */}
            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 text-xs text-text-secondary flex items-start gap-3">
              <UserIcon className="w-5 h-5 text-primary mt-0.5 shrink-0" />
              <div>
                <strong className="text-text-primary">Jaminan Privasi Akademik:</strong> Catatan belajar, interval
                fokus, judul tugas pribadi, dan catatan gaya hidup Anda terisolasi secara ketat. Administrator Mindora
                tidak dapat memeriksa catatan pribadi mahasiswa.
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}

