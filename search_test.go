package main

import (
	"os"
	"strings"
	"testing"

	"github.com/mrchip53/gta-tools/rage/img"
	"github.com/mrchip53/gta-tools/rage/script"
	"github.com/mrchip53/gta-tools/rage/util"
)

func TestSearchScriptsInImg(t *testing.T) {
	searchTerm := "ambdealer" // Change this to your desired search term

	exef := "/home/mrchip/.local/share/Steam/steamapps/common/Grand Theft Auto IV/GTAIV/GTAIV.exe"
	exeBytes, err := os.ReadFile(exef)
	if err != nil {
		t.Fatalf("Failed to read exef file: %v", err)
	}
	util.FindAesKey(exeBytes)

	scriptf := "/home/mrchip/.local/share/Steam/steamapps/common/Grand Theft Auto IV/GTAIV/common/data/cdimages/script_orig.img"
	scriptBytes, err := os.ReadFile(scriptf)
	if err != nil {
		t.Fatalf("Failed to read scriptf file: %v", err)
	}

	imgFile := img.LoadImgFile(scriptBytes)

	for _, entry := range imgFile.Entries() {
		if strings.HasSuffix(entry.Name(), ".sco") {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic while processing script %s: %v", entry.Name(), r)
					}
				}()

				rageScript := script.NewRageScript(entry)
				if rageScript.Unsupported {
					t.Logf("Skipping unsupported script: %s", entry.Name())
					return
				}

				for _, instruction := range rageScript.Opcodes {
					instructionString := instruction.String("", nil)
					if strings.Contains(instructionString, searchTerm) {
						t.Logf("Found '%s' in %s at offset 0x%04X", searchTerm, entry.Name(), instruction.GetOffset())
					}
				}
			}()
		}
	}
}
