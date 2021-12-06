package GoNUSD

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTitleID                 = errors.New("invalid title ID")
	ErrCommonKeyCipher                = errors.New("failed to create a cipher using the common key")
	ErrTMDInvalidSignatureTypeFailure = errors.New("GoNUSD only supports signature type RSA 2048") // TODO: other consoles might use a different signature bit length, make sure this gets support down the line (just read signature type, then the signature and then the rest)

	ErrFileNotFound = func(filename string) error {
		return fmt.Errorf("file %s was not found on NUS", filename)
	}

	ErrServerError = func(filename string, error int) error {
		return fmt.Errorf("server returned %d upon download of %s", error, filename)
	}

	ErrTicketDecryptionFailure = func(err error) error {
		return fmt.Errorf("failed to decode the ticket: %s", err.Error())
	}

	ErrInvalidHash = func(id uint32) error {
		return fmt.Errorf("hash for content %08x did not match the hash stored in its record", id)
	}

	ErrHTTPFailure = func(url string, err error) error {
		return fmt.Errorf("failed to retrieve %s: %s", url, err.Error())
	}

	ErrBufferReadFailure = func(name string, err error) error {
		return fmt.Errorf("failed to read %s: %s", name, err.Error())
	}

	ErrEncodingFailure = func(name string, err error) error {
		return fmt.Errorf("failed to encode %s: %s", name, err.Error())
	}
)
