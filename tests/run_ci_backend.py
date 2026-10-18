import subprocess
import sys
import time
import urllib.request
import os

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def main():
    gateway_script = os.path.join(BASE_DIR, "NHIT Backend", "gateway_swagger.py")
    test_script = os.path.join(BASE_DIR, "tests", "test_full_suite.py")

    print("=" * 70)
    print(" NHIT CI BACKEND RUNNER: Gateway Daemon + Test Suite")
    print("=" * 70)
    print(f"Launching Gateway: {gateway_script}")

    # Inherit environment with test mode enabled
    env = os.environ.copy()
    env["ENVIRONMENT"] = "test"
    env["PORT"] = "8080"
    env["API_BASE_URL"] = "http://localhost:8080/api/v1"

    gateway_proc = subprocess.Popen(
        [sys.executable, gateway_script],
        cwd=BASE_DIR,
        env=env
    )
    print(f"Gateway process launched with PID: {gateway_proc.pid}")

    health_url = "http://localhost:8080/api/v1/health"
    print(f"Waiting for Gateway to respond on {health_url}...")
    ready = False
    for i in range(30):
        # Check if gateway died prematurely
        if gateway_proc.poll() is not None:
            print(f"ERROR: Gateway process exited unexpectedly with code {gateway_proc.returncode}")
            sys.exit(1)
        try:
            with urllib.request.urlopen(health_url, timeout=2) as resp:
                if resp.status == 200:
                    print(f"Gateway is UP and healthy on port 8080 (attempt {i+1})!")
                    ready = True
                    break
        except Exception:
            time.sleep(1)

    if not ready:
        print("ERROR: Gateway failed to report healthy within 30 seconds.")
        gateway_proc.kill()
        sys.exit(1)

    print("\nExecuting Full Integration Test Suite...")
    test_res = subprocess.run([sys.executable, test_script], cwd=BASE_DIR, env=env)

    print(f"\nTerminating Gateway daemon (PID: {gateway_proc.pid})...")
    gateway_proc.terminate()
    try:
        gateway_proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        gateway_proc.kill()

    print("=" * 70)
    if test_res.returncode == 0:
        print(" CI BACKEND VERIFICATION COMPLETE: ALL TESTS PASSED!")
    else:
        print(f" CI BACKEND VERIFICATION FAILED: Test suite returned code {test_res.returncode}")
    print("=" * 70)
    sys.exit(test_res.returncode)

if __name__ == "__main__":
    main()
