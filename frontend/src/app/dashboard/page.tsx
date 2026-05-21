'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function DashboardPage() {
  const router = useRouter();

  useEffect(() => {
    // Client-side redirect to briefs page
    router.push('/dashboard/briefs');
  }, [router]);

  return (
    <div className='flex min-h-screen items-center justify-center bg-background'>
      <div className='text-center'>
        <div className='mx-auto h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary border-r-transparent'></div>
        <p className='mt-4 text-muted-foreground'>Loading dashboard...</p>
      </div>
    </div>
  );
}