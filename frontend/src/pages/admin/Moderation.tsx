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
import { AlertTriangle, CheckCircle, XCircle, FileVideo, FileText, MessageSquare } from 'lucide-react';

const STATUS_TABS = [
  { id: 'pending', label: 'Pending' },
  { id: 'approved', label: 'Approved' },
  { id: 'rejected', label: 'Rejected' },
];

const TYPE_ICONS = {
  video: FileVideo,
  brief: FileText,
  comment: MessageSquare,
};

export default function Moderation() {
  const [activeTab, setActiveTab] = useState('pending');

  const { data: moderationItems, isLoading } = useQuery({
    queryKey: ['moderation'],
    queryFn: () => api.getModerationQueue(),
  });

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'approved':
        return 'success';
      case 'rejected':
        return 'danger';
      case 'pending':
        return 'warning';
      default:
        return 'default';
    }
  };

  const getTypeBadgeVariant = (type: string) => {
    switch (type) {
      case 'video':
        return 'info';
      case 'brief':
        return 'warning';
      case 'comment':
        return 'default';
      default:
        return 'default';
    }
  };

  const filteredItems = moderationItems?.filter(
    (item) => activeTab === 'pending' || item.status === activeTab
  );

  const handleApprove = (itemId: string) => {
    console.log('Approving:', itemId);
  };

  const handleReject = (itemId: string) => {
    console.log('Rejecting:', itemId);
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
        <h1 className="text-3xl font-bold text-gray-900">Moderation Queue</h1>
        <p className="text-gray-500 mt-2">
          Review flagged content and user reports
        </p>
      </div>

      <Tabs tabs={STATUS_TABS} activeTab={activeTab} onChange={setActiveTab}>
        <div />
      </Tabs>

      <Card>
        <CardHeader>
          <CardTitle>
            {activeTab === 'pending' && 'Pending Reviews'}
            {activeTab === 'approved' && 'Approved Items'}
            {activeTab === 'rejected' && 'Rejected Items'}
          </CardTitle>
        </CardHeader>
        <CardBody className="p-0">
          {filteredItems && filteredItems.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Content ID</TableHead>
                  <TableHead>Flagged By</TableHead>
                  <TableHead>Reason</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredItems.map((item) => {
                  const TypeIcon = TYPE_ICONS[item.type as keyof typeof TYPE_ICONS] || FileText;
                  return (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <TypeIcon className="h-5 w-5 text-gray-400" />
                          <Badge variant={getTypeBadgeVariant(item.type) as 'default' | 'success' | 'warning' | 'danger' | 'info'}>
                            {item.type}
                          </Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="font-mono text-sm">{item.contentId}</span>
                      </TableCell>
                      <TableCell>{item.flaggedBy}</TableCell>
                      <TableCell className="max-w-xs">
                        <p className="truncate">{item.reason}</p>
                      </TableCell>
                      <TableCell className="text-gray-500">
                        {formatDate(item.createdAt)}
                      </TableCell>
                      <TableCell>
                        <Badge variant={getStatusBadgeVariant(item.status) as 'default' | 'success' | 'warning' | 'danger' | 'info'}>
                          {item.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {item.status === 'pending' && (
                          <div className="flex items-center gap-2">
                            <Button
                              size="sm"
                              variant="primary"
                              onClick={() => handleApprove(item.id)}
                            >
                              <CheckCircle className="h-4 w-4 mr-1" />
                              Approve
                            </Button>
                            <Button
                              size="sm"
                              variant="danger"
                              onClick={() => handleReject(item.id)}
                            >
                              <XCircle className="h-4 w-4 mr-1" />
                              Reject
                            </Button>
                          </div>
                        )}
                        {item.status === 'approved' && (
                          <div className="flex items-center text-emerald-600">
                            <CheckCircle className="h-4 w-4 mr-1" />
                            <span className="text-sm">Approved</span>
                          </div>
                        )}
                        {item.status === 'rejected' && (
                          <div className="flex items-center text-rose-600">
                            <XCircle className="h-4 w-4 mr-1" />
                            <span className="text-sm">Rejected</span>
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-12">
              <AlertTriangle className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500">No items found</p>
              <p className="text-sm text-gray-400">
                {activeTab === 'pending'
                  ? 'All caught up! No pending reviews.'
                  : `No ${activeTab} items`}
              </p>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
}