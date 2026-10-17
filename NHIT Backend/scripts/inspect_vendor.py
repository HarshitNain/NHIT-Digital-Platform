import sys

with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

lines = content.splitlines(keepends=True)
for i in range(1615, min(1700, len(lines))):
    line_repr = lines[i].rstrip().encode('ascii', 'backslashreplace').decode('ascii')
    print(f"{i+1}: {line_repr}")
