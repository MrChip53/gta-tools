package img

import (
	"os"
	"testing"

	"github.com/mrchip53/gta-tools/rage/util"
)

func TestCompareIdenticalImgFilesLoadedFromBytes(t *testing.T) {
	exef := "/home/mrchip/.local/share/Steam/steamapps/common/Grand Theft Auto IV/GTAIV/GTAIV.exe"
	exeBytes, err := os.ReadFile(exef)
	if err != nil {
		t.Fatalf("Failed to read exef file: %v", err)
	}

	util.FindAesKey(exeBytes)

	oivf := "/home/mrchip/.local/share/Steam/steamapps/common/Grand Theft Auto IV/GTAIV/common/data/cdimages/script_oiv.img"
	oivBytes, err := os.ReadFile(oivf)
	if err != nil {
		t.Fatalf("Failed to read oivf file: %v", err)
	}

	scriptf := "/home/mrchip/.local/share/Steam/steamapps/common/Grand Theft Auto IV/GTAIV/common/data/cdimages/script.img"
	scriptBytes, err := os.ReadFile(scriptf)
	if err != nil {
		t.Fatalf("Failed to read scriptf file: %v", err)
	}

	// 3. Load two ImgFile instances from the same byte slice
	// Note: The LoadImgFile function in the prompt seems to have a potential issue where it uses a global 'data' variable
	// for slicing entry data instead of the 'data' parameter passed to it. This might need correction in img_file.go.
	// Assuming LoadImgFile works correctly with its input 'data' parameter for this test.
	img1 := LoadImgFile(oivBytes)
	img2 := LoadImgFile(scriptBytes)

	// 4. Assert that they have the same number of entries
	if len(img1.Entries()) != len(img2.Entries()) {
		t.Fatalf("Expected same number of entries, got %d and %d", len(img1.Entries()), len(img2.Entries()))
	}
	if len(img1.Entries()) == 0 {
		t.Fatalf("Test requires at least one entry to compare, but loaded files are empty.")
	}

	// 5. Compare entries by index: name and hash should be the same
	for i := range img1.Entries() {
		e1 := img1.Entries()[i]
		e2 := img2.Entries()[i]

		if e1.Name() != e2.Name() {
			t.Errorf("Entry %d: Names do not match. Got '%s' and '%s'", i, e1.Name(), e2.Name())
		}
		if e1.Hash() != e2.Hash() {
			t.Errorf("Entry %d ('%s'): Hashes do not match. Got %s and %s", i, e1.Name(), e1.Hash(), e2.Hash())
		}
		if e1.Toc().Flags != e2.Toc().Flags {
			t.Errorf("Entry %d ('%s'): Toc flags do not match. Got %d and %d", i, e1.Name(), e1.Toc().Flags, e2.Toc().Flags)
		}
		if e1.Toc().Size != e2.Toc().Size {
			t.Errorf("Entry %d ('%s'): Toc entry size do not match. Got %d and %d", i, e1.Name(), e1.Toc().Size, e2.Toc().Size)
		}
		if e1.Toc().OffsetBlock != e2.Toc().OffsetBlock {
			t.Errorf("Entry %d ('%s'): Toc offset do not match. Got %d and %d", i, e1.Name(), e1.Toc().OffsetBlock, e2.Toc().OffsetBlock)
		}
		if e1.Toc().UsedBlocks != e2.Toc().UsedBlocks {
			t.Errorf("Entry %d ('%s'): Toc used count do not match. Got %d and %d", i, e1.Name(), e1.Toc().UsedBlocks, e2.Toc().UsedBlocks)
		}
		if e1.Toc().entrySize != e2.Toc().entrySize {
			t.Errorf("Entry %d ('%s'): Toc offset do not match. Got %d and %d", i, e1.Name(), e1.Toc().entrySize, e2.Toc().entrySize)
		}
	}
}
