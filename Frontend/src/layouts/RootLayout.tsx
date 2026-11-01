import { Outlet } from 'react-router-dom';

export default function RootLayout() {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 flex flex-col justify-center items-center p-6">
      <main className="w-full max-w-lg">
        <Outlet />
      </main>
    </div>
  );
}
