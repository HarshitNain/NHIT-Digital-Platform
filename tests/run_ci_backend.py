import os
import sys
import time
import threading
import traceback
import urllib.request

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BACKEND_DIR = os.path.join(BASE_DIR, "NHIT Backend")
TESTS_DIR = os.path.join(BASE_DIR, "tests")

# Configure environment for testing
os.environ["ENVIRONMENT"] = "test"
os.environ["API_BASE_URL"] = "http://127.0.0.1:8080/api/v1"
if "DB_HOST" not in os.environ:
    os.environ["DB_HOST"] = "127.0.0.1"
if "DB_PORT" not in os.environ:
    os.environ["DB_PORT"] = "5432"
if "DB_NAME" not in os.environ:
    os.environ["DB_NAME"] = "nhit_db"
if "DB_USER" not in os.environ:
    os.environ["DB_USER"] = "postgres"
if "DB_PASSWORD" not in os.environ:
    os.environ["DB_PASSWORD"] = "password"

# Ensure both backend and tests are importable
sys.path.insert(0, BACKEND_DIR)
sys.path.insert(0, TESTS_DIR)

def main():
    print("=" * 70, flush=True)
    print(" NHIT IN-PROCESS CI TEST RUNNER", flush=True)
    print("=" * 70, flush=True)
    print(f"Backend Directory: {BACKEND_DIR}", flush=True)
    print(f"Database Config  : {os.environ.get('DB_USER')}@{os.environ.get('DB_HOST')}:{os.environ.get('DB_PORT')}/{os.environ.get('DB_NAME')}", flush=True)

    # 1. Import Gateway App
    print("Importing FastAPI Gateway from gateway_swagger.py...", flush=True)
    try:
        import gateway_swagger
        # Guarantee DB_CONFIG alignment
        gateway_swagger.DB_CONFIG["host"] = os.environ["DB_HOST"]
        gateway_swagger.DB_CONFIG["port"] = int(os.environ["DB_PORT"])
        gateway_swagger.DB_CONFIG["dbname"] = os.environ["DB_NAME"]
        gateway_swagger.DB_CONFIG["user"] = os.environ["DB_USER"]
        gateway_swagger.DB_CONFIG["password"] = os.environ["DB_PASSWORD"]
        print(f"Gateway imported successfully with DB {gateway_swagger.DB_CONFIG['host']}:{gateway_swagger.DB_CONFIG['port']}!", flush=True)
    except Exception as e:
        print(f"FATAL: Failed to import gateway_swagger: {e}", flush=True)
        traceback.print_exc()
        sys.exit(1)

    # 2. Start Uvicorn in a daemon thread
    import uvicorn
    print("Starting Uvicorn HTTP server on 127.0.0.1:8080...", flush=True)
    config = uvicorn.Config(
        gateway_swagger.app,
        host="127.0.0.1",
        port=8080,
        log_level="info"
    )
    server = uvicorn.Server(config)
    
    server_error = []
    def run_server():
        try:
            server.run()
        except Exception as e:
            server_error.append(e)
            print(f"\n[SERVER THREAD CRASH] {e}", flush=True)
            traceback.print_exc()

    server_thread = threading.Thread(target=run_server, daemon=True)
    server_thread.start()

    # 3. Wait for HTTP readiness
    health_url = "http://127.0.0.1:8080/api/v1/health"
    ready = False
    print(f"Polling readiness endpoint at {health_url}...", flush=True)
    for i in range(25):
        time.sleep(0.4)
        if server_error:
            print(f"FATAL: Uvicorn server thread died with: {server_error[0]}", flush=True)
            sys.exit(1)
        try:
            with urllib.request.urlopen(health_url, timeout=2) as resp:
                if resp.status == 200:
                    body = resp.read().decode("utf-8")
                    print(f"Gateway server is UP and responding on port 8080 (attempt {i+1})! Health: {body[:100]}", flush=True)
                    ready = True
                    break
        except Exception as err:
            pass

    if not ready:
        print("FATAL: Gateway server failed to respond on 127.0.0.1:8080 within 10 seconds.", flush=True)
        if server_error:
            print(f"Thread exception: {server_error[0]}", flush=True)
        sys.exit(1)

    # 4. Import and execute the full test suite
    print("\nExecuting End-to-End Verification Suite...", flush=True)
    try:
        import test_full_suite
        test_full_suite.run_tests()
        print("\nALL 18 INTEGRATION & END-TO-END TESTS PASSED!", flush=True)
    except Exception as e:
        print(f"\nFATAL: Integration test failed with exception: {e}", flush=True)
        traceback.print_exc()
        sys.exit(1)

    print("=" * 70, flush=True)
    print(" CI TEST RUNNER COMPLETED WITH 100% SUCCESS!", flush=True)
    print("=" * 70, flush=True)
    sys.exit(0)

if __name__ == "__main__":
    main()
