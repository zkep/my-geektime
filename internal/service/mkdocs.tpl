site_name: {{ .SiteName }}

theme:
  name: material

  palette:
    # Palette toggle for light mode
    - media: "(prefers-color-scheme: light)"
      scheme: default
      toggle:
        icon: material/brightness-4
        name: Switch to dark mode

    # Palette toggle for dark mode
    - media: "(prefers-color-scheme: dark)"
      scheme: slate
      toggle:
        icon: material/brightness-7
        name: Switch to light mode

  features:
    - navigation.footer
    - content.action.edit
    - navigation.sections
    - content.code.copy

  icon:
    edit: material/pencil
    logo: material/console-line

extra:
  homepage:
  version:
    provider: mike

copyright: Copyright &copy; 2024-Present zkep
repo_url: https://github.com/zkep/my-geektime

markdown_extensions:
  - attr_list
  - md_in_html
  - pymdownx.highlight:
      anchor_linenums: true
  - pymdownx.inlinehilite
  - pymdownx.snippets
  - pymdownx.superfences

nav:
  {{- range $idx, $nav := .Navs }}
   {{ if $nav.Name }}
    - {{ $nav.Name  }}:
   {{- end }}
       {{- range $idx, $item := $nav.Items -}}
          - {{  $item  -}}
       {{- end }}
  {{ end }}