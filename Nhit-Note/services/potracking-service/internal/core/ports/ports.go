package ports

import (
	"context"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
)

type POTrackingRepository interface {
	Create(ctx context.Context, params domain.CreatePOTrackingParams) error
	GetByOrderNo(ctx context.Context, orderNo string) (*domain.POTracking, error)
	GetByID(ctx context.Context, id int64) (*domain.POTracking, error)
	UpdateUtilization(ctx context.Context, id int64, utilized float64, remaining float64, percentage float64, count int32) error
	ListAll(ctx context.Context) ([]*domain.POTracking, error)
}

type POTrackingService interface {
	GetOrCreatePOTracking(ctx context.Context, params domain.CreatePOTrackingParams) (*domain.POTracking, error)
	GetPOByOrderNumber(ctx context.Context, orderNo string) (*domain.POTracking, error)
	CalculateExpenditures(ctx context.Context, orderNo string, invoiceValue float64) (map[string]interface{}, error)
	UpdatePOUtilization(ctx context.Context, poTrackingID int64, greenNoteID int64) error
	UpdatePOUtilizationByOrderNo(ctx context.Context, orderNo string, amount float64) error
	ValidatePOBudget(ctx context.Context, orderNo string, invoiceValue float64, excludeGreenNoteID int64) (bool, string, *domain.POTracking, error)
	GetPOSummary(ctx context.Context, orderNo string) (map[string]interface{}, error)
	RecalculateAll(ctx context.Context) (int, error)
}

type PaymentRepository interface {
	CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment) error
	GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error)
	GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error)
	List(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error)
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	UpdatePaymentGroupStatus(ctx context.Context, slNo string, status string) error
	DeletePayment(ctx context.Context, id int64) error
	DeletePaymentGroup(ctx context.Context, slNo string) error
	LinkVendorAccount(ctx context.Context, paymentID int64, vendorID string, vendorAccountID *int64) error
	CreateShortcut(ctx context.Context, shortcut *domain.PaymentShortcut) (*domain.PaymentShortcut, error)
	GetShortcut(ctx context.Context, id int64) (*domain.PaymentShortcut, error)
	ListShortcuts(ctx context.Context, userID string) ([]*domain.PaymentShortcut, error)
	GenerateSerialNumber(ctx context.Context, prefix string) (string, error)
	AddBankLetterLog(ctx context.Context, log *domain.BankLetterApprovalLog) error
	GetBankLetterLogs(ctx context.Context, slNo string) ([]*domain.BankLetterApprovalLog, error)
}

type PaymentService interface {
	CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment, createShortcut bool, shortcutName string) (*domain.PaymentGroup, error)
	GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error)
	ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error)
	UpdatePaymentGroup(ctx context.Context, slNo string, status string, vendors []domain.Payment) (*domain.PaymentGroup, error)
	DeletePayment(ctx context.Context, id int64, slNo string) error
	DeletePaymentItem(ctx context.Context, slNo string, id int64) error
	GenerateSerialNumber(ctx context.Context) (string, error)
	CreateBankLetter(ctx context.Context, paymentNoteIDs []int64, actorID string) (string, int, float64, error)
	ProcessBankLetterLog(ctx context.Context, slNo string, status string, remarks string, reviewerID string) error
}
