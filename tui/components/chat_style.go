package components

// monokaiGlamourStyle is a custom glamour JSON theme inspired by Monokai —
// vibrant syntax highlighting with high contrast on a dark background.
const monokaiGlamourStyle = `{
  "document": {
    "block_prefix": "",
    "block_suffix": "",
    "color": "#E6EDF3",
    "background_color": "#0A0A0A"
  },

  "paragraph": {
    "color": "#E6EDF3"
  },

  "text": {
    "color": "#E6EDF3"
  },

  "heading": {
    "color": "#79C0FF",
    "bold": true
  },

  "h1": {
    "color": "#79C0FF",
    "bold": true
  },

  "h2": {
    "color": "#79C0FF",
    "bold": true
  },

  "h3": {
    "color": "#79C0FF",
    "bold": true
  },

  "h4": {
    "color": "#79C0FF",
    "bold": true
  },

  "h5": {
    "color": "#79C0FF",
    "bold": true
  },

  "h6": {
    "color": "#79C0FF",
    "bold": true
  },

  "emphasis": {
    "color": "#C9D1D9",
    "italic": true
  },

  "strong": {
    "color": "#FF7B72",
    "bold": true
  },

  "blockquote": {
    "color": "#8B949E",
    "indent": 2,
    "indent_token": "│ "
  },

  "block_quote": {
    "color": "#8B949E",
    "indent": 2,
    "indent_token": "│ "
  },

  "list": {
    "color": "#E6EDF3",
    "level_indent": 2
  },

  "item": {
    "color": "#79C0FF",
    "block_prefix": "• "
  },

  "enumeration": {
    "color": "#FFB86C",
    "block_prefix": ". "
  },

  "link": {
    "color": "#79C0FF",
    "underline": true
  },

  "link_text": {
    "color": "#79C0FF",
    "underline": true
  },

  "hr": {
    "color": "#30363D",
    "format": "\n──────────────────────────────\n"
  },

  "table": {
    "color": "#E6EDF3",
    "background_color": "#111111"
  },

  "code": {
    "color": "#FFB86C",
    "background_color": "#111111",
    "bold": false
  },

  "code_block": {
    "color": "#E6EDF3",
    "background_color": "#0D1117",
    "margin": 0,

    "chroma": {
      "background": {
        "color": "#0D1117"
      },

      "text": {
        "color": "#E6EDF3"
      },

      "error": {
        "color": "#FF7B72"
      },

      "comment": {
        "color": "#6E7681",
        "italic": true
      },

      "comment_preproc": {
        "color": "#6E7681"
      },

      "keyword": {
        "color": "#FF7B72"
      },

      "keyword_constant": {
        "color": "#FF7B72"
      },

      "keyword_declaration": {
        "color": "#FF7B72"
      },

      "keyword_namespace": {
        "color": "#FF7B72"
      },

      "keyword_pseudo": {
        "color": "#FF7B72"
      },

      "keyword_reserved": {
        "color": "#FF7B72"
      },

      "keyword_type": {
        "color": "#79C0FF"
      },

      "operator": {
        "color": "#FF7B72"
      },

      "operator_word": {
        "color": "#FF7B72"
      },

      "punctuation": {
        "color": "#E6EDF3"
      },

      "name": {
        "color": "#E6EDF3"
      },

      "name_attribute": {
        "color": "#79C0FF"
      },

      "name_builtin": {
        "color": "#79C0FF"
      },

      "name_builtin_pseudo": {
        "color": "#79C0FF"
      },

      "name_class": {
        "color": "#79C0FF"
      },

      "name_constant": {
        "color": "#79C0FF"
      },

      "name_decorator": {
        "color": "#79C0FF"
      },

      "name_entity": {
        "color": "#79C0FF"
      },

      "name_exception": {
        "color": "#79C0FF"
      },

      "name_function": {
        "color": "#FFB86C"
      },

      "name_function_magic": {
        "color": "#FFB86C"
      },

      "name_label": {
        "color": "#79C0FF"
      },

      "name_namespace": {
        "color": "#79C0FF"
      },

      "name_other": {
        "color": "#E6EDF3"
      },

      "name_property": {
        "color": "#79C0FF"
      },

      "name_tag": {
        "color": "#FF7B72"
      },

      "name_variable": {
        "color": "#E6EDF3"
      },

      "name_variable_class": {
        "color": "#E6EDF3"
      },

      "name_variable_global": {
        "color": "#E6EDF3"
      },

      "name_variable_instance": {
        "color": "#E6EDF3"
      },

      "literal": {
        "color": "#A5D6FF"
      },

      "literal_date": {
        "color": "#A5D6FF"
      },

      "literal_string": {
        "color": "#E6EDF3"
      },

      "literal_string_affix": {
        "color": "#E6EDF3"
      },

      "literal_string_backtick": {
        "color": "#E6EDF3"
      },

      "literal_string_char": {
        "color": "#E6EDF3"
      },

      "literal_string_delimiter": {
        "color": "#E6EDF3"
      },

      "literal_string_doc": {
        "color": "#6E7681"
      },

      "literal_string_double": {
        "color": "#E6EDF3"
      },

      "literal_string_escape": {
        "color": "#79C0FF"
      },

      "literal_string_heredoc": {
        "color": "#E6EDF3"
      },

      "literal_string_interpol": {
        "color": "#79C0FF"
      },

      "literal_string_other": {
        "color": "#E6EDF3"
      },

      "literal_string_regex": {
        "color": "#A5D6FF"
      },

      "literal_string_single": {
        "color": "#E6EDF3"
      },

      "literal_string_symbol": {
        "color": "#E6EDF3"
      },

      "literal_number": {
        "color": "#A5D6FF"
      },

      "generic_deleted": {
        "color": "#FF7B72"
      },

      "generic_emph": {
        "italic": true
      },

      "generic_inserted": {
        "color": "#7EE787"
      },

      "generic_strong": {
        "bold": true
      },

      "generic_subheading": {
        "color": "#79C0FF"
      }
    }
  }
}`
