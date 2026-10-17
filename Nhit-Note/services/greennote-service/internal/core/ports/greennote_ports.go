package ports

import (
	"context"
	"errors"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
)

// ErrNotFound is a sentinel error used when an entity cannot be found.
var ErrNotFound = errors.New("not found")

// GreenNoteRepository defines the persistence contract for GreenNotes in terms of the
// minimal greennote.proto API. It is responsible for persisting the core payload and
// returning list/detail views used by the service layer.
type GreenNoteRepository interface {
	// List returns a lightweight paginated list of green notes based on the
	// filtering and pagination options in the request.
	List(ctx context.Context, req *greennotepb.ListGreenNotesRequest, orgID, tenantID string) (*greennotepb.ListGreenNotesResponse, error)

	// Get returns the full payload for a single green note by its identifier.
	// The identifier is modeled as a UUID string at the API boundary.
	Get(ctx context.Context, id string) (*greennotepb.GreenNotePayload, string, string, error)

	// Create persists a new green note and returns its generated identifier as a
	// UUID string.
	Create(ctx context.Context, payload *greennotepb.GreenNotePayload, orgID, tenantID string) (string, error)

	// Update applies changes from payload to an existing green note, addressed
	// by its UUID string identifier.
	Update(ctx context.Context, id string, payload *greennotepb.GreenNotePayload, orgID, tenantID string) error

	// Cancel transitions a green note into the cancelled status.
	Cancel(ctx context.Context, id string, reason string, orgID, tenantID string) error

	// AddSupportingDocuments persists metadata for multiple documents associated with a green note.
	AddSupportingDocuments(ctx context.Context, noteID string, docs []*greennotepb.SupportingDocument, orgID, tenantID string) error

	// LinkPaymentNote updates the payment_note_id for a green note.
	LinkPaymentNote(ctx context.Context, id string, paymentNoteID string) error
}

// GreenNoteApprovedEvent is emitted when a GreenNote is fully approved and a payment note draft is created.
type GreenNoteApprovedEvent struct {
	GreenNoteID string  `json:"green_note_id"`
	OrderNo     string  `json:"order_no"`
	PoNumber    string  `json:"po_number"`
	NetAmount   float64 `json:"net_amount"`
	Status      string  `json:"status"`
	Comments    string  `json:"comments"`
	ApprovedAt  string  `json:"approved_at"`
}

// EventPublisher publishes domain events (Kafka implementation lives in adapters).
type EventPublisher interface {
	PublishGreenNoteApproved(ctx context.Context, event GreenNoteApprovedEvent) error
}

// DocumentStorage abstracts where supporting documents are stored (e.g. MinIO or in-memory).
type DocumentStorage interface {
	Save(ctx context.Context, objectName string, content []byte, contentType string) error
	Load(ctx context.Context, objectName string) ([]byte, string, error)
}
