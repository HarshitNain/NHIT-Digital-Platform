with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
methods = [
    'create_green_note', 'list_green_notes', 'get_green_note',
    'create_payment_note', 'list_payment_notes', 'get_payment_note',
    'create_reimbursement_note', 'list_reimbursement_notes', 'get_reimbursement_note'
]

for m in methods:
    found = re.search(rf'async def {m}\(.*?\) ->.*?|async def {m}\(.*?:\n(?:    [^\n]+\n)+', content)
    if found:
        print(f"=== {m} ===")
        print(found.group(0)[:300])
    else:
        print(f"=== {m}: NOT FOUND ===")
