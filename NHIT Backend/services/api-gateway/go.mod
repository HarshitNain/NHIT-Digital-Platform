module github.com/ShristiRnr/NHIT_Backend/services/api-gateway

go 1.24.11

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb v0.0.0-00010101000000-000000000000
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb v0.0.0
	github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb v0.0.0
	github.com/ShristiRnr/Nhit-Note v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/minio/minio-go/v7 v7.0.97
	golang.org/x/time v0.14.0
	google.golang.org/grpc v1.77.0
)

replace (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb => ../../api/pb/authpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb => ../../api/pb/departmentpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb => ../../api/pb/designationpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb => ../../api/pb/organizationpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb => ../../api/pb/projectpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/userpb => ../../api/pb/userpb
	github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb => ../../api/pb/vendorpb
	github.com/ShristiRnr/NHIT_Backend/pkg/logger => ../../pkg/logger
	github.com/ShristiRnr/Nhit-Note => ../../../Nhit-Note
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/minio/crc64nvme v1.1.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251124214823-79d6a2a48846 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
