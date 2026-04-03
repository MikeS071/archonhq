import type { Metadata } from 'next';
import { Inter, Bricolage_Grotesque, JetBrains_Mono } from 'next/font/google';
import './globals.css';
import { Providers } from './providers';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const bricolage = Bricolage_Grotesque({ subsets: ['latin'], variable: '--font-bricolage', weight: ['400', '500', '600', '700', '800'] });
const jetbrains = JetBrains_Mono({ subsets: ['latin'], variable: '--font-jetbrains', weight: ['400', '500', '600'] });

const BASE_URL = 'https://archonhq.ai';

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  title: {
    default: 'ArchonHQ — AI Automation for Small Businesses',
    template: '%s | ArchonHQ',
  },
  description:
    'Practical AI automation that saves time and effort for busy small businesses and solopreneurs. Cut manual work by 50-80% on repetitive tasks.',
  keywords: [
    'AI automation',
    'small business automation',
    'solopreneur tools',
    'workflow automation',
    'AI templates',
    'automation templates',
  ],
  authors: [{ name: 'ArchonHQ', url: BASE_URL }],
  creator: 'ArchonHQ',
  openGraph: {
    type: 'website',
    locale: 'en_AU',
    url: BASE_URL,
    siteName: 'Mission Control',
    title: 'Mission Control — AI Agent Coordination Dashboard',
    description:
      'One dashboard to coordinate your AI agents, track costs, and cut your LLM spend by up to 50% with automatic smart routing.',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Mission Control — AI Agent Coordination Dashboard',
    description:
      'One dashboard to coordinate your AI agents, track costs, and cut LLM spend by up to 50%.',
    creator: '@teaser380',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  alternates: {
    canonical: BASE_URL,
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${bricolage.variable} ${jetbrains.variable} font-sans`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
