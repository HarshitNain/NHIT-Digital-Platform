import React, { useState } from 'react';
import {
  Shield,
  ArrowRight,
  CheckCircle2,
  ChevronRight,
  Zap,
  Globe,
  Layers,
  Cpu,
  Database,
  Activity,
  FileText,
  CreditCard,
  Building2,
  Terminal,
  ExternalLink,
  Sparkles,
  TrendingUp,
  Lock,
  RefreshCw,
  BarChart3,
  Check,
  Users,
  MapPin,
  Sliders,
  DollarSign,
  Briefcase,
  Eye
} from 'lucide-react';

interface LandingViewProps {
  onEnterFieldPortal: () => void;
  onEnterAdminPortal: () => void;
  onOpenLogin: (mode?: 'admin' | 'user') => void;
}

export const LandingView: React.FC<LandingViewProps> = ({ 
  onEnterFieldPortal, 
  onEnterAdminPortal, 
  onOpenLogin 
}) => {
  const [activeWorkflowTab, setActiveWorkflowTab] = useState<'admin' | 'user' | 'dev'>('admin');
  const [activeCorridor, setActiveCorridor] = useState<number>(0);

  const corridors = [
    {
      code: 'NH-48',
      name: 'Delhi - Mumbai Expressway Corridor (Pkg II)',
      length: '412 km',
      contractor: 'L&T Infrastructure Ltd',
      budget: '₹78.50 Cr',
      progress: 84,
      status: 'Active Concession',
      tollStatus: 'FASTag 99.8% Digital',
    },
    {
      code: 'NH-44',
      name: 'Hyderabad - Bengaluru Section',
      length: '560 km',
      contractor: 'Afcons Infrastructure Ltd',
      budget: '₹142.00 Cr',
      progress: 96,
      status: 'Maintenance & O&M',
      tollStatus: 'FASTag 99.9% Digital',
    },
    {
      code: 'NH-19',
      name: 'Delhi - Agra - Varanasi Economic Corridor',
      length: '240 km',
      contractor: 'Dilip Buildcon Ltd',
      budget: '₹32.50 Cr',
      progress: 68,
      status: 'Widening & Overlay',
      tollStatus: 'Multi-lane Free Flow',
    },
    {
      code: 'NE-02',
      name: 'Eastern Peripheral Expressway Ring',
      length: '135 km',
      contractor: 'NHAI SPV Managed',
      budget: '₹95.00 Cr',
      progress: 100,
      status: 'Fully Operational',
      tollStatus: 'Smart Tolling Active',
    },
  ];

  return (
    <div className="landing-page-root">
      {/* Background ambient lighting orbs */}
      <div className="landing-ambient-glow glow-1" />
      <div className="landing-ambient-glow glow-2" />
      <div className="landing-ambient-glow glow-3" />

      {/* 1. INSTITUTIONAL TOP NAVIGATION */}
      <header className="landing-nav-bar">
        <div className="landing-nav-container">
          <div className="landing-brand-group">
            <div className="landing-logo-badge">
              <Shield size={24} className="text-white" />
            </div>
            <div className="landing-brand-text">
              <div className="brand-title-row">
                <span className="brand-name">NHIT</span>
                <span className="brand-version-pill">OS v2.4</span>
              </div>
              <span className="brand-sub">National Highways Infra Trust</span>
            </div>
          </div>

          <nav className="landing-nav-links">
            <a href="#features" className="nav-item">Platform Pillars</a>
            <a href="#corridors" className="nav-item">Corridor Portfolio</a>
            <a href="#workflow" className="nav-item">Capital Pipeline</a>
            <a 
              href="http://localhost:8080/docs" 
              target="_blank" 
              rel="noopener noreferrer" 
              className="nav-item swagger-nav-link"
              title="Interactive OpenAPI 3.0 documentation on Port 8080"
            >
              <Terminal size={14} />
              <span>Swagger API (:8080)</span>
              <ExternalLink size={12} className="opacity-60" />
            </a>
          </nav>

          <div className="landing-nav-actions">
            <button 
              onClick={onEnterFieldPortal}
              className="btn-glass-subtle"
              title="Open Field Site Engineer Portal directly"
            >
              <Briefcase size={15} />
              <span>Field Portal</span>
            </button>
            <button 
              onClick={onEnterAdminPortal}
              className="btn-primary-gradient"
              title="Open Admin Executive Command Center directly"
            >
              <Shield size={15} />
              <span>Admin Portal</span>
            </button>
            <button
              onClick={() => onOpenLogin('admin')}
              className="btn-glass-subtle"
              style={{ padding: '8px 12px' }}
              title="Sign in with specific credentials"
            >
              <Lock size={14} />
              <span>Sign In</span>
            </button>
          </div>
        </div>
      </header>

      {/* 2. HERO SECTION */}
      <section className="landing-hero-section">
        <div className="landing-hero-container">
          {/* Pill Announcement */}
          <div className="hero-announcement-pill">
            <Sparkles size={14} className="text-cyan" />
            <span>Sovereign Infrastructure Capital Flow & Governance Engine</span>
            <span className="pill-dot" />
            <span className="pill-status">Go Mesh Online</span>
          </div>

          {/* Main Hero Headline */}
          <h1 className="hero-headline">
            Institutional Capital Flow, <br />
            <span className="gradient-text-accent">Engineered for National Scale.</span>
          </h1>

          <p className="hero-description">
            The next-generation digital operating system for National Highways Infra Trust (NHIT).
            Automating green note certifications, Section 194C statutory tax deductions, 
            multi-tier financial approvals, and direct banking UTR reconciliation across 
            India's arterial expressway network.
          </p>

          {/* Hero CTAs */}
          <div className="hero-cta-group">
            <button 
              onClick={onEnterAdminPortal}
              className="btn-hero-primary"
              title="Launch Admin Command Center directly"
            >
              <Shield size={18} />
              <span>Launch Admin Command Center</span>
              <ArrowRight size={18} />
            </button>
            <button 
              onClick={onEnterFieldPortal}
              className="btn-hero-secondary"
              title="Enter Site Field Portal directly"
            >
              <Users size={18} />
              <span>Enter Site Field Portal</span>
            </button>
            <a 
              href="http://localhost:8080/docs"
              target="_blank"
              rel="noopener noreferrer"
              className="btn-hero-outline"
            >
              <Terminal size={17} />
              <span>OpenAPI Swagger (:8080)</span>
              <ExternalLink size={14} />
            </a>
          </div>

          {/* HERO LIVE SIMULATOR / INTERACTIVE PREVIEW */}
          <div className="hero-simulator-wrapper">
            <div className="hero-simulator-card">
              <div className="simulator-top-bar">
                <div className="window-dots">
                  <span className="dot red" />
                  <span className="dot yellow" />
                  <span className="dot green" />
                </div>
                <div className="simulator-title">
                  <Activity size={14} className="text-emerald" />
                  <span>NHIT Central Capital Gateway • Live Telemetry Stream</span>
                </div>
                <div className="simulator-badge-live">
                  <span className="pulse-ping" />
                  <span>REST & gRPC Active (Port 8080)</span>
                </div>
              </div>

              {/* Live Ticker Stats */}
              <div className="simulator-metrics-grid">
                <div className="sim-metric-card">
                  <div className="metric-header">
                    <span className="metric-label">Live AUM Portfolio</span>
                    <TrendingUp size={16} className="text-emerald" />
                  </div>
                  <div className="metric-val">₹14,850.42 Cr</div>
                  <div className="metric-meta text-emerald">↑ +4.8% YoY Trust Growth</div>
                </div>

                <div className="sim-metric-card">
                  <div className="metric-header">
                    <span className="metric-label">Green Notes In-Flight</span>
                    <FileText size={16} className="text-cyan" />
                  </div>
                  <div className="metric-val">₹25.30 Cr</div>
                  <div className="metric-meta text-cyan">4 Pending Statutory Clearance</div>
                </div>

                <div className="sim-metric-card">
                  <div className="metric-header">
                    <span className="metric-label">Settled Today (UTR Confirmed)</span>
                    <CheckCircle2 size={16} className="text-emerald" />
                  </div>
                  <div className="metric-val">₹13.91 Cr</div>
                  <div className="metric-meta text-muted">100% Section 194C Compliant</div>
                </div>

                <div className="sim-metric-card">
                  <div className="metric-header">
                    <span className="metric-label">Gateway Mesh Latency</span>
                    <Cpu size={16} className="text-purple" />
                  </div>
                  <div className="metric-val mono">3.4 ms</div>
                  <div className="metric-meta text-muted">Go Backend Microservices</div>
                </div>
              </div>

              {/* Simulated Live Ledger Strip */}
              <div className="simulator-ledger-preview">
                <div className="ledger-preview-header">
                  <span>REAL-TIME CAPITAL CLEARANCE QUEUE</span>
                  <span className="mono text-dim">SYNC_ID: NHIT-TX-2026-9921</span>
                </div>
                <div className="ledger-row highlight">
                  <div className="ledger-badge approved">APPROVED</div>
                  <div className="ledger-code mono">GN-4030</div>
                  <div className="ledger-title">NH-44 Hyderabad Section • Afcons Infrastructure Ltd</div>
                  <div className="ledger-amt mono">₹1,42,00,000</div>
                  <div className="ledger-tax mono">TDS: ₹2,84,000</div>
                  <div className="ledger-utr mono text-emerald">UTR: HDFCR5202604149021</div>
                </div>
                <div className="ledger-row">
                  <div className="ledger-badge pending">IN REVIEW</div>
                  <div className="ledger-code mono">GN-4029</div>
                  <div className="ledger-title">NH-48 Corridor Expansion Pkg II • L&T Infrastructure Ltd</div>
                  <div className="ledger-amt mono">₹78,50,000</div>
                  <div className="ledger-tax mono">TDS: ₹1,57,000</div>
                  <div className="ledger-utr mono text-muted">Awaiting Controller Sign-off</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 3. ENTERPRISE NUMBERS TICKER */}
      <section className="landing-stats-strip">
        <div className="stats-strip-container">
          <div className="stat-item">
            <div className="stat-big-num">₹14,850+ <span className="stat-unit">Cr</span></div>
            <div className="stat-sub-label">Assets Under Management (AUM)</div>
          </div>
          <div className="stat-divider" />
          <div className="stat-item">
            <div className="stat-big-num">12+ <span className="stat-unit">Corridors</span></div>
            <div className="stat-sub-label">National Expressways Under Trust</div>
          </div>
          <div className="stat-divider" />
          <div className="stat-item">
            <div className="stat-big-num">100% <span className="stat-unit">Digital</span></div>
            <div className="stat-sub-label">Section 194C TDS & Statutory Audit</div>
          </div>
          <div className="stat-divider" />
          <div className="stat-item">
            <div className="stat-big-num">99.98% <span className="stat-unit">SLA</span></div>
            <div className="stat-sub-label">Go Microservices Mesh Uptime</div>
          </div>
        </div>
      </section>

      {/* 4. PLATFORM PILLARS (BENTO GRID) */}
      <section id="features" className="landing-section">
        <div className="landing-section-header">
          <div className="section-pill">
            <Layers size={14} />
            <span>Platform Capabilities</span>
          </div>
          <h2 className="section-title">Built for Sovereign Scale & Institutional Rigor</h2>
          <p className="section-subtitle">
            Every transaction, milestone certificate, and vendor disbursal is guarded by 
            tamper-proof cryptographic logs and automated statutory governance.
          </p>
        </div>

        <div className="bento-grid">
          {/* Card 1: Green Note Certification */}
          <div className="bento-card card-span-2">
            <div className="card-top-icon green-glow">
              <FileText size={28} className="text-emerald" />
            </div>
            <h3 className="card-title">Digital Green Note Lifecycle</h3>
            <p className="card-text">
              Eliminate physical paperwork delays. Field engineers upload milestone measurements, 
              geo-tagged inspection images, and BOQ line items directly from remote highway construction packages.
            </p>
            <ul className="card-feature-list">
              <li><Check size={16} className="text-emerald" /> Multi-tier financial delegations (FC, CGM, Trustee)</li>
              <li><Check size={16} className="text-emerald" /> Cryptographic timestamping and tamper-evident audit logs</li>
              <li><Check size={16} className="text-emerald" /> Automated variance checking against contract allowances</li>
            </ul>
          </div>

          {/* Card 2: Automated Statutory TDS */}
          <div className="bento-card">
            <div className="card-top-icon cyan-glow">
              <CreditCard size={28} className="text-cyan" />
            </div>
            <h3 className="card-title">Statutory TDS & Treasury Engine</h3>
            <p className="card-text">
              Real-time calculation of Section 194C / 194J income tax withholdings, 
              net payable derivation, and automatic payment voucher compilation with formal QR verification.
            </p>
          </div>

          {/* Card 3: Dual-Persona Portal */}
          <div className="bento-card">
            <div className="card-top-icon purple-glow">
              <Users size={28} className="text-purple" />
            </div>
            <h3 className="card-title">Dual-Role Portal Architecture</h3>
            <p className="card-text">
              Strict multi-tenant role isolation. Executive controllers command treasury and disbursements, 
              while site engineers and contractor liaisons operate in focused claim filing portals.
            </p>
          </div>

          {/* Card 4: Microservices Mesh */}
          <div className="bento-card card-span-2">
            <div className="card-top-icon indigo-glow">
              <Cpu size={28} className="text-primary" />
            </div>
            <h3 className="card-title">High-Throughput Go Microservices</h3>
            <p className="card-text">
              Modular Go microservices orchestrating Authentication, Green Note Ledger, Treasury Payments, 
              and Reimbursements with sub-5ms response latencies and zero memory leaks.
            </p>
            <div className="card-code-pill mono">
              <code>GET /api/v1/health → {"{ status: \"UP\", gateway_port: 8080, latency: \"3.2ms\" }"}</code>
            </div>
          </div>

          {/* Card 5: Open API & Swagger */}
          <div className="bento-card">
            <div className="card-top-icon amber-glow">
              <Terminal size={28} className="text-amber" />
            </div>
            <h3 className="card-title">OpenAPI 3.0 on Port 8080</h3>
            <p className="card-text">
              Dedicated developer Swagger UI for partner banks, CAG auditors, and NHAI SPVs to inspect, 
              test, and integrate via standardized JSON schemas.
            </p>
          </div>

          {/* Card 6: Vendor KYC Registry */}
          <div className="bento-card">
            <div className="card-top-icon rose-glow">
              <Building2 size={28} className="text-rose" />
            </div>
            <h3 className="card-title">Vendor KYC & Compliance Registry</h3>
            <p className="card-text">
              Unified database of major highway concessionaires (L&T, Afcons, Dilip Buildcon) 
              with GSTIN status, MSME tiering, and verified bank account validation.
            </p>
          </div>
        </div>
      </section>

      {/* 5. INTERACTIVE CAPITAL WORKFLOW PIPELINE */}
      <section id="workflow" className="landing-section bg-contrast-section">
        <div className="landing-section-header">
          <div className="section-pill">
            <Sliders size={14} />
            <span>Interactive Operational Workflow</span>
          </div>
          <h2 className="section-title">From Site Milestone to Central Disbursal</h2>
          <p className="section-subtitle">
            Experience how capital flows securely through the NHIT operating ecosystem.
          </p>
        </div>

        {/* 4 Pipeline Stages */}
        <div className="workflow-pipeline-grid">
          <div className="pipeline-step">
            <div className="step-number">01</div>
            <div className="step-title">Site Milestone Filing</div>
            <div className="step-desc">
              Site Engineer measures highway completion, attaches vendor invoice, and triggers Green Note draft.
            </div>
            <div className="step-portal-tag user-tag">Field Site Portal</div>
          </div>

          <div className="pipeline-connector" />

          <div className="pipeline-step">
            <div className="step-number">02</div>
            <div className="step-title">Controller Sign-Off</div>
            <div className="step-desc">
              Financial Controller verifies contract allowances, validates GST amounts, and approves voucher release.
            </div>
            <div className="step-portal-tag admin-tag">Admin Command Center</div>
          </div>

          <div className="pipeline-connector" />

          <div className="pipeline-step">
            <div className="step-number">03</div>
            <div className="step-title">Statutory Deduction</div>
            <div className="step-desc">
              Treasury engine auto-withholds Section 194C TDS (2%) and issues formal payment voucher.
            </div>
            <div className="step-portal-tag engine-tag">Statutory Engine</div>
          </div>

          <div className="pipeline-connector" />

          <div className="pipeline-step">
            <div className="step-number">04</div>
            <div className="step-title">Bank Settlement (UTR)</div>
            <div className="step-desc">
              Treasury dispatches payment via HDFC/SBI and records RBI UTR reference for immutable proof.
            </div>
            <div className="step-portal-tag bank-tag">Banking Ledger</div>
          </div>
        </div>

        {/* Interactive Workspace View Switcher */}
        <div className="workflow-interactive-viewer">
          <div className="viewer-tab-nav">
            <button
              onClick={() => setActiveWorkflowTab('admin')}
              className={`viewer-tab ${activeWorkflowTab === 'admin' ? 'active' : ''}`}
            >
              <Shield size={16} />
              <span>Admin Command View</span>
            </button>
            <button
              onClick={() => setActiveWorkflowTab('user')}
              className={`viewer-tab ${activeWorkflowTab === 'user' ? 'active' : ''}`}
            >
              <Briefcase size={16} />
              <span>Site Field Portal View</span>
            </button>
            <button
              onClick={() => setActiveWorkflowTab('dev')}
              className={`viewer-tab ${activeWorkflowTab === 'dev' ? 'active' : ''}`}
            >
              <Terminal size={16} />
              <span>Developer Swagger Gateway (:8080)</span>
            </button>
          </div>

          <div className="viewer-display-card">
            {activeWorkflowTab === 'admin' && (
              <div className="viewer-content-pane">
                <div className="pane-header">
                  <div className="pane-title-group">
                    <h4>Executive Controller Command Dashboard</h4>
                    <span className="pane-subtitle">Central governance console for financial heads and directors</span>
                  </div>
                  <button 
                    onClick={onEnterAdminPortal}
                    className="btn-pane-action"
                  >
                    <span>Launch Admin Portal</span>
                    <ArrowRight size={14} />
                  </button>
                </div>
                <div className="pane-mock-grid">
                  <div className="mock-widget">
                    <span className="mock-widget-label">DISBURSEMENT APPROVAL QUEUE</span>
                    <div className="mock-list-item">
                      <div>
                        <strong>GN-4029: NH-48 Corridor Expansion</strong>
                        <p>L&T Infrastructure • ₹78,50,000</p>
                      </div>
                      <span className="badge-chip warning">Pending Approval</span>
                    </div>
                    <div className="mock-list-item">
                      <div>
                        <strong>GN-4031: NH-19 Delhi-Agra Overlay</strong>
                        <p>Dilip Buildcon • ₹32,50,000</p>
                      </div>
                      <span className="badge-chip info">In Audit Review</span>
                    </div>
                  </div>
                  <div className="mock-widget">
                    <span className="mock-widget-label">STATUTORY COMPLIANCE LEDGER</span>
                    <div className="statutory-progress">
                      <div className="statutory-row">
                        <span>Sec 194C Contractor TDS</span>
                        <span className="text-emerald mono font-semibold">100% Reconciled</span>
                      </div>
                      <div className="statutory-row">
                        <span>GST Input Tax Credit</span>
                        <span className="text-cyan mono font-semibold">₹1.84 Cr Verified</span>
                      </div>
                      <div className="statutory-row">
                        <span>Voucher Print Engine</span>
                        <span className="text-muted mono">Ready with QR stamp</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeWorkflowTab === 'user' && (
              <div className="viewer-content-pane">
                <div className="pane-header">
                  <div className="pane-title-group">
                    <h4>Site Field Portal & Claim Submission</h4>
                    <span className="pane-subtitle">Lightweight, fast interface for project engineers and site managers</span>
                  </div>
                  <button 
                    onClick={onEnterFieldPortal}
                    className="btn-pane-action"
                  >
                    <span>Launch Field Portal</span>
                    <ArrowRight size={14} />
                  </button>
                </div>
                <div className="pane-mock-grid">
                  <div className="mock-widget">
                    <span className="mock-widget-label">SUBMIT NEW GREEN NOTE</span>
                    <div className="mock-form-preview">
                      <div className="preview-field">
                        <span>Project:</span>
                        <strong>NH-48 Corridor Expansion Package II</strong>
                      </div>
                      <div className="preview-field">
                        <span>Milestone:</span>
                        <strong>Bridge Foundation Concrete Pouring 240m</strong>
                      </div>
                      <div className="preview-field">
                        <span>Invoice Amount:</span>
                        <strong className="text-emerald">₹78,50,000</strong>
                      </div>
                    </div>
                  </div>
                  <div className="mock-widget">
                    <span className="mock-widget-label">RECENT CLAIM STATUS</span>
                    <div className="mock-list-item">
                      <div>
                        <strong>GN/2026/04/0030</strong>
                        <p>Bituminous wearing course milestone</p>
                      </div>
                      <span className="badge-chip success">Payment Disbursed</span>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeWorkflowTab === 'dev' && (
              <div className="viewer-content-pane">
                <div className="pane-header">
                  <div className="pane-title-group">
                    <h4>OpenAPI 3.0 / Swagger UI Interface</h4>
                    <span className="pane-subtitle">Serving on Port 8080 for system developers and institutional audits</span>
                  </div>
                  <a 
                    href="http://localhost:8080/docs" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="btn-pane-action"
                  >
                    <span>Open Swagger UI</span>
                    <ExternalLink size={14} />
                  </a>
                </div>
                <div className="pane-mock-grid">
                  <div className="mock-widget code-widget">
                    <span className="mock-widget-label">GO MICROSERVICES ROUTING TABLE</span>
                    <div className="swagger-routes">
                      <div className="route-row get">
                        <span className="route-method">GET</span>
                        <span className="route-path mono">/api/v1/green-notes</span>
                        <span className="route-desc">Fetch paginated green notes</span>
                      </div>
                      <div className="route-row post">
                        <span className="route-method">POST</span>
                        <span className="route-path mono">/api/v1/green-notes/{"{id}"}/approve</span>
                        <span className="route-desc">Execute financial delegation</span>
                      </div>
                      <div className="route-row post">
                        <span className="route-method">POST</span>
                        <span className="route-path mono">/api/v1/payment-notes/{"{id}"}/record-utr</span>
                        <span className="route-desc">Record bank UTR number</span>
                      </div>
                      <div className="route-row get">
                        <span className="route-method">GET</span>
                        <span className="route-path mono">/api/v1/telemetry/health</span>
                        <span className="route-desc">Inspect Go service cluster health</span>
                      </div>
                    </div>
                  </div>
                  <div className="mock-widget">
                    <span className="mock-widget-label">DEVELOPER PORTAL DETAILS</span>
                    <div className="dev-meta-list">
                      <div className="meta-item">
                        <span>Gateway Port:</span>
                        <strong className="mono">8080</strong>
                      </div>
                      <div className="meta-item">
                        <span>Frontend Port:</span>
                        <strong className="mono">5173</strong>
                      </div>
                      <div className="meta-item">
                        <span>Specification:</span>
                        <strong>OpenAPI 3.0 / Swagger UI</strong>
                      </div>
                      <div className="meta-item">
                        <span>Auth Standard:</span>
                        <strong>JWT Bearer with Role Scopes</strong>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </section>

      {/* 6. CORRIDOR ASSET PORTFOLIO SHOWCASE */}
      <section id="corridors" className="landing-section">
        <div className="landing-section-header">
          <div className="section-pill">
            <MapPin size={14} />
            <span>Infrastructure Asset Portfolio</span>
          </div>
          <h2 className="section-title">Major National Expressway Corridors</h2>
          <p className="section-subtitle">
            Managing capital maintenance and green note governance for India's strategic economic lifelines.
          </p>
        </div>

        <div className="corridor-showcase-container">
          <div className="corridor-nav-list">
            {corridors.map((c, idx) => (
              <button
                key={c.code}
                onClick={() => setActiveCorridor(idx)}
                className={`corridor-nav-item ${activeCorridor === idx ? 'active' : ''}`}
              >
                <div className="corridor-nav-code mono">{c.code}</div>
                <div className="corridor-nav-info">
                  <div className="corridor-nav-title">{c.name}</div>
                  <div className="corridor-nav-sub">{c.length} • {c.contractor}</div>
                </div>
                <ChevronRight size={18} className="chevron" />
              </button>
            ))}
          </div>

          <div className="corridor-detail-panel">
            {(() => {
              const current = corridors[activeCorridor];
              return (
                <div className="corridor-card">
                  <div className="corridor-card-header">
                    <div>
                      <span className="corridor-pill mono">{current.code} CORRIDOR</span>
                      <h3 className="corridor-card-title">{current.name}</h3>
                    </div>
                    <span className="badge-chip success">{current.status}</span>
                  </div>

                  <div className="corridor-stats-grid">
                    <div className="c-stat">
                      <span className="c-label">Section Length</span>
                      <strong className="c-val mono">{current.length}</strong>
                    </div>
                    <div className="c-stat">
                      <span className="c-label">Prime Concessionaire</span>
                      <strong className="c-val">{current.contractor}</strong>
                    </div>
                    <div className="c-stat">
                      <span className="c-label">Total Concession Budget</span>
                      <strong className="c-val mono text-emerald">{current.budget}</strong>
                    </div>
                    <div className="c-stat">
                      <span className="c-label">Tolling Digitalization</span>
                      <strong className="c-val text-cyan">{current.tollStatus}</strong>
                    </div>
                  </div>

                  <div className="corridor-progress-box">
                    <div className="progress-info-row">
                      <span>Milestone Execution Progress</span>
                      <span className="mono font-semibold">{current.progress}% Complete</span>
                    </div>
                    <div className="progress-track">
                      <div 
                        className="progress-fill" 
                        style={{ width: `${current.progress}%` }} 
                      />
                    </div>
                  </div>

                  <div className="corridor-action-footer">
                    <button 
                      onClick={onEnterAdminPortal}
                      className="btn-corridor-inspect"
                    >
                      <Eye size={15} />
                      <span>Inspect Corridor Ledger in Admin Portal</span>
                      <ArrowRight size={15} />
                    </button>
                  </div>
                </div>
              );
            })()}
          </div>
        </div>
      </section>

      {/* 7. PRE-LOGIN ONBOARDING GATEWAY (CALLOUT) */}
      <section className="landing-cta-banner">
        <div className="cta-banner-content">
          <div className="cta-brand-icon">
            <Shield size={36} />
          </div>
          <h2 className="cta-title">Ready to Access the NHIT Operating Portal?</h2>
          <p className="cta-desc">
            Enter the portal directly as a Field Site Engineer or Financial Controller, 
            or sign in with custom officer credentials.
          </p>

          <div className="cta-buttons-wrapper">
            <button 
              onClick={onEnterAdminPortal}
              className="btn-cta-admin"
              title="Enter Admin Command Center directly"
            >
              <Shield size={18} />
              <span>Enter Admin Command Center</span>
            </button>
            <button 
              onClick={onEnterFieldPortal}
              className="btn-cta-user"
              title="Enter Site Field Portal directly"
            >
              <Users size={18} />
              <span>Enter Site Field Portal</span>
            </button>
            <button 
              onClick={() => onOpenLogin('admin')}
              className="btn-cta-swagger"
              title="Sign in with custom officer email & password"
            >
              <Lock size={16} />
              <span>Sign In with Credentials</span>
            </button>
            <a 
              href="http://localhost:8080/swagger" 
              target="_blank" 
              rel="noopener noreferrer"
              className="btn-cta-swagger"
            >
              <Terminal size={18} />
              <span>Developer Swagger (:8080)</span>
              <ExternalLink size={14} />
            </a>
          </div>
        </div>
      </section>

      {/* 8. ENTERPRISE FOOTER */}
      <footer className="landing-footer">
        <div className="landing-footer-container">
          <div className="footer-top">
            <div className="footer-brand">
              <div className="brand-title-row">
                <div className="landing-logo-badge sm">
                  <Shield size={16} />
                </div>
                <span className="brand-name">NHIT</span>
              </div>
              <p className="footer-tagline">
                National Highways Infra Trust — Sovereign infrastructure asset investment trust 
                sponsored by the National Highways Authority of India (NHAI).
              </p>
            </div>

            <div className="footer-columns">
              <div className="footer-col">
                <h5>Product</h5>
                <a href="#features">Platform Pillars</a>
                <a href="#corridors">Corridor Portfolio</a>
                <a href="#workflow">Capital Pipeline</a>
                <button onClick={onEnterAdminPortal} className="text-left-link">
                  Admin Portal
                </button>
                <button onClick={() => onOpenLogin('admin')} className="text-left-link">
                  Sign In with Credentials
                </button>
              </div>

              <div className="footer-col">
                <h5>Developers</h5>
                <a href="http://localhost:8080/swagger" target="_blank" rel="noopener noreferrer">
                  Swagger UI (:8080)
                </a>
                <a href="http://localhost:8080/openapi.json" target="_blank" rel="noopener noreferrer">
                  OpenAPI Spec
                </a>
                <button onClick={onEnterFieldPortal} className="text-left-link">
                  Field User Portal
                </button>
              </div>

              <div className="footer-col">
                <h5>Governance</h5>
                <span>Section 194C TDS Automated</span>
                <span>NHAI SPV Compliance</span>
                <span>CAG Audit Readiness</span>
                <span>RBI UTR Banking Ledger</span>
              </div>
            </div>
          </div>

          <div className="footer-bottom">
            <div className="footer-copy">
              © {new Date().getFullYear()} National Highways Infra Trust (NHIT). All rights reserved.
            </div>
            <div className="footer-status-pill">
              <span className="pulse-ping" />
              <span>Go Gateway Active on Port 8080</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};
