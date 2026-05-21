'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

const STATUS_TABS = [
  { id: 'open', label: 'Open' },
  { id: 'under_review', label: 'Under Review' },
  { id: 'resolved', label: 'Resolved' },
];

export default function DisputesPage() {
  const [activeTab, setActiveTab] = useState('open');

  const { data: disputes, isLoading } = useQuery({
    queryKey: ['disputes'],
    queryFn: () => api.getDisputes(),
  });

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'resolved':
        return 'default';
      case 'under_review':
        return 'secondary';
      case 'open':
        return 'destructive';
      default:
        return 'default';
    }
  };

  const filteredDisputes = disputes?.filter(
    (d) => activeTab === 'open' || d.status === activeTab
  );

  const handleMarkReviewing = (disputeId: string) => {
    console.log('Marking as reviewing:', disputeId);
  };

  const handleResolveReporter = (disputeId: string) => {
    console.log('Resolving in favor of reporter:', disputeId);
  };

  const handleResolveTarget = (disputeId: string) => {
    console.log('Resolving in favor of target:', disputeId);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="h-12 bg-muted rounded"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-48 bg-muted rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dispute Resolution</h1>
        <p className="text-muted-foreground mt-1">
          Manage and resolve user disputes
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          {STATUS_TABS.map((tab) => (
            <TabsTrigger key={tab.id} value={tab.id}>
              {tab.label}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <div className="space-y-4">
        {filteredDisputes && filteredDisputes.length > 0 ? (
          filteredDisputes.map((dispute) => (
            <Card key={dispute.id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Dispute #{dispute.id.slice(-6)}</CardTitle>
                    <CardDescription>
                      Filed on {formatDate(dispute.createdAt)}
                    </CardDescription>
                  </div>
<Badge variant={getStatusBadgeVariant(dispute.status) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                    {dispute.status.replace('_', ' ')}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-rose-100 dark:bg-rose-900 rounded-lg">
                        <Icons.user className="h-5 w-5 text-rose-600" />
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">Reporter</p>
                        <p className="font-medium">{dispute.reporterName}</p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-muted rounded-lg">
                        <Icons.user className="h-5 w-5 text-muted-foreground" />
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">Target</p>
                        <p className="font-medium">{dispute.targetName}</p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-lg">
                        <Icons.alertCircle className="h-5 w-5 text-amber-600" />
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">Reason</p>
                        <p className="font-medium">{dispute.reason}</p>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <p className="text-sm text-muted-foreground mb-2">Evidence</p>
                      <div className="space-y-2">
                        {dispute.evidence.map((link, idx) => (
                          <a
                            key={idx}
                            href={link}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-2 text-sm hover:underline"
                          >
                            <Icons.externalLink className="h-4 w-4" />
                            Evidence {idx + 1}
                          </a>
                        ))}
                      </div>
                    </div>

                    {dispute.status === 'resolved' && dispute.resolution && (
                      <div className="p-3 bg-emerald-50 dark:bg-emerald-950 rounded-lg">
                        <p className="text-sm font-medium text-emerald-600 dark:text-emerald-400">
                          Resolution
                        </p>
                        <p className="text-sm text-emerald-700 dark:text-emerald-300 mt-1">
                          {dispute.resolution}
                        </p>
                        {dispute.resolvedAt && (
                          <p className="text-xs text-emerald-500 dark:text-emerald-500 mt-2">
                            Resolved on {formatDate(dispute.resolvedAt)}
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                </div>

                {dispute.status !== 'resolved' && (
                  <div className="mt-6 pt-6 border-t">
                    <div className="flex flex-wrap gap-3">
                      {dispute.status === 'open' && (
                        <Button
                          variant="secondary"
                          onClick={() => handleMarkReviewing(dispute.id)}
                        >
                          <Icons.clock className="h-4 w-4 mr-2" />
                          Mark as Reviewing
                        </Button>
                      )}
                      {(dispute.status === 'open' ||
                        dispute.status === 'under_review') && (
                        <>
                          <Button
                            onClick={() =>
                              handleResolveReporter(dispute.id)
                            }
                          >
                            <Icons.user className="h-4 w-4 mr-2" />
                            Resolve in Favor of Reporter
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => handleResolveTarget(dispute.id)}
                          >
                            <Icons.user className="h-4 w-4 mr-2" />
                            Resolve in Favor of Target
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        ) : (
          <Card>
            <CardContent>
              <div className="text-center py-12">
                <Icons.check className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">No disputes found</p>
                <p className="text-sm text-muted-foreground/60">
                  {activeTab === 'open'
                    ? 'All disputes have been resolved!'
                    : `No ${activeTab.replace('_', ' ')} disputes`}
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}