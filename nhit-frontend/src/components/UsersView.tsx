import React, { useState, useEffect, useCallback } from 'react';
import { User, AccessRequest } from '../types';
import {
  UserPlus,
  ShieldCheck,
  Mail,
  Building,
  Award,
  Search,
  CheckCircle,
  XCircle,
  Clock,
  ShieldAlert,
  Send,
  UserCheck,
} from 'lucide-react';
import { api } from '../services/api';

interface UsersViewProps {
  users: User[];
  loading: boolean;
  onOpenCreateModal: () => void;
}

export const UsersView: React.FC<UsersViewProps> = ({ users, loading, onOpenCreateModal }) => {
  const [activeTab, setActiveTab] = useState<'users' | 'requests'>('users');
  const [searchTerm, setSearchTerm] = useState('');

  // Access Requests State
  const [requests, setRequests] = useState<AccessRequest[]>([]);
  const [requestsLoading, setRequestsLoading] = useState(false);
  const [actionMsg, setActionMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [roleSelections, setRoleSelections] = useState<Record<string, string>>({});
  const [processingId, setProcessingId] = useState<string | null>(null);

  const fetchAccessRequests = useCallback(async () => {
    setRequestsLoading(true);
    try {
      const data = await api.getAccessRequests();
      setRequests(data);
      // Initialize role selections
      const initialRoles: Record<string, string> = {};
      data.forEach((r) => {
        initialRoles[r.request_id] = r.requested_role || 'super_admin';
      });
      setRoleSelections(initialRoles);
    } catch {
      // ignore
    } finally {
      setRequestsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAccessRequests();
  }, [fetchAccessRequests]);

  const handleApprove = async (req: AccessRequest) => {
    const selectedRole = roleSelections[req.request_id] || req.requested_role || 'super_admin';
    setProcessingId(req.request_id);
    setActionMsg(null);

    try {
      const res = await api.approveAccessRequest(req.request_id, {
        granted_role: selectedRole,
        department: req.department,
      });

      setActionMsg({
        type: 'success',
        text: `Granted access to ${req.name} (${req.email}) with role [${selectedRole.toUpperCase()}]. Activation code dispatched to applicant.`,
      });
      // Refresh requests list
      await fetchAccessRequests();
    } catch (err: any) {
      setActionMsg({
        type: 'error',
        text: err.message || 'Failed to approve access request.',
      });
    } finally {
      setProcessingId(null);
    }
  };

  const handleReject = async (req: AccessRequest) => {
    setProcessingId(req.request_id);
    setActionMsg(null);

    try {
      await api.rejectAccessRequest(req.request_id, 'Administrative review decision');
      setActionMsg({
        type: 'success',
        text: `Declined access request for ${req.name} (${req.email}).`,
      });
      await fetchAccessRequests();
    } catch (err: any) {
      setActionMsg({
        type: 'error',
        text: err.message || 'Failed to reject access request.',
      });
    } finally {
      setProcessingId(null);
    }
  };

  const filteredUsers = users.filter((u) => {
    const term = searchTerm.toLowerCase();
    return (
      u.name?.toLowerCase().includes(term) ||
      u.email?.toLowerCase().includes(term) ||
      u.department_name?.toLowerCase().includes(term) ||
      u.designation_name?.toLowerCase().includes(term)
    );
  });

  const pendingRequests = requests.filter((r) => r.status === 'PENDING_APPROVAL');

  return (
    <div>
      <div className="view-header">
        <div>
          <h2>Identity &amp; Access Governance (RBAC)</h2>
          <p className="subtitle">Enterprise directory of sanctioned officers, administrative approvals, and project directors</p>
        </div>
        <button className="btn btn-primary" onClick={onOpenCreateModal}>
          <UserPlus size={16} />
          <span>Provision User Manually</span>
        </button>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-3" style={{ marginBottom: '1.5rem' }}>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Active Personnel</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--text-primary)' }}>
            {users.length}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--success-color)', marginTop: '0.25rem' }}>
            ✓ Verified with Multi-Tenant Scopes
          </div>
        </div>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Pending Approval Requests</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: pendingRequests.length > 0 ? '#fbbf24' : 'var(--text-primary)' }}>
            {pendingRequests.length}
          </div>
          <div style={{ fontSize: '0.75rem', color: pendingRequests.length > 0 ? '#fbbf24' : 'var(--text-muted)', marginTop: '0.25rem' }}>
            {pendingRequests.length > 0 ? '⚠️ Awaiting Administrator Sanction' : '✓ All Requests Processed'}
          </div>
        </div>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Security Protocol</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--primary-color)' }}>
            Admin Gatekeeper
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Zero Auto-Provisioning &bull; 4-Tier RBAC
          </div>
        </div>
      </div>

      {/* Sub-Tabs: Active Users vs Pending Requests */}
      <div style={{
        display: 'flex',
        gap: '10px',
        borderBottom: '1px solid var(--border-color)',
        marginBottom: '1.25rem',
      }}>
        <button
          type="button"
          onClick={() => setActiveTab('users')}
          style={{
            padding: '10px 18px',
            fontSize: '14px',
            fontWeight: 600,
            border: 'none',
            background: 'none',
            color: activeTab === 'users' ? 'var(--primary-light)' : 'var(--text-muted)',
            borderBottom: activeTab === 'users' ? '2px solid var(--primary-light)' : '2px solid transparent',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <UserCheck size={16} />
          <span>Sanctioned Personnel ({users.length})</span>
        </button>

        <button
          type="button"
          onClick={() => { setActiveTab('requests'); fetchAccessRequests(); }}
          style={{
            padding: '10px 18px',
            fontSize: '14px',
            fontWeight: 600,
            border: 'none',
            background: 'none',
            color: activeTab === 'requests' ? '#fbbf24' : 'var(--text-muted)',
            borderBottom: activeTab === 'requests' ? '2px solid #fbbf24' : '2px solid transparent',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <ShieldAlert size={16} />
          <span>Pending Access Requests</span>
          {pendingRequests.length > 0 && (
            <span style={{
              background: '#fbbf24',
              color: '#0f172a',
              fontSize: '11px',
              fontWeight: 700,
              padding: '2px 7px',
              borderRadius: '9999px',
            }}>
              {pendingRequests.length}
            </span>
          )}
        </button>
      </div>

      {actionMsg && (
        <div style={{
          padding: '12px 16px',
          borderRadius: '8px',
          marginBottom: '1rem',
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          background: actionMsg.type === 'success' ? 'rgba(16, 185, 129, 0.12)' : 'rgba(239, 68, 68, 0.12)',
          border: actionMsg.type === 'success' ? '1px solid rgba(16, 185, 129, 0.3)' : '1px solid rgba(239, 68, 68, 0.3)',
          color: actionMsg.type === 'success' ? '#34d399' : '#f87171',
        }}>
          {actionMsg.type === 'success' ? <CheckCircle size={16} /> : <XCircle size={16} />}
          <span>{actionMsg.text}</span>
        </div>
      )}

      {/* ════════════════ TAB 1: SANCTIONED PERSONNEL ════════════════ */}
      {activeTab === 'users' && (
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
            <div style={{ position: 'relative', width: '320px', maxWidth: '100%' }}>
              <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
              <input
                type="text"
                placeholder="Search by name, email, or dept..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="input-field"
                style={{ paddingLeft: '2.25rem' }}
              />
            </div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
              Showing {filteredUsers.length} of {users.length} accounts
            </div>
          </div>

          <div className="table-container">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Officer / User</th>
                  <th>Email Address</th>
                  <th>Department</th>
                  <th>Designation</th>
                  <th>Tenant Scope</th>
                  <th>Access Tier</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>
                      <div className="spinner" style={{ margin: '0 auto 0.5rem' }}></div>
                      <span style={{ color: 'var(--text-muted)' }}>Querying Auth Microservice directory...</span>
                    </td>
                  </tr>
                ) : filteredUsers.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
                      No users matching criteria.
                    </td>
                  </tr>
                ) : (
                  filteredUsers.map((u) => {
                    const initials = u.name
                      ? u.name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()
                      : 'U';
                    return (
                      <tr key={u.user_id}>
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                            <div
                              style={{
                                width: '32px',
                                height: '32px',
                                borderRadius: '50%',
                                backgroundColor: 'rgba(99, 102, 241, 0.15)',
                                color: 'var(--primary-light)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyCenter: 'center',
                                fontSize: '0.75rem',
                                fontWeight: 700,
                              }}
                            >
                              {initials}
                            </div>
                            <div>
                              <div style={{ fontWeight: 600 }}>{u.name}</div>
                              <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                                ID: {u.user_id.slice(0, 8)}...
                              </div>
                            </div>
                          </div>
                        </td>
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                            <Mail size={14} style={{ color: 'var(--text-muted)' }} />
                            {u.email}
                          </div>
                        </td>
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                            <Building size={14} style={{ color: 'var(--text-muted)' }} />
                            {u.department_name || 'General Operations'}
                          </div>
                        </td>
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                            <Award size={14} style={{ color: 'var(--text-muted)' }} />
                            {u.designation_name || 'Executive Officer'}
                          </div>
                        </td>
                        <td>
                          <span style={{ fontSize: '0.75rem', fontFamily: 'monospace', color: 'var(--text-muted)' }}>
                            {u.tenant_id ? `${u.tenant_id.slice(0, 8)}...` : 'Default Trust'}
                          </span>
                        </td>
                        <td>
                          <span className="badge badge-approved" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                            <ShieldCheck size={12} />
                            Active RBAC
                          </span>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ════════════════ TAB 2: PENDING ACCESS REQUESTS ════════════════ */}
      {activeTab === 'requests' && (
        <div className="card">
          <div style={{ marginBottom: '1.25rem' }}>
            <h3 style={{ margin: '0 0 6px', fontSize: '1.1rem' }}>Pending Administrative Authorization Requests</h3>
            <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-muted)' }}>
              External visitors and new applicants who submitted an access request. Review applicant credentials, assign their authorized role, and grant platform permissions.
            </p>
          </div>

          <div className="table-container">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Applicant</th>
                  <th>Email &amp; Contact</th>
                  <th>Department</th>
                  <th>Requested Scope</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Administrative Action</th>
                </tr>
              </thead>
              <tbody>
                {requestsLoading ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>
                      <div className="spinner" style={{ margin: '0 auto 0.5rem' }}></div>
                      <span style={{ color: 'var(--text-muted)' }}>Loading pending access requests...</span>
                    </td>
                  </tr>
                ) : requests.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2.5rem', color: 'var(--text-muted)' }}>
                      <CheckCircle size={32} style={{ color: '#34d399', margin: '0 auto 8px' }} />
                      <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>No Pending Access Requests</div>
                      <div style={{ fontSize: '0.8rem', marginTop: '4px' }}>All applicants have been reviewed. Random registrations are blocked at the perimeter.</div>
                    </td>
                  </tr>
                ) : (
                  requests.map((r) => {
                    const isPending = r.status === 'PENDING_APPROVAL';
                    const isApproved = r.status === 'APPROVED' || r.status === 'COMPLETED';
                    const isRejected = r.status === 'REJECTED';

                    return (
                      <tr key={r.request_id} style={{ background: isPending ? 'rgba(245, 158, 11, 0.03)' : undefined }}>
                        <td>
                          <div style={{ fontWeight: 600 }}>{r.name}</div>
                          {r.reason && (
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', fontStyle: 'italic', maxWidth: '200px' }}>
                              &ldquo;{r.reason}&rdquo;
                            </div>
                          )}
                        </td>
                        <td>
                          <div style={{ fontSize: '0.85rem' }}>{r.email}</div>
                          {r.phone && <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{r.phone}</div>}
                        </td>
                        <td>
                          <span style={{ fontSize: '0.85rem' }}>{r.department || 'General Operations'}</span>
                        </td>
                        <td>
                          <span style={{
                            fontSize: '0.8rem',
                            fontWeight: 600,
                            color: r.requested_role.toLowerCase().includes('admin') ? '#fbbf24' : 'var(--primary-light)',
                          }}>
                            {r.requested_role}
                          </span>
                        </td>
                        <td>
                          {isPending && (
                            <span className="badge" style={{ background: 'rgba(245, 158, 11, 0.15)', color: '#fbbf24', border: '1px solid #f59e0b' }}>
                              <Clock size={11} style={{ marginRight: 4 }} />
                              Pending Approval
                            </span>
                          )}
                          {isApproved && (
                            <span className="badge badge-approved">
                              <CheckCircle size={11} style={{ marginRight: 4 }} />
                              {r.status === 'COMPLETED' ? 'Active User' : 'Approved (Code Sent)'}
                            </span>
                          )}
                          {isRejected && (
                            <span className="badge" style={{ background: 'rgba(239, 68, 68, 0.15)', color: '#f87171', border: '1px solid #ef4444' }}>
                              <XCircle size={11} style={{ marginRight: 4 }} />
                              Rejected
                            </span>
                          )}
                        </td>
                        <td style={{ textAlign: 'right' }}>
                          {isPending ? (
                            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px' }}>
                              <select
                                value={roleSelections[r.request_id] || 'super_admin'}
                                onChange={(e) => setRoleSelections({ ...roleSelections, [r.request_id]: e.target.value })}
                                style={{
                                  padding: '5px 8px',
                                  fontSize: '12px',
                                  borderRadius: '6px',
                                  background: '#1e293b',
                                  color: '#f8fafc',
                                  border: '1px solid #334155',
                                }}
                              >
                                <option value="super_admin">Role: Super Admin</option>
                                <option value="financial_controller">Role: Financial Controller</option>
                                <option value="project_engineer">Role: Project Engineer</option>
                                <option value="employee">Role: Field Officer</option>
                              </select>

                              <button
                                type="button"
                                className="btn btn-primary"
                                style={{ padding: '6px 12px', fontSize: '12px', background: '#10b981', borderColor: '#059669' }}
                                onClick={() => handleApprove(r)}
                                disabled={processingId === r.request_id}
                                title="Grants authorization and dispatches one-time activation OTP to user"
                              >
                                <Send size={12} />
                                <span>Grant &amp; Approve</span>
                              </button>

                              <button
                                type="button"
                                className="btn"
                                style={{ padding: '6px 10px', fontSize: '12px', background: 'rgba(239, 68, 68, 0.15)', color: '#f87171', border: '1px solid rgba(239, 68, 68, 0.4)' }}
                                onClick={() => handleReject(r)}
                                disabled={processingId === r.request_id}
                              >
                                Decline
                              </button>
                            </div>
                          ) : isApproved ? (
                            <span style={{ fontSize: '0.8rem', color: '#34d399', fontWeight: 500 }}>
                              Assigned: {r.approved_role || 'Admin'}
                            </span>
                          ) : (
                            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Declined</span>
                          )}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
