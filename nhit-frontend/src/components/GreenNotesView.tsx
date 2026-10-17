import React, { useState } from 'react';
import { Download, RefreshCw, Plus, ArrowUpDown } from 'lucide-react';
import { GreenNote } from '../types';
import { formatINR } from './DashboardView';

interface GreenNotesViewProps {
  greenNotes: GreenNote[];
  onRefresh: () => void;
  onOpenNewGreenNote: () => void;
  onOpenApproval: (note: GreenNote) => void;
  searchQuery: string;
}

export const GreenNotesView: React.FC<GreenNotesViewProps> = ({
  greenNotes,
  onRefresh,
  onOpenNewGreenNote,
  onOpenApproval,
  searchQuery,
}) => {
  const [sortAsc, setSortAsc] = useState(false);

  const filtered = greenNotes.filter((n) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      n.sl_no.toLowerCase().includes(q) ||
      n.project_name.toLowerCase().includes(q) ||
      (n.vendor_name || n.supplier_name || '').toLowerCase().includes(q)
    );
  });

  const sorted = [...filtered].sort((a, b) => {
    return sortAsc ? a.amount - b.amount : b.amount - a.amount;
  });

  const exportCSV = () => {
    const headers = ['Note Reference', 'Project Name', 'Contractor', 'Gross Amount', 'Status', 'Creator'];
    const rows = sorted.map((n) => [
      `"${n.sl_no}"`,
      `"${n.project_name}"`,
      `"${n.vendor_name || n.supplier_name || 'EPC Vendor'}"`,
      `"${n.amount}"`,
      `"${n.status}"`,
      `"${n.created_by || 'Resident Engineer'}"`,
    ]);

    const csvContent = '\ufeff' + [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = 'NHIT_GreenNotes_Registry.csv';
    link.click();
    link.remove();
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'APPROVED':
        return <span className="badge badge-approved">APPROVED</span>;
      case 'PENDING':
      case 'IN_REVIEW':
        return <span className="badge badge-pending">PENDING</span>;
      case 'REJECTED':
        return <span className="badge badge-rejected">REJECTED</span>;
      default:
        return <span className="badge badge-draft">DRAFT</span>;
    }
  };

  return (
    <div className="table-card">
      <div className="table-header-bar">
        <h2>
          <span>GreenNotes Approval Registry</span>
        </h2>
        <div className="table-header-actions">
          <button className="btn-secondary" onClick={exportCSV}>
            <Download size={14} />
            <span>Export CSV</span>
          </button>
          <button className="btn-secondary" onClick={onRefresh}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </button>
          <button className="btn-primary" onClick={onOpenNewGreenNote}>
            <Plus size={14} />
            <span>Create Green Note</span>
          </button>
        </div>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>Note Reference</th>
            <th>Project Details</th>
            <th>Contractor</th>
            <th className="sortable" onClick={() => setSortAsc(!sortAsc)}>
              Gross Amount (INR) <ArrowUpDown size={12} className="sort-icon" />
            </th>
            <th>Initiator</th>
            <th>Status</th>
            <th>Review & Signoff</th>
          </tr>
        </thead>
        <tbody>
          {sorted.length === 0 ? (
            <tr>
              <td colSpan={7}>
                <div className="empty-state">
                  <h4>No Green Notes Found</h4>
                  <p>Try clearing your search query or draft a new note.</p>
                </div>
              </td>
            </tr>
          ) : (
            sorted.map((note) => (
              <tr key={note.id}>
                <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                  {note.sl_no}
                </td>
                <td>
                  <div style={{ fontWeight: 600 }}>{note.project_name}</div>
                </td>
                <td>{note.vendor_name || note.supplier_name || 'EPC Vendor'}</td>
                <td className="mono" style={{ fontWeight: 700 }}>
                  {formatINR(note.amount)}
                </td>
                <td>{note.created_by || 'Resident Engineer'}</td>
                <td>{getStatusBadge(note.status)}</td>
                <td>
                  <button
                    className="btn-primary"
                    style={{ padding: '4px 12px', fontSize: '0.75rem' }}
                    onClick={() => onOpenApproval(note)}
                  >
                    Signoff Stepper
                  </button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};
