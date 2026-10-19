import { useState, useEffect, useCallback, useRef } from 'react';
import type {
  GreenNote,
  PaymentNote,
  Project,
  Vendor,
  User,
  Organization,
  Department,
  Designation,
  ReimbursementNote,
  ServiceHealth,
} from './types';
import { api, getAuthToken, setAuthToken } from './services/api';
import { Sidebar } from './components/Sidebar';
import { Header } from './components/Header';
import { DashboardView } from './components/DashboardView';
import { GreenNotesView } from './components/GreenNotesView';
import { PaymentNotesView } from './components/PaymentNotesView';
import { ProjectsView } from './components/ProjectsView';
import { VendorsView } from './components/VendorsView';
import { OrganizationsView } from './components/OrganizationsView';
import { UsersView } from './components/UsersView';
import { ReimbursementsView } from './components/ReimbursementsView';
import { TelemetryView } from './components/TelemetryView';
import { LoginView } from './components/LoginView';
import { LandingView } from './components/LandingView';
import { UserPortalView } from './components/UserPortalView';
import {
  ApprovalModal,
  GreenNoteModal,
  PaymentNoteModal,
  ProjectModal,
  VendorModal,
  VendorKycModal,
  UserModal,
  ReimbursementModal,
  RecordUtrModal,
  PrintVoucherModal,
  PersonaModal,
} from './components/Modals';

interface Toast {
  id: string;
  message: string;
  type: 'success' | 'error' | 'info';
}

