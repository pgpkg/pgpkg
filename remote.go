package pgpkg

// FIXME: UNTESTED CODE COPIED DIRECTLY FROM ChatGPT.
// This is code that (supposedly) downloads a ZIP file of
// a Git archive. It's intended to allow pgpkg to implement
// package downloads from GitHub.
//
// This code has never been executed and might not work.
//

//func remote() {
//	// Set up the URL for the GitHub archive download
//	repo := "https://github.com/username/repo"
//	branch := "main"
//	url := fmt.Sprintf("%s/archive/refs/heads/%s.zip", repo, branch)
//
//	// Send an HTTP GET request to download the archive
//	resp, err := http.Get(url)
//	if err != nil {
//		fmt.Printf("Error downloading archive: %v\n", err)
//		return
//	}
//	defer resp.Body.Close()
//
//	// Create a new zip reader to read the archive
//	zipReader, err := zip.NewReader(resp.Body, resp.ContentLength)
//	if err != nil {
//		fmt.Printf("Error reading archive: %v\n", err)
//		return
//	}
//
//	// Extract each file from the archive to a corresponding file on disk
//	for _, file := range zipReader.File {
//		path := filepath.Join(".", file.Name)
//		if file.FileInfo().IsDir() {
//			os.MkdirAll(path, file.Mode())
//		} else {
//			reader, err := file.Open()
//			if err != nil {
//				fmt.Printf("Error opening file in archive: %v\n", err)
//				return
//			}
//			defer reader.Close()
//
//			writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
//			if err != nil {
//				fmt.Printf("Error creating file on disk: %v\n", err)
//				return
//			}
//			defer writer.Close()
//
//			_, err = io.Copy(writer, reader)
//			if err != nil {
//				fmt.Printf("Error extracting file: %v\n", err)
//				return
//			}
//		}
//	}
//
//	fmt.Println("Archive download and extraction complete.")
//}
