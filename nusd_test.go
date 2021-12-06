package GoNUSD

import (
	"testing"
)

const (
	titleID = 0x0000000100000002
	version = 514
)

// TODO: hash checks
func TestDownload(t *testing.T) {
	wad, err := Download(0x0000000100000002, 514, false)
	if err != nil { // System Menu 4.3E (Wii)
		t.Fatalf("Failed downloading: \"%s\".", err.Error())
	}

	tmd := wad.TMD
	if tmd.TitleID != titleID {
		t.Fatalf("Title ID: %016x != %016x", tmd.TitleID, titleID)
	}

	if tmd.TitleVersion != version {
		t.Fatalf("Version: %d != %d", tmd.TitleVersion, version)
	}

	t.Logf("Successfully downloaded %016x (v %d)", tmd.TitleID, tmd.TitleVersion)
}
