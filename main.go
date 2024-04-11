package main

import (
	"context"
	_ "embed"
	"log"
	"select/plugins/filescan"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	WIDTH  = 1000
	HEIGHT = 400
)

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

	input := widget.NewEntry()
	input.OnChanged = func(searchTerm string) {
		if err := dirList.Set([]string{}); err != nil {
			log.Fatalln(err.Error())
		}

		searchTerm = strings.Trim(searchTerm, " ")
		if searchTerm == "" {
			scrollableList.Hide()
			return
		}

		scrollableList.Resize(
			fyne.NewSize(
				window.Canvas().Size().Width,
				window.Canvas().Size().Height-input.Size().Height,
			),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pathsChannel := make(chan string)
		go filescan.FileScan(".", searchTerm, pathsChannel, ctx)

		for path := range pathsChannel {
			dirList.Append(path)
		}

		scrollableList.Show()
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
