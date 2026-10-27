import {
  GreenNote,
  PaymentNote,
  Project,
  Vendor,
  VendorAccount,
  User,
  Organization,
  Department,
  Designation,
  ReimbursementNote,
  ServiceHealth,
  LoginResponse,
  AccessRequest,
} from '../types';

// Target API Gateway URL for production deployments (Render)
export const PRODUCTION_API_GATEWAY = 'https://nhit-api-gateway.onrender.com';

declare const __API_BASE__: string;

export function resolveApiBaseUrl(): string {
  // 1. Injected at build time by Vite from VITE_API_GATEWAY_URL
  if (typeof __API_BASE__ !== 'undefined' && __API_BASE__ && typeof __API_BASE__ === 'string' && __API_BASE__.trim() !== '') {
    return `${__API_BASE__.replace(/\/+$/, '')}/api/v1`;
  }
  // 2. Browser runtime detection: if running on Vercel or any live host (not localhost), use Render gateway
  if (typeof window !== 'undefined' && window.location) {
    const host = window.location.hostname;
    const isLocalhost = host === 'localhost' || host === '127.0.0.1' || host === '[::1]' || host.endsWith('.local');
    if (!isLocalhost) {
      return `${PRODUCTION_API_GATEWAY}/api/v1`;
    }
  }
  // 3. Local development fallback (Vite proxy /api/* -> http://localhost:8080 or port 8000)
  return '/api/v1';
}

const BASE_URL = resolveApiBaseUrl();


let cachedToken: string | null = null;

export function getAuthToken(): string | null {
  if (cachedToken) return cachedToken;
  try {
    cachedToken = sessionStorage.getItem('nhit_session_token') || localStorage.getItem('nhit_auth_token');
    return cachedToken;
  } catch {
    return null;
  }
}

export function setAuthToken(token: string | null) {
  cachedToken = token;
  try {
    if (token) {
      sessionStorage.setItem('nhit_session_token', token);
      localStorage.setItem('nhit_auth_token', token);
    } else {
      sessionStorage.removeItem('nhit_session_token');
      localStorage.removeItem('nhit_auth_token');
    }
  } catch {
    // Graceful fallback for strict storage environments
  }
}

