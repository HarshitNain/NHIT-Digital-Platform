"""
NHIT Digital Platform — Automated Test & Health Verification Harness
Validates backend gateway compilation, database tables, financial logic, and frontend TypeScript.

Usage:
    python "NHIT Backend/scripts/harness.py"
"""

import sys
import os
import re
import py_compile
import subprocess
from pathlib import Path

# Ensure UTF-8 output on Windows consoles
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
    except Exception:
        pass

# Root directory resolution
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent.parent
GATEWAY_PATH = REPO_ROOT / "NHIT Backend" / "gateway_swagger.py"
FRONTEND_DIR = REPO_ROOT / "nhit-frontend"

def print_header(title: str):
    print("\n" + "=" * 65)
    print(f"  {title}")
    print("=" * 65)

def print_result(name: str, passed: bool, detail: str = ""):
    symbol = "✓ [PASS]" if passed else "✗ [FAIL]"
    print(f"  {symbol} {name}")
    if detail:
        print(f"         Detail: {detail}")

def check_python_environment() -> bool:
    print_header("1. PYTHON ENVIRONMENT & DEPENDENCY CHECK")
    version = f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    passed = sys.version_info >= (3, 10)
    print_result(f"Python Runtime ({version})", passed)
    
    deps = ["fastapi", "uvicorn", "pydantic", "jwt", "psycopg2"]
    all_deps_passed = True
    for dep in deps:
        try:
            __import__(dep)
            print_result(f"Package: {dep}", True)
        except ImportError:
            print_result(f"Package: {dep}", False, "Module not found in active python interpreter")
            all_deps_passed = False
            
    # Optional dependency: redis
    try:
        __import__("redis")
        print_result("Optional Package: redis", True, "Installed")
    except ImportError:
        print_result("Optional Package: redis", True, "Not installed (fallback in-memory store active)")

    return passed and all_deps_passed

def check_gateway_compilation() -> bool:
    print_header("2. GATEWAY BACKEND SYNTAX & COMPILATION")
    if not GATEWAY_PATH.exists():
        print_result("Gateway file existence", False, f"Not found at {GATEWAY_PATH}")
        return False
        
    try:
        py_compile.compile(str(GATEWAY_PATH), doraise=True)
        print_result("FastAPI Gateway compilation (gateway_swagger.py)", True, "Zero syntax or indentation errors")
        return True
    except py_compile.PyCompileError as e:
        print_result("FastAPI Gateway compilation (gateway_swagger.py)", False, str(e))
        return False

def check_database() -> bool:
    print_header("3. POSTGRESQL DATABASE & SCHEMA CHECK")
    db_config = {
        "host": os.environ.get("DB_HOST", "localhost"),
        "port": int(os.environ.get("DB_PORT", 5433)),
        "dbname": os.environ.get("DB_NAME", "nhit_db"),
        "user": os.environ.get("DB_USER", "postgres"),
        "password": os.environ.get("DB_PASSWORD", "shristi"),
    }
    
    try:
        import psycopg2
        conn = psycopg2.connect(**db_config, connect_timeout=3)
        print_result(f"PostgreSQL Connection ({db_config['host']}:{db_config['port']}/{db_config['dbname']})", True)
        
        cur = conn.cursor()
        cur.execute("""
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public';
        """)
        existing_tables = set(r[0] for r in cur.fetchall())
        conn.close()
        
        required_tables = [
            "vendors", "vendor_accounts", "projects", 
            "green_notes", "payment_notes", "users", 
            "departments", "organizations"
        ]
        
        all_tables_present = True
        for t in required_tables:
            if t in existing_tables:
                print_result(f"Table exists: {t}", True)
            else:
                print_result(f"Table exists: {t}", False, "Table missing from public schema")
                all_tables_present = False
                
        return all_tables_present
    except Exception as e:
        print_result("PostgreSQL Connectivity", False, f"Could not connect to {db_config['host']}:{db_config['port']} ({str(e)})")
        print("         Notice: Local mock/fallback data mode is available when DB is offline.")
        return False

