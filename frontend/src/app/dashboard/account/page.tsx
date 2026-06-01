'use client';

import { useState } from 'react';
import { useAuthStore } from '@/stores/auth-store';
import PageContainer from '@/components/layout/page-container';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Icons } from '@/components/icons';

export default function AccountPage() {
  const user = useAuthStore((state) => state.user);

  const [name, setName] = useState(user?.name ?? '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [marketingEmails, setMarketingEmails] = useState(false);

  const handleSaveProfile = () => {
    console.log('Saving profile:', { name });
  };

  const handleUpdatePassword = () => {
    console.log('Updating password:', { currentPassword, newPassword, confirmPassword });
  };

  const handleSavePreferences = () => {
    console.log('Saving preferences:', { emailNotifications, marketingEmails });
  };

  if (!user) {
    return (
      <PageContainer pageTitle="Account" pageDescription="Manage your account settings">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>Account</CardTitle>
            <CardDescription>You are not signed in.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => (window.location.href = '/auth/login')}>
              <Icons.login className="mr-2 h-4 w-4" />
              Sign In
            </Button>
          </CardContent>
        </Card>
      </PageContainer>
    );
  }

  return (
    <PageContainer
      pageTitle="Account"
      pageDescription="Manage your account settings and preferences"
    >
      <div className="mx-auto max-w-2xl space-y-6">
        {/* Profile Information */}
        <Card>
          <CardHeader>
            <CardTitle>Profile Information</CardTitle>
            <CardDescription>Update your personal details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Your full name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" value={user.email} readOnly disabled />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role">Role</Label>
              <div>
                <Badge variant="secondary">{user.role}</Badge>
              </div>
            </div>
            <Button onClick={handleSaveProfile}>
              <Icons.check className="mr-2 h-4 w-4" />
              Save Changes
            </Button>
          </CardContent>
        </Card>

        {/* Change Password */}
        <Card>
          <CardHeader>
            <CardTitle>Change Password</CardTitle>
            <CardDescription>Update your password to keep your account secure</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="current-password">Current Password</Label>
              <Input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Enter current password"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-password">New Password</Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">Confirm New Password</Label>
              <Input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm new password"
              />
            </div>
            <Button onClick={handleUpdatePassword}>
              <Icons.lock className="mr-2 h-4 w-4" />
              Update Password
            </Button>
          </CardContent>
        </Card>

        {/* Preferences */}
        <Card>
          <CardHeader>
            <CardTitle>Preferences</CardTitle>
            <CardDescription>Manage your notification settings</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="email-notifications">Email Notifications</Label>
                <p className="text-muted-foreground text-sm">
                  Receive updates about your projects and briefs
                </p>
              </div>
              <Switch
                id="email-notifications"
                checked={emailNotifications}
                onCheckedChange={setEmailNotifications}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="marketing-emails">Marketing Emails</Label>
                <p className="text-muted-foreground text-sm">
                  Receive tips, product updates, and promotional offers
                </p>
              </div>
              <Switch
                id="marketing-emails"
                checked={marketingEmails}
                onCheckedChange={setMarketingEmails}
              />
            </div>
            <Button onClick={handleSavePreferences}>
              <Icons.check className="mr-2 h-4 w-4" />
              Save Preferences
            </Button>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}
