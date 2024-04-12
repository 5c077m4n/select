// main
package main

import (
	"context"
	_ "embed"
	"log"
	"select/plugins/appscan"
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
	width  = 1000
	height = 400
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
	window.Resize(fyne.NewSize(width, height))
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
	list.OnSelected = func(id int) {
		selected, err := dirList.GetValue(id)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println(selected)
		window.Close()
	}
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

		appsChannel := make(chan string, 100)
		go appscan.AppScan(searchTerm, appsChannel)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pathsChannel := make(chan string, 100)
		go filescan.FileScan(ctx, ".", searchTerm, pathsChannel)

		scrollableList.Resize(
			fyne.NewSize(
				window.Canvas().Size().Width,
				window.Canvas().Size().Height-input.Size().Height,
			),
		)
		scrollableList.Show()

		for app := range appsChannel {
			dirList.Append(app)
		}
		for path := range pathsChannel {
			dirList.Append(path)
		}

		list.Refresh()
	}
	input.OnSubmitted = func(_content string) {
		window.Canvas().Focus(list)
	}

	content := container.NewVBox(input, scrollableList)
	content.Resize(window.Content().Size())

	window.SetContent(content)
	window.Canvas().Focus(input)

	window.ShowAndRun()
}
