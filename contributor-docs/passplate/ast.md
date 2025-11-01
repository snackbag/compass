# AST

Passplate generates an AST to be interpreted/rendered. Pseudo:

```json lines
[
  {"name": "Text", "value": "<h1>hey</h1><p>"},
  {"name": "Variable", "id": "name"},
  {"name": "Text", "id": "</p>"}
]

/*

<h1>hey</h1>
<% if name == "cool" %>
  <p>your name is cool</p>
<% else %>
  <p>ur not cool</p>
<% end %>

*/

[
  {"name": "Text", "value": "<h1>hey</h1>\n"},
  {"name": "Statement", 
    "expr": { // name == "cool"
      "a": {"type": "variable", "name": "name"}, 
      "b": {"type": "string", "value": "cool"},
      "operator": "=="
    },

    "pass": [{"name": "Text", "value": "<p>your name is cool</p>"}],
    "else": [{"name": "Text", "value": "<p>ur not cool</p>"}]
  }
]

```