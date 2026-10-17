import sys
sys.stdout.reconfigure(encoding='utf-8')

with open('gateway_swagger.py', 'r', encoding='utf-8') as f:
    content = f.read()

old_create = """async def create_payment_note(body: CreatePaymentNoteRequest):
    return {"note": {}}"""

new_create = """async def create_payment_note(body: CreatePaymentNoteRequest):
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
    return {"note": item, "message": "Payment note created successfully"}"""

old_list = """async def list_payment_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None, description="Search by note number, vendor, or project"),
):
    return {"notes": [], "pagination": {}}"""

new_list = """async def list_payment_notes(
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

if old_create in content and old_list in content:
    content = content.replace(old_create, new_create, 1)
    content = content.replace(old_list, new_list, 1)
    with open('gateway_swagger.py', 'w', encoding='utf-8') as f:
        f.write(content)
    print("Successfully patched create_payment_note and list_payment_notes!")
else:
    print("ERROR: strings not found.")
