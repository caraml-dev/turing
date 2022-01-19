package utils

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

// DownloadFile will download a given url into a local file.
func DownloadFile(url *url.URL, filepath string) error {

	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
