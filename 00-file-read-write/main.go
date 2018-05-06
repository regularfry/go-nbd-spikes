package main

import (
	"fmt"
	"os"
	"strings"
)

func createDummyFile(filename string, data string) uint32 {
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	fatalIf(err)
	_, err = file.WriteString(data)
	fatalIf(err)
	stat, err := file.Stat()
	fatalIf(err)

	actualSize := uint32(stat.Size())

	file.Close()

	return actualSize
}

func main() {
	if len(os.Args) < 2 {
		panic("Filename needed")
	}
	filename := os.Args[1]

	actualSize := createDummyFile(filename, "this is some dummy data\n")

	d := NewDiscImage(filename)

	d.Read(1, 17, os.Stdout)
	fmt.Println("")
	d.Write(1, 3, strings.NewReader("hat"))
	d.Read(0, actualSize, os.Stdout)
	d.Read(0, actualSize+1, os.Stdout)
}
