package main

import (
	"fmt"

	"github.com/wii-tools/GoNUSD"
)

const (
	titleID = 0x0000000100000002
	version = 514
)

func main() {
	// TODO: interpret command line

	if _, err := GoNUSD.Download(titleID, version, false, true); err != nil { // Mii Channel (Wii)
		fmt.Printf("Failed to download title %016x (%d): \"%s\".\n", titleID, version, err.Error())
		return
	}

	fmt.Printf("Successfully downloaded %016x (%d).\n", titleID, version)
}
