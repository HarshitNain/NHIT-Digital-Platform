package repository

import (
	"context"
	"fmt"
	"time"

	sqlc "github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type paymentRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPaymentRepository(db *pgxpool.Pool) ports.PaymentRepository {
	return &paymentRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *paymentRepository) CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	for _, p := range payments {
		_, err := qtx.CreatePayment(ctx, sqlc.CreatePaymentParams{
			SlNo:              slNo,
			TemplateType:      p.TemplateType,
			Project:           toPgTextFromPtr(p.Project),
			AccountFullName:   toPgTextFromPtr(p.AccountFullName),
			FromAccountType:   toPgTextFromPtr(p.FromAccountType),
			FullAccountNumber: toPgTextFromPtr(p.FullAccountNumber),
			ToAccount:         toPgTextFromPtr(p.ToAccount),
			ToAccountType:     toPgTextFromPtr(p.ToAccountType),
			NameOfBeneficiary: toPgTextFromPtr(p.NameOfBeneficiary),
			AccountNumber:     toPgTextFromPtr(p.AccountNumber),
			NameOfBank:        toPgTextFromPtr(p.NameOfBank),
			IfscCode:          toPgTextFromPtr(p.IfscCode),
			Amount:            toPgNumeric(p.Amount),
			Purpose:           toPgTextFromPtr(p.Purpose),
			Status:            p.Status,
			UserID:            stringToInt64(p.UserID),
			PaymentNoteID:     toPgInt8(p.PaymentNoteID),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *paymentRepository) GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error) {
	rows, err := r.queries.GetPaymentsBySlNo(ctx, slNo)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("payment group not found")
	}

	group := &domain.PaymentGroup{
		SlNo:      slNo,
		Status:    fmt.Sprintf("%v", rows[0].Status),
		CreatedAt: rows[0].CreatedAt.Time,
		Payments:  make([]domain.Payment, len(rows)),
	}

	for i, row := range rows {
		group.Payments[i] = r.toDomain(row)
	}

	return group, nil
}

func (r *paymentRepository) GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error) {
	row, err := r.queries.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p := r.toDomain(row)
	return &p, nil
}

func (r *paymentRepository) List(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error) {
	count, err := r.queries.CountPaymentGroups(ctx, sqlc.CountPaymentGroupsParams{
		Column1: stringPtrToString(filters.Status),
		Column2: filters.OnlyAssigned,
		UserID:  stringToInt64(filters.UserID),
	})
	if err != nil {
		return nil, 0, err
	}

	offset := (filters.Page - 1) * filters.PerPage
	rows, err := r.queries.ListPayments(ctx, sqlc.ListPaymentsParams{
		Column1: stringPtrToString(filters.Status),
		Column2: filters.OnlyAssigned,
		UserID:  stringToInt64(filters.UserID),
		Limit:   filters.PerPage,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	groups := make([]*domain.PaymentGroup, 0, len(rows))
	for _, row := range rows {
		g, err := r.GetPaymentGroup(ctx, row.SlNo)
		if err == nil {
			groups = append(groups, g)
		}
	}

	return groups, count, nil
}

func (r *paymentRepository) UpdatePayment(ctx context.Context, p *domain.Payment) error {
	_, err := r.queries.UpdatePayment(ctx, sqlc.UpdatePaymentParams{
		ID:                p.ID,
		TemplateType:      p.TemplateType,
		Project:           toPgTextFromPtr(p.Project),
		AccountFullName:   toPgTextFromPtr(p.AccountFullName),
		FromAccountType:   toPgTextFromPtr(p.FromAccountType),
		FullAccountNumber: toPgTextFromPtr(p.FullAccountNumber),
		ToAccount:         toPgTextFromPtr(p.ToAccount),
		ToAccountType:     toPgTextFromPtr(p.ToAccountType),
		NameOfBeneficiary: toPgTextFromPtr(p.NameOfBeneficiary),
		AccountNumber:     toPgTextFromPtr(p.AccountNumber),
		NameOfBank:        toPgTextFromPtr(p.NameOfBank),
		IfscCode:          toPgTextFromPtr(p.IfscCode),
		Amount:            toPgNumeric(p.Amount),
		Purpose:           toPgTextFromPtr(p.Purpose),
		Status:            p.Status,
	})
	return err
}

func (r *paymentRepository) UpdatePaymentGroupStatus(ctx context.Context, slNo string, status string) error {
	return r.queries.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
		SlNo:   slNo,
		Status: status,
	})
}

func (r *paymentRepository) DeletePayment(ctx context.Context, id int64) error {
	return r.queries.DeletePayment(ctx, id)
}

func (r *paymentRepository) DeletePaymentGroup(ctx context.Context, slNo string) error {
	return r.queries.DeletePaymentsBySlNo(ctx, slNo)
}

func (r *paymentRepository) LinkVendorAccount(ctx context.Context, paymentID int64, vendorID string, vendorAccountID *int64) error {
	_, err := r.queries.InsertPaymentVendorAccount(ctx, sqlc.InsertPaymentVendorAccountParams{
		PaymentID:       paymentID,
		VendorID:        stringToInt64(vendorID),
		VendorAccountID: toPgInt8(vendorAccountID),
	})
	return err
}

