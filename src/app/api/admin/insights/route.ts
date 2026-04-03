import { NextResponse } from 'next/server';
import { auth } from '@/lib/auth';
import { db } from '@/lib/db';
import { insights } from '@/db/schema';
import { eq } from 'drizzle-orm';

export async function GET() {
  const session = await auth();
  if (!session?.user?.email) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const allInsights = await db.select().from(insights);
  return NextResponse.json(allInsights);
}

export async function POST(request: Request) {
  const session = await auth();
  if (!session?.user?.email) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const body = await request.json();
    const { slug, title, description, contentMd, sourceUrl, imageUrl, publishedAt } = body;

    if (!slug || !title || !description || !contentMd) {
      return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
    }

    const [newInsight] = await db.insert(insights).values({
      slug,
      title,
      description,
      contentMd,
      sourceUrl: sourceUrl || null,
      imageUrl: imageUrl || null,
      publishedAt: publishedAt ? new Date(publishedAt) : new Date(),
    }).returning();

    return NextResponse.json(newInsight, { status: 201 });
  } catch (err) {
    console.error('Create insight error:', err);
    return NextResponse.json({ error: 'Failed to create insight' }, { status: 500 });
  }
}
