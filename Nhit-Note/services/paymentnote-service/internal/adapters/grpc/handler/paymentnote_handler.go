package handler

import (
	"context"
	"strconv"
	
	"github.com/ShristiRnr/Nhit-Note/api/pb/commonpb"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/ports"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentNoteHandler struct {
	paymentnotepb.UnimplementedPaymentNoteServiceServer
	service ports.PaymentNoteService
}

// NewPaymentNoteHandler creates a new payment note handler
func NewPaymentNoteHandler(service ports.PaymentNoteService) *PaymentNoteHandler {
	return &PaymentNoteHandler{
		service: service,
	}
}

// CreatePaymentNote creates a new payment note
func (h *PaymentNoteHandler) CreatePaymentNote(ctx context.Context, req *paymentnotepb.CreatePaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	note, err := h.service.CreatePaymentNote(ctx, h.service.ProtoToDomain(req.GetNote()))
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// GetPaymentNote retrieves a payment note by ID
func (h *PaymentNoteHandler) GetPaymentNote(ctx context.Context, req *paymentnotepb.GetPaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	note, err := h.service.GetPaymentNoteByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// ListPaymentNotes lists payment notes with filters
func (h *PaymentNoteHandler) ListPaymentNotes(ctx context.Context, req *paymentnotepb.ListPaymentNotesRequest) (*paymentnotepb.ListPaymentNotesResponse, error) {
	filters := domain.PaymentNoteFilters{
		Status:  &req.Status,
		Search:  &req.Search,
		Page:    req.GetPage(),
		PerPage: req.GetPerPage(),
	}
	
	notes, total, err := h.service.ListPaymentNotes(ctx, filters)
	if err != nil {
		return nil, err
	}
	
	protoNotes := make([]*paymentnotepb.PaymentNoteSummary, len(notes))
	for i, note := range notes {
		protoNotes[i] = h.service.DomainToSummaryProto(note)
	}
	
	return &paymentnotepb.ListPaymentNotesResponse{
		Notes: protoNotes,
		Pagination: &commonpb.Pagination{
			TotalItems: total,
		},
	}, nil
}

// UpdatePaymentNote updates a payment note
func (h *PaymentNoteHandler) UpdatePaymentNote(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	note, err := h.service.UpdatePaymentNote(ctx, req.GetId(), h.service.ProtoToDomain(req.GetNote()))
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// DeletePaymentNote deletes a payment note
func (h *PaymentNoteHandler) DeletePaymentNote(ctx context.Context, req *paymentnotepb.DeletePaymentNoteRequest) (*emptypb.Empty, error) {
	err := h.service.DeletePaymentNote(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	
	return &emptypb.Empty{}, nil
}

// GeneratePaymentNoteOrderNumber generates an order number
func (h *PaymentNoteHandler) GeneratePaymentNoteOrderNumber(ctx context.Context, _ *emptypb.Empty) (*paymentnotepb.GeneratePaymentNoteOrderNumberResponse, error) {
	orderNo, err := h.service.GeneratePaymentNoteOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.GeneratePaymentNoteOrderNumberResponse{
		OrderNumber: orderNo,
	}, nil
}

// PutPaymentNoteOnHold puts a payment note on hold
func (h *PaymentNoteHandler) PutPaymentNoteOnHold(ctx context.Context, req *paymentnotepb.PutPaymentNoteOnHoldRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	userID, _ := strconv.ParseInt(req.GetUserId(), 10, 64)
	note, err := h.service.PutOnHold(ctx, req.GetId(), req.GetHoldReason(), userID)
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// RemovePaymentNoteFromHold removes a payment note from hold
func (h *PaymentNoteHandler) RemovePaymentNoteFromHold(ctx context.Context, req *paymentnotepb.RemovePaymentNoteFromHoldRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	note, err := h.service.RemoveFromHold(ctx, req.GetId(), req.GetNewStatus())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// UpdatePaymentNoteUTR updates UTR information
func (h *PaymentNoteHandler) UpdatePaymentNoteUTR(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteUtrRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	note, err := h.service.UpdateUTR(ctx, req.GetId(), req.GetUtrNo(), req.GetUtrDate())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}
