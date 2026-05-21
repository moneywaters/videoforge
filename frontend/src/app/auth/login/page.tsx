'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuthStore } from '@/stores/auth-store';
import { api } from '@/lib/api';

export default function LoginPage() {
  const router = useRouter();
  const { loginWithGoogle } = useAuthStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleGoogleLogin = () => {
    loginWithGoogle();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await api.login(email, password);
      router.push('/dashboard');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className='min-h-screen bg-muted flex items-center justify-center p-4'>
      <Card className='w-full max-w-md'>
        <CardHeader className='text-center'>
          <CardTitle className='text-2xl'>Welcome Back</CardTitle>
          <CardDescription>Sign in to your VideoForge account</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className='space-y-4'>
            {error && (
              <div className='p-3 text-sm text-destructive bg-destructive/10 rounded-md'>
                {error}
              </div>
            )}
            <div className='space-y-2'>
              <Label htmlFor='email'>Email</Label>
              <Input
                id='email'
                type='email'
                placeholder='you@example.com'
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='password'>Password</Label>
              <Input
                id='password'
                type='password'
                placeholder='Enter your password'
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <Button type='submit' className='w-full' disabled={isLoading} isLoading={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign In'}
            </Button>
            <div className='relative py-2'>
              <div className='absolute inset-0 flex items-center'>
                <span className='w-full border-t' />
              </div>
              <div className='relative flex justify-center text-sm'>
                <span className='bg-card px-2 text-muted-foreground'>or</span>
              </div>
            </div>
            <Button
              type='button'
              variant='outline'
              onClick={handleGoogleLogin}
              disabled={isLoading}
              className='w-full'
            >
              <Icons.logo className='size-5' />
              <span>Sign in with Google</span>
            </Button>
            <p className='text-center text-sm text-muted-foreground'>
              Don&apos;t have an account?{' '}
              <Link href='/auth/register' className='text-primary hover:text-primary/80 font-medium'>
                Create account
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}