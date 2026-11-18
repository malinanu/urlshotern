import Layout from '@/components/layout/Layout';
import MotionProvider from '@/components/motion/MotionProvider';
import TrustStrip from '@/components/marketing/TrustStrip';
import FeaturesHero from '@/components/features/FeaturesHero';
import WhatsNew from '@/components/features/WhatsNew';
import CardStudio from '@/components/features/CardStudio';
import PagesShowcase from '@/components/features/PagesShowcase';
import QrPlayground from '@/components/features/QrPlayground';
import PlatformGrid from '@/components/features/PlatformGrid';
import UseCases from '@/components/features/UseCases';
import PlansGlance from '@/components/features/PlansGlance';
import FeaturesClosing from '@/components/features/FeaturesClosing';
import { MICROSITES_ENABLED } from '@/lib/microsites';

/**
 * /features: shows the product with real pieces of it (the card templates, a live QR code,
 * the page designs) and asks for one thing, a free account. Every plan label and limit on
 * this page comes from internal/services/billing.go and internal/middleware/plan_gating.go.
 */
export default function FeaturesPage() {
  return (
    <Layout>
      <MotionProvider>
        <FeaturesHero />
        <TrustStrip />
        <WhatsNew />
        <CardStudio />
        {MICROSITES_ENABLED && <PagesShowcase />}
        <QrPlayground />
        <PlatformGrid />
        <UseCases />
        <PlansGlance />
        <FeaturesClosing />
      </MotionProvider>
    </Layout>
  );
}
