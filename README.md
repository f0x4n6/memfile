# memfile
[![Go Reference](https://pkg.go.dev/badge/github.com/cuhsat/memfile.svg)](https://pkg.go.dev/github.com/cuhsat/memfile)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuhsat/memfile?style=flat-square)](https://goreportcard.com/report/github.com/cuhsat/memfile)

In-memory `os.File` abstraction for Go.

```console
go get github.com/cuhsat/memfile
```

## Example
```go
f := New("example")

f.WriteString("Hello World")

f.Seek(0, io.SeekStart)

b, _ := io.ReadAll(f)

fmt.Println(string(b))
```

## License
Released under the [MIT License](LICENSE.md).
