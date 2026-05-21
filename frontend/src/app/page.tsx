'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function Page() {
  const router = useRouter();

  useEffect(() => {
    // Client-side redirect to auth page
    router.push('/auth/login');
  }, [router]);

  return (
    <div className='flex min-h-screen items-center justify-center bg-background'>
      <div className='text-center'>
        <div className='mx-auto h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary border-r-transparent'></div>
        <p className='mt-4 text-muted-foreground'>Loading...</p>
      </div>
    </div>
  );
}