import type { Metadata } from 'next';
import Link from 'next/link';
import { db } from '@/lib/db';
import { insights } from '@/db/schema';
import { desc } from 'drizzle-orm';

export const metadata: Metadata = {
  title: 'Manage Insights — Admin',
};

export default async function AdminInsightsPage() {
  const allInsights = await db.select().from(insights).orderBy(desc(insights.publishedAt));

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          Insights
        </h1>
        <Link
          href="/admin/insights/new"
          className="rounded-full px-4 py-2 text-sm font-semibold transition hover:opacity-90"
          style={{ background: '#2dd47a', color: '#0a1a12', fontFamily: 'var(--font-jetbrains, monospace)' }}
        >
          + New Insight
        </Link>
      </div>

      {allInsights.length === 0 ? (
        <p className="text-sm" style={{ color: '#6a7f6f' }}>No insights yet.</p>
      ) : (
        <div className="space-y-4">
          {allInsights.map((insight) => (
            <div key={insight.id} className="flex items-center justify-between rounded-xl border border-white/5 bg-white/[0.03] p-4">
              <div>
                <h2 className="font-semibold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
                  {insight.title}
                </h2>
                <p className="mt-1 text-sm" style={{ color: '#6a7f6f' }}>
                  {insight.slug} — {insight.publishedAt?.toLocaleDateString()}
                </p>
              </div>
              <div className="flex gap-2">
                <Link
                  href={`/admin/insights/${insight.id}/edit`}
                  className="rounded-lg border border-white/10 px-3 py-1.5 text-xs transition hover:border-[#2dd47a] hover:text-[#2dd47a]"
                  style={{ color: '#f1f5f0', fontFamily: 'var(--font-jetbrains, monospace)' }}
                >
                  Edit
                </Link>
                <Link
                  href={`/insights/${insight.slug}`}
                  target="_blank"
                  className="rounded-lg border border-white/10 px-3 py-1.5 text-xs transition hover:border-[#2dd47a] hover:text-[#2dd47a]"
                  style={{ color: '#f1f5f0', fontFamily: 'var(--font-jetbrains, monospace)' }}
                >
                  View
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
