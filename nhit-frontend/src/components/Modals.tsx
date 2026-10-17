import React, { useState } from 'react';
import { GreenNote, PaymentNote, Project, Vendor, Organization, Department, Designation } from '../types';
import { X, CheckCircle, AlertTriangle, Printer, Shield, Building, UserCheck } from 'lucide-react';

interface BaseModalProps {
  isOpen: boolean;
  onClose: () => void;
}

// ---------------- 1. 4-TIER APPROVAL STEPPER MODAL ----------------
interface ApprovalModalProps extends BaseModalProps {
  note: GreenNote | null;
  onApprove: (id: string, status: string, remarks: string) => Promise<void>;
}

export const ApprovalModal: React.FC<ApprovalModalProps> = ({ isOpen, onClose, note, onApprove }) => {
  const [remarks, setRemarks] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen || !note) return null;

  const currentStatus = note.status;
  const isApproved = currentStatus === 'APPROVED';

  const handleAction = async (newStatus: string) => {
    if (!remarks.trim() && newStatus !== 'APPROVED') {
      alert('Please enter remarks for audit compliance.');
      return;
    }
    setSubmitting(true);
    try {
      await onApprove(note.id, newStatus, remarks || 'Sanctioned under NHIT Delegated Financial Powers 2024');
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  const steps = [
    { title: 'Prepared', role: 'Project Engineer', done: true, current: false },
    { title: 'Checked', role: 'Dy. General Manager', done: true, current: false },
    { title: 'Verified', role: 'Financial Controller', done: isApproved, current: !isApproved },
    { title: 'Sanctioned', role: 'CGM / Trustee Board', done: isApproved, current: false },
  ];

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '640px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Shield size={20} style={{ color: 'var(--primary-light)' }} />
            <h3 style={{ margin: 0 }}>4-Tier Statutory Approval Stepper</h3>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        {/* Stepper Visual */}
        <div className="stepper-container" style={{ margin: '1rem 0 1.5rem' }}>
          {steps.map((st, i) => (
            <div key={st.title} className="stepper-step">
              <div className={`stepper-node ${st.done ? 'done' : st.current ? 'current' : 'pending'}`}>
                {st.done ? '✓' : i + 1}
              </div>
              <div className="stepper-label">{st.title}</div>
              <div className="stepper-sublabel">{st.role}</div>
            </div>
          ))}
        </div>

        {/* Note Snapshot */}
        <div style={{ background: 'var(--bg-card-subtle)', padding: '1rem', borderRadius: 'var(--radius-sm)', marginBottom: '1rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>GreenNote Ref:</span>
            <span style={{ fontWeight: 600, fontFamily: 'monospace', color: 'var(--primary-light)' }}>{note.sl_no || note.id}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Corridor Project:</span>
            <span style={{ fontWeight: 600 }}>{note.project_name}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Contractor / Vendor:</span>
            <span>{note.supplier_name || note.vendor_name || 'Designated EPC Vendor'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', paddingTop: '0.5rem', borderTop: '1px solid var(--border-color)' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.9rem', fontWeight: 600 }}>Sanctioned Gross Value:</span>
            <span style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--success-color)' }}>
              ₹{Number(note.amount || note.taxable_amount || 0).toLocaleString('en-IN', { maximumFractionDigits: 2 })}
            </span>
          </div>
        </div>

        {/* Remarks Input */}
        <div style={{ marginBottom: '1.25rem' }}>
          <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, marginBottom: '0.4rem', color: 'var(--text-secondary)' }}>
            Audit & Compliance Remarks <span style={{ color: 'var(--text-muted)' }}>(Recorded in Go microservice audit log)</span>
          </label>
          <textarea
            className="input-field"
            rows={3}
            placeholder="e.g. Verified against milestone completion certificate and bill of quantities. Recommended for sanction."
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
          />
        </div>

        {/* Action Buttons */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
          <button
            className="btn btn-secondary"
            onClick={() => handleAction('REJECTED')}
            disabled={submitting}
            style={{ color: 'var(--danger-color)' }}
          >
            Reject Note
          </button>
          <button
            className="btn btn-primary"
            onClick={() => handleAction('APPROVED')}
            disabled={submitting}
          >
            {submitting ? 'Sanctioning...' : 'Approve & Sanction Note'}
          </button>
        </div>
      </div>
    </div>
  );
};

