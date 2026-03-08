'use client';

import { useState } from 'react';
import Layout from '@/components/layout/Layout';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  EnvelopeIcon,
  ChatBubbleLeftRightIcon,
  PaperAirplaneIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  UserIcon
} from '@heroicons/react/24/outline';
import { useCMSPage, getContentBlock, getCustomField } from '@/hooks/useCMSPage';

interface FormData {
  name: string;
  email: string;
  company: string;
  subject: string;
  message: string;
  inquiryType: string;
  priority: string;
}

type FormStatus = 'idle' | 'sending' | 'sent' | 'error';

// Fallback data (used if CMS fetch fails)
const fallbackContactMethods = [
  {
    icon: 'EnvelopeIcon',
    title: 'Email Support',
    description: 'Get detailed help via email',
    contact: 'hello@urlshorter.com',
    responseTime: 'Within 24 hours',
    color: 'text-primary'
  },
  {
    icon: 'ChatBubbleLeftRightIcon',
    title: 'Live Chat',
    description: 'Instant help when you need it',
    contact: 'Start Chat',
    responseTime: 'Immediate',
    color: 'text-green-500'
  }
];

const fallbackInquiryTypes = [
  { value: 'general', label: 'General Inquiry' },
  { value: 'technical', label: 'Technical Support' },
  { value: 'billing', label: 'Billing Question' },
  { value: 'sales', label: 'Sales Inquiry' },
  { value: 'partnership', label: 'Partnership' },
  { value: 'feedback', label: 'Feedback' }
];

const fallbackPriorityLevels = [
  { value: 'low', label: 'Low Priority' },
  { value: 'medium', label: 'Medium Priority' },
  { value: 'high', label: 'High Priority' },
  { value: 'urgent', label: 'Urgent' }
];

const iconMap: Record<string, any> = {
  EnvelopeIcon,
  ChatBubbleLeftRightIcon,
  UserIcon,
};

