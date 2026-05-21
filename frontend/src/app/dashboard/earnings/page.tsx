'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

const CURRENT_USER_ID = 'usr-editor-001';

export default function EarningsPage() {
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
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-32 bg-muted rounded"></div>
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
        return 'default';
      case 'pending':
        return 'secondary';
      case 'failed':
        return 'destructive';
      default:
        return 'default';
    }
  };

  const tierProgress = earnings?.tierProgress || 0;
  const tierPercentage = Math.min(100, (tierProgress / 500) * 100);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">My Earnings</h1>
        <p className="text-muted-foreground mt-1">Track your earnings and payout history</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-100 dark:bg-emerald-900 rounded-lg">
                <Icons.trendingUp className="h-5 w-5 text-emerald-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Earned</p>
                <p className="text-2xl font-semibold">{formatCurrency(earnings?.totalEarned || 0)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-lg">
                <Icons.clock className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Pending Balance</p>
                <p className="text-2xl font-semibold">{formatCurrency(earnings?.pendingBalance || 0)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                <Icons.check className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Lifetime Sales</p>
                <p className="text-2xl font-semibold">{earnings?.lifetimeSales || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-lg">
                <Icons.pro className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Current Fee Tier</p>
                <p className="text-xl font-semibold">
                  {(earnings?.feeRate || 0) === 0 ? '0% Fee' : '5% Fee'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Platform Fee Structure</CardTitle>
          <CardDescription>How VideoForge platform fees work</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="bg-muted rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium">
                First $500: <span className="text-emerald-600">0% platform fee</span>
              </span>
              <span className="text-sm text-muted-foreground">{Math.min(tierProgress, 500)} / $500</span>
            </div>
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm font-medium">
                After $500: <span className="text-foreground">5% platform fee</span>
              </span>
              <span className="text-sm text-muted-foreground">
                {tierProgress >= 500 ? 'Tier achieved!' : 'Locked'}
              </span>
            </div>
            <div className="relative h-3 bg-muted-foreground/20 rounded-full overflow-hidden">
              <div
                className="absolute left-0 top-0 h-full bg-gradient-to-r from-emerald-400 to-emerald-600 rounded-full transition-all duration-500"
                style={{ width: `${tierPercentage}%` }}
              ></div>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              {tierProgress >= 500
                ? `You've earned over $500! All future earnings are processed at 0% fee.`
                : `Earn $${500 - tierProgress} more to unlock the 0% fee tier.`}
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
          <CardDescription>Your latest payout activity</CardDescription>
        </CardHeader>
        <CardContent>
          {payouts && payouts.length > 0 ? (
            <div className="space-y-4">
              {payouts.map((payout) => (
                <div
                  key={payout.id}
                  className="flex items-center justify-between p-4 bg-muted rounded-lg"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-background rounded-lg border">
                      <Icons.creditCard className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <p className="font-medium">{formatCurrency(payout.amount)}</p>
                      <p className="text-sm text-muted-foreground">
                        Fee: {formatCurrency(payout.fee)} &bull; Net: {formatCurrency(payout.netAmount)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <Badge variant={getStatusBadgeVariant(payout.status) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                      {payout.status.charAt(0).toUpperCase() + payout.status.slice(1)}
                    </Badge>
                    <span className="text-sm text-muted-foreground">{formatDate(payout.createdAt)}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <Icons.clock className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No transactions yet</p>
              <p className="text-sm text-muted-foreground/60">Complete projects to start earning</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}