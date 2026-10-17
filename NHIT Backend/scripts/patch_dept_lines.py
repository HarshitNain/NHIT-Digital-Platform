with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    lines = f.readlines()

dept_idx = None
for idx, line in enumerate(lines):
    if "DEPARTMENT SERVICE" in line:
        dept_idx = idx
        break

print(f"Found DEPARTMENT SERVICE at line {dept_idx + 1}")

# We want to replace from dept_idx - 1 (the border line) through the line that contains '@app.post("/api/v1/designations"'
# Let's see what line contains '@app.post("/api/v1/designations"'
desig_idx = None
for idx in range(dept_idx, dept_idx + 5):
    if '/api/v1/designations' in lines[idx]:
        desig_idx = idx
        break

print(f"Found designations endpoint at line {desig_idx + 1}")

dept_code = """# ══════════════════════════════════════════════════════════════
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

@app.post("/api/v1/designations", tags=["📋 Designation Service"], summary="Create Designation",
"""

# Extract the rest of the desig line after summary="Create Designation",
old_desig_line = lines[desig_idx]
post_idx = old_desig_line.find('@app.post("/api/v1/designations"')
# We replace lines[dept_idx - 1 : desig_idx + 1] with dept_code

new_lines = lines[:dept_idx - 1] + [dept_code] + lines[desig_idx + 1:]

with open('gateway_swagger.py', 'w', encoding='utf-8') as f:
    f.writelines(new_lines)

print("Saved cleanly!")
