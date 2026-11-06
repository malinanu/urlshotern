import MicroPageBuilderClient from './_client';

// Required by `output: 'export'`; the client reads the real id from the URL.
export function generateStaticParams() {
  return [{ id: '_placeholder' }];
}

export default function Page() {
  return <MicroPageBuilderClient />;
}
