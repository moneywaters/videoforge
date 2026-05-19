import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Card, CardBody } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Tabs } from '@/components/ui/Tabs';
import type { LeaderboardEntry } from '@/types/index';

const rankLabel = (rank: number) => {
  switch (rank) {
    case 1: return '🥇';
    case 2: return '🥈';
    case 3: return '🥉';
    default: return `#${rank}`;
  }
};

function LeaderboardRow({ entry, isGlobal }: { entry: LeaderboardEntry; isGlobal: boolean }) {
  const displayRank = isGlobal ? rankLabel(entry.rank) : `#${entry.rank}`;

  return (
    <TableRow>
      <TableCell>
        <span className="font-semibold">{displayRank}</span>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-brand-100 flex items-center justify-center">
            <span className="text-sm font-medium text-brand-700">
              {entry.editorName.charAt(0)}
            </span>
          </div>
          <span className="font-medium text-gray-900">
            {isGlobal ? entry.editorName : `Editor #${entry.rank + 100}`}
          </span>
        </div>
      </TableCell>
      <TableCell>{entry.videoCount}</TableCell>
      <TableCell>{entry.totalSales}</TableCell>
      <TableCell className="font-semibold">${entry.revenue.toLocaleString()}</TableCell>
    </TableRow>
  );
}

function LeaderboardSkeleton() {
  return (
    <Card>
      <CardBody>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Rank</TableHead>
              <TableHead>Editor</TableHead>
              <TableHead>Videos</TableHead>
              <TableHead>Total Sales</TableHead>
              <TableHead>Revenue</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[...Array(5)].map((_, i) => (
              <TableRow key={i}>
                <TableCell><Skeleton className="h-6 w-12" /></TableCell>
                <TableCell><Skeleton className="h-6 w-32" /></TableCell>
                <TableCell><Skeleton className="h-6 w-12" /></TableCell>
                <TableCell><Skeleton className="h-6 w-12" /></TableCell>
                <TableCell><Skeleton className="h-6 w-24" /></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardBody>
    </Card>
  );
}

export function Leaderboard() {
  const [activeTab, setActiveTab] = useState('global');

  const { data: leaderboard, isLoading } = useQuery({
    queryKey: ['leaderboard'],
    queryFn: () => api.getLeaderboard(),
  });

  const { data: briefs } = useQuery({
    queryKey: ['briefs'],
    queryFn: () => api.getBriefs(),
  });

  const tabs = [
    { id: 'global', label: 'Global Rankings' },
    { id: 'brief', label: 'By Brief' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Leaderboard</h1>
      </div>

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab}>
        {isLoading ? (
          <LeaderboardSkeleton />
        ) : activeTab === 'global' ? (
          <Card>
            <CardBody>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Rank</TableHead>
                    <TableHead>Editor</TableHead>
                    <TableHead>Videos</TableHead>
                    <TableHead>Total Sales</TableHead>
                    <TableHead>Revenue</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {leaderboard?.map((entry) => (
                    <LeaderboardRow key={entry.rank} entry={entry} isGlobal={true} />
                  ))}
                </TableBody>
              </Table>
            </CardBody>
          </Card>
        ) : (
          <Card>
            <CardBody>
              <p className="text-gray-500 text-center py-8">
                Select a brief to view its rankings
              </p>
              <div className="mt-4 space-y-2">
                {briefs?.filter((b) => b.status === 'closed').map((brief) => (
                  <div
                    key={brief.id}
                    className="p-3 border border-gray-200 rounded hover:bg-gray-50 cursor-pointer"
                  >
                    <p className="font-medium text-gray-900">{brief.title}</p>
                    <p className="text-sm text-gray-500">
                      {brief.currentSubmissions} submissions • ${brief.bountyBudget.toLocaleString()} bounty
                    </p>
                  </div>
                ))}
              </div>
            </CardBody>
          </Card>
        )}
      </Tabs>
    </div>
  );
}