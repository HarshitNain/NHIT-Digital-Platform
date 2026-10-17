with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
m = re.search(r'class GreenNotePayload.*?(?=class |\n#)', content, re.DOTALL)
if m:
    print(m.group(0))
