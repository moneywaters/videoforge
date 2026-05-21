import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Card, CardBody, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Tabs } from '@/components/ui/Tabs';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/Table';
import { DollarSign, Users, Clock, CheckCircle, XCircle } from 'lucide-react';

const STATUS_FILTERS = [
  { id: 'all', label: 'All' },
  { id: 'pending', label: 'Pending' },
  { id: 'processing', label: 'Processing' },
  { id: 'completed', label: 'Completed' },
  { id: 'failed', label: 'Failed' },
];

export default function PayoutHistory() {
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
        return 'success';
      case 'processing':
        return 'info';
      case 'pending':
        return 'warning';
      case 'failed':
        return 'danger';
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
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-48"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Payout History</h1>
        <p className="text-gray-500 mt-2">Manage and process creator payouts</p>
      </div>

      <Tabs
        tabs={STATUS_FILTERS}
        activeTab={activeStatus}
        onChange={setActiveStatus}
      >
        <div />
      </Tabs>

      <Card>
        <CardHeader>
          <CardTitle>Payouts</CardTitle>
        </CardHeader>
        <CardBody className="p-0">
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
                        <div className="h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center">
                          <Users className="h-4 w-4 text-gray-400" />
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{payout.userName}</p>
                          <p className="text-sm text-gray-500">{payout.userId}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="font-medium">
                      {formatCurrency(payout.amount)}
                    </TableCell>
                    <TableCell className="text-gray-500">
                      {formatCurrency(payout.fee)}
                    </TableCell>
                    <TableCell className="font-semibold text-gray-900">
                      {formatCurrency(payout.netAmount)}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          getStatusBadgeVariant(payout.status) as
                            | 'default'
                            | 'success'
                            | 'warning'
                            | 'danger'
                            | 'info'
                        }
                      >
                        {payout.status.charAt(0).toUpperCase() +
                          payout.status.slice(1)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-gray-500">
                      {formatDate(payout.createdAt)}
                    </TableCell>
                    <TableCell>
                      {payout.status === 'pending' && (
                        <Button
                          size="sm"
                          variant="primary"
                          onClick={() => handleProcessPayout(payout.id)}
                        >
                          Process
                        </Button>
                      )}
                      {payout.status === 'processing' && (
                        <Button size="sm" variant="secondary" disabled>
                          <Clock className="h-4 w-4 mr-1" />
                          Processing...
                        </Button>
                      )}
                      {payout.status === 'completed' && (
                        <div className="flex items-center text-emerald-600">
                          <CheckCircle className="h-4 w-4 mr-1" />
                          <span className="text-sm">Done</span>
                        </div>
                      )}
                      {payout.status === 'failed' && (
                        <div className="flex items-center text-rose-600">
                          <XCircle className="h-4 w-4 mr-1" />
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
              <DollarSign className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500">No payouts found</p>
              <p className="text-sm text-gray-400">
                {activeStatus === 'all'
                  ? 'All payouts will appear here'
                  : `No ${activeStatus} payouts`}
              </p>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
}