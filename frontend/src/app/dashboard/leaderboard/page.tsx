'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import type { LeaderboardEntry, Brief } from '@/types';

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
          <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
            <span className="text-sm font-medium text-primary">
              {entry.editorName.charAt(0)}
            </span>
          </div>
          <span className="font-medium">
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
      <CardContent>
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
      </CardContent>
    </Card>
  );
}

export default function LeaderboardPage() {
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
        <h1 className="text-2xl font-bold">Leaderboard</h1>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="global">Global Rankings</TabsTrigger>
          <TabsTrigger value="brief">By Brief</TabsTrigger>
        </TabsList>
        <TabsContent value="global">
          {isLoading ? (
            <LeaderboardSkeleton />
          ) : (
            <Card>
              <CardContent>
                {leaderboard && leaderboard.length > 0 ? (
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
                      {leaderboard.map((entry) => (
                        <LeaderboardRow key={entry.rank} entry={entry} isGlobal={true} />
                      ))}
                    </TableBody>
                  </Table>
                ) : (
                  <div className="text-center py-12 text-muted-foreground">
                    No leaderboard data available yet. Check back later!
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </TabsContent>
        <TabsContent value="brief">
          <Card>
            <CardContent>
              <p className="text-muted-foreground text-center py-8">
                Select a brief to view its rankings
              </p>
              <div className="mt-4 space-y-2">
                {briefs?.filter((b) => b.status === 'closed').map((brief) => (
                  <div
                    key={brief.id}
                    className="p-3 border rounded-md hover:bg-muted cursor-pointer"
                  >
                    <p className="font-medium">{brief.title}</p>
                    <p className="text-sm text-muted-foreground">
                      {brief.currentSubmissions} submissions &bull; ${brief.bountyBudget.toLocaleString()} bounty
                    </p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}