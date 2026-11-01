export default function HomePage() {
  return (
    <div className="bg-white rounded-2xl shadow-sm border border-slate-200 p-8 text-center space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-slate-900">
          Mindora
        </h1>
        <p className="text-sm font-medium text-slate-500">
          Student Productivity Application
        </p>
      </div>

      <div className="inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full bg-emerald-50 border border-emerald-200 text-emerald-700 text-xs font-semibold">
        <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
        Frontend is running
      </div>
    </div>
  );
}