async function fetchJSON<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${BASE_URL}${endpoint}`;
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(options?.headers as Record<string, string> || {}),
  };

  const res = await fetch(url, {
    ...options,
    headers,
  });
  if (!res.ok) {
    if (res.status === 401 && endpoint !== '/auth/login' && endpoint !== '/auth/login-otp/verify' && token) {
      setAuthToken(null);
      localStorage.removeItem('nhit_auth_user');
      window.dispatchEvent(new CustomEvent('nhit:auth_expired', {
        detail: 'Your session has expired or authentication is invalid. Please log in again.'
      }));
    }
    let errMessage = `API Error ${res.status}`;
    try {
      const err = await res.json();
      errMessage = err.detail || err.message || errMessage;
    } catch {
      // ignore
    }
    throw new Error(errMessage);
  }
  return res.json();
}

export const api = {
  // Authentication & 2-Way OTP / SSO
  login: (credentials: { login: string; password: string; enable_2fa?: boolean }) =>
    fetchJSON<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    }),

  verifyLoginOTP: (session_temp_token: string, otp: string) =>
    fetchJSON<LoginResponse>('/auth/login-otp/verify', {
      method: 'POST',
      body: JSON.stringify({ session_temp_token, otp }),
    }),

  requestLoginOTP: (email: string) =>
    fetchJSON<LoginResponse>('/auth/login-otp/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  googleSSO: (data: { credential?: string; code?: string; email?: string; name?: string }) =>
    fetchJSON<LoginResponse>('/auth/sso/google', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  changePassword: (data: { new_password: string; confirm_password?: string; session_temp_token?: string }) =>
    fetchJSON<LoginResponse>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  validateToken: (token: string) =>
    fetchJSON<{
      valid: boolean;
      user_id: string;
      email: string;
      name: string;
      roles: string[];
      permissions: string[];
      auth_method?: string;
      expires_at: number;
    }>('/auth/validate-token', {
      method: 'POST',
      body: JSON.stringify({ token }),
    }),

  // Access Requests & Admin Approval Workflow
  submitAccessRequest: (data: {
    name: string;
    email: string;
    phone?: string;
    requested_role?: string;
    department?: string;
    reason?: string;
  }) =>
    fetchJSON<{ status: string; message: string; email: string }>('/auth/access-request', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  registerWithApprovedOTP: (data: { email: string; otp: string; password: string }) =>
    fetchJSON<LoginResponse>('/auth/register-approved', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getAccessRequests: async (status?: string) => {
    const query = status ? `?status=${encodeURIComponent(status)}` : '';
    const res = await fetchJSON<{ requests: AccessRequest[]; total: number; pending_count: number }>(`/admin/access-requests${query}`);
    return res.requests || [];
  },

  approveAccessRequest: (requestId: string, data: { granted_role: string; department?: string }) =>
    fetchJSON<{ success: boolean; status: string; message: string; activation_otp?: string }>(`/admin/access-requests/${requestId}/approve`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  rejectAccessRequest: (requestId: string, reason?: string) =>
    fetchJSON<{ success: boolean; status: string; message: string }>(`/admin/access-requests/${requestId}/reject`, {
      method: 'POST',
      body: JSON.stringify({ reason: reason || 'Administrative discretion' }),
    }),

  // System Health
  getHealth: () => fetchJSON<ServiceHealth>('/health'),
  getServices: () => fetchJSON<{ services: Record<string, { port: string; description: string }> }>('/services'),

  // GreenNotes
  getGreenNotes: async () => {
    const data = await fetchJSON<{ notes: GreenNote[] }>('/green-notes');
    return data.notes || [];
  },
  createGreenNote: (note: {
    project_name: string;
    supplier_name: string;
    base_value: number;
    gst: number;
    total_amount: number;
    remarks: string;
  }) => fetchJSON<{ success: boolean; id: string }>('/green-notes', {
    method: 'POST',
    body: JSON.stringify({ note }),
  }),
  updateGreenNoteStatus: (id: string, status: string, remarks: string) =>
    fetchJSON<{ success: boolean; status: string }>(`/green-notes/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, remarks }),
    }),

  // PaymentNotes
  getPaymentNotes: async () => {
    const data = await fetchJSON<{ notes: PaymentNote[]; payment_notes?: PaymentNote[] }>('/payment-notes');
    return data.notes || data.payment_notes || [];
  },
  createPaymentNote: (note: {
    green_note_id: string;
    vendor_name: string;
    project_name: string;
    invoice_no: string;
    gross_amount: number;
    net_payable_amount: number;
    tds_amount: number;
    status: string;
  }) => fetchJSON<{ success: boolean }>('/payment-notes', {
    method: 'POST',
    body: JSON.stringify({ note }),
  }),
  updatePaymentNoteStatus: (id: string | number, status: string, remarks: string) =>
    fetchJSON<{ success: boolean; status: string }>(`/payment-notes/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, remarks }),
    }),
  updatePaymentNoteUTR: (id: string | number, utr_number: string, payment_date: string) =>
    fetchJSON<{ success: boolean; utr_number: string }>(`/payment-notes/${id}/utr`, {
      method: 'PATCH',
      body: JSON.stringify({ utr_number, payment_date }),
    }),

  // Highway Projects
  getProjects: async (orgId: string = '550e8400-e29b-41d4-a716-446655440001') => {
    const data = await fetchJSON<{ projects: Project[] }>(`/organizations/${orgId}/projects`);
    return data.projects || [];
  },
  createProject: (project: { project_name: string; org_id: string; tenant_id?: string; created_by?: string }) =>
    fetchJSON<{ success: boolean }>('/projects', {
      method: 'POST',
      body: JSON.stringify({
        tenant_id: '550e8400-e29b-41d4-a716-446655440000',
        created_by: '550e8400-e29b-41d4-a716-446655440001',
        ...project,
      }),
    }),

  // Vendors
  getVendors: async () => {
    const data = await fetchJSON<{ vendors: Vendor[] }>('/vendors');
    return data.vendors || [];
  },
  createVendor: (vendor: {
    name: string;
    vendor_code: string;
    msme_classification: string;
    bank_name: string;
    bank_account_number: string;
    bank_ifsc: string;
  }) => fetchJSON<{ success: boolean }>('/vendors', {
    method: 'POST',
    body: JSON.stringify(vendor),
  }),

  getVendorAccounts: async (vendorId: string) => {
    const data = await fetchJSON<{ accounts: VendorAccount[]; total_count?: number }>(`/vendors/${vendorId}/accounts`);
    return data.accounts || [];
  },

  createVendorAccount: (vendorId: string, account: {
    account_name: string;
    account_number: string;
    name_of_bank: string;
    ifsc_code: string;
    account_type?: string;
    branch_name?: string;
    is_primary?: boolean;
    remarks?: string;
  }) => fetchJSON<{ success: boolean; account: VendorAccount }>(`/vendors/${vendorId}/accounts`, {
    method: 'POST',
    body: JSON.stringify({
      created_by: '550e8400-e29b-41d4-a716-446655440001',
      ...account,
    }),
  }),

  // Organizations & Departments
  getOrganizations: async () => {
    const data = await fetchJSON<{ organizations: Organization[] }>('/organizations');
    return data.organizations || [];
  },
  getDepartments: async () => {
    const data = await fetchJSON<{ departments: Department[] }>('/departments');
    return data.departments || [];
  },
  getDesignations: async () => {
    const data = await fetchJSON<{ designations: Designation[] }>('/designations');
    return data.designations || [];
  },

  // Users & RBAC
  getUsers: async () => {
    const data = await fetchJSON<{ users: User[] }>('/users');
    return data.users || [];
  },
  createUser: (user: {
    name: string;
    email: string;
    password?: string;
    must_change_password?: boolean;
    department_id?: string;
    designation_id?: string;
    org_id?: string;
    tenant_id?: string;
  }) => fetchJSON<{ success: boolean; temporary_password?: string; user_id?: string; must_change_password?: boolean }>('/users', {
    method: 'POST',
    body: JSON.stringify({
      tenant_id: '550e8400-e29b-41d4-a716-446655440000',
      org_id: '550e8400-e29b-41d4-a716-446655440001',
      ...user,
    }),
  }),

  // Reimbursements
  getReimbursements: async () => {
    const data = await fetchJSON<{ notes: ReimbursementNote[] }>('/reimbursement-notes');
    return data.notes || [];
  },
  createReimbursement: (claim: {
    purpose_of_travel: string;
    account_holder: string;
    bank_name: string;
    bank_account: string;
    ifsc_code: string;
    amount: number;
  }) => fetchJSON<{ success: boolean }>('/reimbursement-notes', {
    method: 'POST',
    body: JSON.stringify({
      note: {
        purpose_of_travel: claim.purpose_of_travel,
        account_holder: claim.account_holder,
        bank_name: claim.bank_name,
        bank_account: claim.bank_account,
        ifsc_code: claim.ifsc_code,
        status: 'D',
      },
      expenses: [
        {
          description: claim.purpose_of_travel,
          amount: claim.amount,
          expense_date: new Date().toISOString().slice(0, 10),
          category: 'Travel',
        },
      ],
    }),
  }),
};
