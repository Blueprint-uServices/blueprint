# Blueprint API Documentation

Deprecated: do not use this any more.

Blueprint API documentation for the core plugins, external plugins, and blueprint runtime packages is available in [api](api/).

## Updating Documentation

### Dependencies

The documentations is generated using ```go doc``` and converted to markdown using godoc2markdown.
Installation instructions for installing godoc2markdown can be found [here](https://git.sr.ht/~humaid/godoc2markdown).

### Re-generating documentation

To regenerate the documentation, execute the following script from the root blueprint directory:

```bash
python scripts/gen_docs.py
```

### Updating Website docs

To update the documentation on the website, please follow the instructions [here](https://github.com/blueprint-uservices/Blueprint-uServices.github.io/tree/main#updating-documentation).