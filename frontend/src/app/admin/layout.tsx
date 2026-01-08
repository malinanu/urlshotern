'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { adminAuthService } from '@/lib/admin-auth-service';
import { API_ENDPOINTS } from '@/lib/api-config';

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const checkAuth = async () => {
      // Skip check for login page
      if (pathname === '/admin/login') {
        setIsLoading(false);
        return;
      }

      const token = adminAuthService.getAdminToken();

      if (!token) {
        router.push('/admin/login');
        return;
      }

      // Validate token with backend
      try {
        const response = await fetch(API_ENDPOINTS.admin.dashboard, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error('Invalid token');
        }

        // Token is valid, allow access
        setIsLoading(false);
      } catch (error) {
        // Token is invalid, clear storage and redirect to login
        adminAuthService.logout();
        router.push('/admin/login');
      }
    };

    checkAuth();
  }, [pathname, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
