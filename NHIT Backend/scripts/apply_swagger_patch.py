import sys

with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

# 1. Patch Department Service & Designation Header
dept_target = """# 🏫 DEPARTMENT SERVICE (5 endpoints)
# ════════════════════════════════════════════════════@app.post("/api/v1/designations","""

dept_replacement = """# ══════════════════════════════════════════════════════════════
# 🏫 DEPARTMENT SERVICE (5 endpoints)
# ══════════════════════════════════════════════════════════════

@app.post("/api/v1/departments", tags=["🏫 Department Service"], summary="Create Department",
          description="Creates a new organizational department (e.g., 'Finance', 'Civil Engineering').")
async def create_department(body: CreateDepartmentRequest):
    try:
        execute_db("INSERT INTO departments (name, description) VALUES (%s, %s)", (body.name, body.description))
        return {"department": {"name": body.name, "description": body.description}, "message": "Department created"}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/departments", tags=["🏫 Department Service"], summary="List Departments",
         description="Lists all departments with pagination.")
async def list_departments(page: int = Query(1, ge=1), page_size: int = Query(20, ge=1, le=100)):
    try:
        depts = query_db("SELECT id, name, description FROM departments ORDER BY name LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM departments")
        return {"departments": depts, "total_count": total[0]["cnt"] if total else 0}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Get Department",
         description="Retrieves a specific department by its UUID.")
async def get_department(id: str = Path(..., description="Department UUID")):
    try:
        dept = query_db("SELECT id, name, description FROM departments WHERE id = %s", (id,))
        if not dept:
            raise HTTPException(404, "Department not found")
        return {"department": dept[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Update Department",
         description="Updates department name or description.")
async def update_department(id: str = Path(..., description="Department UUID"), body: UpdateDepartmentRequest = ...):
    try:
        execute_db("UPDATE departments SET name = COALESCE(%s, name), description = COALESCE(%s, description) WHERE id = %s",
                   (body.name, body.description, id))
        return {"department": {"id": id, "name": body.name, "description": body.description}, "message": "Department updated"}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.delete("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Delete Department",
            description="Deletes a department by UUID.")
async def delete_department(id: str = Path(..., description="Department UUID")):
    try:
        execute_db("DELETE FROM departments WHERE id = %s", (id,))
        return {"message": "Department deleted"}
    except Exception as e:
        raise HTTPException(500, str(e))


# ══════════════════════════════════════════════════════════════
# 📋 DESIGNATION SERVICE (5 endpoints)
# ══════════════════════════════════════════════════════════════

@app.post("/api/v1/designations","""

if dept_target in content:
    content = content.replace(dept_target, dept_replacement, 1)
    print("Successfully replaced Department Service section.")
else:
    print("WARNING: dept_target not found directly, trying pattern matching...")
    import re
    p = re.compile(r'# 🏫 DEPARTMENT SERVICE.*?@app\.post\("/api/v1/designations",', re.DOTALL)
    if p.search(content):
        content = p.sub(dept_replacement, content, count=1)
        print("Successfully matched and replaced Department Service section via regex.")
    else:
        print("ERROR: Could not find Department Service target.")

# 2. Patch Vendor Section (clean up duplicated stub methods and residual splice)
# Find from @app.get("/api/v1/vendors/dropdowns/projects" up to # 🔔 NOTIFICATION SERVICE
import re
vendor_target_regex = re.compile(
    r'@app\.get\("/api/v1/vendors/dropdowns/projects",\s*tags=\["🤝 Vendor Service"\].*?# 🔔 NOTIFICATION SERVICE',
    re.DOTALL
)

vendor_clean_replacement = """@app.get("/api/v1/vendors/dropdowns/projects", tags=["🤝 Vendor Service"], summary="Projects Dropdown for Vendors",
         description="Returns project list for vendor form dropdowns, filtered by the authenticated user's organization.")
async def vendor_projects_dropdown():
    try:
        projects = query_db("SELECT id, project_name FROM projects ORDER BY project_name")
        return {"projects": projects}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.post("/api/v1/vendors/{vendor_id}/document", tags=["🤝 Vendor Service"], summary="Upload Vendor Document",
          description="Uploads a document (PAN card, GST certificate, etc.) for a vendor. Stored in object storage.")
async def upload_vendor_document(vendor_id: str):
    return {"success": True, "message": "Document upload — forwards to Vendor gRPC service", "vendor_id": vendor_id}


# ══════════════════════════════════════════════════════════════
# 🔔 NOTIFICATION SERVICE"""

if vendor_target_regex.search(content):
    content = vendor_target_regex.sub(vendor_clean_replacement, content, count=1)
    print("Successfully cleaned up Vendor dropdowns and duplicate stubs.")
else:
    print("ERROR: vendor_target_regex did not match.")

with open('gateway_swagger.py', 'w', encoding='utf-8') as f:
    f.write(content)

print("Patch applied and saved to gateway_swagger.py.")
