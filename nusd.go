package GoNUSD

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net/http"

	"github.com/wii-tools/wadlib"
)

const nusdUrl = "http://ccs.cdn.wup.shop.nintendo.net/ccs/download"

func getContentIV(index uint16) []byte {
	iv := [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	iv[0] = uint8(index >> 8)
	iv[1] = uint8(index & 0xff)
	return iv[:]
}

func downloadContents(tmd wadlib.BinaryTMD, records []wadlib.ContentRecord, calculateSHA1 bool, ticket *wadlib.Ticket) ([][]byte, error) {
	buffer := make([][]byte, tmd.NumberOfContents)

	if ticket != nil {
		if err := ticket.DecryptKey(); err != nil {
			return nil, ErrTicketDecryptionFailure(err)
		}
	}

	var i uint16 = 0
	for ; i < tmd.NumberOfContents; i++ {
		contentURL := fmt.Sprintf("%s/%016x/%08x", nusdUrl, tmd.TitleID, records[i].ID)
		contentResponse, err := http.Get(contentURL)

		if err != nil {
			return nil, ErrHTTPFailure(contentURL, err)
		}

		if contentResponse.StatusCode != http.StatusOK {
			return nil, ErrContentNotFound(tmd.NumberOfContents, records[i].ID)
		}

		buffer[i] = make([]byte, contentResponse.ContentLength)
		if err = binary.Read(contentResponse.Body, binary.BigEndian, buffer[i]); err != nil {
			return nil, ErrBufferReadFailure(fmt.Sprintf("Content %08x", records[i].ID), err)
		}

		if ticket != nil {
			block, err := aes.NewCipher(ticket.TitleKey[:])
			if err != nil {
				return nil, ErrCommonKeyCipher
			}

			blockMode := cipher.NewCBCDecrypter(block, getContentIV(records[i].Index))

			decryptedData := make([]byte, len(buffer[i]))
			blockMode.CryptBlocks(decryptedData, buffer[i])

			if calculateSHA1 {
				sha1sum := sha1.Sum(decryptedData)
				if !bytes.Equal(records[i].Hash[:], sha1sum[:]) { // TODO: sometimes they're not equal, figure out why
					return nil, ErrInvalidHash(records[i].ID)
				}
			}

			copy(buffer[i], decryptedData)
		}
	}

	return buffer, nil
}

// Download downloads the specified title. Requires both the high and low title ID to be concatenated.
// Returns: TMD, Ticket (if available) plus all the contents (all raw, big endian).
// TODO: allow downloading content for other systems (Wii U, DSi + 3DS)
func Download(titleID uint64, version uint16, calculateSHA1, decodeAutomatically bool) ([][]byte, error) {
	tmdURL := fmt.Sprintf("%s/%016x/tmd", nusdUrl, titleID)
	if version != 0 {
		tmdURL += fmt.Sprintf(".%d", version)
	}

	tmdResponse, err := http.Get(tmdURL)
	if err != nil {
		return nil, ErrHTTPFailure(tmdURL, err)
	}

	if tmdResponse.StatusCode != http.StatusOK {
		return nil, ErrTitleNotFound
	}

	rawTMD := make([]byte, tmdResponse.ContentLength)
	if err = binary.Read(tmdResponse.Body, binary.BigEndian, rawTMD); err != nil {
		return nil, ErrBufferReadFailure("TMD", err)
	}

	tmdBuffer := bytes.NewBuffer(rawTMD)
	var tmd wadlib.BinaryTMD
	if err = binary.Read(tmdBuffer, binary.BigEndian, &tmd); err != nil {
		return nil, ErrBufferReadFailure("TMD header", err)
	}

	records := make([]wadlib.ContentRecord, tmd.NumberOfContents)
	if err = binary.Read(tmdBuffer, binary.BigEndian, &records); err != nil {
		return nil, ErrBufferReadFailure("TMD content", err)
	}

	if tmd.SignatureType != wadlib.SignatureRSA2048 {
		return nil, ErrTMDInvalidSignatureTypeFailure
	}

	if !decodeAutomatically {
		data, err := downloadContents(tmd, records, calculateSHA1, nil)
		if err != nil {
			return nil, err
		}

		buffer := make([][]byte, 2)

		copy(buffer[0], rawTMD)
		buffer = append(buffer, data...)

		return buffer, nil
	}

	ticketURL := fmt.Sprintf("%s/%016x/cetk", nusdUrl, titleID)
	ticketResponse, err := http.Get(ticketURL)
	if err != nil {
		return nil, ErrHTTPFailure(ticketURL, err)
	}

	if ticketResponse.StatusCode != http.StatusOK {
		return nil, ErrTicketNotFound
	}

	rawTicket := make([]byte, ticketResponse.ContentLength)
	if err = binary.Read(ticketResponse.Body, binary.BigEndian, rawTicket); err != nil {
		return nil, ErrBufferReadFailure("Ticket", err)
	}

	var (
		ticketBuffer = bytes.NewBuffer(rawTicket)
		ticket       wadlib.Ticket
	)
	if err = binary.Read(ticketBuffer, binary.BigEndian, &ticket); err != nil {
		return nil, ErrBufferReadFailure("Ticket contents", err)
	}

	data, err := downloadContents(tmd, records, calculateSHA1, &ticket)
	if err != nil {
		return nil, err
	}

	buffer := make([][]byte, 2)

	buffer[0] = rawTMD
	buffer[1] = rawTicket
	buffer = append(buffer, data...)

	return buffer, nil
}
