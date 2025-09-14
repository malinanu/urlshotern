'use client';

import Layout from '@/components/layout/Layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  LinkIcon,
  ShieldCheckIcon,
  GlobeAltIcon,
  UsersIcon,
  ChartBarIcon,
  ClockIcon,
  ArrowRightIcon,
  BuildingOfficeIcon,
  HeartIcon,
  StarIcon
} from '@heroicons/react/24/outline';

const stats = [
  { label: 'Links Shortened', value: '10M+', icon: LinkIcon, color: 'text-primary-600' },
  { label: 'Uptime', value: '99.9%', icon: ShieldCheckIcon, color: 'text-green-500' },
  { label: 'Countries', value: '150+', icon: GlobeAltIcon, color: 'text-primary-600' },
  { label: 'Active Users', value: '500K+', icon: UsersIcon, color: 'text-primary-600' }
];

const values = [
  {
    title: 'Innovation First',
    description: 'We constantly push the boundaries of what\'s possible in URL management and analytics.',
    icon: LinkIcon
  },
  {
    title: 'Security & Privacy',
    description: 'Your data and links are protected with enterprise-grade security and privacy measures.',
    icon: ShieldCheckIcon
  },
  {
    title: 'Performance',
    description: 'Lightning-fast redirects and real-time analytics ensure optimal user experience.',
    icon: ChartBarIcon
  },
  {
    title: 'Reliability',
    description: '99.9% uptime guarantee ensures your links work when your audience needs them most.',
    icon: ClockIcon
  }
];

const achievements = [
  { year: '2024', title: 'Reached 10M Links Shortened', description: 'Milestone achievement in our growth journey' },
  { year: '2023', title: 'SOC 2 Type II Compliance', description: 'Enterprise-grade security certification' },
  { year: '2023', title: '99.9% Uptime Achievement', description: 'Consistent reliability and performance' },
  { year: '2022', title: 'Global Expansion', description: 'Serving users in over 150 countries worldwide' }
];

