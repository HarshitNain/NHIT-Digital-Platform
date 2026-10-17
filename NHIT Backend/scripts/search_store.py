with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

import re
m = re.search(r'GREEN_NOTES_STORE.*?(?=\n\nclass |\n#)', content, re.DOTALL)
if m:
    print(m.group(0))
else:
    print("Not found")
