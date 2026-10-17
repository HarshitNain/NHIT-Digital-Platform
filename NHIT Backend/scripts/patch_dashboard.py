import sys
sys.stdout.reconfigure(encoding='utf-8')

with open('gateway_swagger.py', 'r', encoding='utf-8') as f:
    content = f.read()

dashboard_start_marker = "# DASHBOARD (HTML)\n# ═══════════════════════════════════════════════════════\n\n@app.get(\"/\", response_class=HTMLResponse, include_in_schema=False)\nasync def dashboard():"
dashboard_end_marker = "# ENTRY POINT\n# ═══════════════════════════════════════════════════════"

dashboard_replacement = """# DASHBOARD (HTML)
# ═══════════════════════════════════════════════════════

@app.get("/", response_class=HTMLResponse, include_in_schema=False)
async def dashboard():
    return HTMLResponse(content=\"\"\"<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>NHIT Backend — Enterprise API Hub & Gateway</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
<style>
:root {
  --bg-dark: #07090e;
  --bg-card: rgba(18, 24, 38, 0.75);
  --bg-card-hover: rgba(28, 38, 60, 0.85);
  --border-subtle: rgba(255, 255, 255, 0.08);
  --border-glow: rgba(99, 102, 241, 0.35);
  --primary: #6366f1;
  --primary-light: #818cf8;
  --cyan: #06b6d4;
  --emerald: #10b981;
  --rose: #f43f5e;
  --amber: #f59e0b;
  --text-main: #f1f5f9;
  --text-muted: #94a3b8;
  --text-dim: #64748b;
}

* { margin:0; padding:0; box-sizing:border-box; }
body {
  font-family: 'Plus Jakarta Sans', sans-serif;
  background: var(--bg-dark);
  color: var(--text-main);
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background-image: 
    radial-gradient(circle at 15% 15%, rgba(99, 102, 241, 0.12) 0%, transparent 40%),
    radial-gradient(circle at 85% 20%, rgba(6, 182, 212, 0.10) 0%, transparent 45%),
    radial-gradient(circle at 50% 80%, rgba(139, 92, 246, 0.08) 0%, transparent 50%);
}

.container {
  max-width: 1360px;
  margin: 0 auto;
  padding: 48px 24px 80px;
}

/* Header */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 56px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-subtle);
}
.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}
.brand-logo {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #4f46e5, #06b6d4);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 1.3rem;
  color: #fff;
  box-shadow: 0 8px 24px rgba(79, 70, 229, 0.35);
}
.brand-name {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: -0.5px;
  background: linear-gradient(to right, #ffffff, #cbd5e1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.brand-badge {
  font-size: 0.7rem;
  padding: 3px 8px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.18);
  border: 1px solid rgba(99, 102, 241, 0.3);
  color: var(--primary-light);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.live-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  padding: 8px 16px;
  border-radius: 20px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #34d399;
  font-weight: 600;
}
.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 12px #10b981;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 8px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

/* Hero Section */
.hero {
  text-align: center;
  margin-bottom: 50px;
}
.hero h1 {
  font-size: 3.2rem;
  font-weight: 800;
  letter-spacing: -1.2px;
  line-height: 1.15;
  margin-bottom: 16px;
  background: linear-gradient(135deg, #ffffff 30%, #94a3b8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.hero h1 span {
  background: linear-gradient(135deg, #818cf8, #06b6d4, #f472b6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.hero p {
  color: var(--text-muted);
  font-size: 1.15rem;
  max-width: 680px;
  margin: 0 auto;
  line-height: 1.6;
}

/* Stats Row */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 44px;
}
.stat-box {
  background: var(--bg-card);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border-subtle);
  border-radius: 20px;
  padding: 26px 20px;
  text-align: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}
.stat-box::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--border-glow), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}
.stat-box:hover {
  transform: translateY(-4px);
  border-color: var(--border-glow);
  box-shadow: 0 12px 30px -10px rgba(99, 102, 241, 0.25);
}
.stat-box:hover::before {
  opacity: 1;
}
.stat-value {
  font-size: 2.6rem;
  font-weight: 800;
  letter-spacing: -1px;
  line-height: 1.1;
  background: linear-gradient(135deg, #ffffff, #cbd5e1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.stat-label {
  color: var(--text-muted);
  font-size: 0.82rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-top: 6px;
}
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 30px;
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.status-pill.connected {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: #34d399;
}
.status-pill.connecting {
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.35);
  color: #fbbf24;
}

/* Action Cards */
.actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 22px;
  margin-bottom: 50px;
}
.action-card {
  background: var(--bg-card);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border-subtle);
  border-radius: 22px;
  padding: 30px;
  text-decoration: none;
  color: inherit;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}
.action-card:hover {
  transform: translateY(-5px);
  background: var(--bg-card-hover);
  border-color: var(--border-glow);
  box-shadow: 0 16px 40px -10px rgba(0, 0, 0, 0.5);
}
.card-icon-wrap {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}
.card-swagger .card-icon-wrap {
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  color: #818cf8;
}
.card-redoc .card-icon-wrap {
  background: rgba(244, 63, 94, 0.15);
  border: 1px solid rgba(244, 63, 94, 0.3);
  color: #fb7185;
}
.card-openapi .card-icon-wrap {
  background: rgba(6, 182, 212, 0.15);
  border: 1px solid rgba(6, 182, 212, 0.3);
  color: #22d3ee;
}
.card-title {
  font-size: 1.35rem;
  font-weight: 700;
  margin-bottom: 10px;
  letter-spacing: -0.3px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-desc {
  color: var(--text-muted);
  font-size: 0.92rem;
  line-height: 1.55;
  flex-grow: 1;
  margin-bottom: 22px;
}
.card-link-arrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.88rem;
  font-weight: 700;
  transition: gap 0.2s ease;
}
.action-card:hover .card-link-arrow {
  gap: 10px;
}
.card-swagger .card-link-arrow { color: #818cf8; }
.card-redoc .card-link-arrow { color: #fb7185; }
.card-openapi .card-link-arrow { color: #22d3ee; }

/* Quick Credentials Drawer */
.creds-banner {
  background: rgba(30, 41, 59, 0.45);
  border: 1px dashed rgba(148, 163, 184, 0.25);
  border-radius: 18px;
  padding: 20px 24px;
  margin-bottom: 50px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}
.creds-info {
  display: flex;
  align-items: center;
  gap: 14px;
}
.creds-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
}
.creds-text h4 {
  font-size: 0.95rem;
  font-weight: 700;
  margin-bottom: 2px;
}
.creds-text p {
  color: var(--text-muted);
  font-size: 0.85rem;
}
.creds-pills {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.copy-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.2);
  padding: 6px 14px;
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
}
.copy-pill:hover {
  background: rgba(30, 41, 59, 0.9);
  border-color: var(--primary-light);
  color: var(--primary-light);
}

/* Section Header */
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}
.section-title {
  font-size: 1.45rem;
  font-weight: 800;
  letter-spacing: -0.4px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.search-box {
  background: rgba(18, 24, 38, 0.8);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  padding: 10px 18px;
  color: var(--text-main);
  font-family: inherit;
  font-size: 0.9rem;
  width: 280px;
  outline: none;
  transition: border-color 0.2s ease;
}
.search-box:focus {
  border-color: var(--primary-light);
}

/* Services Grid */
.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 18px;
}
.svc-card {
  background: var(--bg-card);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border-subtle);
  border-radius: 16px;
  padding: 22px;
  transition: all 0.25s ease;
  display: flex;
  flex-direction: column;
}
.svc-card:hover {
  border-color: rgba(99, 102, 241, 0.35);
  transform: translateY(-2px);
  box-shadow: 0 10px 25px -10px rgba(0, 0, 0, 0.4);
}
.svc-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}
.svc-name {
  color: #fff;
  font-size: 1.05rem;
  font-weight: 700;
}
.svc-badge {
  font-size: 0.72rem;
  font-family: 'JetBrains Mono', monospace;
  padding: 3px 8px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.25);
  color: var(--primary-light);
  font-weight: 600;
}
.svc-port {
  font-family: 'JetBrains Mono', monospace;
  color: var(--cyan);
  font-size: 0.78rem;
  margin-bottom: 10px;
}
.svc-desc {
  color: var(--text-muted);
  font-size: 0.86rem;
  line-height: 1.5;
  flex-grow: 1;
}

@media (max-width: 960px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .actions-grid { grid-template-columns: 1fr; }
  .hero h1 { font-size: 2.4rem; }
}
@media (max-width: 580px) {
  .stats-grid { grid-template-columns: 1fr; }
  .top-nav { flex-direction: column; gap: 16px; align-items: flex-start; }
}
</style>
</head>
<body>
<div class="container">

  <!-- Nav -->
  <div class="top-nav">
    <div class="brand">
      <div class="brand-logo">N</div>
      <div>
        <div style="display:flex;align-items:center;gap:8px">
          <span class="brand-name">NHIT Backend Gateway</span>
          <span class="brand-badge">Hub</span>
        </div>
        <div style="font-size:0.75rem;color:var(--text-dim);margin-top:2px">National Highway Infrastructure Trust</div>
      </div>
    </div>
    <div class="live-badge">
      <div class="pulse-dot"></div>
      <span id="system-status-text">All Services Operational</span>
    </div>
  </div>

  <!-- Hero -->
  <div class="hero">
    <h1>Enterprise API Gateway & <span>Swagger Hub</span></h1>
    <p>Comprehensive interactive API documentation, schema definitions, and testing console for the complete NHIT Backend microservices architecture.</p>
  </div>

  <!-- Stats Grid -->
  <div class="stats-grid">
    <div class="stat-box">
      <div class="stat-value">13</div>
      <div class="stat-label">Microservices</div>
    </div>
    <div class="stat-box">
      <div class="stat-value" id="endpoints-count">159</div>
      <div class="stat-label">API Endpoints</div>
    </div>
    <div class="stat-box">
      <div class="stat-value">v2.0</div>
      <div class="stat-label">API Version</div>
    </div>
    <div class="stat-box" id="db-box">
      <div class="stat-value" id="db-status-container">
        <span class="status-pill connected" id="db-pill">
          <span style="font-size:1.1rem;line-height:1">●</span> CONNECTED
        </span>
      </div>
      <div class="stat-label" id="db-meta-label">PostgreSQL : 5433</div>
    </div>
  </div>

  <!-- Action Cards -->
  <div class="actions-grid">
    <a href="/swagger" class="action-card card-swagger">
      <div class="card-icon-wrap">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
          <line x1="16" y1="13" x2="8" y2="13"></line>
          <line x1="16" y1="17" x2="8" y2="17"></line>
          <polyline points="10 9 9 9 8 9"></polyline>
        </svg>
      </div>
      <div class="card-title">
        <span>Swagger UI Explorer</span>
      </div>
      <div class="card-desc">
        Full interactive API sandbox. Test all 159 endpoints with live requests, Bearer JWT authorization, request payload builders, and parameter validation.
      </div>
      <div class="card-link-arrow">
        Launch Swagger Explorer <span>→</span>
      </div>
    </a>

    <a href="/redoc" class="action-card card-redoc">
      <div class="card-icon-wrap">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
        </svg>
      </div>
      <div class="card-title">
        <span>ReDoc Reference</span>
      </div>
      <div class="card-desc">
        Sleek, three-panel technical documentation manual. Complete schema models, response codes, and clean type hierarchies for frontend engineers.
      </div>
      <div class="card-link-arrow">
        Open Technical Reference <span>→</span>
      </div>
    </a>

    <a href="/openapi.json" class="action-card card-openapi">
      <div class="card-icon-wrap">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
        </svg>
      </div>
      <div class="card-title">
        <span>OpenAPI 3.1 Spec</span>
      </div>
      <div class="card-desc">
        Standard OpenAPI JSON schema. Direct import into Postman, Insomnia, Paw, or CLI code generators for automatic TypeScript / Python SDK builds.
      </div>
      <div class="card-link-arrow">
        Download JSON Spec <span>→</span>
      </div>
    </a>
  </div>

  <!-- Quick Test Credentials -->
  <div class="creds-banner">
    <div class="creds-info">
      <div class="creds-icon">🔑</div>
      <div class="creds-text">
        <h4>Pre-Seeded Test Credentials</h4>
        <p>Click any pill to copy credentials for testing in Swagger UI Authorize dialog</p>
      </div>
    </div>
    <div class="creds-pills">
      <div class="copy-pill" onclick="copyText('admin@nhit.com', this)" title="Click to copy email">
        <span>👤 admin@nhit.com</span>
      </div>
      <div class="copy-pill" onclick="copyText('password123', this)" title="Click to copy password">
        <span>🔒 password123</span>
      </div>
      <div class="copy-pill" onclick="copyText('finance@nhit.com', this)" title="Click to copy finance email">
        <span>👤 finance@nhit.com</span>
      </div>
    </div>
  </div>

  <!-- Microservices Section -->
  <div class="section-head">
    <div class="section-title">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="color:var(--primary-light)">
        <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
        <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
        <line x1="6" y1="6" x2="6.01" y2="6"></line>
        <line x1="6" y1="18" x2="6.01" y2="18"></line>
      </svg>
      <span>Microservices Directory</span>
    </div>
    <input type="text" id="svc-search" class="search-box" placeholder="Filter microservices..." onkeyup="filterServices()">
  </div>

  <div class="services-grid" id="services-grid">
    <!-- Populated dynamically via /api/v1/health -->
  </div>

</div>

<script>
let servicesData = {};

function copyText(text, el) {
  navigator.clipboard.writeText(text);
  const orig = el.innerHTML;
  el.innerHTML = '<span style="color:#10b981">✓ Copied!</span>';
  setTimeout(() => { el.innerHTML = orig; }, 1500);
}

function filterServices() {
  const q = document.getElementById('svc-search').value.toLowerCase();
  const container = document.getElementById('services-grid');
  container.innerHTML = '';
  for (const [name, info] of Object.entries(servicesData)) {
    if (name.toLowerCase().includes(q) || info.description.toLowerCase().includes(q)) {
      renderCard(name, info);
    }
  }
}

function renderCard(name, info) {
  const container = document.getElementById('services-grid');
  const card = document.createElement('div');
  card.className = 'svc-card';
  card.innerHTML = `
    <div class="svc-header">
      <div class="svc-name">${name}</div>
      <span class="svc-badge">REST Gateway</span>
    </div>
    <div class="svc-port">${info.port}</div>
    <div class="svc-desc">${info.description}</div>
  `;
  container.appendChild(card);
}

// Fetch live system health
fetch('/api/v1/health')
  .then(res => res.json())
  .then(data => {
    servicesData = data.services || {};
    const pill = document.getElementById('db-pill');
    const meta = document.getElementById('db-meta-label');
    
    if (data.database === 'connected') {
      pill.className = 'status-pill connected';
      pill.innerHTML = '<span style="font-size:1.1rem;line-height:1">●</span> CONNECTED';
      meta.textContent = 'PostgreSQL : 5433 (nhit_db)';
    } else {
      pill.className = 'status-pill connecting';
      pill.innerHTML = '<span style="font-size:1.1rem;line-height:1">●</span> DEGRADED';
      meta.textContent = 'Database Unreachable';
    }

    const container = document.getElementById('services-grid');
    container.innerHTML = '';
    for (const [name, info] of Object.entries(servicesData)) {
      renderCard(name, info);
    }
  })
  .catch(err => {
    const pill = document.getElementById('db-pill');
    if (pill) {
      pill.className = 'status-pill connecting';
      pill.innerHTML = '<span style="font-size:1.1rem;line-height:1">●</span> OFFLINE';
    }
  });
</script>
</body>
</html>\"\"\")

# ═══════════════════════════════════════════════════════
"""

# Replace the dashboard section
import re
pattern = re.compile(r'# DASHBOARD \(HTML\).*?# ═══════════════════════════════════════════════════════\s*if __name__', re.DOTALL)

replacement_full = dashboard_replacement + "\nif __name__"

if pattern.search(content):
    content = pattern.sub(replacement_full, content, count=1)
    print("Dashboard replaced successfully via regex.")
    with open('gateway_swagger.py', 'w', encoding='utf-8') as f:
        f.write(content)
    print("Saved gateway_swagger.py!")
else:
    print("ERROR: Regex did not match dashboard.")
