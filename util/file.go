package util

import (
	"fmt"
	"github.com/twodarek/go-twitter/twitter"
	"io"
	"net/http"
	"os"
)

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetHighestBitrateVariant(variants []twitter.VideoVariant) twitter.VideoVariant {
	highestBitrateVar := variants[0]
	for _, variant := range variants {
		if variant.Bitrate > highestBitrateVar.Bitrate {
			highestBitrateVar = variant
		}
	}
	return highestBitrateVar
}
