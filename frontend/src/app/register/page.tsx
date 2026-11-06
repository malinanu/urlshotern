'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { EyeIcon, EyeSlashIcon, LinkIcon, CheckCircleIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';

// COPPA age gate: the server enforces the same minimum age
const MIN_SIGNUP_AGE = 13;
const AGE_GATE_KEY = 'age_gate_blocked';

function ageFromDOB(dob: string): number {
  const [y, m, d] = dob.split('-').map(Number);
  const now = new Date();
  let age = now.getUTCFullYear() - y;
  const month = now.getUTCMonth() + 1;
  if (month < m || (month === m && now.getUTCDate() < d)) age--;
  return age;
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

const registerSchema = z.object({
  name: z
    .string()
    .min(1, 'Full name is required')
    .min(2, 'Name must be at least 2 characters'),
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number'),
  phone: z
    .string()
    .min(1, 'Phone number is required')
    .min(10, 'Phone number must be at least 10 digits'),
  dateOfBirth: z
    .string()
    .min(1, 'Date of birth is required')
    .refine(val => /^\d{4}-\d{2}-\d{2}$/.test(val) && !Number.isNaN(Date.parse(val)), 'Please enter a valid date')
    .refine(val => val <= todayISO(), 'Date of birth cannot be in the future'),
  confirmPassword: z
    .string()
    .min(1, 'Please confirm your password'),
  acceptTerms: z
    .boolean()
    .refine(val => val === true, 'You must accept the terms and conditions'),
  marketingConsent: z.boolean().optional(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type RegisterFormData = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [ageBlocked, setAgeBlocked] = useState(false);
  const [ageRequiredProvider, setAgeRequiredProvider] = useState<string | null>(null);
  const { register: registerUser, isLoading } = useAuth();
  const router = useRouter();

  // Once someone is turned away by the age gate, keep them blocked on this device
  // so the date can't simply be changed and resubmitted.
  const blockAge = () => {
    try {
      localStorage.setItem(AGE_GATE_KEY, new Date().toISOString());
    } catch {
      // storage unavailable: the block still applies for this visit
    }
    setAgeBlocked(true);
  };

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    let stored: string | null = null;
    try {
      stored = localStorage.getItem(AGE_GATE_KEY);
    } catch {
      stored = null;
    }
    if (params.get('error') === 'age_restricted') {
      blockAge();
    } else if (stored) {
      setAgeBlocked(true);
    }
    if (params.get('error') === 'age_required') {
      setAgeRequiredProvider(params.get('provider') === 'github' ? 'GitHub' : 'Google');
    }
  }, []);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    trigger,
    getValues,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      acceptTerms: false,
      marketingConsent: false,
    },
  });

  const password = watch('password');

  const getPasswordStrength = (pass: string) => {
    if (!pass) return { strength: 0, text: '', color: '' };

    let strength = 0;
    if (pass.length >= 8) strength++;
    if (/[A-Z]/.test(pass)) strength++;
    if (/[a-z]/.test(pass)) strength++;
    if (/[0-9]/.test(pass)) strength++;
    if (/[^A-Za-z0-9]/.test(pass)) strength++;

    const levels = [
      { text: 'Very Weak', color: 'text-red-600' },
      { text: 'Weak', color: 'text-red-500' },
      { text: 'Fair', color: 'text-yellow-500' },
      { text: 'Good', color: 'text-blue-500' },
      { text: 'Strong', color: 'text-green-600' },
    ];

    return { strength, ...levels[Math.min(strength, 4)] };
  };

  const passwordStrength = getPasswordStrength(password || '');

  const continueWithProvider = async (provider: 'google' | 'github') => {
    setError(null);
    // Social sign-up still has to pass the age gate; the DOB travels with the OAuth state
    if (!(await trigger('dateOfBirth'))) return;
    const dob = getValues('dateOfBirth');
    if (ageFromDOB(dob) < MIN_SIGNUP_AGE) {
      blockAge();
      return;
    }
    const base = provider === 'google' ? API_ENDPOINTS.auth.google : API_ENDPOINTS.auth.github;
    window.location.href = `${base}?dob=${encodeURIComponent(dob)}`;
  };

  const onSubmit = async (data: RegisterFormData) => {
    setError(null);

    if (ageFromDOB(data.dateOfBirth) < MIN_SIGNUP_AGE) {
      blockAge();
      return;
    }

    try {
      const registrationData = {
        name: data.name,
        email: data.email,
        password: data.password,
        phone: data.phone,
        date_of_birth: data.dateOfBirth,
        terms_accepted: data.acceptTerms,
        marketing_consent: data.marketingConsent || false,
      };

      const result = await registerUser(registrationData);

      // Check if we got immediate auth tokens or need verification
      if (result.access_token) {
        // User was automatically logged in — redirect to verify-email if not yet verified
        if (!result.user?.email_verified) {
          router.push('/verify-email');
        } else {
          router.push('/');
        }
      } else {
        // No token returned, go to verify-email page
        router.push('/verify-email');
      }
    } catch (err) {
      if ((err as { code?: string })?.code === 'age_restricted') {
        blockAge();
        return;
      }
      setError(err instanceof Error ? err.message : 'Registration failed. Please try again.');
    }
  };

  if (ageBlocked) {
    return (
      <Layout>
        <div className="min-h-[80vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
          <div className="sm:mx-auto sm:w-full sm:max-w-md">
            <div className="bg-card py-8 px-4 border border-border sm:rounded-lg sm:px-10 text-center">
              <h1 className="text-2xl font-bold text-foreground mb-4">
                We can&apos;t create an account for you
              </h1>
              <p className="text-muted-foreground mb-6">
                Sorry, based on the information you provided you aren&apos;t eligible to create a Trunc
                account. We haven&apos;t saved any of the details you entered.
              </p>
              <Link
                href="/"
                className="w-full flex justify-center py-2 px-4 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-card hover:bg-muted"
              >
                Back to Home
              </Link>
            </div>
          </div>
        </div>
      </Layout>
    );
  }

  if (success) {
    return (
      <Layout>
        <div className="min-h-[80vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
          <div className="sm:mx-auto sm:w-full sm:max-w-md">
            <div className="bg-card py-8 px-4 border border-border sm:rounded-lg sm:px-10 text-center">
              <CheckCircleIcon className="h-16 w-16 text-green-600 dark:text-green-500 mx-auto mb-4" />
              <h2 className="text-2xl font-bold text-foreground mb-4">
                Registration Successful!
              </h2>
              <p className="text-muted-foreground mb-6">
                Thank you for registering! We've sent a confirmation email to your inbox.
                Please check your email and click the verification link to activate your account.
              </p>
              <div className="space-y-3">
                <Link
                  href="/login"
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary"
                >
                  Go to Login
                </Link>
                <Link
                  href="/"
                  className="w-full flex justify-center py-2 px-4 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-card hover:bg-muted focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary"
                >
                  Back to Home
                </Link>
              </div>
            </div>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-[80vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="flex justify-center">
            <div className="flex items-center space-x-2">
              <LinkIcon className="h-10 w-10 text-primary" />
              <span className="text-2xl font-bold text-foreground">Trunc</span>
            </div>
          </div>
          <h1 className="mt-6 text-center text-3xl font-bold text-foreground">
            Create your account
          </h1>
          <p className="mt-2 text-center text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link
              href="/login"
              className="font-medium text-primary hover:text-primary/80"
            >
              Sign in here
            </Link>
          </p>
        </div>

        <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-card py-8 px-4 border border-border sm:rounded-lg sm:px-10">
            {ageRequiredProvider && (
              <div className="mb-6 p-3 bg-muted border border-border rounded-md">
                <p className="text-sm text-foreground">
                  To create a new account with {ageRequiredProvider}, enter your date of birth below and
                  choose {ageRequiredProvider} under &quot;Or sign up with&quot;.
                </p>
              </div>
            )}
            <form className="space-y-6" onSubmit={handleSubmit(onSubmit)}>
              {/* Name Field */}
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-foreground">
                  Full name
                </label>
                <div className="mt-1">
                  <input
                    {...register('name')}
                    id="name"
                    type="text"
                    autoComplete="name"
                    className={clsx(
                      'w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground bg-background',
                      errors.name
                        ? 'border-destructive focus:ring-destructive focus:border-destructive'
                        : 'border-border'
                    )}
                    placeholder="John Doe"
                  />
                  {errors.name && (
                    <p className="mt-1 text-sm text-destructive">{errors.name.message}</p>
                  )}
                </div>
              </div>

              {/* Email Field */}
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-foreground">
                  Email address
                </label>
                <div className="mt-1">
                  <input
                    {...register('email')}
                    id="email"
                    type="email"
                    autoComplete="email"
                    data-cs-mask
                    className={clsx(
                      'w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground bg-background',
                      errors.email
                        ? 'border-destructive focus:ring-destructive focus:border-destructive'
                        : 'border-border'
                    )}
                    placeholder="john.doe@example.com"
                  />
                  {errors.email && (
                    <p className="mt-1 text-sm text-destructive">{errors.email.message}</p>
                  )}
                </div>
              </div>

              {/* Phone Field */}
              <div>
                <label htmlFor="phone" className="block text-sm font-medium text-foreground">
                  Phone number
                </label>
                <div className="mt-1">
                  <input
                    {...register('phone')}
                    id="phone"
                    type="tel"
                    autoComplete="tel"
                    data-cs-mask
                    className={clsx(
                      'w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground bg-background',
                      errors.phone
                        ? 'border-destructive focus:ring-destructive focus:border-destructive'
                        : 'border-border'
                    )}
                    placeholder="+1 (555) 123-4567"
                  />
                  {errors.phone && (
                    <p className="mt-1 text-sm text-destructive">{errors.phone.message}</p>
                  )}
                </div>
              </div>

              {/* Date of Birth Field (age gate) */}
              <div>
                <label htmlFor="dateOfBirth" className="block text-sm font-medium text-foreground">
                  Date of birth
                </label>
                <div className="mt-1">
                  <input
                    {...register('dateOfBirth')}
                    id="dateOfBirth"
                    type="date"
                    autoComplete="bday"
                    max={todayISO()}
                    data-cs-mask
                    className={clsx(
                      'w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground bg-background',
                      errors.dateOfBirth
                        ? 'border-destructive focus:ring-destructive focus:border-destructive'
                        : 'border-border'
                    )}
                  />
                  {errors.dateOfBirth && (
                    <p className="mt-1 text-sm text-destructive">{errors.dateOfBirth.message}</p>
                  )}
                </div>
              </div>

              {/* Password Field */}
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-foreground">
                  Password
                </label>
                <div className="mt-1 relative">
                  <input
                    {...register('password')}
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    autoComplete="new-password"
                    data-cs-mask
                    className={clsx(
                      'w-full px-3 py-2 pr-10 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary text-foreground bg-background',
                      errors.password
                        ? 'border-destructive focus:ring-destructive focus:border-destructive'
                        : 'border-border'
                    )}
                    placeholder="Create a strong password"
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
                    ) : (
                      <EyeIcon className="h-5 w-5 text-muted-foreground" />
                    )}
                  </button>
                </div>

                {/* Password Strength Indicator */}
                {password && (
                  <div className="mt-2">
                    <div className="flex justify-between items-center mb-1">
                      <span className="text-xs text-muted-foreground">Password strength:</span>
                      <span className={`text-xs ${passwordStrength.color}`}>
                        {passwordStrength.text}
                      </span>
                    </div>
                    <div className="w-full bg-muted rounded-full h-1">
                      <div
                        className={`h-1 rounded-full transition-all duration-300 ${passwordStrength.strength === 0 ? 'bg-muted' :
                            passwordStrength.strength === 1 ? 'bg-destructive w-1/5' :
                              passwordStrength.strength === 2 ? 'bg-destructive/80 w-2/5' :
                                passwordStrength.strength === 3 ? 'bg-yellow-500 w-3/5' :
                                  passwordStrength.strength === 4 ? 'bg-primary w-4/5' :
                                    'bg-green-600 dark:bg-green-500 w-full'
                          }`}
                      />
                    </div>
                  </div>
                )}

                {errors.password && (
                  <p className="mt-1 text-sm text-destructive">{errors.password.message}</p>
                )}
              </div>

              {/* Confirm Password Field */}
              <div>
                <label htmlFor="confirmPassword" className="block text-sm font-medium text-black">
                  Confirm password
                </label>
                <div className="mt-1 relative">
                  <input
                    {...register('confirmPassword')}
                    id="confirmPassword"
                    type={showConfirmPassword ? 'text' : 'password'}
                    autoComplete="new-password"
                    data-cs-mask
                    className={clsx(
                      'w-full px-3 py-2 pr-10 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 text-black',
                      errors.confirmPassword
                        ? 'border-red-300 focus:ring-red-500 focus:border-red-500'
                        : 'border-gray-300'
                    )}
                    placeholder="Confirm your password"
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  >
                    {showConfirmPassword ? (
                      <EyeSlashIcon className="h-5 w-5 text-gray-400" />
                    ) : (
                      <EyeIcon className="h-5 w-5 text-gray-400" />
                    )}
                  </button>
                  {errors.confirmPassword && (
                    <p className="mt-1 text-sm text-red-600">{errors.confirmPassword.message}</p>
                  )}
                </div>
              </div>

              {/* Terms and Conditions */}
              <div className="space-y-4">
                <div className="flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      {...register('acceptTerms')}
                      id="acceptTerms"
                      type="checkbox"
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label htmlFor="acceptTerms" className="text-black">
                      I agree to the{' '}
                      <Link href="/terms" className="text-primary-600 hover:text-primary-500">
                        Terms and Conditions
                      </Link>{' '}
                      and{' '}
                      <Link href="/privacy" className="text-primary-600 hover:text-primary-500">
                        Privacy Policy
                      </Link>
                    </label>
                    {errors.acceptTerms && (
                      <p className="mt-1 text-red-600">{errors.acceptTerms.message}</p>
                    )}
                  </div>
                </div>

                {/* Marketing Consent */}
                <div className="flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      {...register('marketingConsent')}
                      id="marketingConsent"
                      type="checkbox"
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label htmlFor="marketingConsent" className="text-black">
                      I would like to receive marketing emails and product updates (optional)
                    </label>
                  </div>
                </div>
              </div>

              {/* Error Message */}
              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                  <p className="text-sm text-red-600">{error}</p>
                </div>
              )}

              {/* Submit Button */}
              <div>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? 'Creating account...' : 'Create account'}
                </button>
              </div>
            </form>

            {/* Social sign-up (needs the date of birth above) */}
            <div className="mt-6">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-border" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-2 bg-card text-muted-foreground">Or sign up with</span>
                </div>
              </div>

              <div className="mt-6 grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => continueWithProvider('google')}
                  className="w-full inline-flex justify-center py-2 px-4 border border-border rounded-md shadow-sm bg-background text-sm font-medium text-foreground hover:bg-muted"
                >
                  <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
                    <path
                      fill="currentColor"
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                    />
                    <path
                      fill="currentColor"
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                    />
                    <path
                      fill="currentColor"
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                    />
                    <path
                      fill="currentColor"
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                    />
                  </svg>
                  <span className="ml-2">Google</span>
                </button>

                <button
                  type="button"
                  onClick={() => continueWithProvider('github')}
                  className="w-full inline-flex justify-center py-2 px-4 border border-border rounded-md shadow-sm bg-background text-sm font-medium text-foreground hover:bg-muted"
                >
                  <svg className="h-5 w-5" viewBox="0 0 98 96" fill="currentColor" aria-hidden="true">
                    <path
                      fillRule="evenodd"
                      clipRule="evenodd"
                      d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z"
                    />
                  </svg>
                  <span className="ml-2">GitHub</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}