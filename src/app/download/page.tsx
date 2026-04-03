import type { Metadata } from 'next';
import Link from 'next/link';
import { Suspense } from 'react';

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  title: 'Free Download — 5 Ways AI Can Cut Repetitive Work',
  description: 'Get the free checklist: 5 practical ways small businesses can use AI to cut repetitive work and save hours every week.',
  alternates: { canonical: `${BASE_URL}/download` },
};

function DownloadForm() {
  return (
    <form className="mt-8 space-y-4" action="/api/newsletter/subscribe" method="POST">
      <input
        type="email"
        name="email"
        required
        placeholder="your@email.com"
        className="w-full rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-white placeholder:text-gray-500 focus:border-[#2dd47a] focus:outline-none"
        style={{ fontFamily: 'var(--font-inter, sans-serif)' }}
      />
      <button
        type="submit"
        className="w-full rounded-xl py-3 text-sm font-semibold transition hover:opacity-90"
        style={{ background: '#2dd47a', color: '#0a1a12', fontFamily: 'var(--font-jetbrains, monospace)' }}
      >
        Get Free Checklist
      </button>
    </form>
  );
}

export default function DownloadPage() {
  return (
    <main className="relative min-h-screen" style={{ background: '#0a1a12' }}>
      <div aria-hidden className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute -left-52 -top-40 h-[460px] w-[460px] rounded-full blur-[130px]" style={{ background: 'rgba(255,59,111,0.04)' }} />
        <div className="absolute -bottom-48 -right-24 h-[400px] w-[400px] rounded-full blur-[100px]" style={{ background: 'rgba(45,212,122,0.05)' }} />
      </div>

      <nav className="flex items-center justify-between px-6 py-6 md:px-10">
        <Link href="/" className="text-sm font-semibold tracking-widest uppercase transition hover:text-[#2dd47a]" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          ArchonHQ
        </Link>
        <div className="flex items-center gap-6">
          <Link href="/insights" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Insights</Link>
          <Link href="/products" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Products</Link>
        </div>
      </nav>

      <div className="relative z-10 mx-auto max-w-xl px-6 py-20 text-center md:px-10">
        <p className="text-xs uppercase tracking-[0.4em]" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          Free Download
        </p>
        <h1 className="mt-4 text-3xl font-extrabold leading-tight text-white md:text-5xl" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          5 Ways AI Can Cut Repetitive Work in Small Businesses
        </h1>
        <p className="mt-6 text-base leading-relaxed" style={{ color: '#c4d4c8' }}>
          A practical checklist of ways to use AI automation to save hours on client reports, content creation, and daily admin tasks.
        </p>

        <Suspense fallback={<div className="mt-8 text-sm" style={{ color: '#6a7f6f' }}>Loading...</div>}>
          <DownloadForm />
        </Suspense>

        <p className="mt-4 text-xs" style={{ color: '#6a7f6f' }}>
          No spam. Just the checklist sent to your inbox.
        </p>

        <div className="mt-16">
          <Link href="/insights" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
            ← Back to insights
          </Link>
        </div>
      </div>
    </main>
  );
}
