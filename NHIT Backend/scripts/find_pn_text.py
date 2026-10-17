import sys
sys.stdout.reconfigure(encoding='utf-8')

with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
m = re.search(r'(@app\.post\("/api/v1/payment-notes".*?async def create_payment_note\(.*?\):(?:\n    [^\n]+)+)', content, re.DOTALL)
if m:
    print("Found create_payment_note:")
    print(m.group(0))

m2 = re.search(r'(@app\.get\("/api/v1/payment-notes".*?async def list_payment_notes\(.*?\):(?:\n    [^\n]+)+)', content, re.DOTALL)
if m2:
    print("Found list_payment_notes:")
    print(m2.group(0))
