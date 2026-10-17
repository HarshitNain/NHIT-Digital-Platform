import React from 'react';
import { Organization, Department, Designation } from '../types';

interface OrganizationsViewProps {
  organizations: Organization[];
  departments: Department[];
  designations: Designation[];
}

export const OrganizationsView: React.FC<OrganizationsViewProps> = ({
  organizations,
  departments,
  designations,
}) => {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '24px' }}>
      {/* Organizations */}
      <div className="table-card">
        <div className="table-header-bar">
          <h2>
            <span>Trust Entities & Subsidiaries</span>
          </h2>
        </div>
        <table className="data-table">
          <thead>
            <tr>
              <th>Entity Code</th>
              <th>Trust / Company Name</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {organizations.map((org) => (
              <tr key={org.org_id}>
                <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                  {org.code}
                </td>
                <td style={{ fontWeight: 600 }}>{org.name}</td>
                <td style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                  {org.description || 'Special Purpose Highway Trust'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Departments & Designations */}
      <div className="table-card">
        <div className="table-header-bar">
          <h2>
            <span>Departments & Cost Centers</span>
          </h2>
        </div>
        <table className="data-table">
          <thead>
            <tr>
              <th>Department Name</th>
              <th>Internal UUID</th>
            </tr>
          </thead>
          <tbody>
            {departments.map((dept) => (
              <tr key={dept.id}>
                <td style={{ fontWeight: 600 }}>{dept.name}</td>
                <td className="mono" style={{ color: 'var(--text-dim)', fontSize: '0.8rem' }}>
                  {dept.id.slice(0, 8)}...
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="table-header-bar" style={{ borderTop: '1px solid var(--border-subtle)' }}>
          <h2>
            <span>Designations & Staff Hierarchy</span>
          </h2>
        </div>
        <table className="data-table">
          <thead>
            <tr>
              <th>Designation Title</th>
              <th>Hierarchy ID</th>
            </tr>
          </thead>
          <tbody>
            {designations.map((desig) => (
              <tr key={desig.id}>
                <td style={{ fontWeight: 600 }}>{desig.name}</td>
                <td className="mono" style={{ color: 'var(--text-dim)', fontSize: '0.8rem' }}>
                  {desig.id.slice(0, 8)}...
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
