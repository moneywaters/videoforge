'use client';

import PageContainer from '@/components/layout/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function TeamPage() {
  return (
    <PageContainer
      pageTitle='Team Management'
      pageDescription='Manage your workspace team, members, roles, security and more.'
    >
      <Card className='max-w-md'>
        <CardHeader>
          <CardTitle>Team Management Coming Soon</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='text-muted-foreground'>
            Team management features are coming soon. Invite members and collaborate!
          </p>
          <Button className='mt-4' onClick={() => window.history.back()}>
            Go Back
          </Button>
        </CardContent>
      </Card>
    </PageContainer>
  );
}