module github.com/ShristiRnr/Nhit-Note

go 1.24.2

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/pkg/middleware v0.0.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/jackc/pgx/v5 v5.8.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.73
	github.com/segmentio/kafka-go v0.4.49
	google.golang.org/genproto/googleapis/api v0.0.0-20251124214823-79d6a2a48846
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sqlc-dev/pqtype v0.3.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
)

replace github.com/ShristiRnr/NHIT_Backend => "../NHIT Backend"

replace github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb => "../NHIT Backend/api/pb/projectpb"

replace github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb => "../NHIT Backend/api/pb/vendorpb"

replace github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb => "../NHIT Backend/api/pb/departmentpb"

replace github.com/ShristiRnr/NHIT_Backend/pkg/middleware => "../NHIT Backend/pkg/middleware"

replace github.com/ShristiRnr/NHIT_Backend/api/pb/authpb => "../NHIT Backend/api/pb/authpb"

replace github.com/ShristiRnr/NHIT_Backend/api/pb/userpb => "../NHIT Backend/api/pb/userpb"
