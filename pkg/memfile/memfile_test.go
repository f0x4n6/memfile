package memfile

import (
	"fmt"
	"io"
)

func Example() {
	f := New("example")

	_, err := f.WriteString("Hello World")

	if err != nil {
		panic(err)
	}

	_, err = f.Seek(0, io.SeekStart)

	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(f)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	// Output:
	// Hello World
}
