import os
import sys
import time
import threading
import uvicorn

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BACKEND_DIR = os.path.join(BASE_DIR, "NHIT Backend")
TESTS_DIR = os.path.join(BASE_DIR, "tests")

# Configure environment for testing
os.environ["ENVIRONMENT"] = "test"
os.environ["API_BASE_URL"] = "http://127.0.0.1:8080/api/v1"

# Ensure both backend and tests are importable
sys.path.insert(0, BACKEND_DIR)
sys.path.insert(0, TESTS_DIR)

def main():
    print("=" * 70, flush=True)
    print(" NHIT IN-PROCESS CI TEST RUNNER", flush=True)
    print("=" * 70, flush=True)
    print(f"Backend Directory: {BACKEND_DIR}", flush=True)
    print(f"Database Config  : {os.environ.get('DB_HOST', 'localhost')}:{os.environ.get('DB_PORT', '5432')}", flush=True)

    # 1. Import Gateway App
    print("Importing FastAPI Gateway from gateway_swagger.py...", flush=True)
    try:
        import gateway_swagger
        print("Gateway imported successfully!", flush=True)
    except Exception as e:
        print(f"FATAL: Failed to import gateway_swagger: {e}", flush=True)
        import traceback
        traceback.print_exc()
        sys.exit(1)

    # 2. Start Uvicorn in a daemon thread
    print("Starting Uvicorn HTTP server on 127.0.0.1:8080...", flush=True)
    config = uvicorn.Config(
        gateway_swagger.app,
        host="127.0.0.1",
        port=8080,
        log_level="warning"
    )
    server = uvicorn.Server(config)
    server_thread = threading.Thread(target=server.run, daemon=True)
    server_thread.start()

    # 3. Wait for HTTP readiness
    import urllib.request
    health_url = "http://127.0.0.1:8080/api/v1/health"
    ready = False
    for i in range(20):
        time.sleep(0.5)
        try:
            with urllib.request.urlopen(health_url, timeout=2) as resp:
                if resp.status == 200:
                    print(f"Gateway server is UP and responding on port 8080 (attempt {i+1})!", flush=True)
                    ready = True
                    break
        except Exception:
            pass

    if not ready:
        print("FATAL: Gateway server failed to respond on 127.0.0.1:8080 within 10 seconds.", flush=True)
        sys.exit(1)

    # 4. Import and execute the full test suite
    print("\nExecuting End-to-End Verification Suite...", flush=True)
    try:
        import test_full_suite
        test_full_suite.run_tests()
        print("\nALL 18 INTEGRATION & END-TO-END TESTS PASSED!", flush=True)
    except Exception as e:
        print(f"\nFATAL: Integration test failed with exception: {e}", flush=True)
        import traceback
        traceback.print_exc()
        sys.exit(1)

    print("=" * 70, flush=True)
    print(" CI TEST RUNNER COMPLETED WITH 100% SUCCESS!", flush=True)
    print("=" * 70, flush=True)
    sys.exit(0)

if __name__ == "__main__":
    main()
