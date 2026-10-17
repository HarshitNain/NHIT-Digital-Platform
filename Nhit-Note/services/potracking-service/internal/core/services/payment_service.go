package services

import (
	"context"

	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"
)

type paymentService struct {
	repo ports.PaymentRepository
}

func NewPaymentService(repo ports.PaymentRepository) ports.PaymentService {
	return &paymentService{
		repo: repo,
	}
}

func (s *paymentService) CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment, createShortcut bool, shortcutName string) (*domain.PaymentGroup, error) {
	if slNo == "" {
		newSlNo, err := s.repo.GenerateSerialNumber(ctx, "PAY")
		if err != nil {
			return nil, err
		}
		slNo = newSlNo
	}

	for i := range payments {
		payments[i].SlNo = slNo
		if payments[i].Status == "" {
			payments[i].Status = "D" // Draft
		}
	}

	err := s.repo.CreatePaymentRequests(ctx, slNo, payments)
	if err != nil {
		return nil, err
	}

	if createShortcut {
		// Simplified shortcut creation
		_, _ = s.repo.CreateShortcut(ctx, &domain.PaymentShortcut{
			ShortcutName:    shortcutName,
			SlNo:            &slNo,
			UserID:          payments[0].UserID,
			RequestDataJSON: "{}", // In a real app, serialize the payments array
		})
	}

	return s.repo.GetPaymentGroup(ctx, slNo)
}

func (s *paymentService) GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error) {
	return s.repo.GetPaymentGroup(ctx, slNo)
}

func (s *paymentService) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error) {
	return s.repo.List(ctx, filters)
}

func (s *paymentService) UpdatePaymentGroup(ctx context.Context, slNo string, status string, vendors []domain.Payment) (*domain.PaymentGroup, error) {
	if len(vendors) > 0 {
		for _, v := range vendors {
			err := s.repo.UpdatePayment(ctx, &v)
			if err != nil {
				return nil, err
			}
		}
	}

	if status != "" {
		err := s.repo.UpdatePaymentGroupStatus(ctx, slNo, status)
		if err != nil {
			return nil, err
		}
	}

	return s.repo.GetPaymentGroup(ctx, slNo)
}

func (s *paymentService) DeletePayment(ctx context.Context, id int64, slNo string) error {
	return s.repo.DeletePayment(ctx, id)
}

func (s *paymentService) DeletePaymentItem(ctx context.Context, slNo string, id int64) error {
	return s.repo.DeletePayment(ctx, id)
}

func (s *paymentService) GenerateSerialNumber(ctx context.Context) (string, error) {
	return s.repo.GenerateSerialNumber(ctx, "PAY")
}

func (s *paymentService) CreateBankLetter(ctx context.Context, paymentNoteIDs []int64, actorID string) (string, int, float64, error) {
	slNo, err := s.repo.GenerateSerialNumber(ctx, "BL")
	if err != nil {
		return "", 0, 0, err
	}
	// Logic to aggregate notes and create logs would go here
	return slNo, len(paymentNoteIDs), 0, nil
}

func (s *paymentService) ProcessBankLetterLog(ctx context.Context, slNo string, status string, remarks string, reviewerID string) error {
	return s.repo.AddBankLetterLog(ctx, &domain.BankLetterApprovalLog{
		SlNo:       slNo,
		Status:     status,
		Comments:   &remarks,
		ReviewerID: reviewerID,
	})
}
