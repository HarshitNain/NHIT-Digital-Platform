module github.com/ShristiRnr/NHIT_Backend/pkg/middleware

go 1.22

require (
	github.com/ShristiRnr/NHIT_Backend/api/pb/authpb v0.0.0
	google.golang.org/grpc v1.59.0
)

replace github.com/ShristiRnr/NHIT_Backend/api/pb/authpb => ../../api/pb/authpb
