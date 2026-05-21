import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { UserRole } from '@/types/index';
import { Card, CardBody, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/Table';
import { Search, Shield, Ban, UserCheck } from 'lucide-react';

const ROLE_OPTIONS = [
  { value: 'client', label: 'Client' },
  { value: 'editor', label: 'Editor' },
  { value: 'ad_specialist', label: 'Ad Specialist' },
  { value: 'admin', label: 'Admin' },
];

const ROLE_BADGE_VARIANTS: Record<UserRole, 'default' | 'success' | 'warning' | 'danger' | 'info'> = {
  client: 'info',
  editor: 'success',
  ad_specialist: 'warning',
  admin: 'danger',
  support_ai: 'default',
};

export default function Users() {
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
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-48"></div>
          <div className="h-12 bg-gray-200 rounded"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">User Management</h1>
        <p className="text-gray-500 mt-2">Manage user roles and permissions</p>
      </div>

      <Card className="mb-6">
        <CardBody>
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
            <Input
              placeholder="Search users by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
        </CardHeader>
        <CardBody className="p-0">
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
              {filteredUsers?.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center">
                        <span className="text-sm font-medium text-gray-600">
                          {user.name
                            .split(' ')
                            .map((n) => n[0])
                            .join('')}
                        </span>
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">{user.name}</p>
                        <p className="text-sm text-gray-500">{user.id}</p>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-gray-500">{user.email}</TableCell>
                  <TableCell>
                    {editingRole === user.id ? (
                      <div className="flex items-center gap-2">
                        <Select
                          options={ROLE_OPTIONS}
                          value={selectedRole}
                          onChange={setSelectedRole}
                          className="w-36"
                        />
                        <Button
                          size="sm"
                          variant="primary"
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
                      <Badge variant="danger">Banned</Badge>
                    ) : (
                      <Badge variant="success">Active</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-gray-500">
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
                          <Shield className="h-4 w-4 mr-1" />
                          Edit Role
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant={bannedUsers.has(user.id) ? 'primary' : 'danger'}
                        onClick={() => handleToggleBan(user.id)}
                      >
                        {bannedUsers.has(user.id) ? (
                          <>
                            <UserCheck className="h-4 w-4 mr-1" />
                            Unban
                          </>
                        ) : (
                          <>
                            <Ban className="h-4 w-4 mr-1" />
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
        </CardBody>
      </Card>
    </div>
  );
}