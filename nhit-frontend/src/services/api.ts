import {
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
  LoginResponse,
} from '../types';

const BASE_URL = '/api/v1';

export function getAuthToken(): string | null {
  return localStorage.getItem('nhit_auth_token');
}

export function setAuthToken(token: string | null) {
  if (token) {
    localStorage.setItem('nhit_auth_token', token);
  } else {
    localStorage.removeItem('nhit_auth_token');
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
    department_id?: string;
    designation_id?: string;
    org_id?: string;
    tenant_id?: string;
  }) => fetchJSON<{ success: boolean }>('/users', {
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
