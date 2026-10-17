import React from 'react';
import { Search, Building, FileText, BookOpen, Menu, Globe } from 'lucide-react';
import { Organization } from '../types';

interface HeaderProps {
  currentTab: string;
  organizations: Organization[];
  selectedOrg: string;
  onSelectOrg: (orgId: string) => void;
  searchQuery: string;
  onSearchChange: (q: string) => void;
  currentUser: {
    name: string;
    email: string;
    role: string;
    avatar: string;
  };
  portalMode: 'admin' | 'user';
  onTogglePortal: () => void;
  onLogout: () => void;
  onGoToLanding?: () => void;
  onOpenSwitchUser: () => void;
  onToggleSidebar: () => void;
  searchInputRef: React.RefObject<HTMLInputElement | null>;
}

export const Header: React.FC<HeaderProps> = ({
  currentTab,
  organizations,
  selectedOrg,
  onSelectOrg,
  searchQuery,
  onSearchChange,
  currentUser,
  portalMode,
  onTogglePortal,
  onLogout,
  onGoToLanding,
  onOpenSwitchUser,
  onToggleSidebar,
  searchInputRef,
}) => {
  const titles: Record<string, string> = {
    dashboard: 'Executive Command Center',
    userportal: 'Officer Self-Service User Portal',
    greennotes: 'GreenNotes Approval Hub',
    paymentnotes: 'Payment Notes & Disbursement Vouchers',
    reimbursements: 'Travel & Operational Reimbursements',
    projects: 'National Highway Infrastructure Corridors',
    vendors: 'Vendor Directory & Banking KYC',
    organizations: 'Corporate Trust Entities & Departments',
    users: 'Identity & Role-Based Access Control',
    telemetry: 'System Health & Microservices Pulse',
  };

  return (
    <header id="top-bar">
      <div className="top-left">
        <button
          className="mobile-menu-btn"
          onClick={onToggleSidebar}
          style={{ background: 'none', border: 'none', color: '#fff', cursor: 'pointer', display: 'none' }}
        >
          <Menu size={22} />
        </button>
        <span className="page-title">{titles[currentTab] || 'NHIT Portal'}</span>

        <div className="org-select-pill">
          <Building size={14} />
          <select value={selectedOrg} onChange={(e) => onSelectOrg(e.target.value)}>
            <option value="all">All NHIT Trusts</option>
            {organizations.map((org) => (
              <option key={org.org_id} value={org.org_id}>
                {org.name}
              </option>
            ))}
          </select>
        </div>

        {/* Portal Switch Pill */}
        <button
          className="portal-switch-pill"
          onClick={onTogglePortal}
          title={`Currently in ${portalMode.toUpperCase()} mode. Click to switch.`}
        >
          <span>Switch to {portalMode === 'admin' ? 'User Portal' : 'Admin Portal'}</span>
        </button>
      </div>

      <div className="top-right">
        <div className="search-input-box">
          <Search size={16} className="search-icon" />
          <input
            ref={searchInputRef}
            type="text"
            placeholder="Global Search..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
          <span className="search-kbd-hint mono">Ctrl+K</span>
        </div>

        {onGoToLanding && (
          <button
            className="top-action-btn"
            onClick={onGoToLanding}
            title="Return to NHIT Product Overview"
            style={{ color: 'var(--primary-light)', borderColor: 'rgba(99, 102, 241, 0.35)', background: 'rgba(99, 102, 241, 0.1)' }}
          >
            <Globe size={15} />
            <span>Product Info</span>
          </button>
        )}

        <a href="http://localhost:8080/swagger" target="_blank" rel="noreferrer" className="top-action-btn" title="Open Swagger Documentation on Port 8080">
          <FileText size={15} />
          <span>Swagger 8080</span>
        </a>

        <a href="http://localhost:8080/redoc" target="_blank" rel="noreferrer" className="top-action-btn" title="Open ReDoc API Reference">
          <BookOpen size={15} />
          <span>ReDoc</span>
        </a>

        <div className="user-profile-btn" onClick={onOpenSwitchUser} title="Click to simulate officer role">
          <div className="user-avatar">{currentUser.avatar}</div>
          <div className="user-meta">
            <div className="name">{currentUser.name}</div>
            <div className="role">{currentUser.role}</div>
          </div>
        </div>

        <button
          className="btn-secondary"
          style={{ padding: '6px 12px', fontSize: '0.8rem' }}
          onClick={onLogout}
          title="Sign out of NHIT Portal"
        >
          Sign Out
        </button>
      </div>
    </header>
  );
};

