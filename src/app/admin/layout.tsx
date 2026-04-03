import type { Metadata } from 'next';
import { redirect } from 'next/navigation';
import { auth } from '@/lib/auth';

export const metadata: Metadata = {
  title: 'Admin — ArchonHQ',
};

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const session = await auth();

  if (!session?.user?.email) {
    redirect('/signin?callbackUrl=/admin/insights');
  }

  return (
    <div className="min-h-screen" style={{ background: '#0a1a12' }}>
      <header className="border-b border-white/5 px-6 py-4">
        <div className="mx-auto max-w-6xl">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
              Admin
            </span>
            <div className="flex items-center gap-6">
              <a href="/admin/insights" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                Insights
              </a>
              <form action="/api/auth/signout" method="POST">
                <button type="submit" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                  Sign out
                </button>
              </form>
            </div>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">
        {children}
      </main>
    </div>
  );
}
