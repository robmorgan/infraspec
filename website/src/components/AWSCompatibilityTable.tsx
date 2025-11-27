'use client';

import { useState } from 'react';
import compatibilityData from '../data/aws-compatibility.json';

type ServiceStatus = 'implemented' | 'planned' | 'unsupported';
type Priority = 'high' | 'medium' | 'low';

interface InfraSpecOperation {
  name: string;
  implemented: boolean;
  description?: string;
}

interface VirtualCloudOperation {
  name: string;
  implemented: boolean;
  priority: Priority;
}

interface InfraSpecCoverage {
  status: ServiceStatus;
  operations: InfraSpecOperation[];
}

interface VirtualCloudCoverage {
  status: ServiceStatus;
  coveragePercent: number;
  totalOperations: number;
  implemented: number;
  operations: VirtualCloudOperation[];
}

interface Service {
  name: string;
  fullName: string;
  status: ServiceStatus;
  infraspec?: InfraSpecCoverage;
  virtualCloud?: VirtualCloudCoverage;
}

interface CompatibilityData {
  generatedAt: string;
  services: Service[];
}

function StatusBadge({ status }: { status: ServiceStatus }) {
  const styles = {
    implemented: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    planned: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    unsupported: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
  };

  const labels = {
    implemented: 'Implemented',
    planned: 'Planned',
    unsupported: 'Not Available',
  };

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded-full ${styles[status]}`}>
      {labels[status]}
    </span>
  );
}

function CoverageBar({ percent }: { percent: number }) {
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          className="h-full bg-purple-600 dark:bg-purple-500 rounded-full transition-all"
          style={{ width: `${Math.min(percent, 100)}%` }}
        />
      </div>
      <span className="text-sm text-gray-600 dark:text-gray-400 min-w-[3rem] text-right">
        {percent.toFixed(1)}%
      </span>
    </div>
  );
}

function OperationList({
  operations,
  type
}: {
  operations: (InfraSpecOperation | VirtualCloudOperation)[];
  type: 'infraspec' | 'virtualCloud';
}) {
  const implemented = operations.filter(op => op.implemented);
  const notImplemented = operations.filter(op => !op.implemented);

  return (
    <div className="space-y-2">
      {implemented.length > 0 && (
        <div>
          <h5 className="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1">
            Available ({implemented.length})
          </h5>
          <div className="flex flex-wrap gap-1">
            {implemented.map(op => (
              <span
                key={op.name}
                className="inline-flex items-center px-2 py-0.5 text-xs bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-400 rounded"
                title={'description' in op ? op.description : undefined}
              >
                <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                {op.name.replace(/^Assert/, '')}
              </span>
            ))}
          </div>
        </div>
      )}
      {notImplemented.length > 0 && type === 'virtualCloud' && (
        <div>
          <h5 className="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1">
            Coming Soon ({notImplemented.length})
          </h5>
          <div className="flex flex-wrap gap-1">
            {notImplemented.map(op => (
              <span
                key={op.name}
                className="inline-flex items-center px-2 py-0.5 text-xs bg-gray-50 text-gray-500 dark:bg-gray-800 dark:text-gray-500 rounded"
              >
                {op.name}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ServiceRow({ service }: { service: Service }) {
  const [isExpanded, setIsExpanded] = useState(false);

  const hasDetails =
    (service.infraspec?.operations?.length ?? 0) > 0 ||
    (service.virtualCloud?.operations?.length ?? 0) > 0;

  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
      <button
        onClick={() => hasDetails && setIsExpanded(!isExpanded)}
        className={`w-full px-4 py-3 flex items-center justify-between text-left ${
          hasDetails ? 'hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer' : ''
        }`}
        disabled={!hasDetails}
      >
        <div className="flex items-center gap-3">
          {hasDetails && (
            <svg
              className={`w-4 h-4 text-gray-400 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          )}
          {!hasDetails && <div className="w-4" />}
          <div>
            <div className="font-medium text-gray-900 dark:text-gray-100">
              {service.fullName}
            </div>
            <div className="text-sm text-gray-500 dark:text-gray-400">
              {service.name.toUpperCase()}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <StatusBadge status={service.status} />
        </div>
      </button>

      {isExpanded && hasDetails && (
        <div className="px-4 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/30">
          <div className="grid md:grid-cols-2 gap-6">
            {/* InfraSpec Column */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                  InfraSpec Assertions
                </h4>
                {service.infraspec && (
                  <StatusBadge status={service.infraspec.status} />
                )}
              </div>
              {service.infraspec?.operations && service.infraspec.operations.length > 0 ? (
                <OperationList operations={service.infraspec.operations} type="infraspec" />
              ) : (
                <p className="text-sm text-gray-500 dark:text-gray-400 italic">
                  No assertions available yet
                </p>
              )}
            </div>

            {/* Virtual Cloud Column */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Virtual Cloud API
                </h4>
                {service.virtualCloud && (
                  <StatusBadge status={service.virtualCloud.status} />
                )}
              </div>
              {service.virtualCloud ? (
                <div className="space-y-3">
                  <div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                      Coverage: {service.virtualCloud.implemented} of {service.virtualCloud.totalOperations} operations
                    </div>
                    <CoverageBar percent={service.virtualCloud.coveragePercent} />
                  </div>
                  {service.virtualCloud.operations && service.virtualCloud.operations.length > 0 && (
                    <OperationList operations={service.virtualCloud.operations} type="virtualCloud" />
                  )}
                </div>
              ) : (
                <p className="text-sm text-gray-500 dark:text-gray-400 italic">
                  Not available in Virtual Cloud
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function AWSCompatibilityTable() {
  const data = compatibilityData as CompatibilityData;
  const implementedServices = data.services.filter(s => s.status === 'implemented');
  const plannedServices = data.services.filter(s => s.status === 'planned');

  const lastUpdated = new Date(data.generatedAt).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

  return (
    <div className="space-y-8">
      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-green-50 dark:bg-green-900/20 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-green-700 dark:text-green-400">
            {implementedServices.length}
          </div>
          <div className="text-sm text-green-600 dark:text-green-500">
            Implemented
          </div>
        </div>
        <div className="bg-yellow-50 dark:bg-yellow-900/20 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-yellow-700 dark:text-yellow-400">
            {plannedServices.length}
          </div>
          <div className="text-sm text-yellow-600 dark:text-yellow-500">
            Planned
          </div>
        </div>
        <div className="bg-purple-50 dark:bg-purple-900/20 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-purple-700 dark:text-purple-400">
            {implementedServices.reduce((acc, s) => acc + (s.infraspec?.operations?.length || 0), 0)}
          </div>
          <div className="text-sm text-purple-600 dark:text-purple-500">
            InfraSpec Assertions
          </div>
        </div>
        <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-blue-700 dark:text-blue-400">
            {implementedServices.reduce((acc, s) => acc + (s.virtualCloud?.implemented || 0), 0)}
          </div>
          <div className="text-sm text-blue-600 dark:text-blue-500">
            Virtual Cloud APIs
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
        <div className="flex items-center gap-2">
          <StatusBadge status="implemented" />
          <span>Service is available</span>
        </div>
        <div className="flex items-center gap-2">
          <StatusBadge status="planned" />
          <span>Coming soon</span>
        </div>
      </div>

      {/* Implemented Services */}
      {implementedServices.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Implemented Services
          </h3>
          <div className="space-y-3">
            {implementedServices.map(service => (
              <ServiceRow key={service.name} service={service} />
            ))}
          </div>
        </div>
      )}

      {/* Planned Services */}
      {plannedServices.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Planned Services
          </h3>
          <div className="space-y-3">
            {plannedServices.map(service => (
              <ServiceRow key={service.name} service={service} />
            ))}
          </div>
        </div>
      )}

      {/* Footer */}
      <div className="text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 pt-4">
        Last updated: {lastUpdated}
      </div>
    </div>
  );
}
