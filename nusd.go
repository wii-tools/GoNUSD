package GoNUSD

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/wii-tools/wadlib"
)

const nusdUrl = "http://ccs.cdn.wup.shop.nintendo.net/ccs/download"

// Download downloads the specified title. Requires both the high and low title ID to be concatenated.
// Returns: TMD, Ticket (if available) plus all the contents (all raw, big endian).
// TODO: allow downloading content for other systems (Wii U, DSi + 3DS)
func Download(titleID uint64, version uint16, downloadTicket bool) (*wadlib.WAD, error) {
	var tmdFilename string
	if version != 0 {
		tmdFilename = fmt.Sprintf("tmd.%d", version)
	} else {
		tmdFilename = "tmd"
	}

	var wad wadlib.WAD

	tmdData, err := nusdFetch(titleID, tmdFilename)
	if err != nil {
		return nil, err
	}
	if wad.LoadTMD(tmdData) != nil {
		return nil, err
	}

	// Manually download and insert all contents
	wad.Data = make([]wadlib.WADFile, wad.TMD.NumberOfContents)

	var i uint16 = 0
	for ; i < wad.TMD.NumberOfContents; i++ {
		currentRecord := wad.TMD.Contents[i]

		contentData, err := nusdFetch(titleID, fmt.Sprintf("%08x", currentRecord.ID))
		if err != nil {
			return nil, err
		}

		file := wadlib.WADFile{
			ContentRecord: currentRecord,
			RawData:       contentData,
		}
		wad.Data[currentRecord.Index] = file
	}

	if !downloadTicket {
		// We're all done.
		return &wad, nil
	}

	ticketData, err := nusdFetch(titleID, "cetk")
	if err != nil {
		return nil, err
	}
	if err = wad.LoadTicket(ticketData); err != nil {
		return nil, err
	}

	return &wad, nil
}

func nusdFetch(titleID uint64, filename string) ([]byte, error) {
	url := fmt.Sprintf("%s/%016x/%s", nusdUrl, titleID, filename)

	response, err := http.Get(url)
	if err != nil {
		return nil, ErrHTTPFailure(url, err)
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound(filename)
	}

	if response.StatusCode != http.StatusOK {
		return nil, ErrServerError(filename, response.StatusCode)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, ErrBufferReadFailure(filename, err)
	} else {
		return contents, nil
	}
}
