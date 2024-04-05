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

func fileScan(pathChannel chan string, searchTerm string) {
	walkFn := func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			close(pathChannel)
			return err
		}
		if strings.Contains(path, searchTerm) {
			pathChannel <- path
		}
		return nil
	}

	cwd, _ := filepath.Abs("..")
	if err := filepath.WalkDir(cwd, walkFn); err != nil {
		close(pathChannel)
		log.Fatal(err.Error())
	}
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
	window.Resize(fyne.NewSize(1000, 200))
	window.CenterOnScreen()

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
	list.Resize(
		fyne.NewSize(window.Canvas().Size().Width, window.Canvas().Size().Height-100),
	)
	scrollableList := container.NewVScroll(list)

	input := widget.NewEntry()
	input.Resize(fyne.NewSize(window.Canvas().Size().Width, 100))
	input.OnChanged = func(searchTerm string) {
		if err := dirList.Set([]string{}); err != nil {
			log.Fatal(err.Error())
		}
		if searchTerm == "" {
			return
		}

		pathChannel := make(chan string)
		go fileScan(pathChannel, searchTerm)

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
	content.Resize(window.Canvas().Size())

	window.SetContent(content)
	window.Canvas().Focus(input)

	window.ShowAndRun()
}
