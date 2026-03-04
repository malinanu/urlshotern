import AudiencePageClient from './_client';

export function generateStaticParams() {
  return [{ shortCode: '_placeholder' }];
}

export default function Page() {
  return <AudiencePageClient />;
}
