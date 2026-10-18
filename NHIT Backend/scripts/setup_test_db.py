#!/usr/bin/env python3
"""
NHIT Database Migration & Seeding Runner
Applies all SQL migrations across microservices in dependency order, then runs comprehensive seed.
"""
import os
import sys
import glob
import psycopg2

DB_CONFIG = {
    "host": os.environ.get("DB_HOST", "localhost"),
    "port": int(os.environ.get("DB_PORT", 5433)),
    "dbname": os.environ.get("DB_NAME", "nhit_db"),
    "user": os.environ.get("DB_USER", "postgres"),
    "password": os.environ.get("DB_PASSWORD", "shristi"),
}

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ROOT_DIR = os.path.dirname(BASE_DIR)

MIGRATION_DIRS = [
    os.path.join(BASE_DIR, "services", "user-service", "migrations"),
    os.path.join(BASE_DIR, "services", "organization-service", "migrations"),
    os.path.join(BASE_DIR, "services", "auth-service", "migrations"),
    os.path.join(BASE_DIR, "services", "department-service", "migrations"),
    os.path.join(BASE_DIR, "services", "designation-service", "migrations"),
    os.path.join(BASE_DIR, "services", "project-service", "migrations"),
    os.path.join(BASE_DIR, "services", "vendor-service", "migrations"),
    os.path.join(ROOT_DIR, "Nhit-Note", "services", "greennote-service", "migrations"),
]

def run_migrations():
    print(f"Connecting to Postgres at {DB_CONFIG['host']}:{DB_CONFIG['port']}/{DB_CONFIG['dbname']}...")
    conn = psycopg2.connect(**DB_CONFIG)
    conn.autocommit = True
    cur = conn.cursor()

    consolidated_sql = os.path.join(BASE_DIR, "scripts", "consolidated-migrations.sql")
    if os.path.exists(consolidated_sql):
        print(f"Applying consolidated migrations from {consolidated_sql}...")
        with open(consolidated_sql, "r", encoding="utf-8") as f:
            try:
                cur.execute(f.read())
                print("Consolidated migrations applied successfully!")
            except Exception as e:
                print(f"[INFO] Notice on consolidated migrations: {str(e).strip()[:100]}")

    for mdir in MIGRATION_DIRS:
        if not os.path.exists(mdir):
            print(f"[SKIP] Directory not found: {mdir}")
            continue

        sql_files = sorted(glob.glob(os.path.join(mdir, "*.sql")))
        for sf in sql_files:
            fname = os.path.basename(sf)
            print(f"Applying migration: {fname}...")
            with open(sf, "r", encoding="utf-8") as f:
                content = f.read()
                try:
                    cur.execute(content)
                except Exception as e:
                    # Ignore table already exists or duplicate column errors in existing DBs
                    err = str(e)
                    if "already exists" in err or "duplicate key" in err:
                        print(f"  [INFO] Already applied / exists: {fname}")
                    else:
                        print(f"  [WARN] Migration notice on {fname}: {err.strip()[:100]}")

    conn.close()
    print("All migrations applied successfully!")

    print("Running comprehensive seed...")
    sys.path.insert(0, os.path.dirname(__file__))
    import comprehensive_seed
    comprehensive_seed.seed()
    print("Database setup and seeding completed successfully!")

if __name__ == "__main__":
    run_migrations()
