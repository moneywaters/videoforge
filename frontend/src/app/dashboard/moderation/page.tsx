'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

const STATUS_TABS = [
  { id: 'pending', label: 'Pending' },
  { id: 'approved', label: 'Approved' },
  { id: 'rejected', label: 'Rejected' },
];

const TYPE_ICONS: Record<string, keyof typeof Icons> = {
  video: 'video',
  brief: 'page',
  comment: 'chat',
};

export default function ModerationPage() {
  const [activeTab, setActiveTab] = useState('pending');

  const { data: moderationItems, isLoading } = useQuery({
    queryKey: ['moderation'],
    queryFn: () => api.getModerationQueue(),
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
      case 'approved':
        return 'default';
      case 'rejected':
        return 'destructive';
      case 'pending':
        return 'secondary';
      default:
        return 'default';
    }
  };

  const getTypeBadgeVariant = (type: string) => {
    switch (type) {
      case 'video':
        return 'secondary';
      case 'brief':
        return 'secondary';
      case 'comment':
        return 'default';
      default:
        return 'default';
    }
  };

  const filteredItems = moderationItems?.filter(
    (item) => activeTab === 'pending' || item.status === activeTab
  );

  const handleApprove = (itemId: string) => {
    console.log('Approving:', itemId);
  };

  const handleReject = (itemId: string) => {
    console.log('Rejecting:', itemId);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="h-12 bg-muted rounded"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Moderation Queue</h1>
        <p className="text-muted-foreground mt-1">
          Review flagged content and user reports
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

      <Card>
        <CardHeader>
          <CardTitle>
            {activeTab === 'pending' && 'Pending Reviews'}
            {activeTab === 'approved' && 'Approved Items'}
            {activeTab === 'rejected' && 'Rejected Items'}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {filteredItems && filteredItems.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Content ID</TableHead>
                  <TableHead>Flagged By</TableHead>
                  <TableHead>Reason</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredItems.map((item) => {
                  const TypeIcon = TYPE_ICONS[item.type] ? Icons[TYPE_ICONS[item.type]] : Icons.page;
                  return (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <TypeIcon className="h-5 w-5 text-muted-foreground" />
                          <Badge variant={getTypeBadgeVariant(item.type) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                            {item.type}
                          </Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="font-mono text-sm">{item.contentId}</span>
                      </TableCell>
                      <TableCell>{item.flaggedBy}</TableCell>
                      <TableCell className="max-w-xs">
                        <p className="truncate">{item.reason}</p>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDate(item.createdAt)}
                      </TableCell>
                      <TableCell>
                        <Badge variant={getStatusBadgeVariant(item.status) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                          {item.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {item.status === 'pending' && (
                          <div className="flex items-center gap-2">
                            <Button
                              size="sm"
                              onClick={() => handleApprove(item.id)}
                            >
                              <Icons.check className="h-4 w-4 mr-1" />
                              Approve
                            </Button>
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={() => handleReject(item.id)}
                            >
                              <Icons.circleX className="h-4 w-4 mr-1" />
                              Reject
                            </Button>
                          </div>
                        )}
                        {item.status === 'approved' && (
                          <div className="flex items-center text-emerald-600">
                            <Icons.check className="h-4 w-4 mr-1" />
                            <span className="text-sm">Approved</span>
                          </div>
                        )}
                        {item.status === 'rejected' && (
                          <div className="flex items-center text-rose-600">
                            <Icons.circleX className="h-4 w-4 mr-1" />
                            <span className="text-sm">Rejected</span>
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-12">
              <Icons.check className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No items found</p>
              <p className="text-sm text-muted-foreground/60">
                {activeTab === 'pending'
                  ? 'All caught up! No pending reviews.'
                  : `No ${activeTab} items`}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}