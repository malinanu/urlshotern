import MicroPageBuilderClient from './_client';

export function generateStaticParams() {
  return [{ id: '_placeholder' }];
}

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <MicroPageBuilderClient id={id} />;
}