export default function AboutPage() {
  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary-50 via-white to-primary-50/30 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary-700 bg-primary-100 hover:bg-primary-200 transition-colors">
                <BuildingOfficeIcon className="h-4 w-4 mr-2" />
                About Trunc
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-gray-900">
                We Are{' '}
                <span className="text-primary-600 bg-gradient-to-r from-primary-600 to-primary-700 bg-clip-text text-transparent">
                  Trunc
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
                Pioneering the next generation of URL management with cutting-edge technology,
                unparalleled performance, and a vision that connects the digital world.
              </p>
            </div>
          </div>

          {/* Stats Grid */}
          <div className="mt-16 grid grid-cols-2 lg:grid-cols-4 gap-8">
            {stats.map((stat, index) => (
              <div key={index} className="text-center space-y-3">
                <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary-100 to-primary-200 rounded-2xl flex items-center justify-center">
                  <stat.icon className={`h-8 w-8 ${stat.color}`} />
                </div>
                <div className="space-y-1">
                  <div className="text-3xl font-bold text-gray-900">{stat.value}</div>
                  <div className="text-sm font-medium text-gray-500">{stat.label}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Mission Section */}
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <Badge variant="outline" className="text-primary-600 border-primary-200">
                  Our Mission
                </Badge>
                <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
                  Connecting Every Digital Experience
                </h2>
                <p className="text-xl text-gray-600 leading-relaxed">
                  We envision a world where every URL is intelligent, every click tells a story,
                  and every connection drives meaningful outcomes. Trunc isn't just a tool—it's
                  the bridge between intention and impact in the digital ecosystem.
                </p>
              </div>

              <div className="flex flex-col sm:flex-row gap-4">
                <Button size="lg" className="flex items-center gap-2 bg-primary-600 hover:bg-primary-700">
                  Get Started
                  <ArrowRightIcon className="h-5 w-5" />
                </Button>
                <Button size="lg" variant="outline" className="border-primary-200 text-primary-700 hover:bg-primary-50">
                  Learn More
                </Button>
              </div>
            </div>

            <div className="lg:order-first">
              <Card className="bg-gradient-to-br from-primary-50 to-white border-primary-100">
                <CardContent className="p-8">
                  <div className="grid grid-cols-2 gap-4 mb-6">
                    <div className="space-y-2">
                      <div className="h-3 bg-primary-200 rounded"></div>
                      <div className="h-2 bg-primary-100 rounded w-3/4"></div>
                      <div className="h-2 bg-primary-100 rounded w-1/2"></div>
                    </div>
                    <div className="space-y-2">
                      <div className="h-3 bg-primary-150 rounded"></div>
                      <div className="h-2 bg-primary-100 rounded w-2/3"></div>
                      <div className="h-2 bg-primary-100 rounded w-3/4"></div>
                    </div>
                  </div>
                  <div className="text-center text-sm text-gray-500 font-medium">
                    Intelligent Analytics Dashboard
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>

      {/* Values Section */}
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Our Values
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              What Drives Us Forward
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Our core values shape every decision we make and define how we build
              products that truly serve our customers.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            {values.map((value, index) => (
              <Card
                key={index}
                className="relative group hover:shadow-xl transition-all duration-300 border-0 shadow-md bg-white"
              >
                <CardHeader className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 bg-gradient-to-br from-primary-100 to-primary-200 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                      <value.icon className="h-6 w-6 text-primary-600" />
                    </div>
                    <CardTitle className="text-xl font-bold text-gray-900 group-hover:text-primary-600 transition-colors">
                      {value.title}
                    </CardTitle>
                  </div>
                  <CardDescription className="text-gray-600 leading-relaxed">
                    {value.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2 text-primary-600">
                    <StarIcon className="h-4 w-4" />
                    <span className="text-sm font-medium">Core Principle</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Team Section */}
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Our Team
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Global Remote Team
            </h2>
          </div>

          <Card className="bg-gradient-to-br from-primary-50 to-white border-primary-100 text-center">
            <CardContent className="p-12">
              <div className="space-y-6">
                <div className="mx-auto w-20 h-20 bg-gradient-to-br from-primary-100 to-primary-200 rounded-2xl flex items-center justify-center">
                  <UsersIcon className="h-10 w-10 text-primary-600" />
                </div>
                <div className="space-y-4">
                  <p className="text-xl text-gray-600 max-w-2xl mx-auto leading-relaxed">
                    Our diverse team of engineers, designers, and product experts work from around
                    the world to bring you the best URL shortening experience.
                  </p>
                  <Button variant="outline" size="lg" className="border-primary-200 text-primary-700 hover:bg-primary-50">
                    View Open Positions
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Achievements Section */}
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Achievements
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Key Milestones
            </h2>
          </div>

          <div className="grid gap-6">
            {achievements.map((achievement, index) => (
              <Card key={index} className="border-l-4 border-l-primary-600 bg-white hover:shadow-lg transition-shadow duration-300">
                <CardHeader>
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    <div className="space-y-2">
                      <CardTitle className="text-xl font-bold text-gray-900">{achievement.title}</CardTitle>
                      <CardDescription className="text-gray-600">{achievement.description}</CardDescription>
                    </div>
                    <Badge className="bg-primary-100 text-primary-700 hover:bg-primary-200 w-fit">
                      {achievement.year}
                    </Badge>
                  </div>
                </CardHeader>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="relative py-20 lg:py-28 bg-gradient-to-br from-primary-600 via-primary-700 to-primary-800 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-primary-600/20 to-primary-800/20" />
        <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="space-y-8">
            <div className="space-y-4">
              <Badge className="bg-white/20 text-white hover:bg-white/30 border-0">
                🚀 Ready to Get Started?
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white">
                Join thousands of businesses using Trunc
              </h2>
              <p className="text-xl text-primary-100 max-w-3xl mx-auto leading-relaxed">
                Start shortening URLs and tracking analytics with our powerful platform.
                No credit card required to get started.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-white text-primary-700 hover:bg-gray-50 hover:text-primary-800 font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href="/register">
                  Start Free Trial
                  <ArrowRightIcon className="h-5 w-5 ml-2" />
                </a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="border-2 border-white/30 text-white hover:bg-white/10 hover:border-white/50 font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href="/contact">Contact Sales</a>
              </Button>
            </div>

            <div className="pt-8 flex flex-col sm:flex-row justify-center items-center gap-6 text-primary-100">
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                <span className="text-sm">Free forever plan</span>
              </div>
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                <span className="text-sm">No credit card required</span>
              </div>
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                <span className="text-sm">Setup in 2 minutes</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </Layout>
  );
}