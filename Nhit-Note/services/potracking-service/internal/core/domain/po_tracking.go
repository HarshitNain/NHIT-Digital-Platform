package domain

import (
	"time"
)

type POTracking struct {
	ID                    int64     `json:"id"`
	OrderNo               string    `json:"order_no"`
	VendorID              string    `json:"vendor_id"`
	SupplierID            *string   `json:"supplier_id"`
	OrderDate             time.Time `json:"order_date"`
	OrderValue            float64   `json:"order_value"`
	TotalUtilized         float64   `json:"total_utilized"`
	RemainingBudget       float64   `json:"remaining_budget"`
	UtilizationPercentage float64   `json:"utilization_percentage"`
	GreenNotesCount       int32     `json:"green_notes_count"`
	Description           string    `json:"description"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreatePOTrackingParams struct {
	OrderNo               string    `json:"order_no"`
	VendorID              string    `json:"vendor_id"`
	SupplierID            *string   `json:"supplier_id"`
	OrderDate             time.Time `json:"order_date"`
	OrderValue            float64   `json:"order_value"`
	TotalUtilized         float64   `json:"total_utilized"`
	RemainingBudget       float64   `json:"remaining_budget"`
	UtilizationPercentage float64   `json:"utilization_percentage"`
	GreenNotesCount       int32     `json:"green_notes_count"`
	Description           string    `json:"description"`
	Status                string    `json:"status"`
}
