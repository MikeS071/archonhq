import Link from 'next/link';
import type { Metadata } from 'next';
import { db } from '@/lib/db';
import { insights } from '@/db/schema';
import { desc } from 'drizzle-orm';

export const dynamic = 'force-dynamic';

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  title: 'ArchonHQ — AI Automation for Small Businesses',
  description:
    'Practical AI automation that saves time and effort for busy small businesses and solopreneurs. Cut manual work by 50-80% on repetitive tasks.',
  alternates: { canonical: BASE_URL },
  openGraph: {
    type: 'website',
    url: BASE_URL,
    title: 'ArchonHQ — AI Automation for Small Businesses',
    description:
      'Practical AI automation that saves time and effort for busy small businesses and solopreneurs.',
  },
};

const products = [
  {
    name: 'Coaching Report Automation Template',
    price: '$39–$59',
    href: '/products/coaching-report-automation',
    description: 'A ready-to-use system that automatically pulls client information, creates personalized reports, and sends them by email.',
    benefit: 'Cuts manual report work by up to 75%, saves many hours every week',
  },
  {
    name: 'AI Content & Operations Automation Pack',
    price: '$49–$69',
    href: '/products/ai-content-operations-pack',
    description: 'Simple templates to automatically generate and handle business content plus basic workflow helpers.',
    benefit: 'Reduces time spent on writing and repetitive admin tasks',
  },
];

const services = [
  {
    name: 'Automation Quick-Win Service',
    price: '$2,500–$5,000',
    description:
      'A 60-minute call to identify one repetitive task draining your time, followed by a custom automation built for your business.',
    benefit: 'Delivers measurable time savings (often 50-75% on the chosen task)',
  },
];