export default function ContactPage() {
  // Fetch page content from CMS with fallback
  const { content: page, loading, usedCMS } = useCMSPage('contact', {
    slug: 'contact',
    title: 'Get in Touch',
    content: 'Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc.',
  });

  // Extract content blocks from CMS or use fallback
  const hero = getContentBlock(page, 'hero', {
    badge: 'Contact Us',
    badge_icon: 'ChatBubbleLeftRightIcon',
    title: 'Get in Touch',
    title_highlight: 'Touch',
    tagline: 'Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc.',
  });

  const contactMethodsBlock = getContentBlock(page, 'contact_methods', {
    badge: 'Contact Options',
    heading: 'Choose How to Connect',
    tagline: 'Multiple ways to get the help you need, when you need it.',
    methods: fallbackContactMethods,
  });

  const contactFormBlock = getContentBlock(page, 'contact_form', {
    badge: 'Contact Form',
    heading: 'Send us a Message',
    tagline: 'Fill out the form below and we\'ll get back to you as soon as possible.',
    form_title: 'Let\'s Talk',
    form_description: 'We\'re here to help you succeed with Trunc.',
  });

  const supportResourcesBlock = getContentBlock(page, 'support_resources', {
    badge: 'Additional Resources',
    heading: 'Alternative Support Channels',
    items: [
      {
        icon: 'UserIcon',
        title: 'Documentation',
        description: 'Find answers in our comprehensive guides',
        cta: { text: 'Browse Documentation', url: '/docs' },
      },
    ],
  });

  const ctaBottom = getContentBlock(page, 'cta_bottom', {
    badge: '🚀 Ready to Get Started?',
    heading: 'Start Your Journey with Trunc',
    tagline: 'Join thousands of businesses using Trunc to power their links. No credit card required to get started.',
    cta_primary: { text: 'Start Free Trial', url: '/register' },
    cta_secondary: { text: 'View Pricing', url: '/pricing' },
  });

  // Extract custom fields
  const inquiryTypes = getCustomField(page, 'inquiry_types', fallbackInquiryTypes);
  const priorityLevels = getCustomField(page, 'priority_levels', fallbackPriorityLevels);

  // Map icon names to actual components
  const contactMethods = (contactMethodsBlock.methods || fallbackContactMethods).map((method: any) => ({
    ...method,
    icon: method.icon && typeof method.icon === 'string' ? (iconMap[method.icon] || EnvelopeIcon) : method.icon || EnvelopeIcon,
  }));

  const supportResources = (supportResourcesBlock.items || []).map((item: any) => ({
    ...item,
    icon: item.icon && typeof item.icon === 'string' ? (iconMap[item.icon] || UserIcon) : item.icon || UserIcon,
  }));
  // Form state
  const [formData, setFormData] = useState<FormData>({
    name: '',
    email: '',
    company: '',
    subject: '',
    message: '',
    inquiryType: 'general',
    priority: 'medium'
  });
  const [formStatus, setFormStatus] = useState<FormStatus>('idle');
  const [formErrors, setFormErrors] = useState<Partial<FormData>>({});

  // Loading state
  if (loading) {
    return (
      <Layout>
        <div className="flex items-center justify-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"></div>
        </div>
      </Layout>
    );
  }

  const validateForm = (): boolean => {
    const errors: Partial<FormData> = {};

    if (!formData.name.trim()) errors.name = 'Name is required';
    if (!formData.email.trim()) errors.email = 'Email is required';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) errors.email = 'Valid email is required';
    if (!formData.subject.trim()) errors.subject = 'Subject is required';
    if (!formData.message.trim()) errors.message = 'Message is required';

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));

    if (formErrors[name as keyof FormData]) {
      setFormErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setFormStatus('sending');

    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL || '';
      const res = await fetch(`${apiBase}/api/v1/contact`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          email: formData.email,
          subject: formData.subject,
          message: formData.message,
        }),
      });
      if (!res.ok) throw new Error('Failed to send');
      setFormStatus('sent');

      setFormData({
        name: '',
        email: '',
        company: '',
        subject: '',
        message: '',
        inquiryType: 'general',
        priority: 'medium'
      });

      setTimeout(() => setFormStatus('idle'), 5000);
    } catch (error) {
      setFormStatus('error');
      setTimeout(() => setFormStatus('idle'), 3000);
    }
  };

  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary/5 via-background to-primary/5 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30 dark:opacity-5" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary bg-primary/10 hover:bg-primary/20 transition-colors">
                {hero.badge_icon && iconMap[hero.badge_icon] && (
                  (() => {
                    const IconComponent = iconMap[hero.badge_icon];
                    return <IconComponent className="h-4 w-4 mr-2" />;
                  })()
                )}
                {hero.badge}
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-foreground">
                {hero.title.replace(hero.title_highlight, '')}{' '}
                <span className="text-primary bg-gradient-to-r from-primary to-primary/80 bg-clip-text text-transparent">
                  {hero.title_highlight}
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-muted-foreground max-w-4xl mx-auto leading-relaxed">
                {hero.tagline}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Contact Methods Section */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              {contactMethodsBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              {contactMethodsBlock.heading}
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              {contactMethodsBlock.tagline}
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8 max-w-2xl mx-auto">
            {(contactMethods || []).map((method, index) => (
              <Card
                key={index}
                className="relative group hover:shadow-xl transition-all duration-300 border-border shadow-sm bg-card text-center"
              >
                <CardHeader className="space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <method.icon className={`h-8 w-8 ${method.color}`} />
                  </div>
                  <div className="space-y-2">
                    <CardTitle className="text-lg font-bold text-foreground group-hover:text-primary transition-colors">
                      {method.title}
                    </CardTitle>
                    <CardDescription className="text-muted-foreground">
                      {method.description}
                    </CardDescription>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <div className="font-medium text-foreground">{method.contact}</div>
                    <Badge variant="secondary" className="text-xs bg-primary/10 text-primary">
                      {method.responseTime}
                    </Badge>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full border-primary/20 text-primary hover:bg-primary/10"
                  >
                    <method.icon className="h-4 w-4 mr-2" />
                    Connect Now
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Contact Form Section */}
      <section className="py-20 lg:py-28 bg-muted/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              {contactFormBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              {contactFormBlock.heading}
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              {contactFormBlock.tagline}
            </p>
          </div>

          <div className="max-w-2xl mx-auto">
            <Card className="bg-card border-border shadow-md">
              <CardHeader>
                <CardTitle className="text-2xl text-center text-foreground">{contactFormBlock.form_title}</CardTitle>
                <CardDescription className="text-center text-muted-foreground">
                  {contactFormBlock.form_description}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmit} className="space-y-6">
                  {formStatus === 'sent' && (
                    <div className="flex items-center gap-2 p-4 bg-green-500/10 border border-green-500/20 rounded-lg text-green-600 dark:text-green-500">
                      <CheckCircleIcon className="h-5 w-5" />
                      <span>Message sent successfully! We'll get back to you soon.</span>
                    </div>
                  )}

                  {formStatus === 'error' && (
                    <div className="flex items-center gap-2 p-4 bg-destructive/10 border border-destructive/20 rounded-lg text-destructive">
                      <ExclamationCircleIcon className="h-5 w-5" />
                      <span>Failed to send message. Please try again.</span>
                    </div>
                  )}

                  <div className="grid md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <label htmlFor="name" className="text-sm font-semibold text-foreground">
                        Name *
                      </label>
                      <Input
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleInputChange}
                        placeholder="Your full name"
                        disabled={formStatus === 'sending'}
                        className={`border-input focus:border-primary focus:ring-primary ${formErrors.name ? 'border-destructive' : ''}`}
                      />
                      {formErrors.name && (
                        <p className="text-xs text-destructive">{formErrors.name}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <label htmlFor="email" className="text-sm font-semibold text-foreground">
                        Email *
                      </label>
                      <Input
                        id="email"
                        name="email"
                        type="email"
                        value={formData.email}
                        onChange={handleInputChange}
                        placeholder="your@email.com"
                        disabled={formStatus === 'sending'}
                        className={`border-input focus:border-primary focus:ring-primary ${formErrors.email ? 'border-destructive' : ''}`}
                      />
                      {formErrors.email && (
                        <p className="text-xs text-red-600">{formErrors.email}</p>
                      )}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="company" className="text-sm font-semibold text-foreground">
                      Company
                    </label>
                    <Input
                      id="company"
                      name="company"
                      value={formData.company}
                      onChange={handleInputChange}
                      placeholder="Your company name (optional)"
                      disabled={formStatus === 'sending'}
                      className="border-input focus:border-primary focus:ring-primary"
                    />
                  </div>

                  <div className="grid md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <label htmlFor="inquiryType" className="text-sm font-semibold text-foreground">
                        Inquiry Type
                      </label>
                      <select
                        id="inquiryType"
                        name="inquiryType"
                        value={formData.inquiryType}
                        onChange={handleInputChange}
                        disabled={formStatus === 'sending'}
                        className="w-full px-3 py-2 border border-input bg-background rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground"
                      >
                        {inquiryTypes.map(type => (
                          <option key={type.value} value={type.value}>
                            {type.label}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="space-y-2">
                      <label htmlFor="priority" className="text-sm font-semibold text-foreground">
                        Priority
                      </label>
                      <select
                        id="priority"
                        name="priority"
                        value={formData.priority}
                        onChange={handleInputChange}
                        disabled={formStatus === 'sending'}
                        className="w-full px-3 py-2 border border-input bg-background rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground"
                      >
                        {priorityLevels.map(level => (
                          <option key={level.value} value={level.value}>
                            {level.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="subject" className="text-sm font-semibold text-foreground">
                      Subject *
                    </label>
                    <Input
                      id="subject"
                      name="subject"
                      value={formData.subject}
                      onChange={handleInputChange}
                      placeholder="Brief description of your inquiry"
                      disabled={formStatus === 'sending'}
                      className={`border-input focus:border-primary focus:ring-primary ${formErrors.subject ? 'border-destructive' : ''}`}
                    />
                    {formErrors.subject && (
                      <p className="text-xs text-destructive">{formErrors.subject}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="message" className="text-sm font-semibold text-foreground">
                      Message *
                    </label>
                    <Textarea
                      id="message"
                      name="message"
                      rows={5}
                      value={formData.message}
                      onChange={handleInputChange}
                      placeholder="Please provide details about your inquiry..."
                      disabled={formStatus === 'sending'}
                      className={`border-input focus:border-primary focus:ring-primary ${formErrors.message ? 'border-destructive' : ''}`}
                    />
                    {formErrors.message && (
                      <p className="text-xs text-destructive">{formErrors.message}</p>
                    )}
                  </div>

                  <Button
                    type="submit"
                    disabled={formStatus === 'sending'}
                    size="lg"
                    className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-6"
                  >
                    {formStatus === 'sending' ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2"></div>
                        Sending...
                      </>
                    ) : (
                      <>
                        <PaperAirplaneIcon className="h-5 w-5 mr-2" />
                        Send Message
                      </>
                    )}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Support Resources Section */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              {supportResourcesBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              {supportResourcesBlock.heading}
            </h2>
          </div>

          <div className="flex justify-center">
            {(supportResources || []).map((resource: any, index: number) => (
              <div key={index} className="max-w-md">
                <Card className="bg-gradient-to-br from-primary/5 to-background border-primary/10">
                  <CardHeader className="text-center space-y-4">
                    <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center">
                      <resource.icon className="h-8 w-8 text-primary" />
                    </div>
                    <div className="space-y-2">
                      <CardTitle className="text-xl font-bold text-foreground">{resource.title}</CardTitle>
                      <CardDescription className="text-muted-foreground">{resource.description}</CardDescription>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <Button
                      variant="outline"
                      size="lg"
                      className="w-full border-primary/20 text-primary hover:bg-primary/10"
                      asChild
                    >
                      <a href={resource.cta.url}>{resource.cta.text}</a>
                    </Button>
                  </CardContent>
                </Card>
              </div>
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
                {ctaBottom.badge}
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-primary-foreground">
                {ctaBottom.heading}
              </h2>
              <p className="text-xl text-primary-foreground/80 max-w-3xl mx-auto leading-relaxed">
                {ctaBottom.tagline}
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-background text-primary hover:bg-muted font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href={ctaBottom.cta_primary.url}>{ctaBottom.cta_primary.text}</a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="bg-transparent border-2 border-primary-foreground/30 text-primary-foreground hover:bg-primary-foreground/10 hover:border-primary-foreground/50 hover:text-primary-foreground font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href={ctaBottom.cta_secondary.url}>{ctaBottom.cta_secondary.text}</a>
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Development Indicator */}
      {process.env.NODE_ENV === 'development' && usedCMS && (
        <div className="fixed bottom-4 right-4 bg-green-500 text-white px-3 py-1 rounded-full text-xs font-medium shadow-lg">
          ✓ Using CMS
        </div>
      )}
    </Layout>
  );
}