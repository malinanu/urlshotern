import Layout from '@/components/layout/Layout';
import { CheckIcon, XMarkIcon, StarIcon } from '@heroicons/react/24/outline';

export default function PricingPage() {
  const plans = [
    {
      name: 'Free',
      price: '$0',
      period: '/month',
      description: 'Perfect for getting started',
      features: [
        { name: '10 short links', included: true },
        { name: 'Basic analytics', included: true },
        { name: 'Standard support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: false },
        { name: 'Team collaboration', included: false },
        { name: 'Advanced analytics', included: false },
        { name: 'API access', included: false },
        { name: 'Priority support', included: false }
      ],
      cta: 'Get Started Free',
      popular: false,
      ctaLink: '/'
    },
    {
      name: 'Pro',
      price: '$5',
      period: '/month',
      description: 'For growing businesses',
      features: [
        { name: '1,000 short links', included: true },
        { name: 'Advanced analytics', included: true },
        { name: 'Priority support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: true },
        { name: 'Custom domains', included: true },
        { name: 'Team collaboration (5 users)', included: true },
        { name: 'API access', included: true },
        { name: 'Password protection', included: true }
      ],
      cta: 'Start Pro Plan',
      popular: true,
      ctaLink: '/register'
    },
    {
      name: 'Ultra',
      price: '$29.99',
      period: '/month',
      description: 'For enterprises & power users',
      features: [
        { name: 'Unlimited short links', included: true },
        { name: 'Advanced analytics', included: true },
        { name: '24/7 premium support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: true },
        { name: 'Custom domains', included: true },
        { name: 'Unlimited team members', included: true },
        { name: 'Unlimited API access', included: true },
        { name: 'Password protection', included: true }
      ],
      cta: 'Choose Ultra',
      popular: false,
      ctaLink: '/register'
    }
  ];

  const faqs = [
    {
      question: 'Can I change plans at any time?',
      answer: 'Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and billing is prorated.'
    },
    {
      question: 'What happens if I exceed my link limit?',
      answer: 'On the Free plan, you\'ll need to upgrade. On paid plans, we\'ll notify you and provide options to upgrade or purchase additional capacity.'
    },
    {
      question: 'Do you offer annual billing?',
      answer: 'Yes! Annual billing comes with a 20% discount. Contact us for annual pricing details.'
    },
    {
      question: 'Is there a free trial for paid plans?',
      answer: 'Yes, we offer a 14-day free trial for all paid plans. No credit card required to start.'
    },
    {
      question: 'Can I use my own domain?',
      answer: 'Custom domains are available on Pro and Enterprise plans. You can use your own branded short domain.'
    },
    {
      question: 'What kind of support do you provide?',
      answer: 'Free users get community support. Pro users get priority email support. Enterprise users get 24/7 phone and email support.'
    }
  ];

  return (
    <Layout>
      {/* Hero Section */}
      <section className="bg-gradient-to-br from-primary-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-black mb-6">
              Simple, Transparent{' '}
              <span className="text-primary-600">Pricing</span>
            </h1>
            <p className="text-xl text-black mb-12 max-w-3xl mx-auto">
              Choose the perfect plan for your needs. All plans include our core features 
              with no hidden fees or surprise charges.
            </p>
          </div>
        </div>
      </section>

      {/* Pricing Cards */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-3 gap-8">
            {plans.map((plan) => (
              <div 
                key={plan.name} 
                className={`relative bg-white rounded-2xl border-2 p-8 ${
                  plan.popular 
                    ? 'border-primary-600 shadow-xl transform scale-105' 
                    : 'border-gray-200 shadow-lg'
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                    <div className="bg-primary-600 text-white px-4 py-2 rounded-full text-sm font-medium flex items-center space-x-1">
                      <StarIcon className="h-4 w-4" />
                      <span>Most Popular</span>
                    </div>
                  </div>
                )}

                <div className="text-center">
                  <h3 className="text-2xl font-bold text-black mb-2">{plan.name}</h3>
                  <p className="text-gray-600 mb-6">{plan.description}</p>
                  
                  <div className="mb-6">
                    <span className="text-5xl font-bold text-black">{plan.price}</span>
                    <span className="text-xl text-gray-600">{plan.period}</span>
                  </div>

                  <a
                    href={plan.ctaLink}
                    className={`w-full py-3 px-6 rounded-lg font-semibold transition-colors block ${
                      plan.popular
                        ? 'bg-primary-600 text-white hover:bg-primary-700'
                        : 'bg-gray-100 text-black hover:bg-gray-200'
                    }`}
                  >
                    {plan.cta}
                  </a>
                </div>

                <div className="mt-8">
                  <h4 className="font-semibold text-black mb-4">What's included:</h4>
                  <ul className="space-y-3">
                    {plan.features.map((feature) => (
                      <li key={feature.name} className="flex items-center space-x-3">
                        {feature.included ? (
                          <CheckIcon className="h-5 w-5 text-green-600 flex-shrink-0" />
                        ) : (
                          <XMarkIcon className="h-5 w-5 text-gray-400 flex-shrink-0" />
                        )}
                        <span className={`text-sm ${feature.included ? 'text-black' : 'text-gray-400'}`}>
                          {feature.name}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Feature Comparison Table */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-black mb-4">
              Compare Plans
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              See exactly what's included in each plan to make the right choice for your needs.
            </p>
          </div>

          <div className="bg-white rounded-xl shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-black">Feature</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Free</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Pro</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Ultra</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  <tr>
                    <td className="px-6 py-4 text-sm text-black">Short Links</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">10</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">1,000</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr className="bg-gray-50">
                    <td className="px-6 py-4 text-sm text-black">Team Members</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">1</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">5</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr>
                    <td className="px-6 py-4 text-sm text-black">API Access</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">-</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">✓</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr className="bg-gray-50">
                    <td className="px-6 py-4 text-sm text-black">Support</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Standard</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Priority</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">24/7 Premium</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="py-20 bg-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-black mb-4">
              Frequently Asked Questions
            </h2>
            <p className="text-xl text-gray-600">
              Get answers to common questions about our pricing and plans.
            </p>
          </div>

          <div className="space-y-8">
            {faqs.map((faq, index) => (
              <div key={index} className="bg-gray-50 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-black mb-3">
                  {faq.question}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {faq.answer}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section className="py-20 bg-primary-600">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-white mb-4">
            Ready to Get Started?
          </h2>
          <p className="text-xl text-primary-100 mb-8 max-w-2xl mx-auto">
            Join thousands of users who trust Trunc for their link management needs.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href="/"
              className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-50 transition-colors"
            >
              Start Free Trial
            </a>
            <a
              href="/contact"
              className="border border-primary-400 text-white px-8 py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors"
            >
              Contact Sales
            </a>
          </div>
        </div>
      </section>
    </Layout>
  );
}