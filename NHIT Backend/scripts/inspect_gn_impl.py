with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
m = re.search(r'async def create_green_note.*?(?=@app\.)', content, re.DOTALL)
if m:
    print("--- create_green_note ---")
    print(m.group(0))

m2 = re.search(r'async def list_green_notes.*?(?=@app\.)', content, re.DOTALL)
if m2:
    print("--- list_green_notes ---")
    print(m2.group(0))
