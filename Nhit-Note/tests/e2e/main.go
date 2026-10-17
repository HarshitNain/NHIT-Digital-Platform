package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
)

const (
	greenNoteAddr   = "localhost:50059"
	paymentNoteAddr = "localhost:50056"
	authAddr        = "localhost:50052"
	kafkaAddr       = "localhost:9092"
	orgID           = "e6c3aaca-6311-4f84-8602-7294d1276852"
	tenantID        = "b7f3f213-d7f1-4d72-b14b-98d795bd72ae"
	userEmail       = "rnr@gmail.com"
	userPassword    = "rnr@123"
)

func main() {
	ctx := context.Background()

	// 1. Connect to Auth Service and Login
	authConn, err := grpc.Dial(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Auth service: %v", err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	log.Println("--- AUTH: Logging in ---")
	loginResp, err := authClient.Login(ctx, &authpb.UserLoginRequest{
		Login:    userEmail,
		Password: userPassword,
		TenantId: tenantID,
		OrgId:    orgID,
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	token := loginResp.Token
	log.Println("✅ Logged in successfully")

	// 2. Setup metadata for subsequent calls
	md := metadata.Pairs(
		"authorization", "Bearer "+token,
		"org_id", orgID,
		"tenant_id", tenantID,
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 3. Connect to services
	gnConn, err := grpc.Dial(greenNoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to GreenNote service: %v", err)
	}
	defer gnConn.Close()
	gnClient := greennotepb.NewGreenNoteServiceClient(gnConn)

	pnConn, err := grpc.Dial(paymentNoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Payment Note service: %v", err)
	}
	defer pnConn.Close()
	pnClient := paymentnotepb.NewPaymentNoteServiceClient(pnConn)

	log.Println("\n--- STEP 1: Create GreenNote ---")
	createReq := &greennotepb.CreateGreenNoteRequest{
		Note: &greennotepb.GreenNotePayload{
			ProjectName:     "Project Alpha",
			SupplierName:    "Bright Solutions",
			ExpenseCategory: "General",
			TotalAmount:     5000.0,
			DetailedStatus:  "draft",
			PoNumber:        "PO-123456",
			Invoice: &greennotepb.InvoiceInput{
				InvoiceNumber: "INV-001",
				InvoiceDate:   "2026-01-01",
				TaxableValue:  4500.0,
				Gst:           500.0,
				InvoiceValue:  5000.0,
			},
		},
	}
	createResp, err := gnClient.CreateGreenNote(ctx, createReq)
	if err != nil {
		log.Fatalf("Failed to create GreenNote: %v", err)
	}
	gnID := createResp.Id
	log.Printf("✅ GreenNote created with ID: %s", gnID)

	log.Println("\n--- STEP 2: View GreenNote (Verify transition to Pending) ---")
	detailResp, err := gnClient.GetGreenNote(ctx, &greennotepb.GetGreenNoteRequest{Id: gnID})
	if err != nil {
		log.Fatalf("Failed to get GreenNote: %v", err)
	}
	log.Printf("Status after view: %s (Proto Enum: %s)", detailResp.Data.DetailedStatus, detailResp.Data.Status)
	
	log.Println("\n--- STEP 3: Approve GreenNote ---")
	publishApprovalEvent(gnID, "GREENNOTE", "APPROVED", "Final Approval Granted.")
	time.Sleep(10 * time.Second) // Wait for consumer
	
	detailResp, _ = gnClient.GetGreenNote(ctx, &greennotepb.GetGreenNoteRequest{Id: gnID})
	log.Printf("Final GreenNote Status: %s", detailResp.Data.DetailedStatus)

	log.Println("\n--- STEP 4: Verify Automatic Payment Note Draft Creation ---")
	time.Sleep(15 * time.Second)
	
	// Poll for draft creation (max 10 retries, 3s interval) because Kafka is async
	var pnID int64
	found := false
	maxRetries := 10
	
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempt %d/%d to find auto-created draft...", i+1, maxRetries)
		
		listResp, err := pnClient.ListPaymentNotes(ctx, &paymentnotepb.ListPaymentNotesRequest{
			DraftsOnly: true,
			PerPage:    20, // Fetch more to avoid it being pushed off page 1
		})
		if err != nil {
			log.Printf("Warning: Failed to list Payment Notes: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, note := range listResp.Notes {
			gnRefID := "N/A"
			if note.GreenNote != nil {
				gnRefID = note.GreenNote.Id
			}
			log.Printf("Found Draft Note: ID=%d, GreenNoteID=%s", note.Id, gnRefID)

			if note.GreenNote != nil && note.GreenNote.Id == gnID {
				pnID = note.Id
				found = true
				break
			}
		}
		
		if found {
			break
		}
		time.Sleep(3 * time.Second)
	}

	if !found {
		log.Printf("❌ Could not find automatically created Payment Note for GN-%s after retries", gnID)
	} else {
		log.Printf("✅ Found Payment Note Draft with ID: %d", pnID)
		
		log.Println("\n--- STEP 5: View Payment Note (Verify transition to Pending & PO Enrichment) ---")
		pnResp, err := pnClient.GetPaymentNote(ctx, &paymentnotepb.GetPaymentNoteRequest{Id: pnID})
		if err != nil {
			log.Fatalf("Failed to get Payment Note: %v", err)
		}
		log.Printf("Status after view: %s", pnResp.Note.Status)
		log.Printf("Enriched PO Number: %s", pnResp.Note.LoaPoNo)
		
		log.Println("\n--- STEP 6: Approve Payment Note ---")
		publishApprovalEvent(fmt.Sprintf("%d", pnID), "PAYMENTNOTE", "APPROVED", "PN Approved.")
		time.Sleep(3 * time.Second)
		
		pnResp, _ = pnClient.GetPaymentNote(ctx, &paymentnotepb.GetPaymentNoteRequest{Id: pnID})
		log.Printf("Final Payment Note Status: %s", pnResp.Note.Status)
	}

	log.Println("\n--- STEP 7: Test Direct Payment Note Creation ---")
	directReq := &paymentnotepb.CreatePaymentNoteRequest{
		Note: &paymentnotepb.PaymentNotePayload{
			GreenNoteId: gnID,
			Subject:     "Direct Note Test",
			GrossAmount: 1000.0,
			Status:      "P",
		},
	}
	directResp, err := pnClient.CreatePaymentNote(ctx, directReq)
	if err != nil {
		log.Printf("❌ Direct creation failed: %v", err)
	} else {
		log.Printf("✅ Direct Payment Note created with ID: %d, Status: %s", directResp.Note.Id, directResp.Note.Status)
	}
}

func publishApprovalEvent(sourceID, sourceType, status, remarks string) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaAddr},
		Topic:   "approval_events",
	})
	defer w.Close()

	data := map[string]interface{}{
		"source_id":   sourceID,
		"source_type": sourceType,
		"status":      status,
		"remarks":     remarks,
		"actor_id":    "approver-1",
	}
	dataJSON, _ := json.Marshal(data)

	event := map[string]interface{}{
		"event_type":   "ApprovalStatusChanged",
		"aggregate_id": sourceID,
		"data":         string(dataJSON),
	}
	eventJSON, _ := json.Marshal(event)

	err := w.WriteMessages(context.Background(), kafka.Message{
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Failed to publish Kafka event: %v", err)
	} else {
		log.Printf("📤 Published Kafka Event: %s -> %s", sourceType, status)
	}
}
