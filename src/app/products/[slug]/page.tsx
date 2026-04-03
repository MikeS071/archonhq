import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import Link from 'next/link';

const BASE_URL = 'https://archonhq.ai';

const products: Record<string, {
  name: string;
  price: string;
  description: string;
  longDescription: string;
  benefit: string;
  includes: string[];
  whopProductId: string;
}> = {
  'coaching-report-automation': {
    name: 'Coaching Report Automation Template',
    price: '$39–$59',
    description: 'A ready-to-use system that automatically pulls client information, creates personalized reports, and sends them by email.',
    longDescription: 'Based on real work with a fitness coaching business, this template automates the entire client reporting workflow. Instead of spending hours manually compiling progress reports, the system pulls client data, generates personalized insights, and delivers professional reports automatically.',
    benefit: 'Cuts manual report work by up to 75%. A process that took 3+ hours per client now runs in minutes, freeing you to focus on actual coaching.',
    includes: [
      'Pre-built report templates (editable)',
      'Client data collection forms',
      'Automated email delivery system',
      'Step-by-step setup instructions',
      'Screenshots showing time savings',
      '30-day support for setup questions',
    ],
    whopProductId: process.env.WHOP_COACHING_TEMPLATE_ID || '',
  },
  'ai-content-operations-pack': {
    name: 'AI Content & Operations Automation Pack',
    price: '$49–$69',
    description: 'Simple templates to automatically generate and handle business content plus basic workflow helpers.',
    longDescription: 'A collection of automation templates designed for small business owners who spend too much time on content creation and admin tasks. Includes email automation, content repurposing workflows, and daily operations helpers.',
    benefit: 'Reduces time spent on writing and repetitive admin by 50%. More consistent content output without the burnout.',
    includes: [
      'Email template automation (5 templates)',
      'Content repurposing workflow',
      'Daily operations checklist automation',
      'Setup guide with examples',
      'Customization instructions',
      '30-day support for setup questions',
    ],
    whopProductId: process.env.WHOP_CONTENT_PACK_ID || '',
  },
};

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const product = products[slug];
  if (!product) return { title: 'Product Not Found' };

  return {
    title: `${product.name} — ArchonHQ`,
    description: product.description,
    alternates: { canonical: `${BASE_URL}/products/${slug}` },
  };
}

export default async function ProductDetailPage({ params }: Props) {
  const { slug } = await params;
  const product = products[slug];
  if (!product) notFound();

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

      <div className="relative z-10 mx-auto max-w-3xl px-6 py-16 md:px-10">
        <Link href="/products" className="text-sm transition hover:text-[#ff6b8a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          ← Back to products
        </Link>

        <h1 className="mt-6 text-3xl font-extrabold leading-tight text-white md:text-5xl" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          {product.name}
        </h1>

        <div className="mt-6 flex items-baseline gap-4">
          <span className="text-4xl font-bold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
            {product.price}
          </span>
          <span className="text-sm" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
            One-time payment
          </span>
        </div>

        <p className="mt-6 text-lg leading-relaxed" style={{ color: '#c4d4c8' }}>
          {product.longDescription}
        </p>

        <div className="mt-8 rounded-2xl border border-[#2dd47a]/20 bg-[#2dd47a]/5 p-6">
          <p className="text-sm font-semibold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
            Time Savings:
          </p>
          <p className="mt-2 text-base" style={{ color: '#f1d4c8' }}>
            {product.benefit}
          </p>
        </div>

        <div className="mt-8">
          <h2 className="text-xl font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            What&apos;s Included
          </h2>
          <ul className="mt-4 space-y-3">
            {product.includes.map((item) => (
              <li key={item} className="flex items-start gap-3" style={{ color: '#c4d4c8' }}>
                <span className="mt-0.5" style={{ color: '#2dd47a' }}>✓</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="mt-10">
          {product.whopProductId ? (
            <a
              href={`https://whop.com/checkout/${product.whopProductId}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-block rounded-full px-8 py-4 text-base font-semibold transition hover:opacity-90"
              style={{ background: '#ff3b6f', color: '#fff', fontFamily: 'var(--font-jetbrains, monospace)' }}
            >
              Buy Now — {product.price}
            </a>
          ) : (
            <p className="text-sm" style={{ color: '#6a7f6f' }}>
              Coming soon on Whop
            </p>
          )}
        </div>

        <div className="mt-12 rounded-2xl border border-white/5 bg-white/[0.03] p-6">
          <h3 className="text-lg font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            Have questions?
          </h3>
          <p className="mt-2 text-sm" style={{ color: '#c4d4c8' }}>
            Check out the <Link href="/insights" className="underline transition hover:text-[#2dd47a]">insights section</Link> for examples of how these automations work, or <Link href="/services" className="underline transition hover:text-[#2dd47a]">book a call</Link> to discuss your specific needs.
          </p>
        </div>
      </div>
    </main>
  );
}

export async function generateStaticParams() {
  return Object.keys(products).map((slug) => ({ slug }));
}
