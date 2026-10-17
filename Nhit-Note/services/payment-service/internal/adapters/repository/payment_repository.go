package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/core/ports"
)

type paymentRepository struct {
	db      *sql.DB
	queries *generated.Queries
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sql.DB) ports.PaymentRepository {
	return &paymentRepository{
		db:      db,
		queries: generated.New(db),
	}
}

// CreatePaymentRequests creates multiple payments under the same sl_no
func (r *paymentRepository) CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	for _, payment := range payments {
		_, err := qtx.CreatePayment(ctx, generated.CreatePaymentParams{
			SlNo:              slNo,
			TemplateType:      generated.TemplateType(payment.TemplateType),
			Project:           sqlNullString(payment.Project),
			AccountFullName:   sqlNullString(payment.AccountFullName),
			FromAccountType:   sqlNullString(payment.FromAccountType),
			FullAccountNumber: sqlNullString(payment.FullAccountNumber),
			ToAccount:         sqlNullString(payment.ToAccount),
			ToAccountType:     sqlNullString(payment.ToAccountType),
			NameOfBeneficiary: sqlNullString(payment.NameOfBeneficiary),
			AccountNumber:     sqlNullString(payment.AccountNumber),
			NameOfBank:        sqlNullString(payment.NameOfBank),
			IfscCode:          sqlNullString(payment.IfscCode),
			Amount:            fmt.Sprintf("%.2f", payment.Amount),
			Purpose:           sqlNullString(payment.Purpose),
			Status:            generated.NullPaymentStatus{PaymentStatus: generated.PaymentStatus(payment.Status), Valid: true},
			UserID:            payment.UserID,
			PaymentNoteID:     sqlNullInt64(payment.PaymentNoteID),
		})
		if err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}
	}

	return tx.Commit()
}

// GetPaymentGroup retrieves all payments with the given sl_no
func (r *paymentRepository) GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error) {
	payments, err := r.queries.GetPaymentsBySlNo(ctx, slNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	if len(payments) == 0 {
		return nil, fmt.Errorf("payment group not found")
	}

	group := &domain.PaymentGroup{
		SlNo:      slNo,
		Status:    string(payments[0].Status.PaymentStatus),
		Payments:  make([]domain.Payment, len(payments)),
		CreatedAt: payments[0].CreatedAt,
	}

	for i, p := range payments {
		group.Payments[i] = r.toDomain(&p)
	}

	return group, nil
}

// GetPaymentByID retrieves a single payment by ID
func (r *paymentRepository) GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error) {
	payment, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	result := r.toDomain(&payment)
	return &result, nil
}

