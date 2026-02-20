import os
import shutil
import re

base_dir = "D:/Nagare_Project/nagare/backend"
repo_dir = os.path.join(base_dir, "internal/adapter/repository")
ext_dir = os.path.join(base_dir, "internal/adapter/external")

os.makedirs(ext_dir, exist_ok=True)

to_move = ["monitors", "media", "llm"]

for module in to_move:
    src = os.path.join(repo_dir, module)
    dst = os.path.join(ext_dir, module)
    if os.path.exists(src):
        shutil.move(src, dst)

all_go_files = []
for root, _, files in os.walk(base_dir):
    for f in files:
        if f.endswith('.go'):
            all_go_files.append(os.path.join(root, f))

for filepath in all_go_files:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    orig = content
    
    for module in to_move:
        old_import = f'"nagare/internal/adapter/repository/{module}"'
        new_import = f'"nagare/internal/adapter/external/{module}"'
        content = content.replace(old_import, new_import)
        
        old_import2 = f'"nagare/internal/adapter/repository/{module}/'
        new_import2 = f'"nagare/internal/adapter/external/{module}/'
        content = content.replace(old_import2, new_import2)
        
    if content != orig:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)

print("External integrations moved.")
