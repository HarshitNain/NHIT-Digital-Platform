import React from 'react';
import {
  LayoutDashboard,
  FileCheck2,
  CreditCard,
  Plane,
  Layers,
  Users2,
  Building2,
  ShieldCheck,
  Activity,
  Shield,
} from 'lucide-react';

interface SidebarProps {
  currentTab: string;
  onSelectTab: (tab: string) => void;
  badgeCounts: {
    greenNotes: number;
    paymentNotes: number;
  };
  dbStatus: string;
  isOpen: boolean;
  onClose: () => void;
  portalMode: 'admin' | 'user';
  onTogglePortal: () => void;
  userRoles?: string[];
  userPermissions?: string[];
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentTab,
  onSelectTab,
  badgeCounts,
  dbStatus,
  isOpen,
  portalMode,
  onTogglePortal,
  userRoles = [],
  userPermissions = [],
}) => {
  const isSuperAdmin = userRoles.includes('super_admin') || userPermissions.includes('*');
  const isOrgAdmin = userRoles.includes('org_admin') || userPermissions.some((p) => p.startsWith('users.') || p.startsWith('orgs.'));
  const isFinance = userRoles.includes('finance_reviewer') || userPermissions.includes('payments.process');

  const adminNavItems = [
    { section: 'Command & Control' },
    { id: 'dashboard', label: 'Executive Dashboard', icon: LayoutDashboard },
    
    { section: 'Approval Engine' },
    { id: 'greennotes', label: 'GreenNotes Registry', icon: FileCheck2, badge: badgeCounts.greenNotes },
    ...(isSuperAdmin || isFinance
      ? [{ id: 'paymentnotes', label: 'Payment Vouchers', icon: CreditCard, badge: badgeCounts.paymentNotes }]
      : []),
    { id: 'reimbursements', label: 'Claims Audit', icon: Plane },

    { section: 'Enterprise Assets' },
    { id: 'projects', label: 'Highway Corridors', icon: Layers },
    { id: 'vendors', label: 'Vendors & Banking KYC', icon: Users2 },
    ...(isSuperAdmin || isOrgAdmin
      ? [{ id: 'organizations', label: 'Hierarchy & Entities', icon: Building2 }]
      : []),

    { section: 'Administration' },
    ...(isSuperAdmin || isOrgAdmin
      ? [{ id: 'users', label: 'Identity & RBAC', icon: ShieldCheck }]
      : []),
    { id: 'telemetry', label: 'Cluster Telemetry', icon: Activity },
  ];

  const userNavItems = [
    { section: 'Officer Self-Service' },
    { id: 'userportal', label: 'My Submissions Hub', icon: LayoutDashboard },
    { id: 'greennotes', label: 'Draft / View Notes', icon: FileCheck2, badge: badgeCounts.greenNotes },
    { id: 'reimbursements', label: 'My Expense Claims', icon: Plane },
    { id: 'projects', label: 'Corridors Directory', icon: Layers },
    { id: 'vendors', label: 'Verified Vendors', icon: Users2 },
  ];

  const navItems = portalMode === 'admin' ? adminNavItems : userNavItems;

  return (
    <aside id="sidebar" className={isOpen ? 'open' : ''}>
      <div className="sidebar-header">
        <div className="brand-icon">
          <Shield size={22} />
        </div>
        <div className="brand-text">
          <h1>NHIT Portal</h1>
          <span style={{ cursor: 'pointer' }} onClick={onTogglePortal} title="Click to toggle portal mode">
            {portalMode === 'admin' ? '🛡️ Executive Admin' : '👤 Field Officer'}
          </span>
        </div>
      </div>

      <nav className="nav-sections">
        {navItems.map((item, index) => {
          if (item.section) {
            return (
              <div key={index} className="nav-label">
                {item.section}
              </div>
            );
          }

          const Icon = item.icon!;
          const isActive = currentTab === item.id;

          return (
            <div
              key={item.id}
              className={`nav-item ${isActive ? 'active' : ''}`}
              onClick={() => onSelectTab(item.id!)}
            >
              <Icon size={18} />
              <span>{item.label}</span>
              {item.badge !== undefined && item.badge > 0 && (
                <span className="nav-badge">{item.badge}</span>
              )}
            </div>
          );
        })}
      </nav>

      <div className="sidebar-footer">
        <div className="db-pulse">
          <div
            className="pulse-dot"
            style={{
              background: dbStatus === 'connected' ? 'var(--emerald)' : 'var(--amber)',
            }}
          />
          <span>{dbStatus === 'connected' ? 'nhit_db : 5433 (Live)' : 'Degraded DB'}</span>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '4px' }}>
          <span className="pill-tag admin" style={{ fontSize: '0.62rem', letterSpacing: '0.04em' }}>
            {userRoles[0]?.replace(/_/g, ' ').toUpperCase() || 'OFFICER'}
          </span>
          <span className="mono" style={{ fontSize: '0.7rem', color: 'var(--text-dim)' }}>
            PyJWT HS256
          </span>
        </div>
      </div>
    </aside>
  );
};
