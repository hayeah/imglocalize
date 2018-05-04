# imglocalize

Scan text file for markdown or html images, and rewrite the src as local images.

## Example

To write all remote images as local, run:

```
imglocalize foo/README.md
```

All images will be downloaded in the directory of the text file (i.e. `foo`)
