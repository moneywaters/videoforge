'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import type { Brief, BriefStatus } from '@/types';

const statusBadgeVariant = (
  status: BriefStatus
): 'default' | 'secondary' | 'destructive' | 'outline' => {
  switch (status) {
    case 'published':
      return 'default';
    case 'closed':
      return 'destructive';
    case 'draft':
      return 'secondary';
    default:
      return 'secondary';
  }
};

const statusLabels: Record<BriefStatus, string> = {
  draft: 'Draft',
  published: 'Open',
  closed: 'Closed',
};

interface BriefCardProps {
  brief: Brief;
}

function BriefCard({ brief }: BriefCardProps) {
  const statusVariant = statusBadgeVariant(brief.status);
  const statusLabel = statusLabels[brief.status];
  const deadlineDate = brief.deadline
    ? new Date(brief.deadline).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      })
    : 'No deadline';

  return (
    <Link href={`/dashboard/briefs/${brief.id}`}>
      <Card className='hover:shadow-md transition-shadow cursor-pointer h-full'>
        <CardHeader>
          <div className='flex justify-between items-start gap-2'>
            <CardTitle className='text-lg line-clamp-2'>{brief.title}</CardTitle>
            <Badge variant={statusVariant} className='shrink-0'>
              {statusLabel}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className='space-y-3'>
          <div className='flex items-center justify-between text-sm'>
            <span className='text-muted-foreground'>Bounty</span>
            <span className='font-semibold'>
              ${brief.bountyBudget.toLocaleString()}
            </span>
          </div>
          <div className='flex items-center justify-between text-sm'>
            <span className='text-muted-foreground'>Submissions</span>
            <span className='font-medium'>
              {brief.currentSubmissions}/{brief.submissionLimit}
            </span>
          </div>
          <div className='flex flex-wrap gap-1'>
            {brief.tags.slice(0, 3).map((tag) => (
              <Badge key={tag} variant='outline' className='text-xs'>
                {tag}
              </Badge>
            ))}
          </div>
        </CardContent>
        <CardFooter className='text-sm text-muted-foreground'>
          <Icons.calendar className='mr-2 h-4 w-4' />
          Deadline: {deadlineDate}
        </CardFooter>
      </Card>
    </Link>
  );
}

function BriefCardSkeleton() {
  return (
    <Card className='h-full'>
      <CardHeader>
        <Skeleton className='h-6 w-3/4' />
      </CardHeader>
      <CardContent className='space-y-3'>
        <Skeleton className='h-4 w-1/2' />
        <Skeleton className='h-4 w-1/3' />
        <div className='flex gap-1'>
          <Skeleton className='h-5 w-16' />
          <Skeleton className='h-5 w-16' />
        </div>
      </CardContent>
      <CardFooter>
        <Skeleton className='h-4 w-1/2' />
      </CardFooter>
    </Card>
  );
}

interface TabButtonProps {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}

function TabButton({ active, onClick, children }: TabButtonProps) {
  return (
    <button
      type='button'
      onClick={onClick}
      className={`inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 ${
        active
          ? 'bg-primary text-primary-foreground hover:bg-primary/90'
          : 'hover:bg-accent hover:text-accent-foreground'
      } h-10 px-4 py-2`}
    >
      {children}
    </button>
  );
}

export default function BriefsPage() {
  const [activeTab, setActiveTab] = useState<'all' | 'open' | 'closed' | 'draft'>(
    'all'
  );

  const { data: briefs, isLoading } = useQuery({
    queryKey: ['briefs'],
    queryFn: () => api.getBriefs(),
  });

  const filteredBriefs = briefs?.filter((brief) => {
    if (activeTab === 'all') return true;
    if (activeTab === 'open') return brief.status === 'published';
    if (activeTab === 'closed') return brief.status === 'closed';
    if (activeTab === 'draft') return brief.status === 'draft';
    return true;
  }) ?? [];

  const tabs = [
    { id: 'all', label: `All (${briefs?.length ?? 0})` },
    {
      id: 'open',
      label: `Open (${briefs?.filter((b) => b.status === 'published').length ?? 0})`,
    },
    {
      id: 'closed',
      label: `Closed (${briefs?.filter((b) => b.status === 'closed').length ?? 0})`,
    },
    {
      id: 'draft',
      label: `Draft (${briefs?.filter((b) => b.status === 'draft').length ?? 0})`,
    },
  ];

  return (
    <div className='space-y-6'>
      <div className='flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4'>
        <div>
          <h1 className='text-2xl font-bold'>Briefs</h1>
          <p className='text-muted-foreground'>
            Manage your video briefs and campaigns
          </p>
        </div>
        <Link href='/dashboard/briefs/new'>
          <Button>
            <Icons.add className='mr-2 h-4 w-4' />
            Create Brief
          </Button>
        </Link>
      </div>

      <div className='flex gap-2 border-b'>
        {tabs.map((tab) => (
          <TabButton
            key={tab.id}
            active={activeTab === tab.id}
            onClick={() =>
              setActiveTab(tab.id as 'all' | 'open' | 'closed' | 'draft')
            }
          >
            {tab.label}
          </TabButton>
        ))}
      </div>

      {isLoading ? (
        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
          {[...Array(6)].map((_, i) => (
            <BriefCardSkeleton key={i} />
          ))}
        </div>
      ) : filteredBriefs.length === 0 ? (
        <div className='text-center py-12'>
          <Icons.post className='mx-auto h-12 w-12 text-muted-foreground' />
          <h3 className='mt-4 text-lg font-semibold'>No briefs found</h3>
          <p className='mt-2 text-muted-foreground'>
            Create your first brief to get started.
          </p>
          <Link href='/dashboard/briefs/new'>
            <Button className='mt-4'>
              <Icons.add className='mr-2 h-4 w-4' />
              Create Brief
            </Button>
          </Link>
        </div>
      ) : (
        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
          {filteredBriefs.map((brief) => (
            <BriefCard key={brief.id} brief={brief} />
          ))}
        </div>
      )}
    </div>
  );
}