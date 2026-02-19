import PublicBuilderPageClient from './_client';

export function generateStaticParams() {
  return [{ slug: '_placeholder' }];
}

export default async function Page({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <PublicBuilderPageClient slug={slug} />;
}
