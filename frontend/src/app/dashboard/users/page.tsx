'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import type { UserRole } from '@/types';

const ROLE_OPTIONS = [
  { value: 'client', label: 'Client' },
  { value: 'editor', label: 'Editor' },
  { value: 'ad_specialist', label: 'Ad Specialist' },
  { value: 'admin', label: 'Admin' },
];

const ROLE_BADGE_VARIANTS: Record<UserRole, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  client: 'secondary',
  editor: 'default',
  ad_specialist: 'secondary',
  admin: 'destructive',
  support_ai: 'outline',
};

export default function UsersPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [editingRole, setEditingRole] = useState<string | null>(null);
  const [selectedRole, setSelectedRole] = useState<string>('');
  const [bannedUsers, setBannedUsers] = useState<Set<string>>(new Set());

  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => api.getUsers(),
  });

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const filteredUsers = users?.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleEditRole = (userId: string, currentRole: string) => {
    setEditingRole(userId);
    setSelectedRole(currentRole);
  };

  const handleSaveRole = async (userId: string) => {
    try {
      await api.updateUserRole(userId, selectedRole as UserRole);
      setEditingRole(null);
    } catch (error) {
      console.error('Failed to update role:', error);
    }
  };

  const handleToggleBan = (userId: string) => {
    setBannedUsers((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(userId)) {
        newSet.delete(userId);
      } else {
        newSet.add(userId);
      }
      return newSet;
    });
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
        <h1 className="text-2xl font-bold">User Management</h1>
        <p className="text-muted-foreground mt-1">Manage user roles and permissions</p>
      </div>

      <Card className="max-w-md">
        <CardContent>
          <div className="relative">
            <Icons.search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search users by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {filteredUsers && filteredUsers.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Joined</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-muted flex items-center justify-center">
                          <span className="text-sm font-medium text-muted-foreground">
                            {user.name
                              .split(' ')
                              .map((n) => n[0])
                              .join('')}
                          </span>
                        </div>
                        <div>
                          <p className="font-medium">{user.name}</p>
                          <p className="text-sm text-muted-foreground">{user.id}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{user.email}</TableCell>
                    <TableCell>
                      {editingRole === user.id ? (
                        <div className="flex items-center gap-2">
                          <Select value={selectedRole} onValueChange={setSelectedRole}>
                            <SelectTrigger className="w-36">
                              <SelectValue placeholder="Select role" />
                            </SelectTrigger>
                            <SelectContent>
                              {ROLE_OPTIONS.map((option) => (
                                <SelectItem key={option.value} value={option.value}>
                                  {option.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <Button
                            size="sm"
                            onClick={() => handleSaveRole(user.id)}
                          >
                            Save
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setEditingRole(null)}
                          >
                            Cancel
                          </Button>
                        </div>
                      ) : (
                        <Badge variant={ROLE_BADGE_VARIANTS[user.role]}>
                          {user.role.replace('_', ' ')}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {bannedUsers.has(user.id) ? (
                        <Badge variant="destructive">Banned</Badge>
                      ) : (
                        <Badge variant="default">Active</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatDate(user.createdAt)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {editingRole !== user.id && (
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() => handleEditRole(user.id, user.role)}
                          >
                            <Icons.userPen className="h-4 w-4 mr-1" />
                            Edit Role
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant={bannedUsers.has(user.id) ? 'default' : 'destructive'}
                          onClick={() => handleToggleBan(user.id)}
                        >
                          {bannedUsers.has(user.id) ? (
                            <>
                              <Icons.user className="h-4 w-4 mr-1" />
                              Unban
                            </>
                          ) : (
                            <>
                              <Icons.employee className="h-4 w-4 mr-1" />
                              Ban
                            </>
                          )}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-12">
              <Icons.user className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No users found</p>
              <p className="text-sm text-muted-foreground/60">
                {searchQuery ? 'Try a different search term' : 'Users will appear here'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}