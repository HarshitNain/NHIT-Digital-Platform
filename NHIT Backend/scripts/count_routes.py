import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
sys.stdout.reconfigure(encoding='utf-8')

from gateway_swagger import app

routes = [r for r in app.routes if hasattr(r, 'methods')]
print(f"Total API routes registered: {len(routes)}")

tags = set()
for r in routes:
    for t in getattr(r, 'tags', []):
        tags.add(t)

print(f"Total Services/Tags: {len(tags)}")
for t in sorted(tags):
    tag_routes = [r for r in routes if t in getattr(r, 'tags', [])]
    print(f"  - {t}: {len(tag_routes)} endpoints")
