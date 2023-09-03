project = 'slides'
copyright = '2023, Алексей Юдолевич'
author = 'Алексей Юдолевич'

extensions = [
    'myst_parser',
    'sphinx_revealjs',
    'sphinxcontrib.mermaid',
]

templates_path = ['_templates']
exclude_patterns = [
    '_build', 'Thumbs.db', '.DS_Store',
    'custom_html.md', '*_topic_*.md',
]

language = 'ru'

revealjs_style_theme = 'league'
# revealjs_js_files = ["mermaid.js"]
revealjs_static_path = ['_static']
revealjs_script_plugins = [
    {
        "name": "RevealHighlight",
        "src": "revealjs4/plugin/highlight/highlight.js",
    },
    {
        "name": "mermaid",
        "src": "revealjs4/plugin/mermaid/mermaid.js",
    },
]
revealjs_css_files = [
    "revealjs4/plugin/highlight/zenburn.css",
]
