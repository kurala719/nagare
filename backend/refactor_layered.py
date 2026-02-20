import os
import re
import shutil

base_dir = "D:/Nagare_Project/nagare/backend"
internal_dir = os.path.join(base_dir, "internal")

moves = {
    "model": "core/domain",
    "service": "core/service",
    "api": "adapter/handler",
    "repository": "adapter/repository"
}

pkg_renames = {
    "model": "domain",
    "api": "handler"
}

for old_name, new_path in moves.items():
    old_dir = os.path.join(internal_dir, old_name)
    new_dir = os.path.join(internal_dir, new_path)
    os.makedirs(os.path.dirname(new_dir), exist_ok=True)
    if os.path.exists(old_dir):
        shutil.move(old_dir, new_dir)

all_go_files = []
for root, _, files in os.walk(base_dir):
    for f in files:
        if f.endswith('.go'):
            all_go_files.append(os.path.join(root, f))

for filepath in all_go_files:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    orig_content = content
    
    for old_pkg, new_pkg in pkg_renames.items():
        new_dir = os.path.join(internal_dir, moves[old_pkg]).replace('\\', '/')
        if new_dir in filepath.replace('\\', '/'):
            content = re.sub(r'^package\s+' + old_pkg + r'\b', f'package {new_pkg}', content, count=1, flags=re.MULTILINE)
            content = re.sub(r'^package\s+' + old_pkg + r'_test\b', f'package {new_pkg}_test', content, count=1, flags=re.MULTILINE)

    for old_name, new_path in moves.items():
        content = content.replace(f'"nagare/internal/{old_name}"', f'"nagare/internal/{new_path}"')
        content = content.replace(f'"nagare/internal/{old_name}/', f'"nagare/internal/{new_path}/')

    for old_pkg, new_pkg in pkg_renames.items():
        content = re.sub(rf'\b{old_pkg}\.', f'{new_pkg}.', content)

    if content != orig_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)

print("Layered architecture refactoring completed.")
