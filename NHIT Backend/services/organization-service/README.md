# Organization Service

A microservice for managing organizations in a multi-tenant system, built using **Hexagonal Architecture (Ports & Adapters)** pattern.

## Overview

This service manages organizations and their relationships with users, implementing strong business logic separation from infrastructure concerns. It provides complete organization lifecycle management including creation, updates, status toggling, and user-organization relationships.

## Architecture

### Hexagonal Architecture (Ports & Adapters)

```
┌─────────────────────────────────────────────────────────────┐
│                      Adapters Layer                          │
│  ┌──────────────────┐              ┌───────────────────┐   │
│  │  gRPC Handlers   │              │  PostgreSQL Repos │   │
│  │  (Input Port)    │              │  (Output Port)    │   │
│  └─────────┬────────┘              └─────────┬─────────┘   │
└────────────┼──────────────────────────────────┼─────────────┘
             │                                  │
             │                                  │
┌────────────▼──────────────────────────────────▼─────────────┐
│                      Ports Layer                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Service Interfaces & Repository Interfaces         │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────────┬───────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────┐
│                      Core/Domain Layer                        │
│  ┌────────────────┐  ┌───────────────────────────────┐      │
│  │ Business Logic │  │  Domain Models & Validation   │      │
│  │  (Services)    │  │  (Organization, UserOrg)      │      │
│  └────────────────┘  └───────────────────────────────┘      │
└───────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### Domain/Core Layer
- **Domain Models**: `Organization`, `UserOrganization`
- **Business Logic**: Validation, state management, business rules
- **No External Dependencies**: Pure business logic

#### Ports Layer
- **Service Interfaces**: `OrganizationService`, `UserOrganizationService`
- **Repository Interfaces**: `OrganizationRepository`, `UserOrganizationRepository`
- **Data Transfer Objects**: Pagination params and results

#### Adapters Layer
- **gRPC Handlers**: Handle incoming requests, convert proto to domain models
- **PostgreSQL Repositories**: Implement persistence logic
- **External Integrations**: Database, messaging (future)

## Features

### Organization Management
- ✅ Create organizations with validation
- ✅ Update organization details
- ✅ Delete organizations (with business rules)
- ✅ Toggle organization status (activate/deactivate)
- ✅ Check organization code availability
- ✅ Get organization by ID or code
- ✅ List organizations by tenant (with pagination)
- ✅ List accessible organizations for users (with pagination)

### User-Organization Management
- ✅ Add users to organizations with roles
- ✅ Remove users from organizations
- ✅ Switch user's current organization context
- ✅ Get user's current organization
- ✅ List all organizations for a user
- ✅ Update user's role within an organization
- ✅ List all users in an organization

## Domain Models

### Organization

```go
type Organization struct {
    OrgID        uuid.UUID
    TenantID     uuid.UUID
    Name         string      // 3-255 characters
    Code         string      // 2-10 uppercase alphanumeric
    DatabaseName string      // Generated from code
    Description  string
    Logo         string      // File path/URL
    IsActive     bool
    CreatedBy    uuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Business Rules:**
- Organization code must be unique and uppercase
- Database name is auto-generated from code (e.g., `org_nhit`)
- Name must be 3-255 characters
- Code validation: 2-10 characters, uppercase alphanumeric only
- Database name validation: lowercase alphanumeric with underscores

### UserOrganization

```go
type UserOrganization struct {
    UserID           uuid.UUID
    OrgID            uuid.UUID
    RoleID           uuid.UUID
    IsCurrentContext bool
    JoinedAt         time.Time
    UpdatedAt        time.Time
}
```

**Business Rules:**
- User can belong to multiple organizations
- Only one organization can be the current context per user
- Switching context atomically updates current organization
- User must have a role within the organization

## Business Logic Highlights

### Strong Validation
- Domain models validate themselves on creation and update
- Code uniqueness check before creation/update
- Organization active status validation for access
- Prevent deletion of organizations with users

### Organization Context Switching
- Atomic transaction to ensure only one current context
- Validates organization is active before switching
- Validates user has access to organization
- Logs database context for debugging

### Separation of Concerns
- **Organization Service**: Manages organization lifecycle
- **UserOrganization Service**: Manages user-organization relationships
- Clear boundary between organization management and user associations

## API Endpoints (gRPC)

### Organization Operations
- `CreateOrganization` - Create a new organization
- `GetOrganization` - Get organization by ID
- `GetOrganizationByCode` - Get organization by code
- `UpdateOrganization` - Update organization details
- `DeleteOrganization` - Delete organization
- `ListOrganizationsByTenant` - List organizations for a tenant
- `ListAccessibleOrganizations` - List organizations accessible by user
- `ToggleOrganizationStatus` - Toggle organization active status
- `CheckOrganizationCode` - Check if code is available

## Database Schema

### Organizations Table
```sql
CREATE TABLE organizations (
    org_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL UNIQUE,
    database_name VARCHAR(64) NOT NULL UNIQUE,
    description TEXT,
    logo VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### User Organizations Table
```sql
CREATE TABLE user_organizations (
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    role_id UUID NOT NULL,
    is_current_context BOOLEAN NOT NULL DEFAULT false,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);
```

## Key Differences from PHP Implementation

### Architectural Improvements
1. **Hexagonal Architecture**: Clear separation between business logic and infrastructure
2. **Domain-Driven Design**: Rich domain models with self-validation
3. **Dependency Inversion**: Services depend on interfaces, not concrete implementations
4. **Single Responsibility**: Separate services for organization and user-organization concerns

### Business Logic Enhancements
1. **Stronger Validation**: Domain models validate themselves
2. **Explicit Error Types**: Custom error types for different failure scenarios
3. **Immutable IDs**: UUIDs generated at domain model creation
4. **Transaction Safety**: Atomic operations for critical business logic
5. **Database Name Generation**: Automatic and validated database naming

### Removed PHP-Specific Concerns
- ❌ Database cloning logic (should be infrastructure concern)
- ❌ Cache management (should be separate caching service)
- ❌ Logo upload logic (should be file storage service)
- ❌ Migration jobs (should be background job service)
- ❌ View concerns (frontend responsibility)

## Running the Service

### Prerequisites
- Go 1.24+
- PostgreSQL 13+
- Protocol Buffers compiler

### Setup Database
```bash
# Run migrations
psql -U postgres -d your_database -f migrations/001_create_organizations_tables.sql
```

### Build & Run
```bash
cd services/organization-service

# Install dependencies
go mod download

# Build
go build -o bin/organization-service cmd/server/main.go

# Run
./bin/organization-service
```

### Environment Variables
- `SERVICE_NAME`: organization-service
- `SERVER_PORT`: 50052 (default)
- `DATABASE_URL`: PostgreSQL connection string

## Testing

### Unit Tests
```bash
go test ./internal/core/domain/...
go test ./internal/core/services/...
```

### Integration Tests
```bash
go test ./internal/adapters/repository/... -tags=integration
```

## Future Enhancements

### Planned Features
- [ ] Database provisioning automation
- [ ] Organization data migration between databases
- [ ] Organization usage metrics and quotas
- [ ] Organization settings and preferences
- [ ] Organization invitations and onboarding
- [ ] Audit logging for organization changes
- [ ] Event sourcing for organization lifecycle

### Potential Improvements
- Add circuit breaker for external calls
- Implement distributed tracing
- Add rate limiting per organization
- Implement organization archiving (soft delete)
- Add organization billing integration
- Support organization hierarchies (parent/child)

## Contributing

When adding new features:
1. Add domain logic to `domain/` package
2. Define interfaces in `ports/` package
3. Implement services in `services/` package
4. Add repositories in `adapters/repository/`
5. Add gRPC handlers in `adapters/grpc/`
6. Update proto definitions if needed

## License

Copyright (c) 2025 NHIT Backend Team
