package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
    "strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

    // 1. Run Migrations
    // Order matters: user-service (creates users) -> organization-service (creates user_organizations with FK) -> others
    migrationDirs := []string{
        "services/user-service/migrations",
        "services/organization-service/migrations",
        "services/auth-service/migrations",
        "services/department-service/migrations",
        "services/designation-service/migrations",
        "services/project-service/migrations",
        "services/vendor-service/migrations",
        "../Nhit-Note/services/greennote-service/migrations",
    }

    for _, migrationDir := range migrationDirs {
        fmt.Printf("Processing migrations in %s...\n", migrationDir)
        files, err := os.ReadDir(migrationDir)
        if err != nil {
            fmt.Printf("Warning: Failed to read migration dir %s: %v\n", migrationDir, err)
            continue
        }

        var sqlFiles []string
        for _, f := range files {
            if strings.HasSuffix(f.Name(), ".sql") {
                sqlFiles = append(sqlFiles, f.Name())
            }
        }
        sort.Strings(sqlFiles)

        for _, file := range sqlFiles {
            content, err := os.ReadFile(filepath.Join(migrationDir, file))
            if err != nil {
                 fmt.Printf("Failed to read %s: %v\n", file, err)
                 os.Exit(1)
            }
            _, err = conn.Exec(context.Background(), string(content))
            if err != nil {
                 // Log but don't exit for "already exists" errors ideally, but pgx doesn't make it super easy to check error codes portably in one line.
                 // Assuming "IF NOT EXISTS" is used in SQL files.
                 fmt.Printf("Note: Executing %s returned: %v\n", file, err)
            } else {
                fmt.Printf("Applied %s\n", file)
            }
        }
    }

	// 2. Seed Tenant
    var tenantID string
    err = conn.QueryRow(context.Background(), "SELECT tenant_id FROM tenants WHERE email='system@nhit.com'").Scan(&tenantID)
    if err == pgx.ErrNoRows {
        fmt.Println("Creating system tenant...")
        password, _ := hashPassword("system@123")
        err = conn.QueryRow(context.Background(), 
            "INSERT INTO tenants (name, email, password) VALUES ($1, $2, $3) RETURNING tenant_id",
            "System Tenant", "system@nhit.com", password).Scan(&tenantID)
        if err != nil {
             fmt.Printf("Failed to create tenant: %v\n", err)
             os.Exit(1)
        }
        fmt.Printf("Created tenant: %s\n", tenantID)
    } else if err != nil {
         fmt.Printf("Failed to query tenant: %v\n", err)
         os.Exit(1)
    } else {
        fmt.Printf("Using existing tenant: %s\n", tenantID)
    }

    // 3. Seed User
    var userID string
    err = conn.QueryRow(context.Background(), "SELECT user_id FROM users WHERE email='nhit@gmail.com'").Scan(&userID)
    if err == pgx.ErrNoRows {
        fmt.Println("Creating user nhit@gmail.com...")
        password, _ := hashPassword("nhit@123")
        err = conn.QueryRow(context.Background(), 
            "INSERT INTO users (tenant_id, name, email, password, email_verified_at) VALUES ($1, $2, $3, $4, $5) RETURNING user_id",
            tenantID, "NHIT Admin", "nhit@gmail.com", password, time.Now()).Scan(&userID)
         if err != nil {
             fmt.Printf("Failed to create user: %v\n", err)
             os.Exit(1)
        }
        fmt.Printf("Created user: %s\n", userID)
    } else {
         fmt.Printf("User already exists: %s\n", userID)
    }

    // 4. Get Role ID (SUPER_ADMIN)
    var roleID string
    err = conn.QueryRow(context.Background(), "SELECT role_id FROM roles WHERE name='SUPER_ADMIN' AND tenant_id=$1", tenantID).Scan(&roleID)
    if err == pgx.ErrNoRows {
         // Should have been created in step 3 or already exist
         fmt.Println("Warning: SUPER_ADMIN role not found!")
    } else {
         // Ensure role has ALL permissions
         _, err = conn.Exec(context.Background(), "UPDATE roles SET permissions = ARRAY(SELECT name FROM permissions) WHERE role_id=$1", roleID)
         if err != nil {
             fmt.Printf("Failed to update SUPER_ADMIN permissions: %v\n", err)
         } else {
             fmt.Println("Updated SUPER_ADMIN permissions")
         }

         // Also assign to user_roles (Tenant-level role assignment)
         _, err = conn.Exec(context.Background(), "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, roleID)
         if err != nil {
             fmt.Printf("Failed to assign role in user_roles: %v\n", err)
         } else {
             fmt.Println("Assigned SUPER_ADMIN to user_roles")
         }
    }

    // 5. Seed Organization
    var orgID string
    // Check if Head Office exists by NAME to preserve existing data
    err = conn.QueryRow(context.Background(), "SELECT org_id FROM organizations WHERE name='Head Office' AND tenant_id=$1", tenantID).Scan(&orgID)
    
    if err == pgx.ErrNoRows {
        fmt.Println("Creating Head Office organization...")
        // Use dynamic UUID generation for fresh creation - NO HARDCODING
        err = conn.QueryRow(context.Background(),
            `INSERT INTO organizations (org_id, tenant_id, name, code, database_name, status, super_admin_email) 
             VALUES (gen_random_uuid(), $1, 'Head Office', 'HO-001', 'nhit_db_ho', 0, 'nhit@gmail.com') 
             RETURNING org_id`,
            tenantID).Scan(&orgID)
            
        if err != nil {
             fmt.Printf("Failed to create organization: %v\n", err)
        } else {
            fmt.Printf("Created organization: %s\n", orgID)
        }
    } else if err != nil {
         fmt.Printf("Failed to query organization: %v\n", err)
    } else {
        fmt.Printf("Using existing organization: %s\n", orgID)
    }

    // 6. Link User to Organization
    if orgID != "" && userID != "" && roleID != "" {
        var dummy int
        err = conn.QueryRow(context.Background(), 
            "SELECT 1 FROM user_organizations WHERE user_id=$1 AND org_id=$2", userID, orgID).Scan(&dummy)
        if err == pgx.ErrNoRows {
             fmt.Println("Linking user to organization...")
             _, err = conn.Exec(context.Background(),
                `INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context)
                 VALUES ($1, $2, $3, true)`,
                userID, orgID, roleID)
             if err != nil {
                 fmt.Printf("Failed to link user to org: %v\n", err)
             } else {
                 fmt.Println("User linked to Head Office as SUPER_ADMIN")
             }
        } else {
            fmt.Println("User already linked to organization.")
        }
    }

    // 7. Seed Projects
    if orgID != "" && tenantID != "" {
        var projCount int
        err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM projects WHERE org_id=$1", orgID).Scan(&projCount)
        if err == nil && projCount == 0 {
            fmt.Println("Seeding default projects for Head Office...")
            projects := []string{"PROJECT A", "PROJECT B", "PROJECT 1", "PROJECT 2"}
            for _, pName := range projects {
                _, err = conn.Exec(context.Background(), 
                    "INSERT INTO projects (id, tenant_id, org_id, project_name, created_by) VALUES (gen_random_uuid(), $1, $2, $3, 'system')",
                    tenantID, orgID, pName)
                if err != nil {
                    fmt.Printf("Failed to seed project %s: %v\n", pName, err)
                }
            }
            fmt.Println("Default projects seeded successfully.")
        }
    }
    
    
    fmt.Println("Migration and Seeding Complete!")
}

func hashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", err
    }
    return string(hashedBytes), nil
}
