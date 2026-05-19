import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Card, CardBody, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Tabs } from '@/components/ui/Tabs';
import { FileText, Users, AlertCircle, Clock, Gavel } from 'lucide-react';

const STATUS_TABS = [
  { id: 'open', label: 'Open' },
  { id: 'under_review', label: 'Under Review' },
  { id: 'resolved', label: 'Resolved' },
];

export default function Disputes() {
  const [activeTab, setActiveTab] = useState('open');

  const { data: disputes, isLoading } = useQuery({
    queryKey: ['disputes'],
    queryFn: () => api.getDisputes(),
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
      case 'resolved':
        return 'success';
      case 'under_review':
        return 'info';
      case 'open':
        return 'warning';
      default:
        return 'default';
    }
  };

  const filteredDisputes = disputes?.filter(
    (d) => activeTab === 'open' || d.status === activeTab
  );

  const handleMarkReviewing = (disputeId: string) => {
    console.log('Marking as reviewing:', disputeId);
  };

  const handleResolveReporter = (disputeId: string) => {
    console.log('Resolving in favor of reporter:', disputeId);
  };

  const handleResolveTarget = (disputeId: string) => {
    console.log('Resolving in favor of target:', disputeId);
  };

  if (isLoading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-48"></div>
          <div className="h-12 bg-gray-200 rounded"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-48 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dispute Resolution</h1>
        <p className="text-gray-500 mt-2">
          Manage and resolve user disputes
        </p>
      </div>

      <Tabs
        tabs={STATUS_TABS}
        activeTab={activeTab}
        onChange={setActiveTab}
      >
        <div />
      </Tabs>

      <div className="space-y-4">
        {filteredDisputes && filteredDisputes.length > 0 ? (
          filteredDisputes.map((dispute) => (
            <Card key={dispute.id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Dispute #{dispute.id.slice(-6)}</CardTitle>
                    <CardDescription>
                      Filed on {formatDate(dispute.createdAt)}
                    </CardDescription>
                  </div>
                  <Badge
                    variant={
                      getStatusBadgeVariant(dispute.status) as
                        | 'default'
                        | 'success'
                        | 'warning'
                        | 'danger'
                        | 'info'
                    }
                  >
                    {dispute.status.replace('_', ' ')}
                  </Badge>
                </div>
              </CardHeader>
              <CardBody>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-rose-100 rounded-lg">
                        <Users className="h-5 w-5 text-rose-600" />
                      </div>
                      <div>
                        <p className="text-sm text-gray-500">Reporter</p>
                        <p className="font-medium text-gray-900">
                          {dispute.reporterName}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-gray-100 rounded-lg">
                        <FileText className="h-5 w-5 text-gray-600" />
                      </div>
                      <div>
                        <p className="text-sm text-gray-500">Target</p>
                        <p className="font-medium text-gray-900">
                          {dispute.targetName}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <div className="p-2 bg-amber-100 rounded-lg">
                        <AlertCircle className="h-5 w-5 text-amber-600" />
                      </div>
                      <div>
                        <p className="text-sm text-gray-500">Reason</p>
                        <p className="font-medium text-gray-900">{dispute.reason}</p>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <p className="text-sm text-gray-500 mb-2">Evidence</p>
                      <div className="space-y-2">
                        {dispute.evidence.map((link, idx) => (
                          <a
                            key={idx}
                            href={link}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-2 text-sm text-brand-600 hover:text-brand-700"
                          >
                            <FileText className="h-4 w-4" />
                            Evidence {idx + 1}
                          </a>
                        ))}
                      </div>
                    </div>

                    {dispute.status === 'resolved' && dispute.resolution && (
                      <div className="p-3 bg-emerald-50 rounded-lg">
                        <p className="text-sm text-emerald-600 font-medium">
                          Resolution
                        </p>
                        <p className="text-sm text-emerald-700 mt-1">
                          {dispute.resolution}
                        </p>
                        {dispute.resolvedAt && (
                          <p className="text-xs text-emerald-500 mt-2">
                            Resolved on {formatDate(dispute.resolvedAt)}
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                </div>

                {dispute.status !== 'resolved' && (
                  <div className="mt-6 pt-6 border-t border-gray-200">
                    <div className="flex flex-wrap gap-3">
                      {dispute.status === 'open' && (
                        <Button
                          variant="secondary"
                          onClick={() => handleMarkReviewing(dispute.id)}
                        >
                          <Clock className="h-4 w-4 mr-2" />
                          Mark as Reviewing
                        </Button>
                      )}
                      {(dispute.status === 'open' ||
                        dispute.status === 'under_review') && (
                        <>
                          <Button
                            variant="primary"
                            onClick={() =>
                              handleResolveReporter(dispute.id)
                            }
                          >
                            <Users className="h-4 w-4 mr-2" />
                            Resolve in Favor of Reporter
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={() => handleResolveTarget(dispute.id)}
                          >
                            <FileText className="h-4 w-4 mr-2" />
                            Resolve in Favor of Target
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                )}
              </CardBody>
            </Card>
          ))
        ) : (
          <Card>
            <CardBody>
              <div className="text-center py-12">
                <Gavel className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                <p className="text-gray-500">No disputes found</p>
                <p className="text-sm text-gray-400">
                  {activeTab === 'open'
                    ? 'All disputes have been resolved!'
                    : `No ${activeTab.replace('_', ' ')} disputes`}
                </p>
              </div>
            </CardBody>
          </Card>
        )}
      </div>
    </div>
  );
}