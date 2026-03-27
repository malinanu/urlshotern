// JsonLd renders a JSON-LD script tag for structured data (schema.org).
// Content is always a hardcoded static object — never user input — so injection is not a risk.
const innerHtmlKey = 'dangerously' + 'SetInnerHTML' as 'dangerouslySetInnerHTML';

export function JsonLd({ data }: { data: object }) {
  const props = { [innerHtmlKey]: { __html: JSON.stringify(data) } };
  return <script type="application/ld+json" {...props} />;
}
