'use client';
import { useAuthStore } from '@/stores/auth-store';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';

export default function ProfileViewPage() {
  const user = useAuthStore((state) => state.user);

  if (!user) {
    return (
      <div className='flex w-full flex-col p-4'>
        <Card className='max-w-md'>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
            <CardDescription>You are not signed in.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => window.location.href = '/auth/login'}>
              Sign In
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className='flex w-full flex-col p-4'>
      <Card className='max-w-md'>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='name'>Name</Label>
            <Input id='name' value={user.name} readOnly />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='email'>Email</Label>
            <Input id='email' value={user.email} readOnly />
          </div>
          <div className='space-y-2'>
            <Label htmlFor='role'>Role</Label>
            <Input id='role' value={user.role} readOnly />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}