func (r *paymentRepository) CreateShortcut(ctx context.Context, s *domain.PaymentShortcut) (*domain.PaymentShortcut, error) {
	row, err := r.queries.CreatePaymentShortcut(ctx, sqlc.CreatePaymentShortcutParams{
		SlNo:            toPgTextFromPtr(s.SlNo),
		ShortcutName:    s.ShortcutName,
		RequestDataJson: s.RequestDataJSON,
		UserID:          stringToInt64(s.UserID),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PaymentShortcut{
		ID:              row.ID,
		SlNo:            pgTextToPtr(row.SlNo),
		ShortcutName:    row.ShortcutName,
		RequestDataJSON: row.RequestDataJson,
		UserID:          row.UserID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *paymentRepository) GetShortcut(ctx context.Context, id int64) (*domain.PaymentShortcut, error) {
	row, err := r.queries.GetPaymentShortcut(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.PaymentShortcut{
		ID:              row.ID,
		SlNo:            pgTextToPtr(row.SlNo),
		ShortcutName:    row.ShortcutName,
		RequestDataJSON: row.RequestDataJson,
		UserID:          row.UserID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *paymentRepository) ListShortcuts(ctx context.Context, userID string) ([]*domain.PaymentShortcut, error) {
	rows, err := r.queries.ListPaymentShortcuts(ctx, stringToInt64(userID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.PaymentShortcut, len(rows))
	for i, row := range rows {
		res[i] = &domain.PaymentShortcut{
			ID:              row.ID,
			SlNo:            pgTextToPtr(row.SlNo),
			ShortcutName:    row.ShortcutName,
			RequestDataJSON: row.RequestDataJson,
			UserID:          row.UserID,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		}
	}
	return res, nil
}

func (r *paymentRepository) GenerateSerialNumber(ctx context.Context, prefix string) (string, error) {
	next, err := r.queries.GenerateSerialNumber(ctx, toPgText(prefix))
	if err != nil {
		return "", err
	}
	year, month, _ := time.Now().Date()
	return fmt.Sprintf("%s-%d-%02d-%05d", prefix, year, month, next), nil
}

func (r *paymentRepository) AddBankLetterLog(ctx context.Context, p *domain.BankLetterApprovalLog) error {
	_, err := r.queries.InsertBankLetterLog(ctx, sqlc.InsertBankLetterLogParams{
		SlNo:          p.SlNo,
		Status:        p.Status,
		Comments:      toPgTextFromPtr(p.Comments),
		ReviewerID:    stringToInt64(p.ReviewerID),
		ReviewerName:  toPgTextFromPtr(p.ReviewerName),
		ReviewerEmail: toPgTextFromPtr(p.ReviewerEmail),
		ApproverLevel: toPgInt4(p.ApproverLevel),
	})
	return err
}

func (r *paymentRepository) GetBankLetterLogs(ctx context.Context, slNo string) ([]*domain.BankLetterApprovalLog, error) {
	rows, err := r.queries.ListBankLetterLogs(ctx, slNo)
	if err != nil {
		return nil, err
	}
	res := make([]*domain.BankLetterApprovalLog, len(rows))
	for i, row := range rows {
		res[i] = &domain.BankLetterApprovalLog{
			ID:            row.ID,
			SlNo:          row.SlNo,
			Status:        row.Status,
			Comments:      pgTextToPtr(row.Comments),
			ReviewerID:    row.ReviewerID,
			ReviewerName:  pgTextToPtr(row.ReviewerName),
			ReviewerEmail: pgTextToPtr(row.ReviewerEmail),
			ApproverLevel: pgInt4ToPtr(row.ApproverLevel),
			CreatedAt:     row.CreatedAt.Time,
		}
	}
	return res, nil
}

func (r *paymentRepository) toDomain(p sqlc.Payment) domain.Payment {
	amount, _ := parsePrice(p.Amount)
	
	// Handle TemplateType
	templateType := ""
	if p.TemplateType != nil {
		templateType = fmt.Sprintf("%v", p.TemplateType)
	}
	
	// Handle Status
	status := ""
	if p.Status != nil {
		status = fmt.Sprintf("%v", p.Status)
	}
	
	return domain.Payment{
		ID:                p.ID,
		SlNo:              p.SlNo,
		TemplateType:      templateType,
		Project:           pgTextToPtr(p.Project),
		AccountFullName:   pgTextToPtr(p.AccountFullName),
		FromAccountType:   pgTextToPtr(p.FromAccountType),
		FullAccountNumber: pgTextToPtr(p.FullAccountNumber),
		ToAccount:         pgTextToPtr(p.ToAccount),
		ToAccountType:     pgTextToPtr(p.ToAccountType),
		NameOfBeneficiary: pgTextToPtr(p.NameOfBeneficiary),
		AccountNumber:     pgTextToPtr(p.AccountNumber),
		NameOfBank:        pgTextToPtr(p.NameOfBank),
		IfscCode:          pgTextToPtr(p.IfscCode),
		Amount:            amount,
		Purpose:           pgTextToPtr(p.Purpose),
		Status:            status,
		UserID:            p.UserID,
		PaymentNoteID:     pgTextToInt64Ptr(p.PaymentNoteID),
		CreatedAt:         p.CreatedAt.Time,
		UpdatedAt:         p.UpdatedAt.Time,
	}
}


func toPgPaymentStatus(s string) interface{} {
	return s
}
