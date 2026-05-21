import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Card, CardBody, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { MetricCard } from '@/components/ui/MetricCard';
import { Badge } from '@/components/ui/Badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Button } from '@/components/ui/Button';
import { Skeleton } from '@/components/ui/Skeleton';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';

// Client Dashboard Component
function DashboardClient() {
  const { user } = useAuth();
  
  const { data: briefs, isLoading: briefsLoading } = useQuery({
    queryKey: ['briefs'],
    queryFn: api.getBriefs,
  });

  const { data: notifications, isLoading: notificationsLoading } = useQuery({
    queryKey: ['notifications', user?.id],
    queryFn: () => api.getNotifications(user!.id),
    enabled: !!user?.id,
  });

  const briefsCount = briefs?.filter(b => b.status === 'published').length ?? 0;
  const pendingVideos = briefs?.reduce((sum, b) => sum + b.currentSubmissions, 0) ?? 0;
  const totalSpent = briefs?.reduce((sum, b) => sum + b.bountyBudget, 0) ?? 0;
  const activeCampaigns = 0; // Would come from campaigns API

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Welcome back, {user?.name?.split(' ')[0]}</h1>
        <p className="text-gray-500">Here's an overview of your video projects</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {briefsLoading ? (
          <>
            {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-32" />)}
          </>
        ) : (
          <>
            <MetricCard title="Active Briefs" value={briefsCount} />
            <MetricCard title="Pending Videos" value={pendingVideos} />
            <MetricCard title="Total Spent" value={`$${totalSpent.toLocaleString()}`} />
            <MetricCard title="Campaigns Running" value={activeCampaigns} />
          </>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Recent Briefs</CardTitle>
              <Link to="/briefs">
                <Button variant="ghost" size="sm">View All</Button>
              </Link>
            </div>
          </CardHeader>
          <CardBody>
            {briefsLoading ? (
              <div className="space-y-3">
                {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-12" />)}
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Submissions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {briefs?.slice(0, 3).map((brief) => (
                    <TableRow key={brief.id}>
                      <TableCell className="font-medium">{brief.title}</TableCell>
                      <TableCell>
                        <Badge variant={brief.status === 'published' ? 'success' : 'default'}>
                          {brief.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{brief.currentSubmissions}/{brief.submissionLimit}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
          </CardHeader>
          <CardBody>
            {notificationsLoading ? (
              <div className="space-y-3">
                {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-12" />)}
              </div>
            ) : (
              <div className="space-y-3">
                {notifications?.slice(0, 5).map((notification) => (
                  <div
                    key={notification.id}
                    className={`p-3 rounded-md ${notification.read ? 'bg-gray-50' : 'bg-brand-50'}`}
                  >
                    <p className="text-sm text-gray-900">{notification.message}</p>
                    <p className="text-xs text-gray-500 mt-1">
                      {new Date(notification.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  );
}

// Editor Dashboard Component
function DashboardEditor() {
  const { user } = useAuth();

  const { data: briefs, isLoading: briefsLoading } = useQuery({
    queryKey: ['briefs'],
    queryFn: api.getBriefs,
  });

  const { data: earnings } = useQuery({
    queryKey: ['earnings', user?.id],
    queryFn: () => api.getEarnings(user!.id),
    enabled: !!user?.id,
  });

  const { data: leaderboard } = useQuery({
    queryKey: ['leaderboard'],
    queryFn: api.getLeaderboard,
  });

  const availableBriefs = briefs?.filter(b => b.status === 'published').length ?? 0;
  const mySubmissions = 5; // Would come from videos API
  const earningsTotal = earnings?.totalEarned ?? 0;
  const myRank = leaderboard?.findIndex(e => e.editorName === user?.name) ?? -1;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Editor Dashboard</h1>
        <p className="text-gray-500">Find briefs and track your submissions</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {briefsLoading ? (
          <>
            {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-32" />)}
          </>
        ) : (
          <>
            <MetricCard title="Available Briefs" value={availableBriefs} />
            <MetricCard title="My Submissions" value={mySubmissions} />
            <MetricCard title="Earnings" value={`$${earningsTotal.toLocaleString()}`} />
            <MetricCard title="Leaderboard Rank" value={myRank >= 0 ? `#${myRank + 1}` : 'N/A'} />
          </>
        )}
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Briefs Matching Your Skills</CardTitle>
            <Link to="/briefs">
              <Button variant="ghost" size="sm">View All</Button>
            </Link>
          </div>
        </CardHeader>
        <CardBody>
          {briefsLoading ? (
            <div className="space-y-3">
              {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-16" />)}
            </div>
          ) : (
            <div className="space-y-3">
              {briefs?.filter(b => b.status === 'published').slice(0, 5).map((brief) => (
                <div key={brief.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
                  <div>
                    <p className="font-medium text-gray-900">{brief.title}</p>
                    <p className="text-sm text-gray-500">{brief.tags.join(', ')}</p>
                  </div>
                  <div className="text-right">
                    <p className="font-medium text-gray-900">${brief.bountyBudget.toLocaleString()}</p>
                    <p className="text-sm text-gray-500">{brief.currentSubmissions}/{brief.submissionLimit} submissions</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
}

// Ad Specialist Dashboard Component
function DashboardSpecialist() {
  const { data: campaigns, isLoading: campaignsLoading } = useQuery({
    queryKey: ['campaigns'],
    queryFn: api.getCampaigns,
  });

  const { data: performance } = useQuery({
    queryKey: ['platformPerformance'],
    queryFn: api.getPlatformPerformance,
  });

  const activeCampaigns = campaigns?.filter(c => c.status === 'active').length ?? 0;
  const totalSpend = campaigns?.reduce((sum, c) => sum + c.spent, 0) ?? 0;
  const totalRevenue = performance?.reduce((sum, p) => sum + p.totalRevenue, 0) ?? 0;
  const avgRoas = performance && performance.length > 0
    ? (performance.reduce((sum, p) => sum + p.roas, 0) / performance.length).toFixed(2)
    : '0';

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Ad Specialist Dashboard</h1>
        <p className="text-gray-500">Track your campaign performance</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {campaignsLoading ? (
          <>
            {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-32" />)}
          </>
        ) : (
          <>
            <MetricCard title="Active Campaigns" value={activeCampaigns} />
            <MetricCard title="Total Ad Spend" value={`$${totalSpend.toLocaleString()}`} />
            <MetricCard title="Sales Generated" value={`$${totalRevenue.toLocaleString()}`} />
            <MetricCard title="Avg ROAS" value={`${avgRoas}x`} />
          </>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Campaign Performance</CardTitle>
          <CardDescription>Track how your ads are performing across platforms</CardDescription>
        </CardHeader>
        <CardBody>
          <div className="h-64 flex items-center justify-center bg-gray-50 rounded-md">
            <div className="text-center">
              <p className="text-gray-500">Chart placeholder</p>
              <p className="text-sm text-gray-400">Campaign performance chart will appear here</p>
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  );
}

// Admin Dashboard Component
function DashboardAdmin() {
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: api.getUsers,
  });

  const { data: briefs } = useQuery({
    queryKey: ['briefs'],
    queryFn: api.getBriefs,
  });

  const { data: payouts, isLoading: payoutsLoading } = useQuery({
    queryKey: ['payouts'],
    queryFn: () => api.getPayouts(''),
  });

  const { data: disputes, isLoading: disputesLoading } = useQuery({
    queryKey: ['disputes'],
    queryFn: api.getDisputes,
  });

  const totalUsers = users?.length ?? 0;
  const activeBriefs = briefs?.filter(b => b.status === 'published').length ?? 0;
  const pendingPayouts = payouts?.filter(p => p.status === 'pending' || p.status === 'processing').length ?? 0;
  const openDisputes = disputes?.filter(d => d.status === 'open' || d.status === 'under_review').length ?? 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="text-gray-500">Platform overview and management</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {usersLoading ? (
          <>
            {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-32" />)}
          </>
        ) : (
          <>
            <MetricCard title="Total Users" value={totalUsers} />
            <MetricCard title="Active Briefs" value={activeBriefs} />
            <MetricCard title="Pending Payouts" value={pendingPayouts} />
            <MetricCard title="Open Disputes" value={openDisputes} />
          </>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Quick Links</CardTitle>
          </CardHeader>
          <CardBody className="space-y-2">
            <Link to="/admin/users">
              <Button variant="secondary" className="w-full justify-start">Manage Users</Button>
            </Link>
            <Link to="/admin/moderation">
              <Button variant="secondary" className="w-full justify-start">Moderation Queue</Button>
            </Link>
            <Link to="/admin/disputes">
              <Button variant="secondary" className="w-full justify-start">Disputes</Button>
            </Link>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Pending Payouts</CardTitle>
          </CardHeader>
          <CardBody>
            {payoutsLoading ? (
              <Skeleton className="h-20" />
            ) : (
              <div className="space-y-2">
                {payouts?.filter(p => p.status === 'pending' || p.status === 'processing').slice(0, 3).map((payout) => (
                  <div key={payout.id} className="flex justify-between text-sm">
                    <span>{payout.userName}</span>
                    <span className="font-medium">${payout.netAmount.toLocaleString()}</span>
                  </div>
                ))}
              </div>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Open Disputes</CardTitle>
          </CardHeader>
          <CardBody>
            {disputesLoading ? (
              <Skeleton className="h-20" />
            ) : (
              <div className="space-y-2">
                {disputes?.filter(d => d.status === 'open' || d.status === 'under_review').slice(0, 3).map((dispute) => (
                  <div key={dispute.id} className="text-sm">
                    <p className="font-medium">{dispute.reporterName} vs {dispute.targetName}</p>
                    <p className="text-gray-500">{dispute.reason}</p>
                  </div>
                ))}
              </div>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  );
}

// Main Dashboard Component
export default function Dashboard() {
  const { user, isClient, isEditor, isSpecialist, isAdmin } = useAuth();

  if (!user) {
    return (
      <div className="p-8 text-center">
        <p className="text-gray-500">Please log in to view your dashboard</p>
      </div>
    );
  }

  if (isClient) {
    return <DashboardClient />;
  }
  if (isEditor) {
    return <DashboardEditor />;
  }
  if (isSpecialist) {
    return <DashboardSpecialist />;
  }
  if (isAdmin) {
    return <DashboardAdmin />;
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-gray-900">Welcome, {user.name}</h1>
      <p className="text-gray-500 mt-2">Your dashboard is being prepared</p>
    </div>
  );
}