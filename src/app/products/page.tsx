import type { Metadata } from 'next';
import Link from 'next/link';

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  title: 'Products — ArchonHQ',
  description:
    'Ready-to-use AI automation templates that cut manual work by 50-75%. Simple setups for coaches, content creators, and small business owners.',
  alternates: { canonical: `${BASE_URL}/products` },
};

const products = [
  {
    slug: 'coaching-report-automation',
    name: 'Coaching Report Automation Template',
    price: '$39–$59',
    tag: 'Highest Potential',
    description: 'A ready-to-use system that automatically pulls client information, creates personalized reports, and sends them by email.',
    benefit: 'Cuts manual report work by up to 75%, saves many hours every week, improves client experience, and makes it easy to handle more clients without extra effort.',
    includes: [
      'Pre-built report templates',
      'Client data collection forms',
      'Automated email delivery',
      'Setup instructions & screenshots',
    ],
  },
  {
    slug: 'ai-content-operations-pack',
    name: 'AI Content & Operations Automation Pack',
    price: '$49–$69',
    tag: 'Popular',
    description: 'Simple templates to automatically generate and handle business content (emails, updates, summaries) plus basic workflow helpers.',
    benefit: 'Reduces time spent on writing and repetitive admin tasks, helps maintain consistent output, and frees up hours for growth activities.',
    includes: [
      'Email template automation',
      'Content repurposing workflows',
      'Daily operations helpers',
      'Setup guide & examples',
    ],
  },
];

export default function ProductsPage() {
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
          <Link href="/services" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>Services</Link>
          <Link href="/about" className="text-sm transition hover:text-[#2dd47a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>About</Link>
        </div>
      </nav>

      <div className="relative z-10 mx-auto max-w-4xl px-6 py-16 md:px-10">
        <Link href="/" className="text-sm transition hover:text-[#ff6b8a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          ← Back to home
        </Link>

        <p className="mt-6 text-xs uppercase tracking-[0.4em]" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          Products
        </p>
        <h1 className="mt-4 text-4xl font-extrabold leading-tight text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          Automation Templates
        </h1>
        <p className="mt-4 text-base leading-relaxed" style={{ color: '#c4d4c8' }}>
          Simple, practical automation templates built from real experiments. Each one is designed to save you hours on repetitive tasks.
        </p>

        <div className="mt-12 space-y-8">
          {products.map((product) => (
            <div key={product.slug} className="rounded-2xl border border-white/5 bg-white/[0.03] p-6">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <span className="inline-block rounded-full px-3 py-1 text-xs" style={{ background: 'rgba(255,59,111,0.15)', color: '#ff6b8a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                    {product.tag}
                  </span>
                  <h2 className="mt-3 text-2xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
                    {product.name}
                  </h2>
                  <p className="mt-3 text-base leading-relaxed" style={{ color: '#c4d4c8' }}>
                    {product.description}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-2xl font-bold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                    {product.price}
                  </p>
                </div>
              </div>

              <div className="mt-6">
                <p className="text-sm font-semibold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                  Business Benefit:
                </p>
                <p className="mt-1 text-sm" style={{ color: '#c4d4c8' }}>
                  {product.benefit}
                </p>
              </div>

              <div className="mt-6">
                <p className="text-sm font-semibold" style={{ color: '#f1f5f0', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                  What&apos;s Included:
                </p>
                <ul className="mt-2 space-y-1">
                  {product.includes.map((item) => (
                    <li key={item} className="flex items-center gap-2 text-sm" style={{ color: '#c4d4c8' }}>
                      <span style={{ color: '#2dd47a' }}>✓</span> {item}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="mt-6">
                <Link
                  href={`/products/${product.slug}`}
                  className="inline-block rounded-full px-6 py-2 text-sm font-semibold transition hover:opacity-90"
                  style={{ background: '#ff3b6f', color: '#fff', fontFamily: 'var(--font-jetbrains, monospace)' }}
                >
                  View Details
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}
