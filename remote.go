package pgpkg

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Remote() {
	// Set up the URL for the GitHub archive download
	repo := "https://github.com/pgpkg/pgpkg-test"
	branch := "main"
	url := fmt.Sprintf("%s/archive/refs/heads/%s.zip", repo, branch)

	// Send an HTTP GET request to download the archive
	resp, err := http.Get(url)
	if err != nil {
		Stderr.Printf("Error downloading archive: %v\n", err)
		return
	}
	defer resp.Body.Close()

	output, err := os.Create("/tmp/pgpkg.zip")
	if err != nil {
		panic(err)
	}

	if _, err = io.Copy(output, resp.Body); err != nil {
		panic(err)
	}
}
