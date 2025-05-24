package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrchip53/gta-tools/rage/util"
)

var (
	imgPath  string
	exePath  string
	imgBytes []byte
)

func readFileToBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	var err error
	flag.StringVar(&imgPath, "img", imgPath, "Path to the img file")
	flag.StringVar(&exePath, "exe", exePath, "Path to the exe file")
	flag.Parse()

	exeBytes, err := readFileToBytes(exePath)
	if err != nil {
		panic(err)
	}

	util.FindAesKey(exeBytes)

	imgBytes, err = readFileToBytes(imgPath)
	if err != nil {
		panic(err)
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
