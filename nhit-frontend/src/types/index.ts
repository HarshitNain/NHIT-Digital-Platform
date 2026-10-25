// TypeScript interfaces mapping directly to Go Microservices / Proto structs

export interface GreenNote {
  id: string;
  sl_no: string;
  project_id?: string;
  project_name: string;
  vendor_id?: string;
  vendor_name?: string;
  supplier_name?: string;
  amount: number;
  taxable_amount?: number;
  base_value?: number;
  gst_amount?: number;
  gst?: number;
  status: 'DRAFT' | 'PENDING' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED' | 'CANCELLED';
  created_by?: string;
  created_at?: string;
  remarks?: string;
}

export interface PaymentNote {
  id: number | string;
  payment_note_number?: string;
  sl_no?: string;
  green_note_id?: string;
  vendor_name: string;
  vendor_account_number?: string;
  project_name?: string;
  bank_name?: string;
  ifsc_code?: string;
  net_payable_amount: number;
  payable_amount?: number;
  gross_amount?: number;
  tds_amount?: number;
  status: 'ACTIVE' | 'HOLD' | 'PAID' | 'D' | 'P' | 'A';
  utr_number?: string | null;
  created_at?: string;
}

export interface Project {
  project_id: string;
  name: string;
  code: string;
  budget?: number;
  status?: string;
  org_id?: string;
}

export interface Vendor {
  id?: string;
  vendor_id: string;
  vendor_code: string;
  name: string;
  msme_classification?: string;
  bank_name?: string;
  bank_account_number?: string;
  bank_ifsc?: string;
  gstin?: string;
  pan?: string;
  is_verified?: boolean;
}

export interface VendorAccount {
  id: string;
  vendor_id: string;
  account_name: string;
  account_number: string;
  account_type?: string;
  name_of_bank: string;
  branch_name?: string;
  ifsc_code: string;
  swift_code?: string;
  is_primary?: boolean;
  is_active?: boolean;
  remarks?: string;
}

export interface User {
  user_id: string;
  name: string;
  email: string;
  department_name?: string;
  designation_name?: string;
  tenant_id?: string;
  last_login_at?: string;
  must_change_password?: boolean;
}

export interface Organization {
  org_id: string;
  name: string;
  code: string;
  description?: string;
}

export interface Department {
  id: string;
  name: string;
}

export interface Designation {
  id: string;
  name: string;
}

export interface ReimbursementNote {
  id: number | string;
  note_no?: string;
  sl_no?: string;
  user_name?: string;
  purpose_of_travel?: string;
  total_amount: number;
  bank_name?: string;
  bank_account?: string;
  ifsc_code?: string;
  status: string;
}

export interface ServiceHealth {
  status: string;
  database: string;
  version: string;
  services: Record<string, { port: string; description: string }>;
}

export interface AuthUser {
  user_id: string;
  email: string;
  name: string;
  roles: string[];
  permissions: string[];
  tenant_id?: string;
  auth_method?: 'password' | 'password_otp' | 'google_sso';
  portalMode: 'admin' | 'user';
}

export interface LoginResponse {
  status?: 'SUCCESS' | '2FA_REQUIRED' | 'PASSWORD_CHANGE_REQUIRED' | 'PENDING_APPROVAL' | 'EXISTS';
  session_temp_token?: string;
  message?: string;
  email?: string;
  expires_in?: number;
  dev_otp_preview?: string;
  token?: string;
  refresh_token?: string;
  user_id?: string;
  name?: string;
  tenant_id?: string;
  roles?: string[];
  permissions?: string[];
  auth_method?: 'password' | 'password_otp' | 'google_sso';
  token_expires_at?: number;
}

export interface AccessRequest {
  request_id: string;
  name: string;
  email: string;
  phone?: string;
  requested_role: string;
  department: string;
  reason?: string;
  status: 'PENDING_APPROVAL' | 'APPROVED' | 'REJECTED' | 'COMPLETED';
  approved_by?: string;
  approved_role?: string;
  activation_otp?: string;
  created_at: string;
}

