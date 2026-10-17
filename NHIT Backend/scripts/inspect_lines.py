import sys

with open('gateway_swagger.py', 'r', encoding='utf-8', errors='replace') as f:
    content = f.read()

# Let's inspect line around 1335-1365 and 1590-1630
lines = content.splitlines(keepends=True)
print(f"Total lines: {len(lines)}")

# Let's check where the splice is around line 1330-1360
for i in range(1330, min(1365, len(lines))):
    line_repr = lines[i].rstrip().encode('ascii', 'backslashreplace').decode('ascii')
    print(f"{i+1}: {line_repr}")

print("--- Check around 1590-1630 ---")
for i in range(1590, min(1630, len(lines))):
    line_repr = lines[i].rstrip().encode('ascii', 'backslashreplace').decode('ascii')
    print(f"{i+1}: {line_repr}")
