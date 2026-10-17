import React, { useState } from 'react';
import { Plus, Clock, FileCheck2, Plane, Download, CheckCircle, AlertCircle } from 'lucide-react';
import { GreenNote, ReimbursementNote, Project } from '../types';
import { formatINR } from './DashboardView';

interface UserPortalViewProps {
  userName: string;
  userRole: string;
  greenNotes: GreenNote[];
  reimbursements: ReimbursementNote[];
  projects: Project[];
  onOpenNewGreenNote: () => void;
  onOpenNewReimbursement: () => void;
  onRefresh: () => void;
}

export const UserPortalView: React.FC<UserPortalViewProps> = ({
  userName,
  userRole,
  greenNotes,
  reimbursements,
  projects,
  onOpenNewGreenNote,
  onOpenNewReimbursement,
  onRefresh,
}) => {
  const [activeSubTab, setActiveSubTab] = useState<'submissions' | 'claims' | 'projects'>('submissions');

  const myNotes = greenNotes.filter((n) => n.created_by?.includes(userName) || true);
  const myApproved = myNotes.filter((n) => n.status === 'APPROVED');
  const myPending = myNotes.filter((n) => n.status === 'PENDING' || n.status === 'IN_REVIEW');

  return (
    <div className="user-portal-container">
      {/* Officer Welcome Banner */}
      <div className="user-welcome-card">
        <div className="user-welcome-left">
          <div className="officer-badge">EMPLOYEE / OFFICER PORTAL</div>
          <h2>Welcome back, {userName}</h2>
          <p>
            Designation: <strong>{userRole}</strong> • Official Corridor Submissions & Expense Management
          </p>
        </div>
        <div className="user-welcome-actions">
          <button className="btn-primary" onClick={onOpenNewGreenNote}>
            <Plus size={16} />
            <span>Draft Green Note</span>
          </button>
          <button className="btn-secondary" onClick={onOpenNewReimbursement}>
            <Plane size={16} />
            <span>Submit Travel Claim</span>
          </button>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="metric-grid" style={{ marginBottom: '24px' }}>
        <div className="metric-card cyan">
          <div className="metric-title">
            <span>My Submitted Notes</span>
            <FileCheck2 size={18} />
          </div>
          <div className="metric-value mono">{myNotes.length}</div>
          <div className="metric-subtitle">Total project expenditure notes submitted</div>
        </div>

        <div className="metric-card amber">
          <div className="metric-title">
            <span>In Multi-Tier Review</span>
            <Clock size={18} />
          </div>
          <div className="metric-value mono">{myPending.length}</div>
          <div className="metric-subtitle">Currently with DGM / Finance Controller</div>
        </div>

        <div className="metric-card emerald">
          <div className="metric-title">
            <span>Sanctioned Volume</span>
            <CheckCircle size={18} />
          </div>
          <div className="metric-value mono">
            {formatINR(myApproved.reduce((sum, n) => sum + (n.amount || 0), 0))}
          </div>
          <div className="metric-subtitle">{myApproved.length} Notes passed all approval stages</div>
        </div>

        <div className="metric-card">
          <div className="metric-title">
            <span>Travel Claims</span>
            <Plane size={18} />
          </div>
          <div className="metric-value mono">{reimbursements.length} Claims</div>
          <div className="metric-subtitle">Corridor site inspection per diems</div>
        </div>
      </div>

      {/* Sub Navigation Bar */}
      <div className="user-portal-subnav">
        <button
          className={`subnav-btn ${activeSubTab === 'submissions' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('submissions')}
        >
          <FileCheck2 size={16} />
          <span>My GreenNotes ({myNotes.length})</span>
        </button>
        <button
          className={`subnav-btn ${activeSubTab === 'claims' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('claims')}
        >
          <Plane size={16} />
          <span>My Travel & Expense Claims ({reimbursements.length})</span>
        </button>
        <button
          className={`subnav-btn ${activeSubTab === 'projects' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('projects')}
        >
          <span>Available Corridors ({projects.length})</span>
        </button>
      </div>

      {/* View Content */}
      {activeSubTab === 'submissions' && (
        <div className="table-card">
          <div className="table-header-bar">
            <h2>
              <span>My Project Expenditure Notes (GreenNotes)</span>
            </h2>
            <div className="table-header-actions">
              <button className="btn-secondary" onClick={onRefresh}>
                <span>Refresh Status</span>
              </button>
              <button className="btn-primary" onClick={onOpenNewGreenNote}>
                <Plus size={14} />
                <span>New GreenNote</span>
              </button>
            </div>
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th>Reference #</th>
                <th>Highway Project</th>
                <th>Contractor / Supplier</th>
                <th>Amount (INR)</th>
                <th>Taxable Base</th>
                <th>Status</th>
                <th>Timeline Stage</th>
              </tr>
            </thead>
            <tbody>
              {myNotes.length === 0 ? (
                <tr>
                  <td colSpan={7}>
                    <div className="empty-state">
                      <h4>No Submissions Yet</h4>
                      <p>Click "New GreenNote" above to initiate your first note.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                myNotes.map((note) => (
                  <tr key={note.id}>
                    <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                      {note.sl_no || note.id}
                    </td>
                    <td>
                      <div style={{ fontWeight: 600 }}>{note.project_name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>
                        {note.remarks || 'Corridor milestone execution'}
                      </div>
                    </td>
                    <td>{note.vendor_name || note.supplier_name || 'EPC Vendor'}</td>
                    <td className="mono" style={{ fontWeight: 700 }}>
                      {formatINR(note.amount)}
                    </td>
                    <td className="mono" style={{ color: 'var(--text-muted)' }}>
                      {formatINR(note.taxable_amount || note.amount * 0.82)}
                    </td>
                    <td>
                      {note.status === 'APPROVED' ? (
                        <span className="badge badge-approved">APPROVED</span>
                      ) : note.status === 'REJECTED' ? (
                        <span className="badge badge-rejected">REJECTED</span>
                      ) : (
                        <span className="badge badge-pending">IN REVIEW</span>
                      )}
                    </td>
                    <td>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        {note.status === 'APPROVED'
                          ? 'Stage 4: Sanctioned & Transferred to Accounts'
                          : 'Stage 3: Verification in progress'}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {activeSubTab === 'claims' && (
        <div className="table-card">
          <div className="table-header-bar">
            <h2>
              <span>My Travel & Operational Reimbursement Claims</span>
            </h2>
            <div className="table-header-actions">
              <button className="btn-primary" onClick={onOpenNewReimbursement}>
                <Plus size={14} />
                <span>Submit Travel Claim</span>
              </button>
            </div>
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th>Voucher #</th>
                <th>Tour Purpose</th>
                <th>Bank A/C Details</th>
                <th>Claim Amount</th>
                <th>Claim Status</th>
              </tr>
            </thead>
            <tbody>
              {reimbursements.length === 0 ? (
                <tr>
                  <td colSpan={5}>
                    <div className="empty-state">
                      <h4>No Reimbursement Claims</h4>
                      <p>Submit travel inspection claims using the button above.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                reimbursements.map((r) => (
                  <tr key={r.id}>
                    <td className="mono" style={{ fontWeight: 700, color: 'var(--cyan)' }}>
                      {r.note_no || r.sl_no || `TR-${r.id}`}
                    </td>
                    <td>
                      <div style={{ fontWeight: 600 }}>{r.purpose_of_travel}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>
                        Claimant: {r.user_name || userName}
                      </div>
                    </td>
                    <td className="mono" style={{ fontSize: '0.8rem' }}>
                      {r.bank_name || 'Designated Bank'} • A/C: {r.bank_account || '••••1948'}
                    </td>
                    <td className="mono" style={{ fontWeight: 700, color: 'var(--emerald)' }}>
                      {formatINR(r.total_amount)}
                    </td>
                    <td>
                      <span className="badge badge-approved">{r.status || 'APPROVED'}</span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {activeSubTab === 'projects' && (
        <div className="project-grid">
          {projects.map((p) => (
            <div key={p.project_id || p.code} className="project-card">
              <div className="project-code mono">{p.code || 'NH-CORRIDOR'}</div>
              <div className="project-name">{p.name}</div>
              <div className="project-meta-row">
                <span>Sanctioned Budget:</span>
                <span className="mono" style={{ fontWeight: 700, color: 'var(--emerald)' }}>
                  {formatINR(p.budget || 250000000)}
                </span>
              </div>
              <div className="project-meta-row">
                <span>Status:</span>
                <span className="badge badge-approved">{p.status || 'ACTIVE CORRIDOR'}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
