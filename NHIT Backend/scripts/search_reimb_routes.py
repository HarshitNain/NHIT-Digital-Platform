with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
matches = re.findall(r'@app\.(?:get|post|put|delete)\("([^"]*reimburse[^"]*)"', content)
print("Reimbursement routes:", matches)
