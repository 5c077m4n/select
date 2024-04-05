package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type empty struct{}

const (
	WIDTH  = 1000
	HEIGHT = 400
)

func fileScan(searchTerm string, pathChannel chan<- string, abortChannel <-chan empty) {
	cwd, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	walkFn := func(path string, _ fs.DirEntry, err error) error {
		for {
			select {
			case <-abortChannel:
				close(pathChannel)
				return filepath.SkipAll
			default:
				if err != nil {
					close(pathChannel)
					return err
				}

				if strings.Contains(path, searchTerm) {
					pathChannel <- path
				}
				return nil
			}
		}
	}

	if err := filepath.WalkDir(cwd, walkFn); err != nil {
		log.Fatal(err.Error())
	}
	close(pathChannel)
}

func main() {
	a := app.New()

	window := a.NewWindow("Select")
	window.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyReturn || e.Name == fyne.KeyEscape {
			window.Close()
		}
	})
	window.SetPadded(false)
	window.Resize(fyne.NewSize(WIDTH, HEIGHT))
	window.CenterOnScreen()
	window.SetFixedSize(true)
	window.SetMaster()

	dirList := binding.NewStringList()
	list := widget.NewListWithData(
		dirList,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(item binding.DataItem, object fyne.CanvasObject) {
			if label, isLabel := object.(*widget.Label); isLabel {
				if boundString, isString := item.(binding.String); isString {
					label.Bind(boundString)
				}
			}
		},
	)
	scrollableList := container.NewVScroll(list)
	scrollableList.Hide()

	abortFileScanChannel := make(chan empty)

	input := widget.NewEntry()
	input.OnChanged = func(searchTerm string) {
		if err := dirList.Set([]string{}); err != nil {
			log.Fatal(err.Error())
		}

		searchTerm = strings.Trim(searchTerm, " ")
		if searchTerm == "" {
			scrollableList.Hide()
			return
		} else {
			scrollableList.Show()
			scrollableList.Resize(
				fyne.NewSize(
					window.Canvas().Size().Width,
					window.Canvas().Size().Height-input.Size().Height,
				),
			)
		}

		pathChannel := make(chan string)
		go fileScan(searchTerm, pathChannel, abortFileScanChannel)

		for path := range pathChannel {
			if err := dirList.Append(path); err != nil {
				log.Fatal(err)
			}
		}
	}
	input.OnSubmitted = func(content string) {
		window.Canvas().Focus(list)
	}

	content := container.NewVBox(input, scrollableList)
	content.Resize(window.Content().Size())

	window.SetContent(content)
	window.Canvas().Focus(input)

	window.ShowAndRun()
}
