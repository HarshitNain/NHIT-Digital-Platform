import React from 'react';
import { CircleDollarSign, Clock, Layers, Users, Plus, ArrowUpDown } from 'lucide-react';
import { GreenNote } from '../types';

interface DashboardViewProps {
  greenNotes: GreenNote[];
  onOpenNewGreenNote: () => void;
  onOpenApproval: (note: GreenNote) => void;
  searchQuery: string;
}

export const formatINR = (val: number | undefined | null): string => {
  if (val === undefined || val === null || isNaN(val)) return '₹ 0';
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(val);
};

export const DashboardView: React.FC<DashboardViewProps> = ({
  greenNotes,
  onOpenNewGreenNote,
  onOpenApproval,
  searchQuery,
}) => {
  const [sortAsc, setSortAsc] = React.useState(false);

  // Compute metrics
  const totalApproved = greenNotes
    .filter((n) => n.status === 'APPROVED')
    .reduce((acc, curr) => acc + (curr.amount || 0), 0);

  const pendingAmount = greenNotes
    .filter((n) => n.status === 'PENDING' || n.status === 'IN_REVIEW')
    .reduce((acc, curr) => acc + (curr.amount || 0), 0);

  // Filter notes
  const filtered = greenNotes.filter((n) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      n.sl_no.toLowerCase().includes(q) ||
      n.project_name.toLowerCase().includes(q) ||
      (n.vendor_name || n.supplier_name || '').toLowerCase().includes(q)
    );
  });

  // Sort notes
  const sorted = [...filtered].sort((a, b) => {
    return sortAsc ? a.amount - b.amount : b.amount - a.amount;
  });

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
    <div>
      {/* Metric Cards Grid */}
      <div className="metric-grid">
        <div className="metric-card emerald">
          <div className="metric-title">
            <span>Total Disbursed</span>
            <CircleDollarSign size={18} />
          </div>
          <div className="metric-value mono">{formatINR(totalApproved || 18200000)}</div>
          <div className="metric-subtitle">Across approved highway expenditure notes</div>
        </div>

        <div className="metric-card amber">
          <div className="metric-title">
            <span>In-Review Approvals</span>
            <Clock size={18} />
          </div>
          <div className="metric-value mono">{formatINR(pendingAmount || 4850000)}</div>
          <div className="metric-subtitle">GreenNotes requiring CFO / Director signoff</div>
        </div>

        <div className="metric-card cyan">
          <div className="metric-title">
            <span>Active Corridors</span>
            <Layers size={18} />
          </div>
          <div className="metric-value mono">5 Sites</div>
          <div className="metric-subtitle">Eastern Peripheral, Delhi-Agra, Bhopal-Indore</div>
        </div>

        <div className="metric-card">
          <div className="metric-title">
            <span>Verified Vendors</span>
            <Users size={18} />
          </div>
          <div className="metric-value mono">6 Vendors</div>
          <div className="metric-subtitle">L&T, Dilip Buildcon, Techno Survey, etc.</div>
        </div>
      </div>

      {/* Priority Approval Feed */}
      <div className="table-card">
        <div className="table-header-bar">
          <h2>
            <span>Priority Approval Pipeline (Live Feed)</span>
          </h2>
          <div className="table-header-actions">
            <button className="btn-primary" onClick={onOpenNewGreenNote}>
              <Plus size={15} />
              <span>Draft New GreenNote</span>
            </button>
          </div>
        </div>

        <table className="data-table">
          <thead>
            <tr>
              <th>Serial / Note #</th>
              <th>Project Corridor</th>
              <th>Vendor / Contractor</th>
              <th className="sortable" onClick={() => setSortAsc(!sortAsc)}>
                Gross Amount <ArrowUpDown size={12} className="sort-icon" />
              </th>
              <th>Taxable & GST</th>
              <th>Current Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 ? (
              <tr>
                <td colSpan={7}>
                  <div className="empty-state">
                    <h4>No Notes Found</h4>
                    <p>Draft your first green note using the button above.</p>
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
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{note.id}</div>
                  </td>
                  <td>
                    <div style={{ fontWeight: 600 }}>
                      {note.vendor_name || note.supplier_name || 'EPC Vendor'}
                    </div>
                  </td>
                  <td className="mono" style={{ fontWeight: 700, color: '#fff' }}>
                    {formatINR(note.amount)}
                  </td>
                  <td className="mono" style={{ fontSize: '0.78rem', color: 'var(--text-dim)' }}>
                    Base: {formatINR(note.taxable_amount || note.base_value || note.amount * 0.82)}
                    <br />
                    GST: {formatINR(note.gst_amount || note.gst || note.amount * 0.18)}
                  </td>
                  <td>{getStatusBadge(note.status)}</td>
                  <td>
                    <button
                      className="btn-secondary"
                      style={{ padding: '4px 10px', fontSize: '0.75rem' }}
                      onClick={() => onOpenApproval(note)}
                    >
                      Review & Action
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