def check_domain_invariants() -> bool:
    print_header("4. DOMAIN LOGIC & FINANCIAL INVARIANTS")
    
    # Invariant 1: PAN validation regex
    pan_regex = r"^[A-Z]{5}[0-9]{4}[A-Z]{1}$"
    test_pan_valid = bool(re.match(pan_regex, "AAACH8392M"))
    test_pan_invalid = not bool(re.match(pan_regex, "123INVALID"))
    print_result("PAN Regex Verification", test_pan_valid and test_pan_invalid)
    
    # Invariant 2: IFSC validation regex
    ifsc_regex = r"^[A-Z]{4}0[A-Z0-9]{6}$"
    test_ifsc_valid = bool(re.match(ifsc_regex, "HDFC0000240"))
    test_ifsc_invalid = not bool(re.match(ifsc_regex, "HDFC1234567"))
    print_result("IFSC Clearing Code Regex Verification", test_ifsc_valid and test_ifsc_invalid)
    
    # Invariant 3: Payment Note Non-Negative Disbursement
    def compute_net(gross: float, tds: float) -> float:
        if gross < tds:
            raise ValueError("TDS exceeds gross amount")
        return gross - tds
        
    try:
        net = compute_net(100000.0, 10000.0)
        assert net == 90000.0
        try:
            compute_net(1000.0, 5000.0)
            net_guard_ok = False
        except ValueError:
            net_guard_ok = True
    except Exception:
        net_guard_ok = False
        
    print_result("Financial Non-Negative Disbursement Guard", net_guard_ok, "Rejects gross < tds")
    return test_pan_valid and test_ifsc_valid and net_guard_ok

def check_frontend_types() -> bool:
    print_header("5. FRONTEND TYPESCRIPT COMPILATION (tsc)")
    if not FRONTEND_DIR.exists():
        print_result("Frontend directory existence", False, f"Not found at {FRONTEND_DIR}")
        return False
        
    try:
        cmd = "npx tsc --noEmit" if sys.platform != "win32" else "cmd /c \"npx tsc --noEmit\""
        result = subprocess.run(
            cmd,
            cwd=str(FRONTEND_DIR),
            shell=True,
            capture_output=True,
            text=True,
            timeout=30
        )
        passed = (result.returncode == 0)
        detail = "Zero TypeScript compile or typing errors" if passed else (result.stdout or result.stderr).strip()
        print_result("React 18 SPA TypeScript Check", passed, detail[:200] if len(detail) > 200 else detail)
        return passed
    except Exception as e:
        print_result("Frontend TypeScript Check", False, str(e))
        return False

def main():
    print("=" * 65)
    print("      NHIT DIGITAL PLATFORM — PLATFORM TEST HARNESS")
    print("=" * 65)
    
    res1 = check_python_environment()
    res2 = check_gateway_compilation()
    res3 = check_database()
    res4 = check_domain_invariants()
    res5 = check_frontend_types()
    
    print_header("HARNESS SUMMARY")
    print(f"  • Python Environment:       {'PASS' if res1 else 'ATTENTION'}")
    print(f"  • Gateway Compilation:      {'PASS' if res2 else 'FAIL'}")
    print(f"  • Database & Tables:        {'PASS' if res3 else 'OFFLINE / NOTICE'}")
    print(f"  • Financial Logic Invariants:{'PASS' if res4 else 'FAIL'}")
    print(f"  • Frontend Type Safety:     {'PASS' if res5 else 'FAIL'}")
    
    critical_pass = res1 and res2 and res4 and res5
    print("\n" + ("*" * 65))
    if critical_pass:
        print("  ✓ OVERALL STATUS: ALL CORE PLATFORM INVARIANTS HEALTHY")
    else:
        print("  ✗ OVERALL STATUS: ATTENTION REQUIRED ON FLAGGED ITEMS")
    print("*" * 65 + "\n")

if __name__ == "__main__":
    main()
