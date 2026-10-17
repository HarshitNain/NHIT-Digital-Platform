package services

import (
	"context"
	"fmt"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"
)

type poTrackingService struct {
	repo ports.POTrackingRepository
}

func NewPOTrackingService(repo ports.POTrackingRepository) ports.POTrackingService {
	return &poTrackingService{
		repo: repo,
	}
}

func (s *poTrackingService) GetOrCreatePOTracking(ctx context.Context, params domain.CreatePOTrackingParams) (*domain.POTracking, error) {
	existing, err := s.repo.GetByOrderNo(ctx, params.OrderNo)
	if err == nil {
		return existing, nil
	}

	// Calculate initial budget etc
	params.RemainingBudget = params.OrderValue - params.TotalUtilized
	if params.OrderValue > 0 {
		params.UtilizationPercentage = (params.TotalUtilized / params.OrderValue) * 100
	} else {
		params.UtilizationPercentage = 0
	}

	if err := s.repo.Create(ctx, params); err != nil {
		return nil, err
	}

	return s.repo.GetByOrderNo(ctx, params.OrderNo)
}

func (s *poTrackingService) GetPOByOrderNumber(ctx context.Context, orderNo string) (*domain.POTracking, error) {
	return s.repo.GetByOrderNo(ctx, orderNo)
}

func (s *poTrackingService) CalculateExpenditures(ctx context.Context, orderNo string, invoiceValue float64) (map[string]interface{}, error) {
	po, err := s.repo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	actualExpenditure := invoiceValue
	budgetExpenditure := po.RemainingBudget
	expenditureOverBudget := "No"
	if actualExpenditure > budgetExpenditure {
		expenditureOverBudget = "Yes"
	}

	remainingBudget := po.RemainingBudget - actualExpenditure
	utilizationPercentage := 0.0
	if po.OrderValue > 0 {
		utilizationPercentage = ((po.TotalUtilized + actualExpenditure) / po.OrderValue) * 100
	}

	return map[string]interface{}{
		"actual_expenditure":      fmt.Sprintf("%.2f", actualExpenditure),
		"budget_expenditure":      fmt.Sprintf("%.2f", budgetExpenditure),
		"expenditure_over_budget": expenditureOverBudget,
		"remaining_budget":        fmt.Sprintf("%.2f", remainingBudget),
		"utilization_percentage":  utilizationPercentage,
	}, nil
}

func (s *poTrackingService) UpdatePOUtilization(ctx context.Context, poTrackingID int64, greenNoteID int64) error {
	po, err := s.repo.GetByID(ctx, poTrackingID)
	if err != nil {
		return err
	}

	// For manual triggers, we might need to recalculate from scratch or use provided data.
	// For now, let's just trigger a full recalculation for this PO.
	return s.updatePO(ctx, po)
}

func (s *poTrackingService) UpdatePOUtilizationByOrderNo(ctx context.Context, orderNo string, amount float64) error {
	po, err := s.repo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}

	po.TotalUtilized += amount
	po.GreenNotesCount++
	
	return s.updatePO(ctx, po)
}

func (s *poTrackingService) updatePO(ctx context.Context, po *domain.POTracking) error {
	po.RemainingBudget = po.OrderValue - po.TotalUtilized
	if po.OrderValue > 0 {
		po.UtilizationPercentage = (po.TotalUtilized / po.OrderValue) * 100
	} else {
		po.UtilizationPercentage = 0
	}

	return s.repo.UpdateUtilization(ctx, po.ID, po.TotalUtilized, po.RemainingBudget, po.UtilizationPercentage, po.GreenNotesCount)
}

func (s *poTrackingService) ValidatePOBudget(ctx context.Context, orderNo string, invoiceValue float64, excludeGreenNoteID int64) (bool, string, *domain.POTracking, error) {
	po, err := s.repo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return false, "PO Tracking record not found", nil, nil
	}

	if po.RemainingBudget < invoiceValue {
		return false, fmt.Sprintf("Insufficient budget. Available: %.2f, Requested: %.2f", po.RemainingBudget, invoiceValue), po, nil
	}

	return true, "Budget valid", po, nil
}

func (s *poTrackingService) GetPOSummary(ctx context.Context, orderNo string) (map[string]interface{}, error) {
	po, err := s.repo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return map[string]interface{}{"exists": false, "message": "PO not found"}, nil
	}

	return map[string]interface{}{
		"exists":                 true,
		"order_no":               po.OrderNo,
		"order_value":            po.OrderValue,
		"total_utilized":         po.TotalUtilized,
		"remaining_budget":       po.RemainingBudget,
		"utilization_percentage": po.UtilizationPercentage,
		"green_notes_count":      po.GreenNotesCount,
		"status":                 po.Status,
	}, nil
}

func (s *poTrackingService) RecalculateAll(ctx context.Context) (int, error) {
	pos, err := s.repo.ListAll(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, po := range pos {
		// Recalculate logic here
		po.RemainingBudget = po.OrderValue - po.TotalUtilized
		if po.OrderValue > 0 {
			po.UtilizationPercentage = (po.TotalUtilized / po.OrderValue) * 100
		}
		
		err := s.repo.UpdateUtilization(ctx, po.ID, po.TotalUtilized, po.RemainingBudget, po.UtilizationPercentage, po.GreenNotesCount)
		if err == nil {
			count++
		}
	}

	return count, nil
}
