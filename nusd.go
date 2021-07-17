package GoNUSD

import (
	"fmt"
	"github.com/wii-tools/wadlib"
	"io/ioutil"
	"os"
)

var nusdUrl = "http://nus.cdn.shop.wii.com/ccs/download"


func NUSD(titleID string, outPath string, WAD bool) error {
	// Check if titleID is 16 characters
	if len(titleID) != 16 {
		fmt.Println("All title ID's must be 16 characters long.")
		return nil
	}
	
	// The ticket variable is here so that if a ticket could not be downloaded, we know not to pack the contents.
	var ticket = true
	err := os.Mkdir(outPath, 0777)
	if err != nil {
		if os.IsExist(err) {
			// The requested directory already exists. We do not want to take any chances with packing the wrong content.
			// As such, we will tell the user to delete the folder then try again.
			fmt.Println("The requested directory already exists. Please choose another directory or delete the directory you wish to download to.")
		}
		return err
	}

	// We will download the TMD first in order to get the contents we need to download
	file := fmt.Sprintf("%s/%s/tmd", nusdUrl, titleID)
	statusCode, err := DownloadFile(fmt.Sprintf("%s/tmd", outPath), file)
	if err != nil {
		return err
	}

	// Title does not exist
	if statusCode == 404 {
		fmt.Println("Requested title does not exist on NUS.")
		return nil
	}

	// Parse the TMD file
	contents, err := ioutil.ReadFile(fmt.Sprintf("%s/tmd", outPath))
	if err != nil {
		return err
	}

	wad := wadlib.WAD{}
	err = wad.LoadTMD(contents)
	if err != nil {
		return err
	}

	// Download ticket, if it exists
	file = fmt.Sprintf("%s/%s/cetk", nusdUrl, titleID)
	statusCode, err = DownloadFile(fmt.Sprintf("%s/cetk", outPath), file)
	if err != nil {
		return nil
	}

	// Ticket does not exist. We will not pack the contents
	if statusCode != 200 {
		ticket = false
		if titleID != "0001000148434d50" {
			fmt.Println("Ticket either failed to download or doesn't exist. A WAD will not be created.")
		}
	}

	if ticket == true {
		// Parse ticket
		contents, err = ioutil.ReadFile(fmt.Sprintf("%s/cetk", outPath))
		if err != nil {
			return err
		}

		err = wad.LoadTicket(contents)
		if err != nil {
			return err
		}

	}

	// Download and decrypt WAD contents
	for _, content := range wad.TMD.Contents {
		file = fmt.Sprintf("%s/%s/%08x", nusdUrl, titleID, content.ID)
		filepath := fmt.Sprintf("%s/%08x", outPath, content.ID)
		// Create IV based on the content's Index
		iv := []byte{0x00, byte(content.Index), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		// Download content from NUSD
		_, err := DownloadFile(filepath, file)
		if err != nil {
			return err
		}

		data, errr := ioutil.ReadFile(filepath)
		if errr != nil {
			return errr
		}

		// Decrypt content
		newFilepath := fmt.Sprintf("%s/%08x.app", outPath, content.Index)
		decryptAESCBC(wad.Ticket.TitleKey[:], iv, data, newFilepath)

		// Delete the encrypted contents
		err = os.Remove(filepath)
		if err != nil {
			return err
		}
	}

	// If a ticket was not downloaded, this would be where we finish.
	if ticket == true {
		// Rename ticket and tmd
		newCetk := fmt.Sprintf("%s/%s.tik", outPath, titleID)
		oldCetk := fmt.Sprintf("%s/cetk", outPath)
		err := os.Rename(oldCetk, newCetk)
		if err != nil {
			return err
		}

		newTMD := fmt.Sprintf("%s/%s.tmd", outPath, titleID)
		oldTMD := fmt.Sprintf("%s/tmd", outPath)
		err = os.Rename(oldTMD, newTMD)
		if err != nil {
			return err
		}

		// Download certificate
		_, err = DownloadFile(fmt.Sprintf("%s/%s.certs", outPath, titleID), "https://sketchmaster2001.github.io/cert")
		if err != nil {
			return err
		}

		// Make footer file
		os.Create(fmt.Sprintf("%s/%s.footer", outPath, titleID))

		if WAD == true {
			// Finally, pack the WAD
			wadDir := fmt.Sprintf("%s/%s.wad", outPath, titleID)
			err = Pack(outPath, wadDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
