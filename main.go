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

		walkFn := func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, searchTerm) {
				if err := dirList.Append(path); err != nil {
					return err
				}
				scrollableList.Resize(
					fyne.NewSize(input.Size().Width, float32(dirList.Length()*100)),
				)
			}
			return nil
		}

		go func() {
			cwd, _ := filepath.Abs(".")
			if err := filepath.WalkDir(cwd, walkFn); err != nil {
				log.Fatal(err.Error())
			}
		}()
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
