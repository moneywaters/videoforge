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

const STATUS_FILTERS = [
  { id: 'all', label: 'All' },
  { id: 'pending', label: 'Pending' },
  { id: 'processing', label: 'Processing' },
  { id: 'completed', label: 'Completed' },
  { id: 'failed', label: 'Failed' },
];

export default function PayoutHistoryPage() {
  const [activeStatus, setActiveStatus] = useState('all');

  const { data: payouts, isLoading } = useQuery({
    queryKey: ['payouts'],
    queryFn: () => api.getPayouts(''),
  });

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'completed':
        return 'default';
      case 'processing':
        return 'secondary';
      case 'pending':
        return 'secondary';
      case 'failed':
        return 'destructive';
      default:
        return 'default';
    }
  };

  const filteredPayouts = payouts?.filter(
    (p) => activeStatus === 'all' || p.status === activeStatus
  ) || [];

  const handleProcessPayout = async (payoutId: string) => {
    console.log('Processing payout:', payoutId);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Payout History</h1>
        <p className="text-muted-foreground mt-1">Manage and process creator payouts</p>
      </div>

      <Tabs value={activeStatus} onValueChange={setActiveStatus}>
        <TabsList>
          {STATUS_FILTERS.map((filter) => (
            <TabsTrigger key={filter.id} value={filter.id}>
              {filter.label}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <Card>
        <CardHeader>
          <CardTitle>Payouts</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {filteredPayouts.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Fee (5%)</TableHead>
                  <TableHead>Net</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPayouts.map((payout) => (
                  <TableRow key={payout.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                          <Icons.user className="h-4 w-4 text-muted-foreground" />
                        </div>
                        <div>
                          <p className="font-medium">{payout.userName}</p>
                          <p className="text-sm text-muted-foreground">{payout.userId}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="font-medium">
                      {formatCurrency(payout.amount)}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatCurrency(payout.fee)}
                    </TableCell>
                    <TableCell className="font-semibold">
                      {formatCurrency(payout.netAmount)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusBadgeVariant(payout.status) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                        {payout.status.charAt(0).toUpperCase() +
                          payout.status.slice(1)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatDate(payout.createdAt)}
                    </TableCell>
                    <TableCell>
                      {payout.status === 'pending' && (
                        <Button
                          size="sm"
                          onClick={() => handleProcessPayout(payout.id)}
                        >
                          Process
                        </Button>
                      )}
                      {payout.status === 'processing' && (
                        <Button size="sm" variant="secondary" disabled>
                          <Icons.spinner className="mr-1 h-4 w-4 animate-spin" />
                          Processing...
                        </Button>
                      )}
                      {payout.status === 'completed' && (
                        <div className="flex items-center text-emerald-600">
                          <Icons.check className="mr-1 h-4 w-4" />
                          <span className="text-sm">Done</span>
                        </div>
                      )}
                      {payout.status === 'failed' && (
                        <div className="flex items-center text-rose-600">
                          <Icons.circleX className="mr-1 h-4 w-4" />
                          <span className="text-sm">Failed</span>
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-12">
              <Icons.creditCard className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No payouts found</p>
              <p className="text-sm text-muted-foreground/60">
                {activeStatus === 'all'
                  ? 'All payouts will appear here'
                  : `No ${activeStatus} payouts`}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}