// ---------------- 2. CREATE GREENNOTE MODAL (WITH LIVE GST MATH) ----------------
interface GreenNoteModalProps extends BaseModalProps {
  projects: Project[];
  vendors: Vendor[];
  onSubmit: (data: {
    project_name: string;
    supplier_name: string;
    base_value: number;
    gst: number;
    total_amount: number;
    remarks: string;
  }) => Promise<void>;
}

export const GreenNoteModal: React.FC<GreenNoteModalProps> = ({ isOpen, onClose, projects, vendors, onSubmit }) => {
  const [projectName, setProjectName] = useState(projects[0]?.name || 'NH-48 Corridor Expansion');
  const [vendorName, setVendorName] = useState(vendors[0]?.name || 'L&T Infrastructure Ltd');
  const [baseValue, setBaseValue] = useState<number>(5000000);
  const [gstRate, setGstRate] = useState<number>(18);
  const [remarks, setRemarks] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const gstAmount = (baseValue * gstRate) / 100;
  const totalAmount = baseValue + gstAmount;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        project_name: projectName,
        supplier_name: vendorName,
        base_value: baseValue,
        gst: gstAmount,
        total_amount: totalAmount,
        remarks: remarks || `Invoice note generated for ${projectName}`,
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '620px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: 'var(--success-color)' }} />
            <h3 style={{ margin: 0 }}>Initiate GreenNote Sanction Request</h3>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Highway Corridor Project</label>
            <select
              className="select-field"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              required
            >
              {projects.length > 0 ? (
                projects.map((p) => (
                  <option key={p.project_id} value={p.name}>
                    {p.name} ({p.code})
                  </option>
                ))
              ) : (
                <option value="NH-48 Corridor Expansion">NH-48 Corridor Expansion</option>
              )}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Contractor / Vendor</label>
            <select
              className="select-field"
              value={vendorName}
              onChange={(e) => setVendorName(e.target.value)}
              required
            >
              {vendors.length > 0 ? (
                vendors.map((v) => (
                  <option key={v.vendor_id} value={v.name}>
                    {v.name} ({v.vendor_code})
                  </option>
                ))
              ) : (
                <option value="L&T Infrastructure Ltd">L&T Infrastructure Ltd</option>
              )}
            </select>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Base Taxable Value (₹)</label>
              <input
                type="number"
                min="100"
                step="100"
                className="input-field"
                value={baseValue}
                onChange={(e) => setBaseValue(Number(e.target.value))}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">GST Bracket</label>
              <select
                className="select-field"
                value={gstRate}
                onChange={(e) => setGstRate(Number(e.target.value))}
              >
                <option value={18}>18% (Standard EPC)</option>
                <option value={12}>12% (Concessionaire)</option>
                <option value={5}>5% (Consultancy)</option>
                <option value={0}>0% (Exempt)</option>
              </select>
            </div>
          </div>

          {/* Auto Computed Summary Card */}
          <div style={{ background: 'var(--bg-card-subtle)', padding: '1rem', borderRadius: 'var(--radius-sm)', margin: '1rem 0' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.35rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Base Amount:</span>
              <span>₹{baseValue.toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.35rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Computed GST ({gstRate}%):</span>
              <span style={{ color: 'var(--accent-color)' }}>+ ₹{gstAmount.toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1rem', fontWeight: 700, paddingTop: '0.5rem', borderTop: '1px solid var(--border-color)' }}>
              <span>Total Gross Sanction:</span>
              <span style={{ color: 'var(--success-color)' }}>₹{totalAmount.toLocaleString('en-IN')}</span>
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Executive Scope & Milestones</label>
            <textarea
              className="input-field"
              rows={2}
              placeholder="e.g. Pavement quality concrete pouring milestone clearance for Ch. 42+000 to 58+000"
              value={remarks}
              onChange={(e) => setRemarks(e.target.value)}
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Transmitting to Go Microservice...' : 'Initiate Note'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 3. CREATE PAYMENT NOTE MODAL ----------------
interface PaymentNoteModalProps extends BaseModalProps {
  greenNotes: GreenNote[];
  onSubmit: (data: {
    green_note_id: string;
    vendor_name: string;
    project_name: string;
    invoice_no: string;
    gross_amount: number;
    net_payable_amount: number;
    tds_amount: number;
    status: string;
  }) => Promise<void>;
}

export const PaymentNoteModal: React.FC<PaymentNoteModalProps> = ({ isOpen, onClose, greenNotes, onSubmit }) => {
  const approvedNotes = greenNotes.filter((n) => n.status === 'APPROVED');
  const initialNote = approvedNotes[0] || greenNotes[0];

  const [selectedNoteId, setSelectedNoteId] = useState(initialNote?.id || '');
  const [invoiceNo, setInvoiceNo] = useState(`INV-NHIT-${Math.floor(1000 + Math.random() * 9000)}`);
  const [tdsRate, setTdsRate] = useState<number>(2); // 2% TDS under section 194C
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const targetNote = greenNotes.find((n) => n.id === selectedNoteId) || initialNote;
  const grossAmount = targetNote ? Number(targetNote.amount || targetNote.taxable_amount || 1000000) : 1000000;
  const tdsAmount = (grossAmount * tdsRate) / 100;
  const netAmount = grossAmount - tdsAmount;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        green_note_id: targetNote?.id || 'GN-1001',
        vendor_name: targetNote?.supplier_name || targetNote?.vendor_name || 'Designated Vendor',
        project_name: targetNote?.project_name || 'Corridor EPC',
        invoice_no: invoiceNo,
        gross_amount: grossAmount,
        tds_amount: tdsAmount,
        net_payable_amount: netAmount,
        status: 'ACTIVE',
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '620px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>Generate Payment Disbursement Voucher</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Select Sanctioned GreenNote</label>
            <select
              className="select-field"
              value={selectedNoteId}
              onChange={(e) => setSelectedNoteId(e.target.value)}
              required
            >
              {greenNotes.map((n) => (
                <option key={n.id} value={n.id}>
                  {n.sl_no || n.id} — {n.project_name} [₹{Number(n.amount || 0).toLocaleString('en-IN')}] ({n.status})
                </option>
              ))}
            </select>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Vendor / Concessionaire</label>
              <input
                type="text"
                className="input-field"
                value={targetNote?.supplier_name || targetNote?.vendor_name || 'Designated Vendor'}
                readOnly
                style={{ opacity: 0.8 }}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Vendor Tax Invoice #</label>
              <input
                type="text"
                className="input-field"
                value={invoiceNo}
                onChange={(e) => setInvoiceNo(e.target.value)}
                required
              />
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Gross Sanction Amount</label>
              <input
                type="text"
                className="input-field"
                value={`₹${grossAmount.toLocaleString('en-IN')}`}
                readOnly
                style={{ opacity: 0.8 }}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Statutory TDS Rate (Sec 194C)</label>
              <select
                className="select-field"
                value={tdsRate}
                onChange={(e) => setTdsRate(Number(e.target.value))}
              >
                <option value={2}>2% (Corporate Contractor)</option>
                <option value={1}>1% (Individual/HUF)</option>
                <option value={10}>10% (Technical / Consultancy)</option>
                <option value={0}>0% (Exempt Certificate)</option>
              </select>
            </div>
          </div>

          {/* Computed Breakdown */}
          <div style={{ background: 'var(--bg-card-subtle)', padding: '1rem', borderRadius: 'var(--radius-sm)', margin: '1rem 0' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.35rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Gross Claim:</span>
              <span>₹{grossAmount.toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.35rem' }}>
              <span style={{ color: 'var(--text-muted)' }}>Less: TDS Deduction ({tdsRate}%):</span>
              <span style={{ color: 'var(--danger-color)' }}>- ₹{tdsAmount.toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1rem', fontWeight: 700, paddingTop: '0.5rem', borderTop: '1px solid var(--border-color)' }}>
              <span>Net Bank Remittance:</span>
              <span style={{ color: 'var(--primary-light)' }}>₹{netAmount.toLocaleString('en-IN')}</span>
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Generating Voucher...' : 'Issue Payment Voucher'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 4. CREATE PROJECT MODAL ----------------
interface ProjectModalProps extends BaseModalProps {
  organizations: Organization[];
  onSubmit: (data: { project_name: string; org_id: string }) => Promise<void>;
}

export const ProjectModal: React.FC<ProjectModalProps> = ({ isOpen, onClose, organizations, onSubmit }) => {
  const [name, setName] = useState('');
  const [orgId, setOrgId] = useState(organizations[0]?.org_id || '550e8400-e29b-41d4-a716-446655440001');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({ project_name: name, org_id: orgId });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '520px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>Register Highway Corridor Asset</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Corridor / Highway Project Name</label>
            <input
              type="text"
              className="input-field"
              placeholder="e.g. NH-66 Mumbai-Goa 4-Laning Package-IV"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label">Governing Trust / SPV Entity</label>
            <select
              className="select-field"
              value={orgId}
              onChange={(e) => setOrgId(e.target.value)}
              required
            >
              {organizations.map((org) => (
                <option key={org.org_id} value={org.org_id}>
                  {org.name} ({org.code})
                </option>
              ))}
            </select>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Registering...' : 'Sanction Corridor Asset'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 5. CREATE VENDOR MODAL ----------------
interface VendorModalProps extends BaseModalProps {
  onSubmit: (data: {
    name: string;
    vendor_code: string;
    msme_classification: string;
    bank_name: string;
    bank_account_number: string;
    bank_ifsc: string;
  }) => Promise<void>;
}

export const VendorModal: React.FC<VendorModalProps> = ({ isOpen, onClose, onSubmit }) => {
  const [name, setName] = useState('');
  const [code, setCode] = useState(`VEND-${Math.floor(100 + Math.random() * 900)}`);
  const [msme, setMsme] = useState('Medium Enterprise');
  const [bankName, setBankName] = useState('State Bank of India');
  const [accountNo, setAccountNo] = useState('');
  const [ifsc, setIfsc] = useState('SBIN0001423');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        name,
        vendor_code: code,
        msme_classification: msme,
        bank_name: bankName,
        bank_account_number: accountNo,
        bank_ifsc: ifsc,
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '580px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>Enroll Vendor / EPC Contractor</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Contractor Legal Entity Name</label>
              <input
                type="text"
                className="input-field"
                placeholder="e.g. Afcons Infrastructure Ltd"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">Vendor Code</label>
              <input
                type="text"
                className="input-field"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">MSME Classification</label>
            <select
              className="select-field"
              value={msme}
              onChange={(e) => setMsme(e.target.value)}
            >
              <option value="Micro Enterprise">Micro Enterprise (&lt; ₹1 Cr)</option>
              <option value="Small Enterprise">Small Enterprise (&lt; ₹10 Cr)</option>
              <option value="Medium Enterprise">Medium Enterprise (&lt; ₹50 Cr)</option>
              <option value="Large / Non-MSME">Large Enterprise / Corporate EPC</option>
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Designated Settlement Bank</label>
            <input
              type="text"
              className="input-field"
              value={bankName}
              onChange={(e) => setBankName(e.target.value)}
              required
            />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Bank Account Number</label>
              <input
                type="text"
                className="input-field"
                placeholder="e.g. 98765432109876"
                value={accountNo}
                onChange={(e) => setAccountNo(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">Bank IFSC</label>
              <input
                type="text"
                className="input-field"
                value={ifsc}
                onChange={(e) => setIfsc(e.target.value)}
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Enrolling...' : 'Register Vendor'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 6. VENDOR KYC DETAILS MODAL ----------------
interface VendorKycModalProps extends BaseModalProps {
  vendor: Vendor | null;
}

export const VendorKycModal: React.FC<VendorKycModalProps> = ({ isOpen, onClose, vendor }) => {
  if (!isOpen || !vendor) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '540px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <CheckCircle size={20} style={{ color: 'var(--success-color)' }} />
            <h3 style={{ margin: 0 }}>Statutory KYC & Banking Profile</h3>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <div style={{ marginBottom: '1.25rem' }}>
          <h4 style={{ fontSize: '1.1rem', margin: '0 0 0.25rem', color: 'var(--text-primary)' }}>{vendor.name}</h4>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontFamily: 'monospace' }}>Code: {vendor.vendor_code}</span>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', background: 'var(--bg-card-subtle)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>MSME Classification:</span>
            <span className="badge badge-active">{vendor.msme_classification || 'Corporate EPC'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Bank Entity:</span>
            <span style={{ fontWeight: 600 }}>{vendor.bank_name || 'HDFC Bank Ltd'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Account Number:</span>
            <span style={{ fontFamily: 'monospace' }}>{vendor.bank_account_number || '3482910482910'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>IFSC Clearing Code:</span>
            <span style={{ fontFamily: 'monospace', color: 'var(--primary-light)' }}>{vendor.bank_ifsc || 'HDFC0000240'}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>PAN Authentication:</span>
            <span style={{ fontFamily: 'monospace' }}>AAACH8392M (Verified)</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>GSTIN Status:</span>
            <span style={{ color: 'var(--success-color)', fontWeight: 600 }}>07AAACH8392M1ZK (Active)</span>
          </div>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.25rem' }}>
          <button className="btn btn-secondary" onClick={onClose}>
            Close Profile
          </button>
        </div>
      </div>
    </div>
  );
};

// ---------------- 7. PROVISION USER MODAL ----------------
interface UserModalProps extends BaseModalProps {
  departments: Department[];
  designations: Designation[];
  organizations: Organization[];
  onSubmit: (data: {
    name: string;
    email: string;
    department_id?: string;
    designation_id?: string;
    org_id?: string;
  }) => Promise<void>;
}

export const UserModal: React.FC<UserModalProps> = ({
  isOpen,
  onClose,
  departments,
  designations,
  organizations,
  onSubmit,
}) => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [deptId, setDeptId] = useState(departments[0]?.id || '');
  const [desigId, setDesigId] = useState(designations[0]?.id || '');
  const [orgId, setOrgId] = useState(organizations[0]?.org_id || '');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        name,
        email,
        department_id: deptId || undefined,
        designation_id: desigId || undefined,
        org_id: orgId || undefined,
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '540px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>Provision Staff Officer (IAM)</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Full Name of Officer</label>
            <input
              type="text"
              className="input-field"
              placeholder="e.g. Rajiv Menon"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label">Official Email Address</label>
            <input
              type="email"
              className="input-field"
              placeholder="e.g. rajiv.menon@nhit.co.in"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Department</label>
              <select className="select-field" value={deptId} onChange={(e) => setDeptId(e.target.value)}>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Designation</label>
              <select className="select-field" value={desigId} onChange={(e) => setDesigId(e.target.value)}>
                {designations.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Organization Scope</label>
            <select className="select-field" value={orgId} onChange={(e) => setOrgId(e.target.value)}>
              {organizations.map((o) => (
                <option key={o.org_id} value={o.org_id}>
                  {o.name}
                </option>
              ))}
            </select>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Provisioning...' : 'Provision Officer'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 8. SUBMIT REIMBURSEMENT MODAL ----------------
interface ReimbursementModalProps extends BaseModalProps {
  onSubmit: (data: {
    purpose_of_travel: string;
    account_holder: string;
    bank_name: string;
    bank_account: string;
    ifsc_code: string;
    amount: number;
  }) => Promise<void>;
}

export const ReimbursementModal: React.FC<ReimbursementModalProps> = ({ isOpen, onClose, onSubmit }) => {
  const [purpose, setPurpose] = useState('');
  const [holder, setHolder] = useState('Shri R. K. Singh');
  const [bank, setBank] = useState('State Bank of India');
  const [account, setAccount] = useState('32918392019');
  const [ifsc, setIfsc] = useState('SBIN0000691');
  const [amount, setAmount] = useState<number>(24500);
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        purpose_of_travel: purpose || 'Official Corridor Inspection',
        account_holder: holder,
        bank_name: bank,
        bank_account: account,
        ifsc_code: ifsc,
        amount: amount,
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '540px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>File Travel / Contingency Claim</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Purpose of Travel / Inspection</label>
            <input
              type="text"
              className="input-field"
              placeholder="e.g. Toll plaza technical audit & bridge structural inspection"
              value={purpose}
              onChange={(e) => setPurpose(e.target.value)}
              required
            />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Officer / Account Holder</label>
              <input
                type="text"
                className="input-field"
                value={holder}
                onChange={(e) => setHolder(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">Claim Amount (₹)</label>
              <input
                type="number"
                min="100"
                step="50"
                className="input-field"
                value={amount}
                onChange={(e) => setAmount(Number(e.target.value))}
                required
              />
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Settlement Bank Account</label>
              <input
                type="text"
                className="input-field"
                value={account}
                onChange={(e) => setAccount(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">IFSC Code</label>
              <input
                type="text"
                className="input-field"
                value={ifsc}
                onChange={(e) => setIfsc(e.target.value)}
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Filing Claim...' : 'Submit Claim Voucher'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 9. RECORD UTR MODAL ----------------
interface RecordUtrModalProps extends BaseModalProps {
  note: PaymentNote | null;
  onSubmit: (id: string | number, utr: string, date: string) => Promise<void>;
}

export const RecordUtrModal: React.FC<RecordUtrModalProps> = ({ isOpen, onClose, note, onSubmit }) => {
  const [utrNumber, setUtrNumber] = useState(`CMS${Math.floor(1000000000 + Math.random() * 9000000000)}`);
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().slice(0, 10));
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen || !note) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!utrNumber.trim()) {
      alert('Please enter a valid bank UTR number');
      return;
    }
    setSubmitting(true);
    try {
      await onSubmit(note.id, utrNumber, paymentDate);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '500px' }}>
        <div className="modal-header">
          <h3 style={{ margin: 0 }}>Record Bank Settlement UTR</h3>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div style={{ background: 'var(--bg-card-subtle)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', marginBottom: '1rem' }}>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Disbursement To:</div>
            <div style={{ fontWeight: 600, fontSize: '0.95rem' }}>{note.vendor_name}</div>
            <div style={{ fontSize: '0.85rem', color: 'var(--primary-light)', fontWeight: 600, marginTop: '0.25rem' }}>
              Amount: ₹{Number(note.net_payable_amount || note.payable_amount || 0).toLocaleString('en-IN')}
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Bank UTR / Transaction Reference (NEFT/RTGS)</label>
            <input
              type="text"
              className="input-field"
              placeholder="e.g. SBIN4202409180029"
              value={utrNumber}
              onChange={(e) => setUtrNumber(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label">Settlement Clearance Date</label>
            <input
              type="date"
              className="input-field"
              value={paymentDate}
              onChange={(e) => setPaymentDate(e.target.value)}
              required
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? 'Updating...' : 'Commit Bank UTR'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// ---------------- 10. PRINTABLE OFFICIAL PAYMENT ADVICE VOUCHER MODAL ----------------
interface PrintVoucherModalProps extends BaseModalProps {
  note: PaymentNote | null;
}

export const PrintVoucherModal: React.FC<PrintVoucherModalProps> = ({ isOpen, onClose, note }) => {
  if (!isOpen || !note) return null;

  const handlePrint = () => {
    window.print();
  };

  const amount = Number(note.net_payable_amount || note.payable_amount || 0);

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '780px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Printer size={18} style={{ color: 'var(--primary-light)' }} />
            <h3 style={{ margin: 0 }}>Statutory Payment Advice Voucher</h3>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        {/* The Printable Advice Document */}
        <div className="printable-voucher" style={{ padding: '1.5rem', background: '#fff', color: '#111827', borderRadius: '4px', margin: '1rem 0' }}>
          {/* Header */}
          <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '2px solid #1e293b', paddingBottom: '0.75rem', marginBottom: '1rem' }}>
            <div>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 800, margin: 0, color: '#0f172a' }}>NATIONAL HIGHWAYS INFRA TRUST</h2>
              <div style={{ fontSize: '0.75rem', color: '#64748b' }}>Statutory Investment Trust | Ministry of Road Transport & Highways, Govt. of India</div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: '0.85rem', fontWeight: 700, color: '#334155' }}>PAYMENT ADVICE SLIP</div>
              <div style={{ fontSize: '0.75rem', color: '#64748b' }}>Date: {new Date().toLocaleDateString('en-IN')}</div>
            </div>
          </div>

          {/* Reference Block */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem', fontSize: '0.85rem' }}>
            <div>
              <div style={{ color: '#64748b' }}>Voucher Reference:</div>
              <div style={{ fontWeight: 700 }}>{note.payment_note_number || note.sl_no || `PN-${note.id}`}</div>
              <div style={{ color: '#64748b', marginTop: '0.5rem' }}>Beneficiary / Vendor:</div>
              <div style={{ fontWeight: 700 }}>{note.vendor_name}</div>
            </div>
            <div>
              <div style={{ color: '#64748b' }}>Linked Project:</div>
              <div style={{ fontWeight: 700 }}>{note.project_name || 'Highway Corridor EPC'}</div>
              <div style={{ color: '#64748b', marginTop: '0.5rem' }}>Bank Settlement Status:</div>
              <div style={{ fontWeight: 700, color: note.utr_number ? '#15803d' : '#b45309' }}>
                {note.utr_number ? `DISBURSED (UTR: ${note.utr_number})` : 'READY FOR CLEARING'}
              </div>
            </div>
          </div>

          {/* Account Details Table */}
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '1rem', fontSize: '0.85rem' }}>
            <thead>
              <tr style={{ background: '#f1f5f9', textAlign: 'left' }}>
                <th style={{ padding: '0.5rem', border: '1px solid #cbd5e1' }}>Disbursement Parameter</th>
                <th style={{ padding: '0.5rem', border: '1px solid #cbd5e1' }}>Recorded Coordinates</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1' }}>Beneficiary Bank</td>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1', fontWeight: 600 }}>{note.bank_name || 'State Bank of India'}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1' }}>Bank Account Number</td>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1', fontFamily: 'monospace' }}>{note.vendor_account_number || '3482910482910'}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1' }}>IFSC Code</td>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1', fontFamily: 'monospace' }}>{note.ifsc_code || 'SBIN0001423'}</td>
              </tr>
              <tr style={{ background: '#f8fafc' }}>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1', fontWeight: 700 }}>Net Disbursed Amount</td>
                <td style={{ padding: '0.5rem', border: '1px solid #cbd5e1', fontWeight: 800, fontSize: '1rem', color: '#0f172a' }}>
                  ₹{amount.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                </td>
              </tr>
            </tbody>
          </table>

          {/* Signatures */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginTop: '2.5rem', textAlign: 'center', fontSize: '0.75rem', color: '#475569' }}>
            <div style={{ borderTop: '1px solid #94a3b8', paddingTop: '0.5rem' }}>
              <strong>Prepared By</strong>
              <div>Accounts Officer</div>
            </div>
            <div style={{ borderTop: '1px solid #94a3b8', paddingTop: '0.5rem' }}>
              <strong>Verified By</strong>
              <div>Financial Controller</div>
            </div>
            <div style={{ borderTop: '1px solid #94a3b8', paddingTop: '0.5rem' }}>
              <strong>Sanction Authority</strong>
              <div>Trustee Board / CGM</div>
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
          <button className="btn btn-secondary" onClick={onClose}>
            Close
          </button>
          <button className="btn btn-primary" onClick={handlePrint}>
            <Printer size={16} />
            <span>Print Official Advice</span>
          </button>
        </div>
      </div>
    </div>
  );
};

// ---------------- 11. PERSONA SWITCHER MODAL ----------------
interface PersonaModalProps extends BaseModalProps {
  currentPersona: string;
  onSelectPersona: (p: { name: string; role: string; dept: string }) => void;
}

export const PersonaModal: React.FC<PersonaModalProps> = ({ isOpen, onClose, currentPersona, onSelectPersona }) => {
  if (!isOpen) return null;

  const personas = [
    { name: 'Shri R. K. Singh', role: 'Financial Controller', dept: 'Finance & Treasury' },
    { name: 'Dr. A. Verma', role: 'Chief General Manager', dept: 'Highway Corridor Projects' },
    { name: 'Ms. Neha Gupta', role: 'Chief Accounts Officer', dept: 'Disbursement & Audit' },
    { name: 'Vikram Malhotra', role: 'Senior Project Engineer', dept: 'EPC Execution' },
  ];

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '480px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <UserCheck size={18} style={{ color: 'var(--primary-light)' }} />
            <h3 style={{ margin: 0 }}>Simulate Officer Persona</h3>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
            <X size={20} />
          </button>
        </div>

        <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '1rem' }}>
          Select an officer persona to test RBAC roles, delegated financial powers, and approval tiers:
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {personas.map((p) => {
            const isSelected = p.name === currentPersona;
            return (
              <div
                key={p.name}
                onClick={() => {
                  onSelectPersona(p);
                  onClose();
                }}
                style={{
                  padding: '0.75rem 1rem',
                  borderRadius: 'var(--radius-sm)',
                  border: isSelected ? '1px solid var(--primary-color)' : '1px solid var(--border-color)',
                  background: isSelected ? 'rgba(99, 102, 241, 0.1)' : 'var(--bg-card)',
                  cursor: 'pointer',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, color: isSelected ? 'var(--primary-light)' : 'var(--text-primary)' }}>{p.name}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{p.role} • {p.dept}</div>
                </div>
                {isSelected && <span className="badge badge-approved">Active</span>}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
