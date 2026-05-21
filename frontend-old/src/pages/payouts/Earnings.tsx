import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Card, CardBody, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { MetricCard } from '@/components/ui/MetricCard';
import { DollarSign, Clock, Award } from 'lucide-react';

const CURRENT_USER_ID = 'usr-editor-001';

export default function Earnings() {
  const { data: earnings, isLoading: earningsLoading } = useQuery({
    queryKey: ['earnings', CURRENT_USER_ID],
    queryFn: () => api.getEarnings(CURRENT_USER_ID),
  });

  const { data: payouts, isLoading: payoutsLoading } = useQuery({
    queryKey: ['payouts', CURRENT_USER_ID],
    queryFn: () => api.getPayouts(CURRENT_USER_ID),
  });

  if (earningsLoading || payoutsLoading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-48"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-32 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

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

  const tierProgress = earnings?.tierProgress || 0;
  const tierPercentage = Math.min(100, (tierProgress / 500) * 100);

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">My Earnings</h1>
        <p className="text-gray-500 mt-2">Track your earnings and payout history</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <MetricCard
          title="Total Earned"
          value={formatCurrency(earnings?.totalEarned || 0)}
          change={12}
          trend="up"
        />
        <MetricCard
          title="Pending Balance"
          value={formatCurrency(earnings?.pendingBalance || 0)}
        />
        <MetricCard
          title="Lifetime Sales"
          value={earnings?.lifetimeSales || 0}
        />
        <Card>
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 rounded-lg">
                <Award className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-500">Current Fee Tier</p>
                <p className="text-xl font-semibold text-gray-900">
                  {(earnings?.feeRate || 0) === 0 ? '0% Fee' : '5% Fee'}
                </p>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>

      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Platform Fee Structure</CardTitle>
          <CardDescription>How VideoForge platform fees work</CardDescription>
        </CardHeader>
        <CardBody>
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-700">
                First $500: <span className="text-emerald-600">0% platform fee</span>
              </span>
              <span className="text-sm text-gray-500">{Math.min(tierProgress, 500)} / $500</span>
            </div>
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm font-medium text-gray-700">
                After $500: <span className="text-gray-900">5% platform fee</span>
              </span>
              <span className="text-sm text-gray-500">
                {tierProgress >= 500 ? 'Tier achieved!' : 'Locked'}
              </span>
            </div>
            <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
              <div
                className="absolute left-0 top-0 h-full bg-gradient-to-r from-emerald-400 to-emerald-600 rounded-full transition-all duration-500"
                style={{ width: `${tierPercentage}%` }}
              ></div>
            </div>
            <p className="text-xs text-gray-500 mt-2">
              {tierProgress >= 500
                ? `You've earned over $500! All future earnings are processed at 0% fee.`
                : `Earn $${500 - tierProgress} more to unlock the 0% fee tier.`}
            </p>
          </div>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
          <CardDescription>Your latest payout activity</CardDescription>
        </CardHeader>
        <CardBody>
          {payouts && payouts.length > 0 ? (
            <div className="space-y-4">
              {payouts.map((payout) => (
                <div
                  key={payout.id}
                  className="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-white rounded-lg border border-gray-200">
                      <DollarSign className="h-5 w-5 text-gray-400" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">
                        {formatCurrency(payout.amount)}
                      </p>
                      <p className="text-sm text-gray-500">
                        Fee: {formatCurrency(payout.fee)} • Net: {formatCurrency(payout.netAmount)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <Badge variant={getStatusBadgeVariant(payout.status) as 'default' | 'success' | 'warning' | 'danger' | 'info'}>
                      {payout.status.charAt(0).toUpperCase() + payout.status.slice(1)}
                    </Badge>
                    <span className="text-sm text-gray-500">{formatDate(payout.createdAt)}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Clock className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500">No transactions yet</p>
              <p className="text-sm text-gray-400">Complete projects to start earning</p>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
}