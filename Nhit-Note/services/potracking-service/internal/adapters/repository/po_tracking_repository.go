package repository

import (
	"context"

	sqlc "github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"

	"strconv"
)

type poTrackingRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
}

func NewPOTrackingRepository(db sqlc.DBTX) ports.POTrackingRepository {
	return &poTrackingRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *poTrackingRepository) Create(ctx context.Context, params domain.CreatePOTrackingParams) error {
	vendorID, _ := strconv.ParseInt(params.VendorID, 10, 64)
	
	return r.queries.CreatePOTracking(ctx, sqlc.CreatePOTrackingParams{
		OrderNo:               params.OrderNo,
		VendorID:              vendorID,
		SupplierID:            toPgInt8FromPtr(params.SupplierID),
		OrderDate:             toPgDate(params.OrderDate),
		OrderValue:            toPgNumeric(params.OrderValue),
		TotalUtilized:         toPgNumeric(params.TotalUtilized),
		RemainingBudget:       toPgNumeric(params.RemainingBudget),
		UtilizationPercentage: toPgNumeric(params.UtilizationPercentage),
		GreenNotesCount:       params.GreenNotesCount,
		Description:           toPgText(params.Description),
		Status:                params.Status,
	})
}

func (r *poTrackingRepository) GetByOrderNo(ctx context.Context, orderNo string) (*domain.POTracking, error) {
	dbPO, err := r.queries.GetPOTrackingByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbPO), nil
}

func (r *poTrackingRepository) GetByID(ctx context.Context, id int64) (*domain.POTracking, error) {
	dbPO, err := r.queries.GetPOTrackingByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbPO), nil
}

func (r *poTrackingRepository) UpdateUtilization(ctx context.Context, id int64, utilized float64, remaining float64, percentage float64, count int32) error {
	return r.queries.UpdatePOUtilization(ctx, sqlc.UpdatePOUtilizationParams{
		ID:                    id,
		TotalUtilized:         toPgNumeric(utilized),
		RemainingBudget:       toPgNumeric(remaining),
		UtilizationPercentage: toPgNumeric(percentage),
		GreenNotesCount:       count,
	})
}

func (r *poTrackingRepository) ListAll(ctx context.Context) ([]*domain.POTracking, error) {
	dbPOs, err := r.queries.ListAllPOTracking(ctx)
	if err != nil {
		return nil, err
	}
	var pos []*domain.POTracking
	for _, dbPO := range dbPOs {
		pos = append(pos, r.toDomain(&dbPO))
	}
	return pos, nil
}

func (r *poTrackingRepository) toDomain(dbPO *sqlc.PoTracking) *domain.POTracking {
	orderValue, _ := parsePrice(dbPO.OrderValue)
	totalUtilized, _ := parsePrice(dbPO.TotalUtilized)
	remainingBudget, _ := parsePrice(dbPO.RemainingBudget)
	utilizationPercentage, _ := parsePrice(dbPO.UtilizationPercentage)

	return &domain.POTracking{
		ID:                    dbPO.ID,
		OrderNo:               dbPO.OrderNo,
		VendorID:              dbPO.VendorID,
		SupplierID:            pgTextToPtr(dbPO.SupplierID),
		OrderDate:             dbPO.OrderDate.Time,
		OrderValue:            orderValue,
		TotalUtilized:         totalUtilized,
		RemainingBudget:       remainingBudget,
		UtilizationPercentage: utilizationPercentage,
		GreenNotesCount:       dbPO.GreenNotesCount,
		Description:           dbPO.Description.String,
		Status:                dbPO.Status,
		CreatedAt:             dbPO.CreatedAt.Time,
		UpdatedAt:             dbPO.UpdatedAt.Time,
	}
}

