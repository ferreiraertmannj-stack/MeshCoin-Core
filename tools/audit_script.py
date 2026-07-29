import os
import json

def audit():
    exclude_dirs = ['build', '.gradle', '.dart_tool', 'node_modules', '.git', '__pycache__', 'PROJECT_AUDIT', '.idea', '.vscode']
    extensions = ['.go', '.py', '.dart', '.md', '.yaml', '.json']
    
    stats = {
        'total_files': 0,
        'loc': 0,
        'todos': 0,
        'fixmes': 0,
        'languages': set()
    }
    
    for root, dirs, files in os.walk('.'):
        dirs[:] = [d for d in dirs if d not in exclude_dirs]
        for f in files:
            ext = os.path.splitext(f)[1]
            if ext in extensions:
                stats['total_files'] += 1
                if ext == '.go': stats['languages'].add('Go')
                elif ext == '.py': stats['languages'].add('Python')
                elif ext == '.dart': stats['languages'].add('Dart')
                elif ext == '.md': stats['languages'].add('Markdown')
                elif ext == '.yaml': stats['languages'].add('YAML')
                
                path = os.path.join(root, f)
                try:
                    with open(path, 'r', encoding='utf-8') as file:
                        for line in file:
                            stats['loc'] += 1
                            if 'TODO' in line: stats['todos'] += 1
                            if 'FIXME' in line: stats['fixmes'] += 1
                except Exception:
                    pass
                    
    stats['languages'] = list(stats['languages'])
    print(json.dumps(stats, indent=2))

if __name__ == '__main__':
    audit()
