import PublicBuilderPageClient from './_client';

export function generateStaticParams() {
  return [];
}

export default function Page({ params }: { params: { slug: string } }) {
  return <PublicBuilderPageClient slug={params.slug} />;
}
