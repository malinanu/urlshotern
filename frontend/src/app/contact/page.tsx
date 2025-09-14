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

const contactMethods = [
  {
    icon: EnvelopeIcon,
    title: 'Email Support',
    description: 'Get detailed help via email',
    contact: 'hello@urlshorter.com',
    responseTime: 'Within 24 hours',
    color: 'text-primary-600'
  },
  {
    icon: ChatBubbleLeftRightIcon,
    title: 'Live Chat',
    description: 'Instant help when you need it',
    contact: 'Start Chat',
    responseTime: 'Immediate',
    color: 'text-green-500'
  }
];

const inquiryTypes = [
  { value: 'general', label: 'General Inquiry' },
  { value: 'technical', label: 'Technical Support' },
  { value: 'billing', label: 'Billing Question' },
  { value: 'sales', label: 'Sales Inquiry' },
  { value: 'partnership', label: 'Partnership' },
  { value: 'feedback', label: 'Feedback' }
];

const priorityLevels = [
  { value: 'low', label: 'Low Priority' },
  { value: 'medium', label: 'Medium Priority' },
  { value: 'high', label: 'High Priority' },
  { value: 'urgent', label: 'Urgent' }
];

export default function ContactPage() {
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
      await new Promise(resolve => setTimeout(resolve, 2000));
      console.log('Form submitted:', formData);
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
      <section className="relative bg-gradient-to-br from-primary-50 via-white to-primary-50/30 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary-700 bg-primary-100 hover:bg-primary-200 transition-colors">
                <ChatBubbleLeftRightIcon className="h-4 w-4 mr-2" />
                Contact Us
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-gray-900">
                Get in{' '}
                <span className="text-primary-600 bg-gradient-to-r from-primary-600 to-primary-700 bg-clip-text text-transparent">
                  Touch
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
                Ready to transform your link management? Our team of experts is here to answer
                your questions and help you unlock the full potential of Trunc.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Contact Methods Section */}
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Contact Options
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Choose How to Connect
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Multiple ways to get the help you need, when you need it.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8 max-w-2xl mx-auto">
            {contactMethods.map((method, index) => (
              <Card
                key={index}
                className="relative group hover:shadow-xl transition-all duration-300 border-0 shadow-md bg-white text-center"
              >
                <CardHeader className="space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary-100 to-primary-200 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <method.icon className={`h-8 w-8 ${method.color}`} />
                  </div>
                  <div className="space-y-2">
                    <CardTitle className="text-lg font-bold text-gray-900 group-hover:text-primary-600 transition-colors">
                      {method.title}
                    </CardTitle>
                    <CardDescription className="text-gray-600">
                      {method.description}
                    </CardDescription>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <div className="font-medium text-gray-900">{method.contact}</div>
                    <Badge variant="secondary" className="text-xs bg-primary-100 text-primary-700">
                      {method.responseTime}
                    </Badge>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full border-primary-200 text-primary-700 hover:bg-primary-50"
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
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Contact Form
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Send us a Message
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Fill out the form below and we'll get back to you as soon as possible.
            </p>
          </div>

          <div className="max-w-2xl mx-auto">
            <Card className="bg-white border-0 shadow-xl">
              <CardHeader>
                <CardTitle className="text-2xl text-center">Let's Talk</CardTitle>
                <CardDescription className="text-center">
                  We're here to help you succeed with Trunc.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmit} className="space-y-6">
                  {formStatus === 'sent' && (
                    <div className="flex items-center gap-2 p-4 bg-green-50 border border-green-200 rounded-lg text-green-800">
                      <CheckCircleIcon className="h-5 w-5" />
                      <span>Message sent successfully! We'll get back to you soon.</span>
                    </div>
                  )}

                  {formStatus === 'error' && (
                    <div className="flex items-center gap-2 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800">
                      <ExclamationCircleIcon className="h-5 w-5" />
                      <span>Failed to send message. Please try again.</span>
                    </div>
                  )}

                  <div className="grid md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <label htmlFor="name" className="text-sm font-semibold text-gray-900">
                        Name *
                      </label>
                      <Input
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleInputChange}
                        placeholder="Your full name"
                        className={`border-gray-200 focus:border-primary-600 focus:ring-primary-600 ${formErrors.name ? 'border-red-500' : ''}`}
                      />
                      {formErrors.name && (
                        <p className="text-xs text-red-600">{formErrors.name}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <label htmlFor="email" className="text-sm font-semibold text-gray-900">
                        Email *
                      </label>
                      <Input
                        id="email"
                        name="email"
                        type="email"
                        value={formData.email}
                        onChange={handleInputChange}
                        placeholder="your@email.com"
                        className={`border-gray-200 focus:border-primary-600 focus:ring-primary-600 ${formErrors.email ? 'border-red-500' : ''}`}
                      />
                      {formErrors.email && (
                        <p className="text-xs text-red-600">{formErrors.email}</p>
                      )}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="company" className="text-sm font-semibold text-gray-900">
                      Company
                    </label>
                    <Input
                      id="company"
                      name="company"
                      value={formData.company}
                      onChange={handleInputChange}
                      placeholder="Your company name (optional)"
                      className="border-gray-200 focus:border-primary-600 focus:ring-primary-600"
                    />
                  </div>

                  <div className="grid md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <label htmlFor="inquiryType" className="text-sm font-semibold text-gray-900">
                        Inquiry Type
                      </label>
                      <select
                        id="inquiryType"
                        name="inquiryType"
                        value={formData.inquiryType}
                        onChange={handleInputChange}
                        className="w-full px-3 py-2 border border-gray-200 bg-white rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-600 focus:border-primary-600"
                      >
                        {inquiryTypes.map(type => (
                          <option key={type.value} value={type.value}>
                            {type.label}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="space-y-2">
                      <label htmlFor="priority" className="text-sm font-semibold text-gray-900">
                        Priority
                      </label>
                      <select
                        id="priority"
                        name="priority"
                        value={formData.priority}
                        onChange={handleInputChange}
                        className="w-full px-3 py-2 border border-gray-200 bg-white rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-600 focus:border-primary-600"
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
                    <label htmlFor="subject" className="text-sm font-semibold text-gray-900">
                      Subject *
                    </label>
                    <Input
                      id="subject"
                      name="subject"
                      value={formData.subject}
                      onChange={handleInputChange}
                      placeholder="Brief description of your inquiry"
                      className={`border-gray-200 focus:border-primary-600 focus:ring-primary-600 ${formErrors.subject ? 'border-red-500' : ''}`}
                    />
                    {formErrors.subject && (
                      <p className="text-xs text-red-600">{formErrors.subject}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="message" className="text-sm font-semibold text-gray-900">
                      Message *
                    </label>
                    <Textarea
                      id="message"
                      name="message"
                      rows={5}
                      value={formData.message}
                      onChange={handleInputChange}
                      placeholder="Please provide details about your inquiry..."
                      className={`border-gray-200 focus:border-primary-600 focus:ring-primary-600 ${formErrors.message ? 'border-red-500' : ''}`}
                    />
                    {formErrors.message && (
                      <p className="text-xs text-red-600">{formErrors.message}</p>
                    )}
                  </div>

                  <Button
                    type="submit"
                    disabled={formStatus === 'sending'}
                    size="lg"
                    className="w-full bg-primary-600 hover:bg-primary-700 text-white font-semibold py-6"
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
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              Additional Resources
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Alternative Support Channels
            </h2>
          </div>

          <div className="flex justify-center">
            <div className="max-w-md">
              <Card className="bg-gradient-to-br from-primary-50 to-white border-primary-100">
                <CardHeader className="text-center space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary-100 to-primary-200 rounded-2xl flex items-center justify-center">
                    <UserIcon className="h-8 w-8 text-primary-600" />
                  </div>
                  <div className="space-y-2">
                    <CardTitle className="text-xl font-bold text-gray-900">Documentation</CardTitle>
                    <CardDescription>Find answers in our comprehensive guides</CardDescription>
                  </div>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" size="lg" className="w-full border-primary-200 text-primary-700 hover:bg-primary-50">
                    Browse Documentation
                  </Button>
                </CardContent>
              </Card>
            </div>
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
                Start Your Journey with Trunc
              </h2>
              <p className="text-xl text-primary-100 max-w-3xl mx-auto leading-relaxed">
                Join thousands of businesses using Trunc to power their links.
                No credit card required to get started.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-white text-primary-700 hover:bg-gray-50 hover:text-primary-800 font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href="/register">Start Free Trial</a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="border-2 border-white/30 text-white hover:bg-white/10 hover:border-white/50 font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href="/pricing">View Pricing</a>
              </Button>
            </div>
          </div>
        </div>
      </section>
    </Layout>
  );
}