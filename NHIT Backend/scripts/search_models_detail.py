with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
m = re.search(r'class CreateDepartmentRequest.*?(?=class |\n#)', content, re.DOTALL)
if m:
    print(m.group(0))

m2 = re.search(r'class UpdateDepartmentRequest.*?(?=class |\n#)', content, re.DOTALL)
if m2:
    print(m2.group(0))
