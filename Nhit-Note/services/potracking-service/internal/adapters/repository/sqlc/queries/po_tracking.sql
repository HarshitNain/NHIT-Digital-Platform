-- name: CreatePOTracking :exec
INSERT INTO po_tracking (
    order_no, vendor_id, supplier_id, order_date, order_value, 
    total_utilized, remaining_budget, utilization_percentage, 
    green_notes_count, description, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
);

-- name: GetPOTrackingByOrderNo :one
SELECT * FROM po_tracking WHERE order_no = $1;

-- name: GetPOTrackingByID :one
SELECT * FROM po_tracking WHERE id = $1;

-- name: UpdatePOUtilization :exec
UPDATE po_tracking
SET total_utilized = $2,
    remaining_budget = $3,
    utilization_percentage = $4,
    green_notes_count = $5
WHERE id = $1;

-- name: ListAllPOTracking :many
SELECT * FROM po_tracking ORDER BY created_at DESC;
