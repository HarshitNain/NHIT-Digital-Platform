import sys
sys.stdout.reconfigure(encoding='utf-8')

with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
routes_to_check = [
    r'@app\.get\("/api/v1/green-notes"',
    r'@app\.get\("/api/v1/payment-notes"',
    r'@app\.get\("/api/v1/reimbursement-notes"'
]

for r in routes_to_check:
    m = re.search(rf'({r}.*?async def \w+\([^)]*\):(?:\n    [^\n]+)+)', content, re.DOTALL)
    if m:
        print("MATCH FOR", r)
        print(m.group(1))
