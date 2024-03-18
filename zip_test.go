package pgpkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func showZip(buf []byte) {
	// Open a zip archive for reading.
	readerAt := bytes.NewReader(buf)

	r, err := zip.NewReader(readerAt, int64(len(buf)))
	if err != nil {
		panic(err)
	}

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(os.Stdout, rc)
		//if err != nil {
		//	log.Fatal(err)
		//}
		rc.Close()
		fmt.Println()
	}
}

func TestZip(t *testing.T) {
	p, err := NewProjectFrom("tests/good/example")
	if err != nil {
		panic(err)
	}

	// in-memory zip
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err = WriteProject(zipWriter, p)
	if err != nil {
		panic(err)
	}

	if err := zipWriter.Close(); err != nil {
		panic(err)
	}

	showZip(buf.Bytes())
}
