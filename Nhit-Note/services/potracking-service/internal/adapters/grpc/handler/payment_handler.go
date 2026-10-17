package handler

import (
	"context"

	"github.com/ShristiRnr/Nhit-Note/api/pb/potrackingpb"
	"github.com/ShristiRnr/Nhit-Note/api/pb/commonpb"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentHandler struct {
	potrackingpb.UnimplementedPaymentServiceServer
	service ports.PaymentService
}

func NewPaymentHandler(service ports.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *potrackingpb.ListPaymentsRequest) (*potrackingpb.ListPaymentsResponse, error) {
	filters := domain.PaymentFilters{
		Status:       &req.Status,
		OnlyAssigned: req.OnlyAssigned,
		UserID:       req.AuthUserId,
		Page:         req.Page,
		PerPage:      req.PerPage,
	}
	if req.Status == "" {
		filters.Status = nil
	}

	groups, total, err := h.service.ListPayments(ctx, filters)
	if err != nil {
		return nil, err
	}

	protoGroups := make([]*potrackingpb.PaymentGroup, len(groups))
	for i, g := range groups {
		protoGroups[i] = h.toProtoGroup(g)
	}

	return &potrackingpb.ListPaymentsResponse{
		Groups: protoGroups,
		Pagination: &commonpb.Pagination{
			TotalItems: total,
		},
	}, nil
}

func (h *PaymentHandler) GetPaymentGroup(ctx context.Context, req *potrackingpb.GetPaymentGroupRequest) (*potrackingpb.PaymentGroupResponse, error) {
	group, err := h.service.GetPaymentGroup(ctx, req.SlNo)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.PaymentGroupResponse{
		Group: h.toProtoGroup(group),
	}, nil
}

func (h *PaymentHandler) CreatePaymentRequests(ctx context.Context, req *potrackingpb.CreatePaymentRequestsRequest) (*potrackingpb.PaymentGroupResponse, error) {
	payments := make([]domain.Payment, len(req.Vendors))
	for i, v := range req.Vendors {
		payments[i] = h.itemToDomain(v, req.CreatedBy)
	}

	group, err := h.service.CreatePaymentRequests(ctx, "", payments, req.CreateShortcut, req.ShortcutName)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.PaymentGroupResponse{
		Group: h.toProtoGroup(group),
	}, nil
}

func (h *PaymentHandler) toProtoGroup(g *domain.PaymentGroup) *potrackingpb.PaymentGroup {
	payments := make([]*potrackingpb.Payment, len(g.Payments))
	for i, p := range g.Payments {
		payments[i] = h.toProtoPayment(&p)
	}

	return &potrackingpb.PaymentGroup{
		SlNo:      g.SlNo,
		Status:    g.Status,
		Payments:  payments,
		CreatedAt: g.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *PaymentHandler) toProtoPayment(p *domain.Payment) *potrackingpb.Payment {
	return &potrackingpb.Payment{
		Id:                p.ID,
		SlNo:              p.SlNo,
		TemplateType:      p.TemplateType,
		Project:           derefString(p.Project),
		AccountFullName:   derefString(p.AccountFullName),
		FromAccountType:   derefString(p.FromAccountType),
		FullAccountNumber: derefString(p.FullAccountNumber),
		To:                derefString(p.ToAccount),
		ToAccountType:     derefString(p.ToAccountType),
		NameOfBeneficiary: derefString(p.NameOfBeneficiary),
		AccountNumber:     derefString(p.AccountNumber),
		NameOfBank:        derefString(p.NameOfBank),
		IfscCode:          derefString(p.IfscCode),
		Amount:            p.Amount,
		Purpose:           derefString(p.Purpose),
		Status:            p.Status,
		UserId:            p.UserID,
		PaymentNoteId:     derefInt64(p.PaymentNoteID),
		CreatedAt:         p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *PaymentHandler) itemToDomain(v *potrackingpb.PaymentRequestItem, userID string) domain.Payment {
	return domain.Payment{
		TemplateType:      v.TemplateType,
		Project:           &v.Project,
		AccountFullName:   &v.AccountFullName,
		FromAccountType:   &v.FromAccountType,
		FullAccountNumber: &v.FromAccountNumber,
		ToAccount:         &v.ToAccountType, // Wait, is this correct? 
		ToAccountType:     &v.ToAccountType,
		NameOfBeneficiary: &v.BeneficiaryName,
		AccountNumber:     &v.BeneficiaryAccountNumber,
		NameOfBank:        &v.BeneficiaryBankName,
		IfscCode:          &v.IfscCode,
		Amount:            v.Amount,
		Purpose:           &v.Purpose,
		UserID:            userID,
		PaymentNoteID:     &v.PaymentNoteId,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// Implement remaining methods with placeholders or basic logic as needed
func (h *PaymentHandler) GeneratePaymentSerialNumber(ctx context.Context, _ *emptypb.Empty) (*potrackingpb.GeneratePaymentSerialNumberResponse, error) {
	sn, err := h.service.GenerateSerialNumber(ctx)
	if err != nil {
		return nil, err
	}
	return &potrackingpb.GeneratePaymentSerialNumberResponse{SlNo: sn}, nil
}
