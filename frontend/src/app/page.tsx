import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

const features = [
  {
    name: 'Lightning Fast',
    description: 'Shorten URLs instantly with our optimized infrastructure.',
    Icon: LinkIcon,
  },
  {
    name: 'Advanced Analytics',
    description: 'Track clicks, locations, and referrers with detailed insights.',
    Icon: ChartBarIcon,
  },
  {
    name: 'Secure & Reliable',
    description: 'Enterprise-grade security with 99.9% uptime guarantee.',
    Icon: ShieldCheckIcon,
  },
];

const stats = [
  { value: '500K+', label: 'URLs shortened' },
  { value: '50M+', label: 'Clicks tracked' },
  { value: '99.9%', label: 'Uptime' },
];

const ctaFeatures = ['Free forever plan', 'No credit card required', 'Setup in 2 minutes'];

export default function HomePage() {
  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary/5 via-background to-primary/5 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30 dark:opacity-5" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary bg-primary/10 hover:bg-primary/20 transition-colors">
                ✨ New: Advanced Analytics Dashboard
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-foreground">
                Free URL Shortener —{' '}
                <span className="text-primary bg-gradient-to-r from-primary to-primary/80 bg-clip-text text-transparent">
                  Shorten Links & Track Clicks
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-muted-foreground max-w-4xl mx-auto leading-relaxed">
                Transform long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more. No credit card required.
              </p>
            </div>
          </div>

          {/* URL Shortener Component */}
          <div className="mt-16">
            <URLShortener />
          </div>

          {/* Social Proof */}
          <div className="mt-16 text-center space-y-6">
            <p className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              Trusted by thousands of businesses worldwide
            </p>
            <div className="flex flex-col sm:flex-row justify-center items-center gap-8 sm:gap-12">
              {stats.map((stat, index) => (
                <>
                  <div key={stat.label} className="flex items-center gap-3">
                    <div className="text-3xl font-bold text-foreground">{stat.value}</div>
                    <div className="text-sm text-muted-foreground font-medium">{stat.label}</div>
                  </div>
                  {index < stats.length - 1 && (
                    <div key={`divider-${index}`} className="hidden sm:block w-px h-8 bg-border" />
                  )}
                </>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              Features
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              Why Choose Trunc?
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              More than just a URL shortener. Get powerful features that help you understand and optimize your links.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature) => (
              <Card
                key={feature.name}
                className="relative group hover:shadow-xl transition-all duration-300 bg-card border-border shadow-sm"
              >
                <CardHeader className="text-center space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <feature.Icon className="h-8 w-8 text-primary" />
                  </div>
                  <CardTitle className="text-xl font-bold text-foreground group-hover:text-primary transition-colors">
                    {feature.name}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-muted-foreground leading-relaxed text-center">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="relative py-20 lg:py-28 bg-primary overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-black/10 to-transparent" />
        <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="space-y-8">
            <div className="space-y-4">
              <Badge className="bg-white/20 text-primary-foreground hover:bg-white/30 border-0">
                🚀 Get Started Today
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-primary-foreground">
                Ready to get started?
              </h2>
              <p className="text-xl text-primary-foreground/80 max-w-3xl mx-auto leading-relaxed">
                Join thousands of businesses using Trunc to track and optimize their links. Start for free today with no credit card required.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-background text-primary hover:bg-muted font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href="/register">Get Started Free</a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="bg-transparent border-2 border-primary-foreground/30 text-primary-foreground hover:bg-primary-foreground/10 hover:border-primary-foreground/50 hover:text-primary-foreground font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href="/pricing">View Pricing</a>
              </Button>
            </div>

            <div className="pt-8 flex flex-col sm:flex-row justify-center items-center gap-6 text-primary-foreground/80">
              {ctaFeatures.map((feature, index) => (
                <div key={index} className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-sm">{feature}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
    </Layout>
  );
}
