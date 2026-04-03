import type { Metadata } from 'next';
import Link from 'next/link';

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  title: 'Services — ArchonHQ',
  description:
    'Automation Quick-Win Service: Get one high-impact automation that actually saves you hours every week. $2,500–$5,000 fixed price.',
  alternates: { canonical: `${BASE_URL}/services` },
};

export default function ServicesPage() {
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
        <Link href="/" className="text-sm transition hover:text-[#ff6b8a]" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          ← Back to home
        </Link>

        <p className="mt-6 text-xs uppercase tracking-[0.4em]" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
          Consulting
        </p>
        <h1 className="mt-4 text-4xl font-extrabold leading-tight text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
          Automation Quick-Win Service
        </h1>

        <p className="mt-6 text-lg leading-relaxed" style={{ color: '#c4d4c8' }}>
          Get one high-impact automation that actually saves you hours every week.
        </p>

        <div className="mt-8 rounded-2xl border border-white/5 bg-white/[0.03] p-6">
          <div className="flex items-baseline justify-between">
            <span className="text-3xl font-bold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
              $2,500–$5,000
            </span>
            <span className="text-sm" style={{ color: '#6a7f6f', fontFamily: 'var(--font-jetbrains, monospace)' }}>
              Fixed price
            </span>
          </div>

          <h2 className="mt-6 text-lg font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            What&apos;s Included
          </h2>
          <ul className="mt-4 space-y-3">
            {[
              '60-minute discovery call to identify your biggest time drain',
              'Custom automation built for your specific workflow',
              'Simple handover so you can run it yourself',
              '30-day support for any questions',
            ].map((item) => (
              <li key={item} className="flex items-start gap-3" style={{ color: '#c4d4c8' }}>
                <span className="mt-0.5" style={{ color: '#2dd47a' }}>✓</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="mt-8 rounded-2xl border border-[#2dd47a]/20 bg-[#2dd47a]/5 p-6">
          <p className="text-sm font-semibold" style={{ color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
            Business Benefit:
          </p>
          <p className="mt-2 text-base" style={{ color: '#c4d4c8' }}>
            Delivers measurable time savings (often 50-75% on the chosen task), reduces weekly effort, and gives you a system you can run yourself.
          </p>
        </div>

        <div className="mt-8">
          <h2 className="text-lg font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            Book a Call
          </h2>
          <p className="mt-2 text-sm" style={{ color: '#c4d4c8' }}>
            Limited to 1 client per month. All work done outside AustralianSuper hours (evenings/weekends AEDT).
          </p>

          <div className="mt-6">
            <iframe
              src="https://cal.com/embed/calendar-widget-embed.html?calendarId=placeholder&embedDomain=archonhq.ai"
              className="w-full rounded-xl"
              style={{ border: '1px solid rgba(45,212,122,0.12)', minHeight: '600px', background: 'transparent' }}
              title="Book a call"
            />
          </div>

          <p className="mt-4 text-xs" style={{ color: '#6a7f6f' }}>
            Prefer email?{' '}
            <a href="mailto:mike@archonhq.ai" className="underline transition hover:text-[#2dd47a]">
              mike@archonhq.ai
            </a>
          </p>
        </div>

        <div className="mt-12 rounded-2xl border border-white/5 bg-white/[0.03] p-6">
          <h3 className="text-lg font-bold text-white" style={{ fontFamily: 'var(--font-bricolage, sans-serif)' }}>
            How It Works
          </h3>
          <div className="mt-4 space-y-4">
            {[
              { step: '1', title: 'Book a call', desc: 'Schedule a 60-minute discovery call via Calendly.' },
              { step: '2', title: 'Identify the problem', desc: 'We pinpoint the one repetitive task eating your time.' },
              { step: '3', title: 'Build the automation', desc: 'I build a custom solution tailored to your workflow.' },
              { step: '4', title: 'Handover & support', desc: 'You get full documentation and 30-day support.' },
            ].map((item) => (
              <div key={item.step} className="flex gap-4">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-bold" style={{ background: 'rgba(45,212,122,0.15)', color: '#2dd47a', fontFamily: 'var(--font-jetbrains, monospace)' }}>
                  {item.step}
                </span>
                <div>
                  <p className="font-semibold text-white" style={{ fontFamily: 'var(--font-jetbrains, monospace)' }}>{item.title}</p>
                  <p className="text-sm" style={{ color: '#c4d4c8' }}>{item.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </main>
  );
}
