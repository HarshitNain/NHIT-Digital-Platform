package handler

import (
	"context"
	"time"

	"github.com/ShristiRnr/Nhit-Note/api/pb/potrackingpb"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"

	"google.golang.org/protobuf/types/known/emptypb"
)

type POTrackingHandler struct {
	potrackingpb.UnimplementedPOTrackingServiceServer
	service ports.POTrackingService
}

func NewPOTrackingHandler(service ports.POTrackingService) *POTrackingHandler {
	return &POTrackingHandler{
		service: service,
	}
}

func (h *POTrackingHandler) GetOrCreatePOTracking(ctx context.Context, req *potrackingpb.GetOrCreatePOTrackingRequest) (*potrackingpb.POTrackingResponse, error) {
	orderDate, _ := time.Parse("2006-01-02", req.OrderDate)
	params := domain.CreatePOTrackingParams{
		OrderNo:               req.OrderNo,
		VendorID:              req.VendorId,
		SupplierID:            &req.SupplierId,
		OrderDate:             orderDate,
		OrderValue:            req.TotalAmount,
		Description:           req.BriefOfGoodsServices,
		Status:                "ACTIVE",
	}

	po, err := h.service.GetOrCreatePOTracking(ctx, params)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.POTrackingResponse{
		PoTracking: h.toProto(po),
	}, nil
}

func (h *POTrackingHandler) GetPOByOrderNumber(ctx context.Context, req *potrackingpb.GetPORequest) (*potrackingpb.POTrackingResponse, error) {
	po, err := h.service.GetPOByOrderNumber(ctx, req.OrderNo)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.POTrackingResponse{
		PoTracking: h.toProto(po),
	}, nil
}

func (h *POTrackingHandler) CalculateExpenditures(ctx context.Context, req *potrackingpb.CalculateExpendituresRequest) (*potrackingpb.CalculateExpendituresResponse, error) {
	res, err := h.service.CalculateExpenditures(ctx, req.OrderNo, req.InvoiceValue)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.CalculateExpendituresResponse{
		ActualExpenditure:      res["actual_expenditure"].(string),
		BudgetExpenditure:      res["budget_expenditure"].(string),
		ExpenditureOverBudget: res["expenditure_over_budget"].(string),
		RemainingBudget:        res["remaining_budget"].(string),
		UtilizationPercentage:  res["utilization_percentage"].(float64),
	}, nil
}

func (h *POTrackingHandler) UpdatePOUtilization(ctx context.Context, req *potrackingpb.UpdatePOUtilizationRequest) (*emptypb.Empty, error) {
	err := h.service.UpdatePOUtilization(ctx, req.PoTrackingId, req.GreenNoteId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *POTrackingHandler) ValidatePOBudget(ctx context.Context, req *potrackingpb.ValidatePOBudgetRequest) (*potrackingpb.ValidatePOBudgetResponse, error) {
	valid, msg, po, err := h.service.ValidatePOBudget(ctx, req.OrderNo, req.InvoiceValue, req.ExcludeGreenNoteId)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.ValidatePOBudgetResponse{
		Valid:                valid,
		Message:              msg,
		PoTracking:           h.toProto(po),
		CurrentUtilization:   po.TotalUtilized,
		RemainingBudget:      po.RemainingBudget,
	}, nil
}

func (h *POTrackingHandler) GetPOSummary(ctx context.Context, req *potrackingpb.GetPORequest) (*potrackingpb.GetPOSummaryResponse, error) {
	res, err := h.service.GetPOSummary(ctx, req.OrderNo)
	if err != nil {
		return nil, err
	}

	if !res["exists"].(bool) {
		return &potrackingpb.GetPOSummaryResponse{
			Exists:  false,
			Message: res["message"].(string),
		}, nil
	}

	return &potrackingpb.GetPOSummaryResponse{
		Exists:                true,
		OrderNo:               res["order_no"].(string),
		OrderValue:            res["order_value"].(float64),
		TotalUtilized:         res["total_utilized"].(float64),
		RemainingBudget:       res["remaining_budget"].(float64),
		UtilizationPercentage: res["utilization_percentage"].(float64),
		GreenNotesCount:      res["green_notes_count"].(int32),
		Status:                res["status"].(string),
	}, nil
}

func (h *POTrackingHandler) RecalculateAllPOTracking(ctx context.Context, _ *emptypb.Empty) (*potrackingpb.RecalculateAllResponse, error) {
	count, err := h.service.RecalculateAll(ctx)
	if err != nil {
		return nil, err
	}

	return &potrackingpb.RecalculateAllResponse{
		RecordsUpdated: int32(count),
	}, nil
}

func (h *POTrackingHandler) toProto(po *domain.POTracking) *potrackingpb.POTracking {
	if po == nil {
		return nil
	}
	
	var supplierID string
	if po.SupplierID != nil {
		supplierID = *po.SupplierID
	}

	return &potrackingpb.POTracking{
		Id:                    po.ID,
		OrderNo:               po.OrderNo,
		VendorId:              po.VendorID,
		SupplierId:            supplierID,
		OrderDate:             po.OrderDate.Format("2006-01-02"),
		OrderValue:            po.OrderValue,
		TotalUtilized:         po.TotalUtilized,
		RemainingBudget:       po.RemainingBudget,
		GreenNotesCount:      po.GreenNotesCount,
		Description:           po.Description,
		Status:                po.Status,
		UtilizationPercentage: po.UtilizationPercentage,
	}
}
