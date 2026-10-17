module github.com/ShristiRnr/NHIT_Backend/services/department-service

go 1.24.11

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/pkg/logger v0.0.0
	github.com/ShristiRnr/NHIT_Backend/pkg/middleware v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
)

replace (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb => ../../api/pb/authpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb => ../../api/pb/departmentpb
	github.com/ShristiRnr/NHIT_Backend/pkg/logger => ../../pkg/logger
	github.com/ShristiRnr/NHIT_Backend/pkg/middleware => ../../pkg/middleware
)
