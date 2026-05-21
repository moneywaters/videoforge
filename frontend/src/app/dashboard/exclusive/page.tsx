'use client';

import PageContainer from '@/components/layout/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function ExclusivePage() {
  return (
    <PageContainer pageTitle='Exclusive Features' pageDescription='Premium features for Pro members'>
      <Card className='max-w-md'>
        <CardHeader>
          <CardTitle>Premium Features Coming Soon</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='text-muted-foreground'>
            Exclusive premium features are coming soon. Upgrade your plan to unlock special
            capabilities!
          </p>
          <Button className='mt-4' onClick={() => window.history.back()}>
            Go Back
          </Button>
        </CardContent>
      </Card>
    </PageContainer>
  );
}