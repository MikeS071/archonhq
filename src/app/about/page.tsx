import type { Metadata } from 'next';
import Link from 'next/link';

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  title: 'About — ArchonHQ',
  description:
    'I build and share straightforward AI automations based on real experiments. Everything here is designed for small teams and solo business owners who want practical results without complexity.',
  alternates: { canonical: `${BASE_URL}/about` },
};

export default function AboutPage() {
  return (
    <main className="relative min-h-screen" style={{ background: '#0a1a12' }}>
      <div aria-hidden className="pointer-events-none fixed inset-0 overflow-hidden">
        <div
          className="absolute -left-52 -top-40 h-[460px] w-[460px] rounded-full blur-[130px]"
          style={{ background: 'rgba(255,59,111,0.04)' }}
        />
        <div
          className="absolute -bottom-48 -right-24 h-[400px] w-[400px] rounded-full blur-[100px]"
          style={{ background: 'rgba(45,212,122,0.05)' }}
        />
      </div>

      <nav className="flex items-center justify-between px-6 py-6 md:px-10">
        <Link
          href="/"
          className="text-sm font-semibold tracking-widest uppercase transition hover:text-[#2dd47a]"
          style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
        >
          ArchonHQ
        </Link>
        <div className="flex items-center gap-6">
          <Link href="/insights" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Insights</Link>
          <Link href="/products" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Products</Link>
          <Link href="/services" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Services</Link>
        </div>
      </nav>

      <div className="relative z-10 mx-auto max-w-3xl px-6 py-16 md:px-10">
        <Link href="/" className="text-sm transition hover:text-[#ff6b8a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          ← Back to home
        </Link>

        <p className="mt-6 text-xs uppercase tracking-[0.4em]" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          About
        </p>
        <h1 className="mt-4 text-4xl font-extrabold leading-tight text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          Mike Szalinski
        </h1>

        <div className="mt-8 space-y-6 text-base leading-7" style={{ color: '#c4d4c8', fontFamily: 'var(--font-inter, sans-serif)' }}>
          <p>
            I build and share straightforward AI automations based on real experiments. One recent example: a coaching report system that reduced manual effort by 75%. Everything here is designed for small teams and solo business owners who want practical results without complexity.
          </p>
          <p>
            This started as personal experiments to solve my own workflow problems. After seeing consistent results — saving hours on repetitive tasks, cutting manual work significantly — I decided to share what worked.
          </p>
          <p>
            Everything is tested in real scenarios before I write about it or package it as a template. If it doesn't deliver measurable time savings, it doesn't make the cut.
          </p>
        </div>

        <div className="mt-10">
          <h2 className="text-xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            Connect
          </h2>
          <div className="mt-4 flex gap-4">
            <a
              href="https://github.com/MikeS071"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              GitHub
            </a>
            <a
              href="https://linkedin.com/in/mszalins"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              LinkedIn
            </a>
          </div>
        </div>

        <div className="mt-12 rounded-2xl border border-white/5 bg-white/[0.03] p-6">
          <h2 className="text-xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            This is a personal experiment
          </h2>
          <p className="mt-3 text-sm leading-relaxed" style={{ color: '#c4d4c8' }}>
            Everything on this site represents personal experiments and learnings, kept completely separate from my day job at AustralianSuper. It's a side project driven by curiosity and a desire to share what actually works.
          </p>
        </div>
      </div>
    </main>
  );
}
