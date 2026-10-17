CREATE TABLE IF NOT EXISTS po_tracking (
    id BIGSERIAL PRIMARY KEY,
    order_no TEXT NOT NULL UNIQUE,
    vendor_id BIGINT NOT NULL,
    supplier_id BIGINT,
    order_date DATE NOT NULL,
    order_value DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_utilized DECIMAL(20, 2) NOT NULL DEFAULT 0,
    remaining_budget DECIMAL(20, 2) NOT NULL DEFAULT 0,
    utilization_percentage DECIMAL(5, 2) NOT NULL DEFAULT 0,
    green_notes_count INT NOT NULL DEFAULT 0,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_po_tracking_order_no ON po_tracking(order_no);
CREATE INDEX IF NOT EXISTS idx_po_tracking_vendor_id ON po_tracking(vendor_id);
CREATE INDEX IF NOT EXISTS idx_po_tracking_status ON po_tracking(status);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_po_tracking_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_po_tracking_modtime
BEFORE UPDATE ON po_tracking
FOR EACH ROW EXECUTE FUNCTION update_po_tracking_modified_column();
