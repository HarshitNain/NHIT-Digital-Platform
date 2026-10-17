package memory

import (
	"context"
	"sync"
	"time"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// Repository is an in-memory implementation of GreenNoteRepository.
type Repository struct {
	mu sync.RWMutex

	// notes are keyed by their UUID string identifier.
	notes map[string]*noteRecord
}

type noteRecord struct {
	payload   *greennotepb.GreenNotePayload
	createdAt time.Time
	updatedAt time.Time
}

// NewInMemoryGreenNoteRepository constructs an in-memory repository.
// The DocumentStorage parameter is currently unused but kept for API compatibility.
func NewInMemoryGreenNoteRepository(_ ports.DocumentStorage) ports.GreenNoteRepository {
	return &Repository{
		notes: make(map[string]*noteRecord),
	}
}

func (r *Repository) List(ctx context.Context, req *greennotepb.ListGreenNotesRequest, orgID, tenantID string) (*greennotepb.ListGreenNotesResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_ = ctx

	type entry struct {
		id  string
		rec *noteRecord
	}

	var entries []entry
	for id, rec := range r.notes {
		// Use DetailedStatus for string-based filtering if it contains the status value,
		// otherwise GetStatus() is an enum.
		// For simplicity in memory repo, we check both.
		if !req.GetIncludeAll() {
			if rec.payload.GetStatus() != req.GetStatus() {
				continue
			}
		}
		entries = append(entries, entry{id: id, rec: rec})
	}

	page := req.GetPage()
	perPage := req.GetPerPage()
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	p := int(page)
	pp := int(perPage)
	start := (p - 1) * pp
	if start > len(entries) {
		start = len(entries)
	}
	end := start + pp
	if end > len(entries) {
		end = len(entries)
	}

	items := make([]*greennotepb.GreenNoteListItem, 0, end-start)
	for _, e := range entries[start:end] {
		payload := e.rec.payload
		items = append(items, &greennotepb.GreenNoteListItem{
			Id:          e.id,
			ProjectName: payload.GetProjectName(),
			VendorName:  payload.GetSupplierName(),
			Amount:      payload.GetTotalAmount(),
			Date:        e.rec.createdAt.Format(time.RFC3339),
			Status:      payload.GetStatus(),
		})
	}

	return &greennotepb.ListGreenNotesResponse{
		Notes:   items,
		Page:    page,
		PerPage: perPage,
		Total:   int64(len(entries)),
	}, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*greennotepb.GreenNotePayload, string, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_ = ctx

	rec, ok := r.notes[id]
	if !ok || rec.payload == nil {
		return nil, "", "", ports.ErrNotFound
	}

	cloned := proto.Clone(rec.payload)
	if cloned == nil {
		return nil, "", "", ports.ErrNotFound
	}
	return cloned.(*greennotepb.GreenNotePayload), "", "", nil
}

func (r *Repository) Create(ctx context.Context, payload *greennotepb.GreenNotePayload, orgID, tenantID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = ctx

	if payload == nil {
		return "", nil
	}

	id := uuid.NewString()
	cloned := proto.Clone(payload).(*greennotepb.GreenNotePayload)
	now := time.Now().UTC()

	r.notes[id] = &noteRecord{
		payload:   cloned,
		createdAt: now,
		updatedAt: now,
	}

	return id, nil
}

func (r *Repository) Update(ctx context.Context, id string, payload *greennotepb.GreenNotePayload, orgID, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = ctx

	rec, ok := r.notes[id]
	if !ok {
		return ports.ErrNotFound
	}
	if payload == nil {
		return nil
	}

	cloned := proto.Clone(payload).(*greennotepb.GreenNotePayload)
	rec.payload = cloned
	rec.updatedAt = time.Now().UTC()
	return nil
}

func (r *Repository) Cancel(ctx context.Context, id string, reason string, orgID, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = ctx
	_ = reason

	rec, ok := r.notes[id]
	if !ok || rec.payload == nil {
		return ports.ErrNotFound
	}

	rec.updatedAt = time.Now().UTC()
	return nil
}

func (r *Repository) AddSupportingDocuments(ctx context.Context, noteID string, docs []*greennotepb.SupportingDocument, orgID, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = ctx

	rec, ok := r.notes[noteID]
	if !ok {
		return ports.ErrNotFound
	}
	if rec.payload == nil {
		return ports.ErrNotFound
	}

	for _, d := range docs {
		newDoc := proto.Clone(d).(*greennotepb.SupportingDocument)
		if newDoc.Id == "" {
			newDoc.Id = uuid.NewString()
		}
		rec.payload.ExistingDocuments = append(rec.payload.ExistingDocuments, newDoc)
	}
	return nil
}

func (r *Repository) LinkPaymentNote(ctx context.Context, id string, paymentNoteID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = ctx

	rec, ok := r.notes[id]
	if !ok {
		return ports.ErrNotFound
	}
	rec.payload.PaymentNoteId = paymentNoteID
	rec.updatedAt = time.Now().UTC()
	return nil
}
