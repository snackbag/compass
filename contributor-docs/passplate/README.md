# Passplate

The Compass templating engine. It consists of the following parts:

- lexer
- parser
- renderer
- interpreter

It brings support for variables and other logic components in normal HTML files, to be evaluated on the server-side. Its
inner workings deviate from the typical programming language interpreter. When a new request for a template is made (and
we exclude all caching magic), it looks similar to the following:

1. Load template file
2. Generate template AST
    - parser goes before the lexer, because the lexer only runs within Passplate blocks
3. Render AST
    - renderer and interpreter are separated, since the renderer mainly focuses around creating cohesive HTML, while the
      interpreter executes the logic

## Useful files

You can find all Passplate files within the `passplate` directory.
