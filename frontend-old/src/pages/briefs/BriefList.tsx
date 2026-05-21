import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardBody, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Tabs } from '@/components/ui/Tabs';
import type { Brief, BriefStatus } from '@/types/index';

const statusBadgeVariant = (status: BriefStatus) => {
  switch (status) {
    case 'published': return 'success';
    case 'closed': return 'danger';
    case 'draft': return 'warning';
    default: return 'default';
  }
};

const statusLabels: Record<BriefStatus, string> = {
  draft: 'Draft',
  published: 'Open',
  closed: 'Closed',
};

function BriefCard({ brief }: { brief: Brief }) {
  const navigate = useNavigate();
  const statusVariant = statusBadgeVariant(brief.status);
  const statusLabel = statusLabels[brief.status];
  const deadlineDate = brief.deadline ? new Date(brief.deadline).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'No deadline';

  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate(`/briefs/${brief.id}`)}>
      <CardHeader>
        <div className="flex justify-between items-start">
          <CardTitle className="text-lg">{brief.title}</CardTitle>
          <Badge variant={statusVariant}>{statusLabel}</Badge>
        </div>
      </CardHeader>
      <CardBody className="space-y-3">
        <div className="flex items-center justify-between text-sm">
          <span className="text-gray-600">Bounty</span>
          <span className="font-semibold text-gray-900">${brief.bountyBudget.toLocaleString()}</span>
        </div>
        <div className="flex items-center justify-between text-sm">
          <span className="text-gray-600">Submissions</span>
          <span className="font-medium text-gray-900">{brief.currentSubmissions}/{brief.submissionLimit}</span>
        </div>
        <div className="flex flex-wrap gap-1">
          {brief.tags.slice(0, 3).map((tag) => (
            <Badge key={tag} variant="info" className="text-xs">{tag}</Badge>
          ))}
        </div>
      </CardBody>
      <CardFooter className="text-sm text-gray-500">
        Deadline: {deadlineDate}
      </CardFooter>
    </Card>
  );
}

function BriefCardSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-6 w-3/4" />
      </CardHeader>
      <CardBody className="space-y-3">
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-4 w-1/3" />
        <div className="flex gap-1">
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-5 w-16" />
        </div>
      </CardBody>
      <CardFooter>
        <Skeleton className="h-4 w-1/2" />
      </CardFooter>
    </Card>
  );
}

export function BriefList() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('all');

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
    { id: 'open', label: `Open (${briefs?.filter((b) => b.status === 'published').length ?? 0})` },
    { id: 'closed', label: `Closed (${briefs?.filter((b) => b.status === 'closed').length ?? 0})` },
    { id: 'draft', label: `Draft (${briefs?.filter((b) => b.status === 'draft').length ?? 0})` },
  ];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Briefs</h1>
        <Button onClick={() => navigate('/briefs/new')}>Create Brief</Button>
      </div>

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab}>
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
              <BriefCardSkeleton key={i} />
            ))}
          </div>
        ) : filteredBriefs.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            No briefs found. Create your first brief to get started.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredBriefs.map((brief) => (
              <BriefCard key={brief.id} brief={brief} />
            ))}
          </div>
        )}
      </Tabs>
    </div>
  );
}