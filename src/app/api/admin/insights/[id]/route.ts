import { NextResponse } from 'next/server';
import { auth } from '@/lib/auth';
import { db } from '@/lib/db';
import { insights } from '@/db/schema';
import { eq } from 'drizzle-orm';

interface Props {
  params: Promise<{ id: string }>;
}

export async function GET(_request: Request, { params }: Props) {
  const session = await auth();
  if (!session?.user?.email) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;
  const [insight] = await db.select().from(insights).where(eq(insights.id, parseInt(id))).limit(1);

  if (!insight) {
    return NextResponse.json({ error: 'Not found' }, { status: 404 });
  }

  return NextResponse.json(insight);
}

export async function PUT(request: Request, { params }: Props) {
  const session = await auth();
  if (!session?.user?.email) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;

  try {
    const body = await request.json();
    const { slug, title, description, contentMd, sourceUrl, imageUrl } = body;

    if (!slug || !title || !description || !contentMd) {
      return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
    }

    const [updated] = await db.update(insights).set({
      slug,
      title,
      description,
      contentMd,
      sourceUrl: sourceUrl || null,
      imageUrl: imageUrl || null,
    }).where(eq(insights.id, parseInt(id))).returning();

    if (!updated) {
      return NextResponse.json({ error: 'Not found' }, { status: 404 });
    }

    return NextResponse.json(updated);
  } catch (err) {
    console.error('Update insight error:', err);
    return NextResponse.json({ error: 'Failed to update insight' }, { status: 500 });
  }
}

export async function DELETE(_request: Request, { params }: Props) {
  const session = await auth();
  if (!session?.user?.email) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { id } = await params;

  try {
    await db.delete(insights).where(eq(insights.id, parseInt(id)));
    return NextResponse.json({ success: true });
  } catch (err) {
    console.error('Delete insight error:', err);
    return NextResponse.json({ error: 'Failed to delete insight' }, { status: 500 });
  }
}
