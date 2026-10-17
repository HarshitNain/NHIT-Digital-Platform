with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    lines = f.readlines()

for idx, line in enumerate(lines):
    if "DEPARTMENT SERVICE" in line:
        for j in range(max(0, idx - 3), min(len(lines), idx + 8)):
            ascii_repr = lines[j].rstrip().encode('ascii', 'backslashreplace').decode('ascii')
            print(f"{j+1}: {ascii_repr}")
        break
