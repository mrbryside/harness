package styles

// glamourTheme generates a Glamour JSON theme that aligns with the
// current theme colors. Backgrounds use theme constants; syntax colors
// inside code blocks remain fixed for optimal highlighting contrast.
func glamourTheme() string {
	return `{
  "document": {
    "block_prefix": "",
    "block_suffix": "",
    "color": "` + hexFromColor(AssistantText) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `"
  },
  "paragraph": {
    "color": "` + hexFromColor(AssistantText) + `"
  },
  "text": {
    "color": "` + hexFromColor(AssistantText) + `"
  },
  "heading": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h1": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h2": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h3": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h4": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h5": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "h6": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "bold": true
  },
  "emphasis": {
    "color": "#E6DB74",
    "italic": true
  },
  "strong": {
    "color": "` + hexFromColor(AccentOrange) + `",
    "bold": true
  },
  "blockquote": {
    "color": "#A6E22E",
    "indent": 2,
    "indent_token": "│ "
  },
  "block_quote": {
    "color": "#A6E22E",
    "indent": 2,
    "indent_token": "│ "
  },
  "list": {
    "color": "` + hexFromColor(AssistantText) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `",
    "level_indent": 2
  },
  "item": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `",
    "block_prefix": "• "
  },
  "enumeration": {
    "color": "` + hexFromColor(ConnectedDot) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `",
    "block_prefix": ". "
  },
  "link": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "underline": true
  },
  "link_text": {
    "color": "` + hexFromColor(ModeBuildColor) + `",
    "underline": true
  },
  "hr": {
    "color": "#AE81FF",
    "format": "\n──────────────────────────────\n"
  },
  "table": {
    "color": "` + hexFromColor(AssistantText) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `"
  },
  "code": {
    "color": "` + hexFromColor(AccentOrange) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `",
    "bold": false
  },
  "code_block": {
    "color": "` + hexFromColor(AssistantText) + `",
    "background_color": "` + hexFromColor(ChatBackground) + `",
    "margin": 0,
    "chroma": {
      "background": {
        "color": "` + hexFromColor(ChatBackground) + `"
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
}

// MonokaiGlamourStyle returns the current theme-aligned Glamour JSON.
var MonokaiGlamourStyle = glamourTheme()
