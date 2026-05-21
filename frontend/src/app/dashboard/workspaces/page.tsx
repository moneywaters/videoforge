'use client';

import PageContainer from '@/components/layout/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function WorkspacesPage() {
  return (
    <PageContainer pageTitle='Workspaces' pageDescription='Manage your workspaces and switch between them'>
      <Card className='max-w-md'>
        <CardHeader>
          <CardTitle>Workspaces Coming Soon</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='text-muted-foreground'>
            Workspace management is coming soon. Create and manage your teams with ease!
          </p>
          <Button className='mt-4' onClick={() => window.history.back()}>
            Go Back
          </Button>
        </CardContent>
      </Card>
    </PageContainer>
  );
}