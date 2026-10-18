import subprocess
import sys
import time
import urllib.request
import os

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BACKEND_DIR = os.path.join(BASE_DIR, "NHIT Backend")

def main():
    gateway_script = os.path.join(BACKEND_DIR, "gateway_swagger.py")
    test_script = os.path.join(BASE_DIR, "tests", "test_full_suite.py")

    print("=" * 70)
    print(" NHIT CI BACKEND RUNNER: Gateway Daemon + Test Suite")
    print("=" * 70)
    print(f"Base Directory   : {BASE_DIR}")
    print(f"Backend Directory: {BACKEND_DIR}")
    print(f"Gateway Script   : {gateway_script}")
    print(f"Test Suite Script: {test_script}")

    # Inherit and enrich environment
    env = os.environ.copy()
    env["ENVIRONMENT"] = "test"
    env["PORT"] = "8080"
    env["API_BASE_URL"] = "http://127.0.0.1:8080/api/v1"
    existing_pythonpath = env.get("PYTHONPATH", "")
    env["PYTHONPATH"] = BACKEND_DIR + (os.pathsep + existing_pythonpath if existing_pythonpath else "")

    # Pipe gateway logs to a file so we can diagnose any crash immediately
    gateway_log_path = os.path.join(BASE_DIR, "gateway_ci.log")
    gateway_log_file = open(gateway_log_path, "w", encoding="utf-8")

    gateway_proc = subprocess.Popen(
        [sys.executable, "-u", gateway_script],
        cwd=BACKEND_DIR,
        env=env,
        stdout=gateway_log_file,
        stderr=subprocess.STDOUT
    )
    print(f"Gateway process launched with PID: {gateway_proc.pid}")

    health_url = "http://127.0.0.1:8080/api/v1/health"
    print(f"Waiting for Gateway to report healthy on {health_url}...")
    ready = False
    for i in range(35):
        # Check if gateway died prematurely
        if gateway_proc.poll() is not None:
            gateway_log_file.flush()
            print(f"ERROR: Gateway process died prematurely with exit code {gateway_proc.returncode}!")
            print("--- Gateway Crash Log ---")
            with open(gateway_log_path, "r", encoding="utf-8", errors="replace") as f:
                print(f.read())
            print("-------------------------")
            sys.exit(1)

        try:
            with urllib.request.urlopen(health_url, timeout=2) as resp:
                if resp.status == 200:
                    data = resp.read().decode("utf-8")
                    print(f"Gateway is UP and healthy on port 8080 (attempt {i+1})! Response: {data[:100]}")
                    ready = True
                    break
        except Exception as e:
            time.sleep(1)

    if not ready:
        gateway_log_file.flush()
        print("ERROR: Gateway failed to report healthy within 35 seconds.")
        print("--- Gateway Diagnostic Log ---")
        with open(gateway_log_path, "r", encoding="utf-8", errors="replace") as f:
            print(f.read())
        print("------------------------------")
        gateway_proc.kill()
        sys.exit(1)

    print("\nExecuting Full Integration Test Suite (18 Routes)...")
    test_res = subprocess.run([sys.executable, test_script], cwd=BASE_DIR, env=env)

    print(f"\nShutting down Gateway daemon (PID: {gateway_proc.pid})...")
    gateway_proc.terminate()
    try:
        gateway_proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        gateway_proc.kill()
    gateway_log_file.close()

    print("=" * 70)
    if test_res.returncode == 0:
        print(" CI BACKEND VERIFICATION COMPLETE: ALL 18 TESTS PASSED WITH 100% SUCCESS!")
    else:
        print(f" CI BACKEND VERIFICATION FAILED: Test suite returned exit code {test_res.returncode}")
        with open(gateway_log_path, "r", encoding="utf-8", errors="replace") as f:
            print("--- Gateway Output during tests ---")
            print(f.read()[-2000:])
            print("-----------------------------------")
    print("=" * 70)
    sys.exit(test_res.returncode)

if __name__ == "__main__":
    main()
