# memfile
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
