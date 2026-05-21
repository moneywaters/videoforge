'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/stores/auth-store';

export default function GoogleCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const setLoading = useAuthStore((state) => state.setLoading);
  const setUser = useAuthStore((state) => state.setUser);
  const { handleGoogleCallback } = useAuthStore();

  useEffect(() => {
    // Read token from URL query parameter
    const token = searchParams.get('token');

    if (!token) {
      setError('No token provided. Please try logging in again.');
      return;
    }

    // Use the handleGoogleCallback from auth store which handles token decoding and user setting
    handleGoogleCallback(token);

    // Navigate to dashboard after handling the callback
    router.push('/dashboard');
  }, [searchParams, router, setLoading, setUser, handleGoogleCallback]);

  if (error) {
    return (
      <div className='min-h-screen bg-muted flex items-center justify-center p-4'>
        <div className='bg-card p-8 rounded-lg shadow-md max-w-md text-center'>
          <div className='text-destructive text-lg font-medium mb-4'>Authentication Error</div>
          <p className='text-muted-foreground mb-4'>{error}</p>
          <Button asChild>
            <a href='/auth/login'>Back to Login</a>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className='min-h-screen bg-muted flex items-center justify-center p-4'>
      <div className='text-center'>
        <Icons.spinner className='size-8 animate-spin text-primary mx-auto mb-4' />
        <p className='text-muted-foreground'>Completing authentication...</p>
      </div>
    </div>
  );
}