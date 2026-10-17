with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
matches = re.findall(r'class .*Department.*\(BaseModel\):', content)
print("Department models:", matches)

matches_desig = re.findall(r'class .*Designation.*\(BaseModel\):', content)
print("Designation models:", matches_desig)
