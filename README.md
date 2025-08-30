# memfile
[![Go Reference](https://pkg.go.dev/badge/github.com/cuhsat/memfile.svg)](https://pkg.go.dev/github.com/cuhsat/memfile)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuhsat/memfile?style=flat-square)](https://goreportcard.com/report/github.com/cuhsat/memfile)

Memfile is an in-memory, thread-safe, dependency-free file abstraction, that supports seek and notify. It tries to be compatible with most of the `io` interfaces that `os.File` supports. Memfile is meant for open-once usage.

## Example
```go
file := memfile.New("example")
file.WriteString("Hello World")

file.Seek(0, io.SeekStart)

buf, _ := io.ReadAll(file)
fmt.Println(string(buf))
```

## License
Released under the [MIT License](LICENSE.md).
