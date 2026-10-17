package grpcadapter

import (
	"context"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/services"
)

// Server is the gRPC adapter that exposes the GreenNoteService over gRPC.
type Server struct {
	greennotepb.UnimplementedGreenNoteServiceServer
	app *services.GreenNoteService
}

// NewGreenNoteGRPCServer wires the core GreenNoteService into a gRPC service implementation.
func NewGreenNoteGRPCServer(app *services.GreenNoteService) greennotepb.GreenNoteServiceServer {
	return &Server{app: app}
}

func (s *Server) ListGreenNotes(ctx context.Context, req *greennotepb.ListGreenNotesRequest) (*greennotepb.ListGreenNotesResponse, error) {
	return s.app.ListGreenNotes(ctx, req)
}

func (s *Server) GetGreenNote(ctx context.Context, req *greennotepb.GetGreenNoteRequest) (*greennotepb.GreenNoteDetailResponse, error) {
	return s.app.GetGreenNote(ctx, req)
}

func (s *Server) CreateGreenNote(ctx context.Context, req *greennotepb.CreateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	return s.app.CreateGreenNote(ctx, req)
}

func (s *Server) UpdateGreenNote(ctx context.Context, req *greennotepb.UpdateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	return s.app.UpdateGreenNote(ctx, req)
}

func (s *Server) CancelGreenNote(ctx context.Context, req *greennotepb.CancelGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	return s.app.CancelGreenNote(ctx, req)
}

func (s *Server) GetOrganizationProjects(ctx context.Context, req *greennotepb.GetOrganizationProjectsRequest) (*greennotepb.GetOrganizationProjectsResponse, error) {
	return s.app.GetOrganizationProjects(ctx, req)
}

func (s *Server) GetOrganizationVendors(ctx context.Context, req *greennotepb.GetOrganizationVendorsRequest) (*greennotepb.GetOrganizationVendorsResponse, error) {
	return s.app.GetOrganizationVendors(ctx, req)
}

func (s *Server) GetOrganizationDepartments(ctx context.Context, req *greennotepb.GetOrganizationDepartmentsRequest) (*greennotepb.GetOrganizationDepartmentsResponse, error) {
	return s.app.GetOrganizationDepartments(ctx, req)
}

func (s *Server) UploadGreenNoteDocuments(ctx context.Context, req *greennotepb.UploadGreenNoteDocumentsRequest) (*greennotepb.UploadGreenNoteDocumentsResponse, error) {
	return s.app.UploadGreenNoteDocuments(ctx, req)
}

func (s *Server) LinkPaymentNote(ctx context.Context, req *greennotepb.LinkPaymentNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	return s.app.LinkPaymentNote(ctx, req)
}