export default async function HomePage() {
  const recentInsights = await db
    .select({
      slug: insights.slug,
      title: insights.title,
      description: insights.description,
      publishedAt: insights.publishedAt,
    })
    .from(insights)
    .orderBy(desc(insights.publishedAt))
    .limit(3);

  return (
    <main className="relative min-h-screen" style={{ background: '#0a1a12' }}>
      {/* Background orbs */}
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

      <div className="relative z-10">
        {/* Navigation */}
        <nav className="flex items-center justify-between px-6 py-6 md:px-10">
          <span
            className="text-sm font-semibold tracking-widest uppercase"
            style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
          >
            ArchonHQ
          </span>
          <div className="flex items-center gap-6">
            <Link
              href="/insights"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              Insights
            </Link>
            <Link
              href="/products"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              Products
            </Link>
            <Link
              href="/services"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              Services
            </Link>
            <Link
              href="/about"
              className="text-sm transition hover:text-[#2dd47a]"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              About
            </Link>
          </div>
        </nav>

        {/* Hero */}
        <section className="mx-auto max-w-4xl px-6 py-24 text-center md:px-10">
          <p
            className="text-xs uppercase tracking-[0.4em]"
            style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
          >
            AI Automation
          </p>
          <h1
            className="mt-6 text-4xl font-extrabold leading-tight text-white md:text-6xl"
            style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
          >
            AI Automation That Saves Time and Effort for Busy Small Businesses and Solopreneurs
          </h1>
          <p
            className="mx-auto mt-8 max-w-2xl text-lg leading-relaxed"
            style={{ color: '#c4d4c8', fontFamily: 'var(--font-inter, sans-serif)' }}
          >
            Practical ways to cut manual work by 50-80% on repetitive tasks like client reports,
            content creation, and daily operations — so you can focus on what matters.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/products"
              className="rounded-full px-8 py-3 text-sm font-semibold transition hover:opacity-90"
              style={{ background: '#ff3b6f', color: '#fff', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              View Products
            </Link>
            <Link
              href="/services"
              className="rounded-full border px-8 py-3 text-sm font-semibold transition hover:border-[#2dd47a] hover:text-[#2dd47a]"
              style={{
                borderColor: 'rgba(45,212,122,0.3)',
                color: '#f1f5f0',
                fontFamily: 'var(--font-jetbrains, monospace)',
              }}
            >
              Try a Service
            </Link>
          </div>
        </section>

        {/* Products */}
        <section className="mx-auto max-w-4xl px-6 py-16 md:px-10">
          <h2
            className="text-2xl font-bold text-white"
            style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
          >
            Ready-to-Use Automation Templates
          </h2>
          <p className="mt-2 text-base" style={{ color: '#c4d4c8' }}>
            Simple setups that deliver immediate time savings.
          </p>
          <div className="mt-8 grid gap-6 md:grid-cols-2">
            {products.map((product) => (
              <Link
                key={product.href}
                href={product.href}
                className="block rounded-2xl border border-white/5 bg-white/[0.03] p-6 transition hover:border-[#2dd47a]/20 hover:bg-white/[0.05]"
              >
                <div className="flex items-start justify-between">
                  <span
                    className="text-xs uppercase tracking-wide"
                    style={{ color: '#ff3b6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
                  >
                    Template
                  </span>
                  <span
                    className="text-sm font-semibold"
                    style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
                  >
                    {product.price}
                  </span>
                </div>
                <h3
                  className="mt-4 text-xl font-bold text-white"
                  style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
                >
                  {product.name}
                </h3>
                <p className="mt-2 text-sm leading-relaxed" style={{ color: '#c4d4c8' }}>
                  {product.description}
                </p>
                <p
                  className="mt-4 text-xs"
                  style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
                >
                  {product.benefit}
                </p>
              </Link>
            ))}
          </div>
          <div className="mt-8 text-center">
            <Link
              href="/products"
              className="text-sm font-semibold transition hover:text-[#2dd47a]"
              style={{ color: '#f1f5f0', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              View all products →
            </Link>
          </div>
        </section>

        {/* Services */}
        <section className="mx-auto max-w-4xl px-6 py-16 md:px-10">
          <h2
            className="text-2xl font-bold text-white"
            style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
          >
            Automation Quick-Win Service
          </h2>
          <div className="mt-6 rounded-2xl border border-white/5 bg-white/[0.03] p-6">
            <div className="flex items-start justify-between">
              <span
                className="text-xs uppercase tracking-wide"
                style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                Consulting
              </span>
              <span
                className="text-sm font-semibold"
                style={{ color: '#ff3b6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                {services[0].price}
              </span>
            </div>
            <h3
              className="mt-4 text-xl font-bold text-white"
              style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
            >
              {services[0].name}
            </h3>
            <p className="mt-2 text-sm leading-relaxed" style={{ color: '#c4d4c8' }}>
              {services[0].description}
            </p>
            <p
              className="mt-4 text-xs"
              style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              {services[0].benefit}
            </p>
            <Link
              href="/services"
              className="mt-6 inline-block rounded-full px-6 py-2 text-sm font-semibold transition hover:opacity-90"
              style={{ background: '#2dd47a', color: '#0a1a12', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              Book a call
            </Link>
          </div>
        </section>

        {/* Recent Insights */}
        {recentInsights.length > 0 && (
          <section className="mx-auto max-w-4xl px-6 py-16 md:px-10">
            <h2
              className="text-2xl font-bold text-white"
              style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
            >
              Latest Insights
            </h2>
            <p className="mt-2 text-base" style={{ color: '#c4d4c8' }}>
              Real experiments, practical results.
            </p>
            <div className="mt-8 space-y-4">
              {recentInsights.map((insight) => (
                <Link
                  key={insight.slug}
                  href={`/insights/${insight.slug}`}
                  className="block rounded-xl border border-white/5 bg-white/[0.03] p-4 transition hover:border-[#2dd47a]/20 hover:bg-white/[0.05]"
                >
                  <h3
                    className="font-semibold text-white"
                    style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}
                  >
                    {insight.title}
                  </h3>
                  <p className="mt-1 text-sm" style={{ color: '#6a7f6f' }}>
                    {insight.description}
                  </p>
                </Link>
              ))}
            </div>
            <div className="mt-6 text-center">
              <Link
                href="/insights"
                className="text-sm font-semibold transition hover:text-[#2dd47a]"
                style={{ color: '#f1f5f0', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                View all insights →
              </Link>
            </div>
          </section>
        )}

        {/* Footer */}
        <footer className="mx-auto max-w-4xl border-t border-white/5 px-6 py-10 md:px-10">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <span
              className="text-sm"
              style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              © {new Date().getFullYear()} ArchonHQ
            </span>
            <div className="flex gap-6">
              <Link
                href="/insights"
                className="text-sm transition hover:text-[#2dd47a]"
                style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                Insights
              </Link>
              <Link
                href="/products"
                className="text-sm transition hover:text-[#2dd47a]"
                style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                Products
              </Link>
              <Link
                href="/services"
                className="text-sm transition hover:text-[#2dd47a]"
                style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                Services
              </Link>
              <Link
                href="/about"
                className="text-sm transition hover:text-[#2dd47a]"
                style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}
              >
                About
              </Link>
            </div>
          </div>
        </footer>
      </div>
    </main>
  );
}
