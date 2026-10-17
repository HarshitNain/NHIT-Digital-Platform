import React, { useState } from 'react';
import { User } from '../types';
import { UserPlus, ShieldCheck, Mail, Building, Award, Search } from 'lucide-react';

interface UsersViewProps {
  users: User[];
  loading: boolean;
  onOpenCreateModal: () => void;
}

export const UsersView: React.FC<UsersViewProps> = ({ users, loading, onOpenCreateModal }) => {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredUsers = users.filter((u) => {
    const term = searchTerm.toLowerCase();
    return (
      u.name?.toLowerCase().includes(term) ||
      u.email?.toLowerCase().includes(term) ||
      u.department_name?.toLowerCase().includes(term) ||
      u.designation_name?.toLowerCase().includes(term)
    );
  });

  return (
    <div>
      <div className="view-header">
        <div>
          <h2>Identity & Access Management (RBAC)</h2>
          <p className="subtitle">Enterprise directory of sanctioned officers, accounts personnel, and project directors</p>
        </div>
        <button className="btn btn-primary" onClick={onOpenCreateModal}>
          <UserPlus size={16} />
          <span>Provision User</span>
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
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Security Protocol</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--primary-color)' }}>
            4-Tier RBAC
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Prepared → Checked → Verified → Sanctioned
          </div>
        </div>
        <div className="card">
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', marginBottom: '0.25rem' }}>Authentication</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--accent-color)' }}>
            Argon2id + JWT
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Go Crypto Microservice Encrypted
          </div>
        </div>
      </div>

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
                              justifyContent: 'center',
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
    </div>
  );
};
