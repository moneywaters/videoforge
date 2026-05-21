'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

export default function AnalyticsPage() {
  const { data: performance, isLoading } = useQuery({
    queryKey: ['performance'],
    queryFn: () => api.getPerformance(),
  });

  const { data: platformPerformance, isLoading: platformLoading } = useQuery({
    queryKey: ['platformPerformance'],
    queryFn: () => api.getPlatformPerformance(),
  });

  if (isLoading || platformLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="h-32 bg-muted rounded"></div>
            <div className="h-32 bg-muted rounded"></div>
            <div className="h-32 bg-muted rounded"></div>
            <div className="h-32 bg-muted rounded"></div>
          </div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  const totalRevenue = performance?.reduce((sum, p) => sum + p.revenue, 0) ?? 0;
  const totalSales = performance?.reduce((sum, p) => sum + p.sales, 0) ?? 0;
  const totalConversions = performance?.reduce((sum, p) => sum + p.conversions, 0) ?? 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Performance Analytics</h1>
        <p className="text-muted-foreground mt-1">Track your video performance across platforms</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-100 dark:bg-emerald-900 rounded-lg">
                <Icons.trendingUp className="h-5 w-5 text-emerald-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Revenue</p>
                <p className="text-2xl font-semibold">${totalRevenue.toLocaleString()}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                <Icons.video className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Sales</p>
                <p className="text-2xl font-semibold">{totalSales}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                <Icons.check className="h-5 w-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Conversions</p>
                <p className="text-2xl font-semibold">{totalConversions}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-lg">
                <Icons.sparkles className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Avg ROAS</p>
                <p className="text-2xl font-semibold">
                  {platformPerformance && platformPerformance.length > 0
                    ? (platformPerformance.reduce((sum, p) => sum + p.roas, 0) / platformPerformance.length).toFixed(2)
                    : '0'}x
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue Over Time</CardTitle>
          <CardDescription>Track your revenue trends</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center bg-muted rounded-lg">
            <div className="text-center">
              <Icons.trendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">Chart placeholder</p>
              <p className="text-sm text-muted-foreground/60">Revenue chart will appear here</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Platform Performance</CardTitle>
            <CardDescription>Performance by platform</CardDescription>
          </CardHeader>
          <CardContent>
            {platformPerformance && platformPerformance.length > 0 ? (
              <div className="space-y-4">
                {platformPerformance.map((p) => (
                  <div key={p.platform} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                    <span className="font-medium capitalize">{p.platform}</span>
                    <div className="text-right">
                      <p className="font-medium">${p.totalRevenue.toLocaleString()}</p>
                      <p className="text-sm text-muted-foreground">{p.roas}x ROAS</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-center py-8 text-muted-foreground">
                No platform data available yet.
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Video Performance</CardTitle>
            <CardDescription>Individual video metrics</CardDescription>
          </CardHeader>
          <CardContent>
            {performance && performance.length > 0 ? (
              <div className="space-y-4">
                {performance.map((p) => (
                  <div key={p.videoId} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                    <span className="font-medium truncate">{p.videoId}</span>
                    <div className="text-right">
                      <p className="font-medium">${p.revenue.toLocaleString()}</p>
                      <p className="text-sm text-muted-foreground">{p.sales} sales</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-center py-8 text-muted-foreground">
                No video performance data available.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}