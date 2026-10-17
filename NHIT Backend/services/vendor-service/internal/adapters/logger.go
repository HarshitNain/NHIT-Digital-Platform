package adapters

import (
	"context"
	"fmt"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/google/uuid"
)

// noOpEventPublisher implements EventPublisher interface with no-ops
type noOpEventPublisher struct{}

// NewNoOpEventPublisher creates a new no-op event publisher
func NewNoOpEventPublisher() ports.EventPublisher {
	return &noOpEventPublisher{}
}

func (p *noOpEventPublisher) PublishVendorCreated(ctx context.Context, vendor *domain.Vendor) error {
	fmt.Println("📢 Event: Vendor created (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishVendorUpdated(ctx context.Context, vendor *domain.Vendor) error {
	fmt.Println("📢 Event: Vendor updated (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishVendorDeleted(ctx context.Context, tenantID, vendorID uuid.UUID) error {
	fmt.Println("📢 Event: Vendor deleted (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountCreated(ctx context.Context, account *domain.VendorAccount) error {
	fmt.Println("📢 Event: Account created (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountUpdated(ctx context.Context, account *domain.VendorAccount) error {
	fmt.Println("📢 Event: Account updated (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountDeleted(ctx context.Context, accountID uuid.UUID) error {
	fmt.Println("📢 Event: Account deleted (no-op)")
	return nil
}
