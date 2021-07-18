package GoNUSD

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTitleID                 = errors.New("invalid title ID")
	ErrTitleNotFound                  = errors.New("title not found")
	ErrTicketNotFound                 = errors.New("ticket not available")
	ErrCommonKeyCipher                = errors.New("failed to create a cipher using the common key")
	ErrTMDInvalidSignatureTypeFailure = errors.New("GoNUSD only supports signature type RSA 2048") // TODO: other consoles might use a different signature bit length, make sure this gets support down the line (just read signature type, then the signature and then the rest)

	ErrContentNotFound = func(numberOfContents uint16, id uint32) error {
		return fmt.Errorf("number of contents is %d, but content %08x is not available", numberOfContents, id)
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
