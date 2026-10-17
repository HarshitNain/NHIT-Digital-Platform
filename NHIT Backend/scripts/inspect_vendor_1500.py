with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    lines = f.readlines()

for i in range(1500, min(1551, len(lines))):
    line_repr = lines[i].rstrip().encode('ascii', 'backslashreplace').decode('ascii')
    print(f"{i+1}: {line_repr}")
