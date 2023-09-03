project = 'infra-course'
copyright = '2023, Алексей Юдолевич'
author = 'Алексей Юдолевич'

extensions = [
    'myst_parser',
    'sphinx_rtd_theme',
    'sphinxcontrib.mermaid',
    'sphinx_copybutton',
]

copybutton_prompt_text = r'$ '
copybutton_only_copy_prompt_lines = True
copybutton_remove_prompts = True
copybutton_copy_empty_lines = False
copybutton_here_doc_delimiter = 'EOF'
copybutton_exclude = '.go'

templates_path = ['_templates']
exclude_patterns = [
    '_build', 'Thumbs.db', '.DS_Store',
    'notes', 'slides',
]

language = 'ru'

html_theme = 'sphinx_rtd_theme'
