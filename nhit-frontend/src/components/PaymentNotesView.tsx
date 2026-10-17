import React, { useState } from 'react';
import { Download, RefreshCw, Plus, ArrowUpDown, FileText } from 'lucide-react';
import { PaymentNote } from '../types';
import { formatINR } from './DashboardView';

interface PaymentNotesViewProps {
  paymentNotes: PaymentNote[];
  onRefresh: () => void;
  onOpenNewPaymentNote: () => void;
  onOpenUTR: (note: PaymentNote) => void;
  onToggleHold: (note: PaymentNote) => void;
  onViewVoucher: (note: PaymentNote) => void;
  searchQuery: string;
}

export const PaymentNotesView: React.FC<PaymentNotesViewProps> = ({
  paymentNotes,
  onRefresh,
  onOpenNewPaymentNote,
  onOpenUTR,
  onToggleHold,
  onViewVoucher,
  searchQuery,
}) => {
  const [sortAsc, setSortAsc] = useState(false);

  const filtered = paymentNotes.filter((p) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    const num = p.payment_note_number || p.sl_no || `PN-${p.id}`;
    return (
      num.toLowerCase().includes(q) ||
      (p.project_name || '').toLowerCase().includes(q) ||
      (p.vendor_name || '').toLowerCase().includes(q)
    );
  });

  const sorted = [...filtered].sort((a, b) => {
    const valA = a.net_payable_amount || a.payable_amount || 0;
    const valB = b.net_payable_amount || b.payable_amount || 0;
    return sortAsc ? valA - valB : valB - valA;
  });

  const exportCSV = () => {
    const headers = ['Voucher Number', 'Project', 'Vendor', 'Payable Amount', 'UTR Reference', 'Status'];
    const rows = sorted.map((p) => [
      `"${p.payment_note_number || p.sl_no || p.id}"`,
      `"${p.project_name || 'Highway Corridor'}"`,
      `"${p.vendor_name || 'EPC Vendor'}"`,
      `"${p.net_payable_amount || p.payable_amount || 0}"`,
      `"${p.utr_number || 'PENDING'}"`,
      `"${p.status}"`,
    ]);

    const csvContent = '\ufeff' + [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = 'NHIT_Payment_Disbursements.csv';
    link.click();
    link.remove();
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'PAID':
        return <span className="badge badge-approved">PAID</span>;
      case 'ACTIVE':
        return <span className="badge badge-approved">ACTIVE</span>;
      case 'HOLD':
        return <span className="badge badge-hold">ON HOLD</span>;
      default:
        return <span className="badge badge-pending">PENDING</span>;
    }
  };

  return (
    <div className="table-card">
      <div className="table-header-bar">
        <h2>
          <span>Payment Notes & Disbursement Vouchers</span>
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
          <button className="btn-primary" onClick={onOpenNewPaymentNote}>
            <Plus size={14} />
            <span>New Payment Voucher</span>
          </button>
        </div>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>Voucher #</th>
            <th>Project / Site</th>
            <th>Vendor Payee</th>
            <th className="sortable" onClick={() => setSortAsc(!sortAsc)}>
              Payable Amount <ArrowUpDown size={12} className="sort-icon" />
            </th>
            <th>Bank UTR Ref</th>
            <th>Disbursement State</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {sorted.length === 0 ? (
            <tr>
              <td colSpan={7}>
                <div className="empty-state">
                  <h4>No Payment Vouchers Found</h4>
                  <p>Issue a new voucher using the button above.</p>
                </div>
              </td>
            </tr>
          ) : (
            sorted.map((p) => {
              const voucherNo = p.payment_note_number || p.sl_no || `PN-2026-${p.id}`;
              const amount = p.net_payable_amount || p.payable_amount || 0;

              return (
                <tr key={p.id}>
                  <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                    {voucherNo}
                  </td>
                  <td>
                    <div style={{ fontWeight: 600 }}>
                      {p.project_name || 'Highway Infrastructure Corridor'}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>
                      GreenNote: {p.green_note_id || 'Linked'}
                    </div>
                  </td>
                  <td>{p.vendor_name || 'EPC Vendor'}</td>
                  <td className="mono" style={{ fontWeight: 700, color: 'var(--emerald)' }}>
                    {formatINR(amount)}
                  </td>
                  <td className="mono" style={{ color: 'var(--text-muted)' }}>
                    {p.utr_number || <span style={{ color: 'var(--amber)' }}>Pending UTR</span>}
                  </td>
                  <td>{getStatusBadge(p.status)}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '6px' }}>
                      <button
                        className="btn-secondary"
                        style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                        title="View Official Printable Voucher"
                        onClick={() => onViewVoucher(p)}
                      >
                        <FileText size={12} />
                        <span>Voucher</span>
                      </button>
                      <button
                        className="btn-secondary"
                        style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                        onClick={() => onOpenUTR(p)}
                      >
                        Record UTR
                      </button>
                      <button
                        className="btn-secondary"
                        style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                        onClick={() => onToggleHold(p)}
                      >
                        {p.status === 'HOLD' ? 'Release' : 'Hold'}
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
};
