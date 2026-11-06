import Layout from '@/components/layout/Layout';
import MotionProvider from '@/components/motion/MotionProvider';
import Hero from '@/components/marketing/Hero';
import TrustStrip from '@/components/marketing/TrustStrip';
import ProductPillars from '@/components/marketing/ProductPillars';
import FeaturesBento from '@/components/marketing/FeaturesBento';
import HowItWorks from '@/components/marketing/HowItWorks';
import FromTheBlog from '@/components/marketing/FromTheBlog';
import CtaBand from '@/components/marketing/CtaBand';

export default function HomePage() {
  return (
    <Layout>
      <MotionProvider>
        <Hero />
        <TrustStrip />
        <ProductPillars />
        <FeaturesBento />
        <HowItWorks />
        <FromTheBlog />
        <CtaBand />
      </MotionProvider>
    </Layout>
  );
}
