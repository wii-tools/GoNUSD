package GoNUSD

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// DownloadFile downloads files from the specified url.
func DownloadFile(filepath string, url string) (int, error) {
	var statusCode int
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return statusCode, err
	}

	statusCode = resp.StatusCode

	if statusCode < 400 {
		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return statusCode, err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
	}

	return statusCode, err
}

// decryptAESCBC is used to decrypt the app files into a format that can be packed into a WAD.
func decryptAESCBC(key []byte, iv []byte, data []byte, outPath string) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(data, data)

	err = ioutil.WriteFile(outPath, data, 0777)
	if err != nil {
		fmt.Printf("An error has occurred when writing the decrypted data: %s\n", err)
		return
	}
}
