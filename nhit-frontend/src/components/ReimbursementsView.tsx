import React, { useState } from 'react';
import { ReimbursementNote } from '../types';
import { PlusCircle, Search, Plane, CreditCard, Building2 } from 'lucide-react';

interface ReimbursementsViewProps {
  claims: ReimbursementNote[];
  loading: boolean;
  onOpenCreateModal: () => void;
}

export const ReimbursementsView: React.FC<ReimbursementsViewProps> = ({
  claims,
  loading,
  onOpenCreateModal,
}) => {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredClaims = claims.filter((c) => {
    const term = searchTerm.toLowerCase();
    return (
      (c.purpose_of_travel && c.purpose_of_travel.toLowerCase().includes(term)) ||
      (c.user_name && c.user_name.toLowerCase().includes(term)) ||
      (c.bank_name && c.bank_name.toLowerCase().includes(term)) ||
      (c.note_no && c.note_no.toLowerCase().includes(term))
    );
  });

  const totalSanctioned = claims.reduce((acc, c) => acc + (Number(c.total_amount) || 0), 0);

  return (
    <div>
      <div className="view-header">
        <div>
          <h2>Operational & Travel Reimbursements</h2>
          <p className="subtitle">Official corridor inspection claims, per diem, and field contingency disbursements</p>
        </div>
        <button className="btn btn-primary" onClick={onOpenCreateModal}>
          <PlusCircle size={16} />
          <span>Submit Travel Claim</span>
        </button>
      </div>

      <div className="grid grid-3" style={{ marginBottom: '1.5rem' }}>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Total Claims Filed</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--text-primary)' }}>
            {claims.length}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Site inspections & audits
          </div>
        </div>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Total Disbursed Volume</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--primary-light)' }}>
            ₹{totalSanctioned.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--success-color)', marginTop: '0.25rem' }}>
            Processed via gRPC Claims Engine
          </div>
        </div>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Compliance Policy</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--accent-color)' }}>
            NHIT Trust 2024
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Direct Bank NEFT / RTGS settlement
          </div>
        </div>
      </div>

      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div style={{ position: 'relative', width: '320px', maxWidth: '100%' }}>
            <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
            <input
              type="text"
              placeholder="Filter by purpose, officer, or bank..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input-field"
              style={{ paddingLeft: '2.25rem' }}
            />
          </div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
            Showing {filteredClaims.length} records
          </div>
        </div>

        <div className="table-container">
          <table className="data-table">
            <thead>
              <tr>
                <th>Voucher #</th>
                <th>Claimant / Purpose</th>
                <th>Bank Credential</th>
                <th>Claim Amount</th>
                <th>Audit Status</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>
                    <div className="spinner" style={{ margin: '0 auto 0.5rem' }}></div>
                    <span style={{ color: 'var(--text-muted)' }}>Loading reimbursement claims...</span>
                  </td>
                </tr>
              ) : filteredClaims.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
                    No reimbursement claims found.
                  </td>
                </tr>
              ) : (
                filteredClaims.map((c) => (
                  <tr key={c.id}>
                    <td>
                      <div style={{ fontWeight: 600, color: 'var(--primary-light)', fontFamily: 'monospace' }}>
                        {c.note_no || c.sl_no || `RN-${c.id}`}
                      </div>
                    </td>
                    <td>
                      <div>
                        <div style={{ fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                          <Plane size={14} style={{ color: 'var(--accent-color)' }} />
                          {c.purpose_of_travel || 'Highway Project Audit'}
                        </div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                          Claimant: {c.user_name || 'Designated Field Officer'}
                        </div>
                      </div>
                    </td>
                    <td>
                      <div style={{ fontSize: '0.85rem' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                          <Building2 size={13} style={{ color: 'var(--text-muted)' }} />
                          {c.bank_name || 'SBI Corporate'}
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                          <CreditCard size={12} />
                          A/C: {c.bank_account || '••••••••'} | IFSC: {c.ifsc_code || 'SBIN0001234'}
                        </div>
                      </div>
                    </td>
                    <td>
                      <div style={{ fontWeight: 700, fontSize: '0.95rem' }}>
                        ₹{Number(c.total_amount).toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                      </div>
                    </td>
                    <td>
                      <span className={`badge ${c.status === 'A' || c.status === 'APPROVED' ? 'badge-approved' : 'badge-pending'}`}>
                        {c.status === 'A' ? 'SANCTIONED' : c.status || 'SUBMITTED'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
