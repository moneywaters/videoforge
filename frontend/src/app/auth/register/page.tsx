'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import zxcvbn from 'zxcvbn';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { UserRole } from '@/types';

import { useAuthStore } from '@/stores/auth-store';
import { api } from '@/lib/api';

export default function RegisterPage() {
  const router = useRouter();
  const setUser = useAuthStore((state) => state.setUser);
  const { loginWithGoogle } = useAuthStore();
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [confirmPassword, setConfirmPassword] = useState('');
  const [role, setRole] = useState<UserRole>('client');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const passwordStrengthLabels: Record<number, string> = {
    0: 'Very Weak',
    1: 'Very Weak',
    2: 'Weak',
    3: 'Fair',
    4: 'Good',
    5: 'Strong',
  };
  const strengthColors: Record<number, string> = {
    0: 'bg-destructive',
    1: 'bg-destructive',
    2: 'bg-orange-500',
    3: 'bg-yellow-500',
    4: 'bg-green-500',
    5: 'bg-green-500',
  };

  const handleGoogleSignup = () => {
    loginWithGoogle();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    if (firstName.trim() === '' || lastName.trim() === '') {
      setError('First name and last name are required');
      return;
    }

    setIsLoading(true);

    try {
      await api.register(email, password, firstName, lastName, role);
      // Log the user in immediately so we have a session
      const loginData = await api.login(email, password);
      setUser(loginData);
      router.push('/onboarding');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className='min-h-screen bg-muted flex items-center justify-center p-4'>
      <Card className='w-full max-w-md'>
        <CardHeader className='text-center'>
          <CardTitle className='text-2xl'>Create Account</CardTitle>
          <CardDescription>Join VideoForge and start creating</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            type='button'
            variant='outline'
            onClick={handleGoogleSignup}
            disabled={isLoading}
            className='w-full'
          >
            <Icons.logo className='size-5' />
            <span>Sign up with Google</span>
          </Button>
          <div className='relative py-2'>
            <div className='absolute inset-0 flex items-center'>
              <span className='w-full border-t' />
            </div>
            <div className='relative flex justify-center text-sm'>
              <span className='bg-card px-2 text-muted-foreground'>or sign up with email</span>
            </div>
          </div>
          <form onSubmit={handleSubmit} className='space-y-4'>
            {error && (
              <div className='p-3 text-sm text-destructive bg-destructive/10 rounded-md'>
                {error}
              </div>
            )}
            <div className='space-y-2'>
              <Label htmlFor='firstName'>First Name</Label>
              <Input
                id='firstName'
                type='text'
                placeholder='John'
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='lastName'>Last Name</Label>
              <Input
                id='lastName'
                type='text'
                placeholder='Doe'
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
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
                placeholder='Create a password'
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  setPasswordStrength(zxcvbn(e.target.value).score);
                }}
                required
                disabled={isLoading}
              />
              {password && (
                <div className='space-y-1'>
                  <div className='w-full h-2 rounded-full bg-muted overflow-hidden'>
                    <div
                      className={`h-full ${strengthColors[passwordStrength]} transition-all duration-300`}
                      style={{ width: `${(passwordStrength / 5) * 100}%` }}
                    />
                  </div>
                  <p className='text-sm text-muted-foreground'>
                    Strength: {passwordStrengthLabels[passwordStrength]}
                  </p>
                </div>
              )}
            </div>
            <div className='space-y-2'>
              <Label htmlFor='role'>Role</Label>
              <Select value={role} onValueChange={(value) => setRole(value as UserRole)}>
                <SelectTrigger id='role'>
                  <SelectValue placeholder='Select a role' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='client'>Client (Brand)</SelectItem>
                  <SelectItem value='editor'>Editor / Creator</SelectItem>
                  <SelectItem value='ad_specialist'>Ad Specialist</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-2'>
              <Label htmlFor='confirmPassword'>Confirm Password</Label>
              <Input
                id='confirmPassword'
                type='password'
                placeholder='Confirm your password'
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <Button type='submit' className='w-full' disabled={isLoading} isLoading={isLoading}>
              {isLoading ? 'Creating account...' : 'Create Account'}
            </Button>
            <p className='text-center text-sm text-muted-foreground'>
              Already have an account?{' '}
              <Link href='/auth/login' className='text-primary hover:text-primary/80 font-medium'>
                Sign in
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}