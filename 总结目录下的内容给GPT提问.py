import os

def read_file(file_path):
    print(f'Reading file: {file_path}')
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            return file.read()
    except UnicodeDecodeError:
        # 尝试使用不同的编码读取或忽略错误
        return None

def get_project_structure(root_dir, ignored_paths, indent=''):
    structure = ''
    for item in os.listdir(root_dir):
        item_path = os.path.join(root_dir, item)
        if item_path in ignored_paths or '.git' in item_path:
            continue
        if os.path.isdir(item_path):
            structure += f'{indent}{item}/\n'
            structure += get_project_structure(item_path, ignored_paths, indent + '    ')
        else:
            structure += f'{indent}{item}\n'
    return structure

def get_all_files(root_dir, files_list, ignored_paths):
    for item in os.listdir(root_dir):
        item_path = os.path.join(root_dir, item)
        if item_path in ignored_paths or '.git' in item_path:
            continue
        if os.path.isdir(item_path):
            get_all_files(item_path, files_list, ignored_paths)
        else:
            files_list.append(item_path)

def read_gitignore(project_dir, gitignore_path):
    ignored_paths = set()
    if os.path.exists(gitignore_path):
        with open(gitignore_path, 'r', encoding='utf-8') as file:
            for line in file:
                line = line.strip()
                if line and not line.startswith('#'):
                    ignored_paths.add(os.path.join(project_dir, line))
    return ignored_paths

def main():
    project_dir = '/Users/didi/workspace/OfferCat/Resume analysis'  # 修改为你的想要总结的项目路径
    output_dir = '/Users/didi/workspace/OfferCat/GPT_tmp'
    output_file = os.path.join(output_dir, os.path.basename(project_dir) + '.txt')

    # 读取.gitignore文件内容
    gitignore_path = os.path.join(project_dir, '.gitignore')
    ignored_paths = read_gitignore(project_dir, gitignore_path)

    # 读取README文件内容
    readme_content = ''
    for readme_name in ['README.md', 'readme.md', 'README.txt', 'readme.txt']:
        readme_path = os.path.join(project_dir, readme_name)
        if os.path.exists(readme_path):
            readme_content = read_file(readme_path)
            break

    # 获取项目结构
    project_structure = get_project_structure(project_dir, ignored_paths)

    # 获取所有文件的内容
    all_files = []
    get_all_files(project_dir, all_files, ignored_paths)
    files_content = ''
    for file_path in all_files:
        relative_path = os.path.relpath(file_path, project_dir)
        content = read_file(file_path)
        if content is None:
            print(f'Error reading file: {file_path}')
            continue
        files_content += f'\n{"-" * 80}\nFile: {relative_path}\n{"-" * 80}\n'
        files_content += content

    # 输出到txt文件
    with open(output_file, 'w', encoding='utf-8') as out_file:
        out_file.write('README Content:\n')
        out_file.write('=' * 80 + '\n')
        out_file.write(readme_content + '\n\n')
        out_file.write('Project Structure:\n')
        out_file.write('=' * 80 + '\n')
        out_file.write(project_structure + '\n\n')
        out_file.write('Files Content:\n')
        out_file.write('=' * 80 + '\n')
        out_file.write(files_content)
    print(f'Output to {output_file}')

if __name__ == '__main__':
    main()