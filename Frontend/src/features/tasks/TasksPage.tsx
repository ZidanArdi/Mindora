import React, { useState, useEffect, useCallback } from 'react';
import { DashboardTemplate } from '../../components/templates';
import { Button, Input, PlusIcon, CloseIcon } from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import type { TaskItem, TaskPriority, TaskStatus } from '../../types';
import { useAuth } from '../../hooks';
import { translatePriority, translateCategory, formatDate } from '../../utils/i18n';

export default function TasksPage() {
  const { user } = useAuth();
  const [tasks, setTasks] = useState<TaskItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters & Sorting
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');
  const [categoryFilter, setCategoryFilter] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [viewMode, setViewMode] = useState<'list' | 'card' | 'kanban'>('list');

  // Modals
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<TaskItem | null>(null);
  const [deletingTask, setDeletingTask] = useState<TaskItem | null>(null);

  // Form State
  const [formTitle, setFormTitle] = useState('');
  const [formDesc, setFormDesc] = useState('');
  const [formCategory, setFormCategory] = useState('Study');
  const [formPriority, setFormPriority] = useState<TaskPriority>('medium');
  const [formDeadline, setFormDeadline] = useState('');
  const [formSubmitting, setFormSubmitting] = useState(false);

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await productivityService.listTasks({
        status: statusFilter || undefined,
        priority: priorityFilter || undefined,
        category: categoryFilter || undefined,
        search: searchQuery || undefined,
        sort_by: sortBy,
        sort_order: sortBy === 'deadline' ? 'asc' : 'desc',
        limit: 100,
      });
      setTasks(data.tasks);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load tasks';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [statusFilter, priorityFilter, categoryFilter, searchQuery, sortBy]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleOpenCreate = () => {
    setFormTitle('');
    setFormDesc('');
    setFormCategory('Study');
    setFormPriority('medium');
    setFormDeadline('');
    setIsCreateOpen(true);
  };

  const handleOpenEdit = (task: TaskItem) => {
    setEditingTask(task);
    setFormTitle(task.title);
    setFormDesc(task.description || '');
    setFormCategory(task.category);
    setFormPriority(task.priority || 'medium');
    setFormDeadline(task.deadline ? task.deadline.slice(0, 16) : '');
  };

  const handleSaveTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formTitle.trim()) return;

    try {
      setFormSubmitting(true);
      const payload = {
        title: formTitle.trim(),
        description: formDesc.trim() || undefined,
        category: formCategory,
        priority: formPriority,
        deadline: formDeadline ? new Date(formDeadline).toISOString() : undefined,
      };

      if (editingTask) {
        await productivityService.updateTask(editingTask.id, payload);
        setEditingTask(null);
      } else {
        await productivityService.createTask(payload);
        setIsCreateOpen(false);
      }
      fetchTasks();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Operation failed');
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleToggleStatus = async (task: TaskItem) => {
    const nextStatus: TaskStatus = task.status === 'completed' ? 'pending' : 'completed';
    try {
      // Optimistic update
      setTasks((prev) =>
        prev.map((t) => (t.id === task.id ? { ...t, status: nextStatus } : t))
      );
      await productivityService.updateTaskStatus(task.id, nextStatus);
    } catch {
      fetchTasks(); // Revert on failure
    }
  };

  const handleDeleteTask = async () => {
    if (!deletingTask) return;
    try {
      await productivityService.deleteTask(deletingTask.id);
      setDeletingTask(null);
      fetchTasks();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  const currentUserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name.slice(0, 2).toUpperCase(),
      }
    : { id: '', name: 'Student', email: '', role: 'user' as const, avatarText: 'ST' };

  return (
    <DashboardTemplate role="user" currentUser={currentUserProfile}>
      <div className="space-y-6 max-w-7xl mx-auto">
        {/* Page Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border">
          <div>
            <h1 className="text-2xl font-bold text-text-primary tracking-tight">
              Manajemen Tugas
            </h1>
            <p className="text-xs text-text-secondary mt-1">
              Atur, prioritaskan, dan pantau tugas akademik serta target fokus harian Anda.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              variant="primary"
              size="md"
              onClick={handleOpenCreate}
              className="gap-2 shadow-soft-sm font-semibold"
            >
              <PlusIcon className="w-4 h-4" />
              <span>Tambah Tugas</span>
            </Button>
          </div>
        </div>

        {/* Filter and View Controls Bar */}
        <div className="bg-white p-4 rounded-2xl border border-border shadow-soft-xs space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 items-center">
            {/* Search Input */}
            <div className="lg:col-span-2">
              <Input
                placeholder="Cari tugas berdasarkan judul atau detail..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full text-xs"
              />
            </div>

            {/* Category Filter */}
            <select
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">Semua Kategori</option>
              <option value="Study">Belajar</option>
              <option value="Academic">Akademik</option>
              <option value="Lab">Praktikum / Lab</option>
              <option value="Project">Proyek</option>
              <option value="Personal">Pribadi</option>
            </select>

            {/* Priority Filter */}
            <select
              value={priorityFilter}
              onChange={(e) => setPriorityFilter(e.target.value)}
              className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">Semua Prioritas</option>
              <option value="high">Prioritas Tinggi</option>
              <option value="medium">Prioritas Sedang</option>
              <option value="low">Prioritas Rendah</option>
            </select>

            {/* Sort Dropdown */}
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="created_at">Terbaru Lebih Dulu</option>
              <option value="deadline">Tenggat Terdekat</option>
              <option value="priority">Peringkat Prioritas</option>
            </select>
          </div>

          {/* Status Pills & View Mode Toggle */}
          <div className="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-border/60">
            <div className="flex items-center gap-1.5 overflow-x-auto no-scrollbar py-1">
              {[
                { label: 'Semua', value: '' },
                { label: 'Menunggu', value: 'pending' },
                { label: 'Sedang Dikerjakan', value: 'in_progress' },
                { label: 'Selesai', value: 'completed' },
              ].map((st) => (
                <button
                  key={st.value}
                  type="button"
                  onClick={() => setStatusFilter(st.value)}
                  className={`px-3 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                    statusFilter === st.value
                      ? 'bg-primary text-white shadow-soft-xs'
                      : 'bg-surface-muted text-text-secondary hover:text-text-primary hover:bg-slate-200'
                  }`}
                >
                  {st.label}
                </button>
              ))}
            </div>

            {/* View Mode Toggle */}
            <div className="flex items-center gap-1 bg-surface-muted p-1 rounded-xl border border-border/60">
              {[
                { mode: 'list' as const, label: 'Daftar' },
                { mode: 'card' as const, label: 'Kartu' },
                { mode: 'kanban' as const, label: 'Kanban' },
              ].map(({ mode, label }) => (
                <button
                  key={mode}
                  type="button"
                  onClick={() => setViewMode(mode)}
                  className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                    viewMode === mode
                      ? 'bg-white text-primary-900 shadow-soft-xs'
                      : 'text-text-secondary hover:text-text-primary'
                  }`}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Content Area */}
        {loading ? (
          <div className="p-12 text-center text-text-secondary animate-pulse">
            Memuat tugas...
          </div>
        ) : error ? (
          <div className="p-6 rounded-2xl bg-red-50 text-red-700 text-xs text-center border border-red-200">
            {error}
          </div>
        ) : tasks.length === 0 ? (
          <div className="p-12 text-center bg-white rounded-2xl border border-border space-y-3">
            <div className="w-12 h-12 rounded-full bg-primary-100 text-primary flex items-center justify-center mx-auto text-xl">
              ðŸ“
            </div>
            <h3 className="text-sm font-bold text-text-primary">Tidak ada tugas ditemukan</h3>
            <p className="text-xs text-text-secondary max-w-sm mx-auto">
              Tidak ada tugas yang cocok dengan filter saat ini. Buat tugas baru agar tetap teratur dan fokus.
            </p>
            <Button variant="primary" size="sm" onClick={handleOpenCreate} className="mt-2 text-xs">
              Tambah Tugas Pertama Anda
            </Button>
          </div>
        ) : viewMode === 'kanban' ? (
          /* Kanban Board View */
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 items-start">
            {(['pending', 'in_progress', 'completed'] as const).map((colStatus) => {
              const colTasks = tasks.filter((t) => t.status === colStatus);
              const colTitles = {
                pending: 'Akan Dikerjakan',
                in_progress: 'Sedang Dikerjakan',
                completed: 'Selesai',
              };
              return (
                <div
                  key={colStatus}
                  className="bg-surface-muted/60 p-4 rounded-2xl border border-border/80 space-y-3"
                >
                  <div className="flex items-center justify-between px-1">
                    <span className="text-xs font-bold text-text-primary uppercase tracking-wider">
                      {colTitles[colStatus]}
                    </span>
                    <span className="text-xs font-bold px-2 py-0.5 rounded-full bg-white border border-border text-text-secondary">
                      {colTasks.length}
                    </span>
                  </div>

                  <div className="space-y-3">
                    {colTasks.map((task) => (
                      <TaskCardItem
                        key={task.id}
                        task={task}
                        onToggle={() => handleToggleStatus(task)}
                        onEdit={() => handleOpenEdit(task)}
                        onDelete={() => setDeletingTask(task)}
                      />
                    ))}
                    {colTasks.length === 0 && (
                      <div className="p-6 text-center text-xs text-text-muted border-2 border-dashed border-border/60 rounded-xl">
                        Belum ada tugas
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        ) : viewMode === 'card' ? (
          /* Card Grid View */
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {tasks.map((task) => (
              <TaskCardItem
                key={task.id}
                task={task}
                onToggle={() => handleToggleStatus(task)}
                onEdit={() => handleOpenEdit(task)}
                onDelete={() => setDeletingTask(task)}
              />
            ))}
          </div>
        ) : (
          /* List Table View */
          <div className="bg-white rounded-2xl border border-border shadow-soft-xs overflow-hidden">
            <div className="divide-y divide-border/60">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className="p-4 flex items-center justify-between gap-4 hover:bg-surface-muted/40 transition-colors group"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <input
                      type="checkbox"
                      checked={task.status === 'completed'}
                      onChange={() => handleToggleStatus(task)}
                      className="w-4 h-4 rounded border-border text-primary focus:ring-primary/20 cursor-pointer"
                    />
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span
                          className={`text-sm font-semibold truncate ${
                            task.status === 'completed'
                              ? 'line-through text-text-muted'
                              : 'text-text-primary'
                          }`}
                        >
                          {task.title}
                        </span>
                        <PriorityPill priority={task.priority || 'medium'} />
                        <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-100 text-slate-700">
                          {translateCategory(task.category)}
                        </span>
                      </div>
                      {task.description && (
                        <p className="text-xs text-text-secondary truncate mt-0.5">
                          {task.description}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-4 shrink-0">
                    {task.deadline && (
                      <span className="text-[11px] font-mono text-text-secondary">
                        ðŸ“… {formatDate(task.deadline)}
                      </span>
                    )}
                    <div className="flex items-center gap-2 opacity-90 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onClick={() => handleOpenEdit(task)}
                        className="px-2.5 py-1 text-xs font-semibold text-text-secondary hover:text-primary rounded-lg hover:bg-surface-muted"
                      >
                        Ubah
                      </button>
                      <button
                        type="button"
                        onClick={() => setDeletingTask(task)}
                        className="px-2.5 py-1 text-xs font-semibold text-red-600 hover:text-red-700 rounded-lg hover:bg-red-50"
                      >
                        Hapus
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Create / Edit Task Modal */}
        {(isCreateOpen || editingTask) && (
          <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
            <div className="bg-white rounded-3xl max-w-lg w-full p-6 shadow-soft-xl border border-border space-y-4 animate-scaleUp">
              <div className="flex items-center justify-between pb-2 border-b border-border">
                <h2 className="text-lg font-bold text-text-primary">
                  {editingTask ? 'Ubah Tugas' : 'Buat Tugas Baru'}
                </h2>
                <button
                  type="button"
                  onClick={() => {
                    setIsCreateOpen(false);
                    setEditingTask(null);
                  }}
                  className="p-1 rounded-lg text-text-secondary hover:text-text-primary"
                >
                  <CloseIcon className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleSaveTask} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Judul *
                  </label>
                  <Input
                    required
                    placeholder="cth. Ringkasan Bab 4 Algoritma & Struktur Data"
                    value={formTitle}
                    onChange={(e) => setFormTitle(e.target.value)}
                    className="w-full text-xs"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Deskripsi (opsional)
                  </label>
                  <textarea
                    rows={3}
                    placeholder="Detail tambahan, tautan referensi, atau catatan..."
                    value={formDesc}
                    onChange={(e) => setFormDesc(e.target.value)}
                    className="w-full p-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-semibold text-text-primary mb-1">
                      Kategori
                    </label>
                    <select
                      value={formCategory}
                      onChange={(e) => setFormCategory(e.target.value)}
                      className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                    >
                      <option value="Study">Belajar</option>
                      <option value="Academic">Akademik</option>
                      <option value="Lab">Praktikum / Lab</option>
                      <option value="Project">Proyek</option>
                      <option value="Personal">Pribadi</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-text-primary mb-1">
                      Prioritas
                    </label>
                    <select
                      value={formPriority}
                      onChange={(e) => setFormPriority(e.target.value as TaskPriority)}
                      className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                    >
                      <option value="low">Prioritas Rendah</option>
                      <option value="medium">Prioritas Sedang</option>
                      <option value="high">Prioritas Tinggi</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Tenggat Waktu (opsional)
                  </label>
                  <input
                    type="datetime-local"
                    value={formDeadline}
                    onChange={(e) => setFormDeadline(e.target.value)}
                    className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div className="flex items-center justify-end gap-3 pt-3 border-t border-border">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setIsCreateOpen(false);
                      setEditingTask(null);
                    }}
                  >
                    Batal
                  </Button>
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    disabled={formSubmitting || !formTitle.trim()}
                  >
                    {formSubmitting ? 'Menyimpan...' : editingTask ? 'Perbarui Tugas' : 'Buat Tugas'}
                  </Button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Delete Confirmation Modal */}
        {deletingTask && (
          <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
            <div className="bg-white rounded-3xl max-w-sm w-full p-6 shadow-soft-xl border border-border space-y-4 text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mx-auto text-xl font-bold">
                ðŸ—‘ï¸
              </div>
              <h3 className="text-base font-bold text-text-primary">Hapus Tugas?</h3>
              <p className="text-xs text-text-secondary leading-relaxed">
                Apakah Anda yakin ingin menghapus <strong className="text-text-primary">"{deletingTask.title}"</strong>? Tindakan ini akan menghapus tugas tetapi tetap menyimpan riwayat belajar terkait.
              </p>
              <div className="flex items-center justify-center gap-3 pt-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setDeletingTask(null)}
                >
                  Batal
                </Button>
                <button
                  type="button"
                  onClick={handleDeleteTask}
                  className="px-4 py-2 rounded-xl text-xs font-semibold bg-red-600 text-white hover:bg-red-700 transition-colors"
                >
                  Hapus
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}

// Subcomponent: Task Card Item for Card and Kanban views
function TaskCardItem({
  task,
  onToggle,
  onEdit,
  onDelete,
}: {
  task: TaskItem;
  onToggle: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="p-4 rounded-2xl bg-white border border-border shadow-soft-xs hover:shadow-soft-sm transition-all space-y-3 group">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-start gap-2.5 min-w-0">
          <input
            type="checkbox"
            checked={task.status === 'completed'}
            onChange={onToggle}
            className="w-4 h-4 mt-0.5 rounded border-border text-primary focus:ring-primary/20 cursor-pointer"
          />
          <h4
            className={`text-xs font-bold leading-tight ${
              task.status === 'completed'
                ? 'line-through text-text-muted'
                : 'text-text-primary'
            }`}
          >
            {task.title}
          </h4>
        </div>
        <PriorityPill priority={task.priority || 'medium'} />
      </div>

      {task.description && (
        <p className="text-[11px] text-text-secondary line-clamp-2 pl-6">
          {task.description}
        </p>
      )}

      <div className="flex items-center justify-between pt-2 border-t border-border/50 text-[10px] text-text-secondary pl-6">
        <span className="px-2 py-0.5 rounded-full bg-slate-100 font-medium text-slate-700">
          {translateCategory(task.category)}
        </span>
        {task.deadline && (
          <span className="font-mono">
            {formatDate(task.deadline)}
          </span>
        )}
      </div>

      <div className="flex items-center justify-end gap-2 pt-1 border-t border-border/30 opacity-80 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
        <button
          type="button"
          onClick={onEdit}
          className="text-[10px] font-semibold text-text-secondary hover:text-primary px-2 py-1 rounded hover:bg-surface-muted"
        >
          Ubah
        </button>
        <button
          type="button"
          onClick={onDelete}
          className="text-[10px] font-semibold text-red-600 hover:text-red-700 px-2 py-1 rounded hover:bg-red-50"
        >
          Hapus
        </button>
      </div>
    </div>
  );
}

function PriorityPill({ priority }: { priority: TaskPriority }) {
  const styles: Record<TaskPriority, string> = {
    urgent: 'bg-red-100 text-red-900 border-red-300',
    high: 'bg-rose-100 text-rose-800 border-rose-200',
    medium: 'bg-amber-100 text-amber-800 border-amber-200',
    low: 'bg-emerald-100 text-emerald-800 border-emerald-200',
  };
  return (
    <span
      className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${
        styles[priority] || styles.medium
      }`}
    >
      {translatePriority(priority)}
    </span>
  );
}

