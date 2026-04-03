'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function NewInsightPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError('');

    const formData = new FormData(e.currentTarget);
    const data = {
      slug: formData.get('slug') as string,
      title: formData.get('title') as string,
      description: formData.get('description') as string,
      contentMd: formData.get('contentMd') as string,
      sourceUrl: formData.get('sourceUrl') as string || null,
      imageUrl: formData.get('imageUrl') as string || null,
      publishedAt: new Date().toISOString(),
    };

    try {
      const res = await fetch('/api/admin/insights', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!res.ok) {
        const { error: err } = await res.json();
        throw new Error(err || 'Failed to create');
      }

      router.push('/admin/insights');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <div className="mb-8">
        <Link href="/admin/insights" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f' }}>
          ← Back to insights
        </Link>
        <h1 className="mt-4 text-2xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          New Insight
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-6">
        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-400">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Slug</label>
          <input name="slug" required pattern="[a-z0-9-]+" className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white placeholder:text-gray-500" style={{ fontFamily: 'var(--font-jetbrains, monospace)' }} />
          <p className="mt-1 text-xs" style={{ color: '#6a7f6f' }}>Lowercase letters, numbers, and hyphens only</p>
        </div>

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Title</label>
          <input name="title" required className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white" />
        </div>

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Description</label>
          <textarea name="description" required rows={2} className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white" />
        </div>

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Content (Markdown)</label>
          <textarea name="contentMd" required rows={15} className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white font-mono text-sm" />
        </div>

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Source URL (optional)</label>
          <input name="sourceUrl" type="url" className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white" />
        </div>

        <div>
          <label className="block text-sm font-medium" style={{ color: '#f1f5f0' }}>Image URL (optional)</label>
          <input name="imageUrl" type="url" className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white" />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="rounded-full px-6 py-2 text-sm font-semibold transition hover:opacity-90 disabled:opacity-50"
          style={{ background: '#2dd47a', color: '#0a1a12', fontFamily: 'var(--font-jetbrains, monospace)' }}
        >
          {loading ? 'Creating...' : 'Create Insight'}
        </button>
      </form>
    </div>
  );
}
