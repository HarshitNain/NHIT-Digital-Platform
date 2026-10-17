package grpc

import (
	"context"

	"github.com/ShristiRnr/Nhit-Note/api/pb/commonpb"
	paymentpb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentpb"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/core/ports"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	service ports.PaymentService
}

func NewPaymentHandler(service ports.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *paymentpb.ListPaymentsRequest) (*paymentpb.ListPaymentsResponse, error) {
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

	protoGroups := make([]*paymentpb.PaymentGroup, len(groups))
	for i, g := range groups {
		protoGroups[i] = h.toProtoGroup(g)
	}

	return &paymentpb.ListPaymentsResponse{
		Groups: protoGroups,
		Pagination: &commonpb.Pagination{
			TotalItems: total,
		},
	}, nil
}

func (h *PaymentHandler) GetPaymentGroup(ctx context.Context, req *paymentpb.GetPaymentGroupRequest) (*paymentpb.PaymentGroupResponse, error) {
	group, err := h.service.GetPaymentGroup(ctx, req.SlNo)
	if err != nil {
		return nil, err
	}

	return &paymentpb.PaymentGroupResponse{
		Group: h.toProtoGroup(group),
	}, nil
}

func (h *PaymentHandler) CreatePaymentRequests(ctx context.Context, req *paymentpb.CreatePaymentRequestsRequest) (*paymentpb.PaymentGroupResponse, error) {
	payments := make([]domain.Payment, len(req.Vendors))
	for i, v := range req.Vendors {
		payments[i] = h.itemToDomain(v, req.CreatedBy)
	}

	group, err := h.service.CreatePaymentRequests(ctx, "", payments, req.CreateShortcut, req.ShortcutName)
	if err != nil {
		return nil, err
	}

	return &paymentpb.PaymentGroupResponse{
		Group: h.toProtoGroup(group),
	}, nil
}

func (h *PaymentHandler) toProtoGroup(g *domain.PaymentGroup) *paymentpb.PaymentGroup {
	payments := make([]*paymentpb.Payment, len(g.Payments))
	for i, p := range g.Payments {
		payments[i] = h.toProtoPayment(&p)
	}

	return &paymentpb.PaymentGroup{
		SlNo:      g.SlNo,
		Status:    g.Status,
		Payments:  payments,
		CreatedAt: g.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *PaymentHandler) toProtoPayment(p *domain.Payment) *paymentpb.Payment {
	return &paymentpb.Payment{
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

func (h *PaymentHandler) itemToDomain(v *paymentpb.PaymentRequestItem, userID int64) domain.Payment {
	return domain.Payment{
		TemplateType:      v.TemplateType,
		Project:           &v.Project,
		AccountFullName:   &v.AccountFullName,
		FromAccountType:   &v.FromAccountType,
		FullAccountNumber: &v.FromAccountNumber,
		ToAccount:         &v.ToAccountType,
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

func (h *PaymentHandler) GeneratePaymentSerialNumber(ctx context.Context, _ *emptypb.Empty) (*paymentpb.GeneratePaymentSerialNumberResponse, error) {
	sn, err := h.service.GenerateSerialNumber(ctx)
	if err != nil {
		return nil, err
	}
	return &paymentpb.GeneratePaymentSerialNumberResponse{SlNo: sn}, nil
}
