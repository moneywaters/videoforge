'use client';

import PageContainer from '@/components/layout/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function BillingPage() {
  return (
    <PageContainer pageTitle='Billing & Plans' pageDescription='Manage your subscription and billing'>
      <Card className='max-w-md'>
        <CardHeader>
          <CardTitle>Billing Coming Soon</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='text-muted-foreground'>
            Billing and subscription management is coming soon. Stay tuned for updates!
          </p>
          <Button className='mt-4' onClick={() => window.history.back()}>
            Go Back
          </Button>
        </CardContent>
      </Card>
    </PageContainer>
  );
}