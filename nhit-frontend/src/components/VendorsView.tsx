import React from 'react';
import { RefreshCw, Plus } from 'lucide-react';
import { Vendor } from '../types';

interface VendorsViewProps {
  vendors: Vendor[];
  onRefresh: () => void;
  onOpenNewVendor: () => void;
  onViewDetails: (vendor: Vendor) => void;
  searchQuery: string;
}

export const VendorsView: React.FC<VendorsViewProps> = ({
  vendors,
  onRefresh,
  onOpenNewVendor,
  onViewDetails,
  searchQuery,
}) => {
  const filtered = vendors.filter((v) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      (v.name || '').toLowerCase().includes(q) ||
      (v.vendor_code || '').toLowerCase().includes(q) ||
      (v.gstin || '').toLowerCase().includes(q) ||
      (v.pan || '').toLowerCase().includes(q)
    );
  });

  return (
    <div className="table-card">
      <div className="table-header-bar">
        <h2>
          <span>Vendor Directory & KYC Bank Details</span>
        </h2>
        <div className="table-header-actions">
          <button className="btn-secondary" onClick={onRefresh}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </button>
          <button className="btn-primary" onClick={onOpenNewVendor}>
            <Plus size={14} />
            <span>Onboard New Vendor</span>
          </button>
        </div>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>Vendor Code</th>
            <th>Company Name</th>
            <th>MSME Category</th>
            <th>Banking & IFSC</th>
            <th>Tax ID (GST/PAN)</th>
            <th>KYC Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {filtered.length === 0 ? (
            <tr>
              <td colSpan={7}>
                <div className="empty-state">
                  <h4>No Vendors Found</h4>
                  <p>Register a new vendor using the button above.</p>
                </div>
              </td>
            </tr>
          ) : (
            filtered.map((v) => (
              <tr key={v.vendor_id || v.vendor_code}>
                <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                  {v.vendor_code}
                </td>
                <td style={{ fontWeight: 600 }}>{v.name}</td>
                <td>
                  <span
                    className="badge"
                    style={{ background: 'rgba(99,102,241,0.15)', color: 'var(--primary-light)' }}
                  >
                    {v.msme_classification || 'Large EPC'}
                  </span>
                </td>
                <td>
                  <div className="mono" style={{ fontSize: '0.8rem', fontWeight: 600 }}>
                    {v.bank_account_number || 'A/C Verified'}
                  </div>
                  <div className="mono" style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>
                    {v.bank_name || 'Bank'} • {v.bank_ifsc || 'IFSC'}
                  </div>
                </td>
                <td className="mono" style={{ fontSize: '0.8rem' }}>
                  GST: {v.gstin || '27AABCL1234F1Z8'}
                  <br />
                  PAN: {v.pan || 'AABCL1234F'}
                </td>
                <td>
                  <span className="badge badge-approved">KYC VALID</span>
                </td>
                <td>
                  <button
                    className="btn-secondary"
                    style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                    onClick={() => onViewDetails(v)}
                  >
                    Banking Details
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