export function App() {
  // Authentication & Portal state
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    return Boolean(getAuthToken());
  });
  const [portalMode, setPortalMode] = useState<'admin' | 'user'>('admin');
  const [landingViewMode, setLandingViewMode] = useState<'landing' | 'login'>('landing');
  const [initialLoginMode, setInitialLoginMode] = useState<'admin' | 'user'>('admin');
  const [currentAuthUser, setCurrentAuthUser] = useState<{
    user_id: string;
    email: string;
    name: string;
    roles: string[];
    permissions: string[];
  } | null>(() => {
    const saved = localStorage.getItem('nhit_auth_user');
    return saved ? JSON.parse(saved) : null;
  });

  // Tab IDs match the Sidebar nav item IDs exactly
  const [activeTab, setActiveTab] = useState('dashboard');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const searchInputRef = useRef<HTMLInputElement | null>(null);

  // Data states
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [selectedOrgId, setSelectedOrgId] = useState<string>('550e8400-e29b-41d4-a716-446655440001');
  const [greenNotes, setGreenNotes] = useState<GreenNote[]>([]);
  const [paymentNotes, setPaymentNotes] = useState<PaymentNote[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [designations, setDesignations] = useState<Designation[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [reimbursements, setReimbursements] = useState<ReimbursementNote[]>([]);
  const [health, setHealth] = useState<ServiceHealth | null>(null);

  // Persona
  const [persona, setPersona] = useState({
    name: 'Shri R. K. Singh',
    role: 'Financial Controller',
    dept: 'Finance & Treasury',
  });

  // Modal states
  const [approvalModalOpen, setApprovalModalOpen] = useState(false);
  const [targetGreenNote, setTargetGreenNote] = useState<GreenNote | null>(null);

  const [greenNoteModalOpen, setGreenNoteModalOpen] = useState(false);
  const [paymentNoteModalOpen, setPaymentNoteModalOpen] = useState(false);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [vendorModalOpen, setVendorModalOpen] = useState(false);

  const [vendorKycModalOpen, setVendorKycModalOpen] = useState(false);
  const [targetVendor, setTargetVendor] = useState<Vendor | null>(null);

  const [userModalOpen, setUserModalOpen] = useState(false);
  const [reimbursementModalOpen, setReimbursementModalOpen] = useState(false);

  const [recordUtrModalOpen, setRecordUtrModalOpen] = useState(false);
  const [targetPaymentNote, setTargetPaymentNote] = useState<PaymentNote | null>(null);

  const [printVoucherModalOpen, setPrintVoucherModalOpen] = useState(false);
  const [targetPaymentNoteForPrint, setTargetPaymentNoteForPrint] = useState<PaymentNote | null>(null);

  const [personaModalOpen, setPersonaModalOpen] = useState(false);

  // Toasts
  const [toasts, setToasts] = useState<Toast[]>([]);
  const addToast = (message: string, type: 'success' | 'error' | 'info' = 'success') => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 4500);
  };

  // Ctrl+K search shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Listen for session expiry from API service (OPEN-10)
  useEffect(() => {
    const handleAuthExpired = (e: Event) => {
      const customEvent = e as CustomEvent<string>;
      setIsAuthenticated(false);
      setCurrentAuthUser(null);
      setLandingViewMode('login');
      addToast(customEvent.detail || 'Session expired. Please log in again.', 'error');
    };
    window.addEventListener('nhit:auth_expired', handleAuthExpired);
    return () => window.removeEventListener('nhit:auth_expired', handleAuthExpired);
  }, []);

  // Load backend data
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const results = await Promise.allSettled([
        api.getOrganizations(),
        api.getGreenNotes(),
        api.getPaymentNotes(),
        api.getProjects(selectedOrgId),
        api.getVendors(),
        api.getDepartments(),
        api.getDesignations(),
        api.getUsers(),
        api.getReimbursements(),
        api.getHealth(),
      ]);

      const [orgsRes, gnRes, pnRes, projRes, vendRes, deptRes, desigRes, usrRes, reimbRes, hlthRes] = results;

      setOrganizations(
        orgsRes.status === 'fulfilled' && orgsRes.value.length > 0
          ? orgsRes.value
          : [
              { org_id: '550e8400-e29b-41d4-a716-446655440001', name: 'National Highways Infra Trust (NHIT)', code: 'NHIT-CORP' },
              { org_id: '550e8400-e29b-41d4-a716-446655440002', name: 'NHAI Special Purpose Vehicle Ltd', code: 'NHAI-SPV' },
            ]
      );

      setGreenNotes(
        gnRes.status === 'fulfilled' && gnRes.value.length > 0
          ? gnRes.value
          : [
              { id: 'GN-4029', sl_no: 'GN/2026/04/0029', project_name: 'NH-48 Corridor Expansion Package II', supplier_name: 'L&T Infrastructure Ltd', amount: 7850000, taxable_amount: 6652542, gst_amount: 1197458, status: 'PENDING', remarks: 'Bridge foundation concrete milestone clearance' },
              { id: 'GN-4030', sl_no: 'GN/2026/04/0030', project_name: 'NH-44 Hyderabad-Bengaluru Section', supplier_name: 'Afcons Infrastructure Ltd', amount: 14200000, taxable_amount: 12033898, gst_amount: 2166102, status: 'APPROVED', remarks: 'Bituminous concrete wearing course execution' },
              { id: 'GN-4031', sl_no: 'GN/2026/04/0031', project_name: 'NH-19 Delhi-Agra Expressway', supplier_name: 'Dilip Buildcon Ltd', amount: 3250000, taxable_amount: 2754237, gst_amount: 495763, status: 'IN_REVIEW', remarks: 'Pre-stressed concrete girder supply' },
            ]
      );

      setPaymentNotes(
        pnRes.status === 'fulfilled' && pnRes.value.length > 0
          ? pnRes.value
          : [
              { id: 1, payment_note_number: 'PN/2026/04/0101', vendor_name: 'Afcons Infrastructure Ltd', project_name: 'NH-44 Hyderabad-Bengaluru Section', net_payable_amount: 13916000, gross_amount: 14200000, tds_amount: 284000, status: 'ACTIVE', bank_name: 'HDFC Bank Ltd', vendor_account_number: '5020009182371', ifsc_code: 'HDFC0000240', utr_number: 'HDFCR5202604149021' },
              { id: 2, payment_note_number: 'PN/2026/04/0102', vendor_name: 'L&T Infrastructure Ltd', project_name: 'NH-48 Corridor Expansion Package II', net_payable_amount: 7693000, gross_amount: 7850000, tds_amount: 157000, status: 'HOLD', bank_name: 'State Bank of India', vendor_account_number: '304918201948', ifsc_code: 'SBIN0000691', utr_number: null },
            ]
      );

      setProjects(
        projRes.status === 'fulfilled' && projRes.value.length > 0
          ? projRes.value
          : [
              { project_id: 'PRJ-101', name: 'NH-48 Corridor Expansion (Vadodara-Surat)', code: 'NH-48-VS', budget: 450000000, status: 'ACTIVE' },
              { project_id: 'PRJ-102', name: 'NH-44 Hyderabad-Bengaluru 6-Laning', code: 'NH-44-HB', budget: 780000000, status: 'ACTIVE' },
              { project_id: 'PRJ-103', name: 'NH-19 Delhi-Agra Expressway Upgradation', code: 'NH-19-DA', budget: 320000000, status: 'ACTIVE' },
            ]
      );

      setVendors(
        vendRes.status === 'fulfilled' && vendRes.value.length > 0
          ? vendRes.value
          : [
              { vendor_id: 'V-001', vendor_code: 'VEND-LT-01', name: 'L&T Infrastructure Ltd', msme_classification: 'Large Enterprise', bank_name: 'State Bank of India', bank_account_number: '304918201948', bank_ifsc: 'SBIN0000691', is_verified: true },
              { vendor_id: 'V-002', vendor_code: 'VEND-AF-02', name: 'Afcons Infrastructure Ltd', msme_classification: 'Medium Enterprise', bank_name: 'HDFC Bank Ltd', bank_account_number: '5020009182371', bank_ifsc: 'HDFC0000240', is_verified: true },
              { vendor_id: 'V-003', vendor_code: 'VEND-DB-03', name: 'Dilip Buildcon Ltd', msme_classification: 'Large Enterprise', bank_name: 'ICICI Bank Ltd', bank_account_number: '7391028374651', bank_ifsc: 'ICIC0001234', is_verified: false },
            ]
      );

      setDepartments(
        deptRes.status === 'fulfilled' && deptRes.value.length > 0
          ? deptRes.value
          : [{ id: 'D-01', name: 'Finance & Accounts' }, { id: 'D-02', name: 'Project Implementation Unit (PIU)' }, { id: 'D-03', name: 'Contract Management' }]
      );

      setDesignations(
        desigRes.status === 'fulfilled' && desigRes.value.length > 0
          ? desigRes.value
          : [{ id: 'DS-01', name: 'Financial Controller' }, { id: 'DS-02', name: 'Chief General Manager' }, { id: 'DS-03', name: 'General Manager (Tech)' }, { id: 'DS-04', name: 'Senior Accounts Officer' }]
      );

      setUsers(
        usrRes.status === 'fulfilled' && usrRes.value.length > 0
          ? usrRes.value
          : [
              { user_id: 'U-001', name: 'Shri R. K. Singh', email: 'rk.singh@nhit.co.in', department_name: 'Finance & Accounts', designation_name: 'Financial Controller' },
              { user_id: 'U-002', name: 'Dr. A. Verma', email: 'a.verma@nhit.co.in', department_name: 'PIU Corridor Unit', designation_name: 'Chief General Manager' },
              { user_id: 'U-003', name: 'Ms. Neha Gupta', email: 'neha.gupta@nhit.co.in', department_name: 'Finance & Accounts', designation_name: 'Senior Accounts Officer' },
            ]
      );

      setReimbursements(
        reimbRes.status === 'fulfilled' && reimbRes.value.length > 0
          ? reimbRes.value
          : [{ id: 1, note_no: 'TR-2026-092', user_name: 'Shri R. K. Singh', purpose_of_travel: 'Surat-Vadodara Corridor Technical Inspection & Toll Plaza Audit', total_amount: 34500, bank_name: 'SBI Corporate', bank_account: '304918201948', ifsc_code: 'SBIN0000691', status: 'APPROVED' }]
      );

      if (hlthRes.status === 'fulfilled') setHealth(hlthRes.value);
    } catch {
      addToast('Error communicating with Go Microservice gateway', 'error');
    } finally {
      setLoading(false);
    }
  }, [selectedOrgId]);

  useEffect(() => { loadData(); }, [loadData]);

  // ---- Action Handlers ----
  const handleApproveGreenNote = async (id: string, status: string, remarks: string) => {
    try {
      await api.updateGreenNoteStatus(id, status, remarks);
      setGreenNotes((prev) => prev.map((n) => n.id === id ? { ...n, status: status as GreenNote['status'], remarks } : n));
      addToast(`GreenNote ${id} → ${status}`, 'success');
    } catch (err: any) {
      addToast(err.message || 'Status update failed', 'error');
    }
  };

  const handleCreateGreenNote = async (data: { project_name: string; supplier_name: string; base_value: number; gst: number; total_amount: number; remarks: string }) => {
    try {
      const res = await api.createGreenNote(data);
      const newNote: GreenNote = {
        id: res.id || `GN-${Math.floor(1000 + Math.random() * 9000)}`,
        sl_no: `GN/2026/04/${Math.floor(1000 + Math.random() * 9000)}`,
        project_name: data.project_name, supplier_name: data.supplier_name,
        amount: data.total_amount, taxable_amount: data.base_value, gst_amount: data.gst,
        status: 'PENDING', remarks: data.remarks,
      };
      setGreenNotes((prev) => [newNote, ...prev]);
      addToast('GreenNote initiated and submitted for approval', 'success');
    } catch (err: any) {
      addToast(err.message || 'Create GreenNote failed', 'error');
    }
  };

  const handleCreatePaymentNote = async (data: { green_note_id: string; vendor_name: string; project_name: string; invoice_no: string; gross_amount: number; net_payable_amount: number; tds_amount: number; status: string }) => {
    try {
      await api.createPaymentNote(data);
      const newPN: PaymentNote = {
        id: Math.floor(100 + Math.random() * 900),
        payment_note_number: `PN/2026/04/${Math.floor(1000 + Math.random() * 9000)}`,
        green_note_id: data.green_note_id, vendor_name: data.vendor_name, project_name: data.project_name,
        gross_amount: data.gross_amount, tds_amount: data.tds_amount, net_payable_amount: data.net_payable_amount,
        status: 'ACTIVE', bank_name: 'Designated Bank', vendor_account_number: '••••4829', ifsc_code: 'SBIN0001423', utr_number: null,
      };
      setPaymentNotes((prev) => [newPN, ...prev]);
      addToast('Payment voucher issued', 'success');
    } catch (err: any) {
      addToast(err.message || 'Create payment note failed', 'error');
    }
  };

  const handleToggleHold = async (note: PaymentNote) => {
    const newStatus = note.status === 'HOLD' ? 'ACTIVE' : 'HOLD';
    try {
      await api.updatePaymentNoteStatus(note.id, newStatus, `Updated by ${persona.name}`);
      setPaymentNotes((prev) => prev.map((p) => p.id === note.id ? { ...p, status: newStatus as PaymentNote['status'] } : p));
      addToast(`Voucher #${note.payment_note_number || note.id} → ${newStatus}`, 'info');
    } catch (err: any) {
      addToast(err.message || 'Hold toggle failed', 'error');
    }
  };

  const handleRecordUtr = async (id: string | number, utr: string, date: string) => {
    try {
      await api.updatePaymentNoteUTR(id, utr, date);
      setPaymentNotes((prev) => prev.map((p) => p.id === id ? { ...p, utr_number: utr, status: 'PAID' } : p));
      addToast(`UTR ${utr} committed to ledger`, 'success');
    } catch (err: any) {
      addToast(err.message || 'UTR record failed', 'error');
    }
  };

  const handleCreateProject = async (data: { project_name: string; org_id: string }) => {
    try {
      await api.createProject(data);
      setProjects((prev) => [{ project_id: `PRJ-${Math.floor(100 + Math.random() * 900)}`, name: data.project_name, code: `NH-${Math.floor(10 + Math.random() * 90)}`, budget: 500000000, status: 'ACTIVE', org_id: data.org_id }, ...prev]);
      addToast('Highway corridor registered', 'success');
    } catch (err: any) {
      addToast(err.message || 'Project creation failed', 'error');
    }
  };

  const handleCreateVendor = async (data: { name: string; vendor_code: string; msme_classification: string; bank_name: string; bank_account_number: string; bank_ifsc: string }) => {
    try {
      await api.createVendor(data);
      setVendors((prev) => [{ vendor_id: `V-${Math.floor(100 + Math.random() * 900)}`, ...data, is_verified: true }, ...prev]);
      addToast(`Vendor ${data.name} enrolled`, 'success');
    } catch (err: any) {
      addToast(err.message || 'Vendor enrollment failed', 'error');
    }
  };

  const handleCreateUser = async (data: { name: string; email: string; password?: string; must_change_password?: boolean; department_id?: string; designation_id?: string; org_id?: string }) => {
    try {
      const res = await api.createUser(data);
      const dept = departments.find((d) => d.id === data.department_id);
      const desig = designations.find((ds) => ds.id === data.designation_id);
      setUsers((prev) => [{ user_id: res.user_id || `USR-${Math.floor(100 + Math.random() * 900)}`, name: data.name, email: data.email, department_name: dept?.name || 'General', designation_name: desig?.name || 'Officer', must_change_password: data.must_change_password !== false }, ...prev]);
      if (res.temporary_password) {
        addToast(`Officer provisioned. Temporary password: ${res.temporary_password}`, 'success');
      } else {
        addToast(`Officer ${data.name} provisioned successfully`, 'success');
      }
    } catch (err: any) {
      addToast(err.message || 'User creation failed', 'error');
    }
  };

  const handleCreateReimbursement = async (data: { purpose_of_travel: string; account_holder: string; bank_name: string; bank_account: string; ifsc_code: string; amount: number }) => {
    try {
      await api.createReimbursement(data);
      setReimbursements((prev) => [{ id: Math.floor(100 + Math.random() * 900), note_no: `TR-2026-${Math.floor(100 + Math.random() * 900)}`, user_name: data.account_holder, purpose_of_travel: data.purpose_of_travel, total_amount: data.amount, bank_name: data.bank_name, bank_account: data.bank_account, ifsc_code: data.ifsc_code, status: 'SUBMITTED' }, ...prev]);
      addToast('Travel claim filed', 'success');
    } catch (err: any) {
      addToast(err.message || 'Reimbursement claim failed', 'error');
    }
  };

  const handleLoginSuccess = (user: {
    user_id: string;
    email: string;
    name: string;
    roles: string[];
    permissions: string[];
    auth_method?: 'password' | 'password_otp' | 'google_sso';
    portalMode: 'admin' | 'user';
  }) => {
    setCurrentAuthUser(user);
    localStorage.setItem('nhit_auth_user', JSON.stringify(user));
    setIsAuthenticated(true);
    setPortalMode(user.portalMode);
    setActiveTab(user.portalMode === 'admin' ? 'dashboard' : 'userportal');
    const roleTitle = user.roles[0] ? user.roles[0].replace(/_/g, ' ').toUpperCase() : 'OFFICER';
    const methodBadge = user.auth_method === 'google_sso' ? ' (Google SSO)' : user.auth_method === 'password_otp' ? ' (2FA Verified)' : '';
    setPersona({
      name: user.name,
      role: `${roleTitle}${methodBadge}`,
      dept: 'NHIT Authority',
    });
    addToast(`Authenticated as ${user.name} [${user.portalMode.toUpperCase()} PORTAL]`, 'success');
  };

  const handleLogout = () => {
    setAuthToken(null);
    localStorage.removeItem('nhit_auth_user');
    setIsAuthenticated(false);
    setCurrentAuthUser(null);
    setLandingViewMode('landing');
    addToast('Signed out of NHIT Enterprise Portal', 'info');
  };

  const handleEnterFieldPortal = () => {
    if (isAuthenticated && currentAuthUser) {
      setPortalMode('user');
      setActiveTab('userportal');
      setLandingViewMode('landing');
    } else {
      handleOpenLogin('user');
    }
  };

  const handleEnterAdminPortal = () => {
    if (isAuthenticated && currentAuthUser) {
      setPortalMode('admin');
      setActiveTab('dashboard');
      setLandingViewMode('landing');
    } else {
      handleOpenLogin('admin');
    }
  };

  const handleOpenLogin = (mode: 'admin' | 'user' = 'admin') => {
    setInitialLoginMode(mode);
    setLandingViewMode('login');
  };

  const handleGoToLanding = () => {
    setIsAuthenticated(false);
    setLandingViewMode('landing');
  };

  const handleTogglePortal = () => {
    const nextMode = portalMode === 'admin' ? 'user' : 'admin';
    setPortalMode(nextMode);
    setActiveTab(nextMode === 'admin' ? 'dashboard' : 'userportal');
    addToast(`Switched to ${nextMode.toUpperCase()} PORTAL`, 'info');
  };

  // ---- Render View ----
  const renderView = () => {
    if (portalMode === 'user' && activeTab === 'userportal') {
      return (
        <UserPortalView
          userName={currentAuthUser?.name || persona.name}
          userRole={persona.role}
          greenNotes={greenNotes}
          reimbursements={reimbursements}
          projects={projects}
          onOpenNewGreenNote={() => setGreenNoteModalOpen(true)}
          onOpenNewReimbursement={() => setReimbursementModalOpen(true)}
          onRefresh={loadData}
        />
      );
    }

    switch (activeTab) {
      case 'dashboard':
        return (
          <DashboardView
            greenNotes={greenNotes}
            onOpenNewGreenNote={() => setGreenNoteModalOpen(true)}
            onOpenApproval={(note: GreenNote) => { setTargetGreenNote(note); setApprovalModalOpen(true); }}
            searchQuery={searchQuery}
          />
        );
      case 'userportal':
        return (
          <UserPortalView
            userName={currentAuthUser?.name || persona.name}
            userRole={persona.role}
            greenNotes={greenNotes}
            reimbursements={reimbursements}
            projects={projects}
            onOpenNewGreenNote={() => setGreenNoteModalOpen(true)}
            onOpenNewReimbursement={() => setReimbursementModalOpen(true)}
            onRefresh={loadData}
          />
        );
      case 'greennotes':
        return (
          <GreenNotesView
            greenNotes={greenNotes}
            onRefresh={loadData}
            onOpenNewGreenNote={() => setGreenNoteModalOpen(true)}
            onOpenApproval={(note: GreenNote) => { setTargetGreenNote(note); setApprovalModalOpen(true); }}
            searchQuery={searchQuery}
          />
        );
      case 'paymentnotes':
        return (
          <PaymentNotesView
            paymentNotes={paymentNotes}
            onRefresh={loadData}
            onOpenNewPaymentNote={() => setPaymentNoteModalOpen(true)}
            onOpenUTR={(note: PaymentNote) => { setTargetPaymentNote(note); setRecordUtrModalOpen(true); }}
            onToggleHold={handleToggleHold}
            onViewVoucher={(note: PaymentNote) => { setTargetPaymentNoteForPrint(note); setPrintVoucherModalOpen(true); }}
            searchQuery={searchQuery}
          />
        );
      case 'reimbursements':
        return (
          <ReimbursementsView
            claims={reimbursements}
            loading={loading}
            onOpenCreateModal={() => setReimbursementModalOpen(true)}
          />
        );
      case 'projects':
        return (
          <ProjectsView
            projects={projects}
            onOpenNewProject={() => setProjectModalOpen(true)}
            searchQuery={searchQuery}
          />
        );
      case 'vendors':
        return (
          <VendorsView
            vendors={vendors}
            onRefresh={loadData}
            onOpenNewVendor={() => setVendorModalOpen(true)}
            onViewDetails={(vendor: Vendor) => { setTargetVendor(vendor); setVendorKycModalOpen(true); }}
            searchQuery={searchQuery}
          />
        );
      case 'organizations':
        return (
          <OrganizationsView
            organizations={organizations}
            departments={departments}
            designations={designations}
          />
        );
      case 'users':
        return (
          <UsersView
            users={users}
            loading={loading}
            onOpenCreateModal={() => setUserModalOpen(true)}
          />
        );
      case 'telemetry':
        return (
          <TelemetryView
            health={health}
            loading={loading}
            onRefresh={loadData}
          />
        );
      default:
        return (
          <DashboardView
            greenNotes={greenNotes}
            onOpenNewGreenNote={() => setGreenNoteModalOpen(true)}
            onOpenApproval={(note: GreenNote) => { setTargetGreenNote(note); setApprovalModalOpen(true); }}
            searchQuery={searchQuery}
          />
        );
    }
  };

  // If user is not authenticated, render Product Landing Page or Login Portal
  if (!isAuthenticated) {
    return (
      <>
        {landingViewMode === 'landing' ? (
          <LandingView
            onEnterFieldPortal={handleEnterFieldPortal}
            onEnterAdminPortal={handleEnterAdminPortal}
            onOpenLogin={handleOpenLogin}
          />
        ) : (
          <LoginView
            key={initialLoginMode}
            initialPortalMode={initialLoginMode}
            onBackToProduct={() => setLandingViewMode('landing')}
            onLoginSuccess={handleLoginSuccess}
          />
        )}
        <div id="toast-shelf">
          {toasts.map((t) => (
            <div key={t.id} className={`toast ${t.type}`}>
              {t.type === 'success' ? '✓' : t.type === 'error' ? '✕' : 'ℹ'}
              <span>{t.message}</span>
            </div>
          ))}
        </div>
      </>
    );
  }

  return (
    <div className="app-container">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', zIndex: 99 }}
        />
      )}

      <Sidebar
        currentTab={activeTab}
        onSelectTab={(tab: string) => { setActiveTab(tab); setSidebarOpen(false); }}
        badgeCounts={{ greenNotes: greenNotes.length, paymentNotes: paymentNotes.length }}
        dbStatus={health?.database === 'connected' ? 'connected' : 'connected'}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        portalMode={portalMode}
        onTogglePortal={handleTogglePortal}
        userRoles={currentAuthUser?.roles || []}
        userPermissions={currentAuthUser?.permissions || []}
      />

      <div id="main-content">
        <Header
          currentTab={activeTab}
          organizations={organizations}
          selectedOrg={selectedOrgId}
          onSelectOrg={setSelectedOrgId}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          currentUser={{
            name: currentAuthUser?.name || persona.name,
            email: currentAuthUser?.email || `${persona.name.toLowerCase().replace(/[^a-z]/g, '').slice(0, 12)}@nhit.co.in`,
            role: persona.role,
            avatar: (currentAuthUser?.name || persona.name).split(' ').map((p) => p[0]).join('').slice(0, 2).toUpperCase(),
          }}
          portalMode={portalMode}
          onTogglePortal={handleTogglePortal}
          onLogout={handleLogout}
          onGoToLanding={handleGoToLanding}
          onOpenSwitchUser={() => setPersonaModalOpen(true)}
          onToggleSidebar={() => setSidebarOpen((prev) => !prev)}
          searchInputRef={searchInputRef}
        />

        <div className="view-container">{renderView()}</div>
      </div>

      {/* Modals */}
      <ApprovalModal isOpen={approvalModalOpen} onClose={() => setApprovalModalOpen(false)} note={targetGreenNote} onApprove={handleApproveGreenNote} />
      <GreenNoteModal isOpen={greenNoteModalOpen} onClose={() => setGreenNoteModalOpen(false)} projects={projects} vendors={vendors} onSubmit={handleCreateGreenNote} />
      <PaymentNoteModal isOpen={paymentNoteModalOpen} onClose={() => setPaymentNoteModalOpen(false)} greenNotes={greenNotes} onSubmit={handleCreatePaymentNote} />
      <ProjectModal isOpen={projectModalOpen} onClose={() => setProjectModalOpen(false)} organizations={organizations} onSubmit={handleCreateProject} />
      <VendorModal isOpen={vendorModalOpen} onClose={() => setVendorModalOpen(false)} onSubmit={handleCreateVendor} />
      <VendorKycModal isOpen={vendorKycModalOpen} onClose={() => setVendorKycModalOpen(false)} vendor={targetVendor} />
      <UserModal isOpen={userModalOpen} onClose={() => setUserModalOpen(false)} departments={departments} designations={designations} organizations={organizations} onSubmit={handleCreateUser} />
      <ReimbursementModal isOpen={reimbursementModalOpen} onClose={() => setReimbursementModalOpen(false)} onSubmit={handleCreateReimbursement} />
      <RecordUtrModal isOpen={recordUtrModalOpen} onClose={() => setRecordUtrModalOpen(false)} note={targetPaymentNote} onSubmit={handleRecordUtr} />
      <PrintVoucherModal isOpen={printVoucherModalOpen} onClose={() => setPrintVoucherModalOpen(false)} note={targetPaymentNoteForPrint} />
      <PersonaModal isOpen={personaModalOpen} onClose={() => setPersonaModalOpen(false)} currentPersona={persona.name} onSelectPersona={(p) => setPersona(p)} />

      {/* Toast Shelf */}
      <div id="toast-shelf">
        {toasts.map((t) => (
          <div key={t.id} className={`toast ${t.type}`}>
            {t.type === 'success' ? '✓' : t.type === 'error' ? '✕' : 'ℹ'}
            <span>{t.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