// List retrieves payment groups with filters
func (r *paymentRepository) List(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error) {
	// Count total groups
	count, err := r.queries.CountPaymentGroups(ctx, generated.CountPaymentGroupsParams{
		Column1: stringPtrToString(filters.Status),
		Column2: filters.OnlyAssigned,
		UserID:  filters.UserID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payment groups: %w", err)
	}

	// Get groups
	offset := (filters.Page - 1) * filters.PerPage
	rows, err := r.queries.ListPayments(ctx, generated.ListPaymentsParams{
		Column1: stringPtrToString(filters.Status),
		Column2: filters.OnlyAssigned,
		UserID:  filters.UserID,
		Limit:   filters.PerPage,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}

	// For each sl_no, get full payment group
	groups := make([]*domain.PaymentGroup, 0)
	for _, row := range rows {
		group, err := r.GetPaymentGroup(ctx, row.SlNo)
		if err != nil {
			continue
		}
		groups = append(groups, group)
	}

	return groups, count, nil
}

// UpdatePayment updates a single payment
func (r *paymentRepository) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	_, err := r.queries.UpdatePayment(ctx, generated.UpdatePaymentParams{
		ID:                payment.ID,
		TemplateType:      generated.TemplateType(payment.TemplateType),
		Project:           sqlNullString(payment.Project),
		AccountFullName:   sqlNullString(payment.AccountFullName),
		FromAccountType:   sqlNullString(payment.FromAccountType),
		FullAccountNumber: sqlNullString(payment.FullAccountNumber),
		ToAccount:         sqlNullString(payment.ToAccount),
		ToAccountType:     sqlNullString(payment.ToAccountType),
		NameOfBeneficiary: sqlNullString(payment.NameOfBeneficiary),
		AccountNumber:     sqlNullString(payment.AccountNumber),
		NameOfBank:        sqlNullString(payment.NameOfBank),
		IfscCode:          sqlNullString(payment.IfscCode),
		Amount:            fmt.Sprintf("%.2f", payment.Amount),
		Purpose:           sqlNullString(payment.Purpose),
		Status:            generated.NullPaymentStatus{PaymentStatus: generated.PaymentStatus(payment.Status), Valid: true},
	})
	return err
}

// UpdatePaymentGroupStatus updates status for all payments in a group
func (r *paymentRepository) UpdatePaymentGroupStatus(ctx context.Context, slNo string, status string) error {
	return r.queries.UpdatePaymentStatus(ctx, generated.UpdatePaymentStatusParams{
		SlNo:   slNo,
		Status: generated.NullPaymentStatus{PaymentStatus: generated.PaymentStatus(status), Valid: true},
	})
}

// DeletePayment deletes a single payment
func (r *paymentRepository) DeletePayment(ctx context.Context, id int64) error {
	return r.queries.DeletePayment(ctx, id)
}

// DeletePaymentGroup deletes all payments in a group
func (r *paymentRepository) DeletePaymentGroup(ctx context.Context, slNo string) error {
	return r.queries.DeletePaymentsBySlNo(ctx, slNo)
}

// LinkVendorAccount links a payment to a vendor account
func (r *paymentRepository) LinkVendorAccount(ctx context.Context, paymentID int64, vendorID int64, vendorAccountID *int64) error {
	_, err := r.queries.InsertPaymentVendorAccount(ctx, generated.InsertPaymentVendorAccountParams{
		PaymentID:       paymentID,
		VendorID:        vendorID,
		VendorAccountID: sqlNullInt64(vendorAccountID),
	})
	return err
}

// CreateShortcut creates a payment shortcut
func (r *paymentRepository) CreateShortcut(ctx context.Context, shortcut *domain.PaymentShortcut) (*domain.PaymentShortcut, error) {
	created, err := r.queries.CreatePaymentShortcut(ctx, generated.CreatePaymentShortcutParams{
		SlNo:            sqlNullString(shortcut.SlNo),
		ShortcutName:    shortcut.ShortcutName,
		RequestDataJson: shortcut.RequestDataJSON,
		UserID:          shortcut.UserID,
	})
	if err != nil {
		return nil, err
	}

	return &domain.PaymentShortcut{
		ID:              created.ID,
		SlNo:            nullStringToPtr(created.SlNo),
		ShortcutName:    created.ShortcutName,
		RequestDataJSON: created.RequestDataJson,
		UserID:          created.UserID,
		CreatedAt:       created.CreatedAt,
		UpdatedAt:       created.UpdatedAt,
	}, nil
}

// GetShortcut retrieves a shortcut by ID
func (r *paymentRepository) GetShortcut(ctx context.Context, id int64) (*domain.PaymentShortcut, error) {
	shortcut, err := r.queries.GetPaymentShortcut(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentShortcut{
		ID:              shortcut.ID,
		SlNo:            nullStringToPtr(shortcut.SlNo),
		ShortcutName:    shortcut.ShortcutName,
		RequestDataJSON: shortcut.RequestDataJson,
		UserID:          shortcut.UserID,
		CreatedAt:       shortcut.CreatedAt,
		UpdatedAt:       shortcut.UpdatedAt,
	}, nil
}

// ListShortcuts lists shortcuts for a user
func (r *paymentRepository) ListShortcuts(ctx context.Context, userID int64) ([]*domain.PaymentShortcut, error) {
	shortcuts, err := r.queries.ListPaymentShortcuts(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.PaymentShortcut, len(shortcuts))
	for i, s := range shortcuts {
		result[i] = &domain.PaymentShortcut{
			ID:              s.ID,
			SlNo:            nullStringToPtr(s.SlNo),
			ShortcutName:    s.ShortcutName,
			RequestDataJSON: s.RequestDataJson,
			UserID:          s.UserID,
			CreatedAt:       s.CreatedAt,
			UpdatedAt:       s.UpdatedAt,
		}
	}

	return result, nil
}

// GenerateSerialNumber generates the next payment serial number
func (r *paymentRepository) GenerateSerialNumber(ctx context.Context, prefix string) (string, error) {
	nextNum, err := r.queries.GenerateSerialNumber(ctx, sql.NullString{String: prefix, Valid: true})
	if err != nil {
		return "", err
	}

	year := time.Now().Year()
	month := time.Now().Month()
	serialNumber := fmt.Sprintf("%s-%d-%02d-%05d", prefix, year, month, nextNum)
	return serialNumber, nil
}

// AddBankLetterLog adds a bank letter approval log
func (r *paymentRepository) AddBankLetterLog(ctx context.Context, log *domain.BankLetterApprovalLog) error {
	_, err := r.queries.InsertBankLetterLog(ctx, generated.InsertBankLetterLogParams{
		SlNo:          log.SlNo,
		Status:        log.Status,
		Comments:      sqlNullString(log.Comments),
		ReviewerID:    log.ReviewerID,
		ReviewerName:  sqlNullString(log.ReviewerName),
		ReviewerEmail: sqlNullString(log.ReviewerEmail),
		ApproverLevel: sqlNullInt32(log.ApproverLevel),
	})
	return err
}

// GetBankLetterLogs retrieves bank letter logs
func (r *paymentRepository) GetBankLetterLogs(ctx context.Context, slNo string) ([]*domain.BankLetterApprovalLog, error) {
	logs, err := r.queries.ListBankLetterLogs(ctx, slNo)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.BankLetterApprovalLog, len(logs))
	for i, log := range logs {
		result[i] = &domain.BankLetterApprovalLog{
			ID:            log.ID,
			SlNo:          log.SlNo,
			Status:        log.Status,
			Comments:      nullStringToPtr(log.Comments),
			ReviewerID:    log.ReviewerID,
			ReviewerName:  nullStringToPtr(log.ReviewerName),
			ReviewerEmail: nullStringToPtr(log.ReviewerEmail),
			ApproverLevel: nullInt32ToPtr(log.ApproverLevel),
			CreatedAt:     log.CreatedAt,
		}
	}

	return result, nil
}

// Helper functions
func (r *paymentRepository) toDomain(p *generated.Payment) domain.Payment {
	amount, _ := parseDecimal(p.Amount)

	return domain.Payment{
		ID:                p.ID,
		SlNo:              p.SlNo,
		TemplateType:      string(p.TemplateType),
		Project:           nullStringToPtr(p.Project),
		AccountFullName:   nullStringToPtr(p.AccountFullName),
		FromAccountType:   nullStringToPtr(p.FromAccountType),
		FullAccountNumber: nullStringToPtr(p.FullAccountNumber),
		ToAccount:         nullStringToPtr(p.ToAccount),
		ToAccountType:     nullStringToPtr(p.ToAccountType),
		NameOfBeneficiary: nullStringToPtr(p.NameOfBeneficiary),
		AccountNumber:     nullStringToPtr(p.AccountNumber),
		NameOfBank:        nullStringToPtr(p.NameOfBank),
		IfscCode:          nullStringToPtr(p.IfscCode),
		Amount:            amount,
		Purpose:           nullStringToPtr(p.Purpose),
		Status:            string(p.Status.PaymentStatus),
		UserID:            p.UserID,
		PaymentNoteID:     nullInt64ToPtr(p.PaymentNoteID),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

func sqlNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func sqlNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func sqlNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullInt64ToPtr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullInt32ToPtr(ni sql.NullInt32) *int32 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int32
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseDecimal(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
