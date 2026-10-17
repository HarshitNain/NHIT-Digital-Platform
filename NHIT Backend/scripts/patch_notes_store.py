import sys
sys.stdout.reconfigure(encoding='utf-8')

with open('gateway_swagger.py', 'r', encoding='utf-8') as f:
    content = f.read()

import re

# 1. Update GreenNote handlers
old_gn = """async def create_green_note(body: CreateGreenNoteRequest):
    return {"id": "generated-uuid", "success": True, "message": "Green note created"}


@app.get("/api/v1/green-notes", tags=["📝 GreenNote Service"], summary="List Green Notes",
         description="Lists green notes with status filter, pagination, and user scoping. Returns lightweight list items (no full payload).")
async def list_green_notes(
    status: Optional[str] = Query(None, description="Filter: PENDING, APPROVED, REJECTED, DRAFT, CANCELLED"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    include_all: bool = Query(False, description="Include all statuses"),
):
    return {"notes": [], "page": page, "per_page": per_page, "total": 0}


@app.get("/api/v1/green-notes/{id}", tags=["📝 GreenNote Service"], summary="Get Green Note Detail",
         description="Retrieves the full green note payload including all compliance fields, invoice details, financial calculations, approval history, and linked documents.")
async def get_green_note(id: str = Path(..., description="Green Note UUID")):
    return {"success": True, "data": {}}"""

new_gn = """async def create_green_note(body: CreateGreenNoteRequest):
    new_id = f"gn-550e8400-{len(GREEN_NOTES_STORE) + 1:04d}"
    item = {
        "id": new_id,
        "sl_no": f"GN-2026-{len(GREEN_NOTES_STORE) + 1:03d}",
        "project_name": body.note.project_name or "NH-44 Highway Infrastructure Project",
        "supplier_name": body.note.supplier_name or "Larsen & Toubro Ltd",
        "amount": body.note.total_amount or 250000.0,
        "status": "PENDING",
        "created_by": "System User",
        "created_at": datetime.now().isoformat(),
        "payload": body.note.model_dump()
    }
    GREEN_NOTES_STORE.insert(0, item)
    return {"id": new_id, "success": True, "message": "Green note created successfully", "data": item}


@app.get("/api/v1/green-notes", tags=["📝 GreenNote Service"], summary="List Green Notes",
         description="Lists green notes with status filter, pagination, and user scoping. Returns lightweight list items (no full payload).")
async def list_green_notes(
    status: Optional[str] = Query(None, description="Filter: PENDING, APPROVED, REJECTED, DRAFT, CANCELLED"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    include_all: bool = Query(False, description="Include all statuses"),
):
    notes = GREEN_NOTES_STORE
    if status and not include_all:
        notes = [n for n in notes if n.get("status") == status]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "page": page, "per_page": per_page, "total": len(notes)}


@app.get("/api/v1/green-notes/{id}", tags=["📝 GreenNote Service"], summary="Get Green Note Detail",
         description="Retrieves the full green note payload including all compliance fields, invoice details, financial calculations, approval history, and linked documents.")
async def get_green_note(id: str = Path(..., description="Green Note UUID")):
    match = next((n for n in GREEN_NOTES_STORE if n.get("id") == id), None)
    if not match:
        raise HTTPException(404, "Green Note not found")
    return {"success": True, "data": match}"""

if old_gn in content:
    content = content.replace(old_gn, new_gn, 1)
    print("Patched Green Note handlers.")
else:
    print("ERROR: old_gn not found.")

# 2. Update PaymentNote handlers
old_pn = """@app.post("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="Create Payment Note",
          description="Creates a payment note manually or from an approved green note.")
async def create_payment_note(body: CreatePaymentNoteRequest):
    return {"note": {}}


@app.get("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="List Payment Notes",
         description="Lists payment notes with status filter, pagination, search, and role-based scoping. Super admins see all; others see only assigned notes.")
async def list_payment_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None, description="Search by note number, vendor, or project"),
):
    return {"notes": [], "pagination": {}}"""

new_pn = """@app.post("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="Create Payment Note",
          description="Creates a payment note manually or from an approved green note.")
async def create_payment_note(body: CreatePaymentNoteRequest):
    new_id = len(PAYMENT_NOTES_STORE) + 1
    item = {
        "id": new_id,
        "payment_note_number": f"PN-2026-{new_id:03d}",
        "vendor_name": body.vendor_name or "Larsen & Toubro Ltd",
        "vendor_account_number": body.vendor_account_number or "50200012345678",
        "bank_name": body.bank_name or "HDFC Bank",
        "ifsc_code": body.ifsc_code or "HDFC0000060",
        "net_payable_amount": body.net_payable_amount or 150000.0,
        "status": "D",
        "created_at": datetime.now().isoformat()
    }
    PAYMENT_NOTES_STORE.insert(0, item)
    return {"note": item, "message": "Payment note created successfully"}


@app.get("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="List Payment Notes",
         description="Lists payment notes with status filter, pagination, search, and role-based scoping. Super admins see all; others see only assigned notes.")
async def list_payment_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None, description="Search by note number, vendor, or project"),
):
    notes = PAYMENT_NOTES_STORE
    if status:
        notes = [n for n in notes if n.get("status") == status]
    if search:
        s = search.lower()
        notes = [n for n in notes if s in str(n.get("vendor_name", "")).lower() or s in str(n.get("payment_note_number", "")).lower()]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "pagination": {"page": page, "per_page": per_page, "total_items": len(notes)}}"""

if old_pn in content:
    content = content.replace(old_pn, new_pn, 1)
    print("Patched Payment Note handlers.")
else:
    print("ERROR: old_pn not found.")

# 3. Update Reimbursement handlers
old_rn = """@app.get("/api/v1/reimbursement-notes", tags=["🧾 Reimbursement Service"], summary="List Reimbursement Notes",
         description="Lists reimbursement notes with status filter, search, pagination, and role-based scoping.")
async def list_reimbursement_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted, B=Bank, PNA=Payment Note Approved"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None),
):
    return {"notes": [], "pagination": {}}"""

new_rn = """@app.get("/api/v1/reimbursement-notes", tags=["🧾 Reimbursement Service"], summary="List Reimbursement Notes",
         description="Lists reimbursement notes with status filter, search, pagination, and role-based scoping.")
async def list_reimbursement_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted, B=Bank, PNA=Payment Note Approved"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None),
):
    notes = REIMBURSEMENTS_STORE
    if status:
        notes = [n for n in notes if n.get("status") == status]
    if search:
        s = search.lower()
        notes = [n for n in notes if s in str(n.get("user_name", "")).lower() or s in str(n.get("note_no", "")).lower()]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "pagination": {"page": page, "per_page": per_page, "total_items": len(notes)}}


@app.get("/api/v1/reimbursements", tags=["🧾 Reimbursement Service"], summary="List Reimbursements (Alias)",
         description="Alias route for /api/v1/reimbursement-notes to support intuitive REST path discovery.")
async def list_reimbursements_alias(
    status: Optional[str] = Query(None), page: int = Query(1, ge=1), per_page: int = Query(20, ge=1, le=100), search: Optional[str] = Query(None)
):
    return await list_reimbursement_notes(status=status, page=page, per_page=per_page, search=search)"""

if old_rn in content:
    content = content.replace(old_rn, new_rn, 1)
    print("Patched Reimbursement handlers.")
else:
    print("ERROR: old_rn not found.")

with open('gateway_swagger.py', 'w', encoding='utf-8') as f:
    f.write(content)

print("Saved gateway_swagger.py successfully!")
