package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type empty struct{}

const (
	WIDTH  = 1000
	HEIGHT = 400
)

//go:embed plugins/file-find.wasm
var pluginWASM []byte

func logString(_ context.Context, mod api.Module, offset, byteCount uint32) {
	if buf, ok := mod.Memory().Read(offset, byteCount); ok {
		fmt.Println(string(buf))
	} else {
		log.Fatalf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
}

func initPluginSystem() (wazero.Runtime, api.Module, context.Context) {
	ctx := context.Background()

	wasmRuntime := wazero.NewRuntime(ctx)

	_, err := wasmRuntime.
		NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(logString).
		Export("log").
		Instantiate(ctx)
	if err != nil {
		log.Panicln(err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, wasmRuntime)

	mod, err := wasmRuntime.Instantiate(ctx, pluginWASM)
	if err != nil {
		log.Fatalf("failed to start module: %v", err)
	}

	return wasmRuntime, mod, ctx
}

func fileScan(mod api.Module, ctx context.Context, searchTerm string) []string {
	malloc := mod.ExportedFunction("malloc")
	free := mod.ExportedFunction("free")

	searchTermSize := uint64(len(searchTerm))
	results, err := malloc.Call(ctx, searchTermSize)
	if err != nil {
		log.Fatal(err)
	}
	searchTermPointer := results[0]
	defer free.Call(ctx, searchTermPointer)

	if !mod.Memory().Write(uint32(searchTermPointer), []byte(searchTerm)) {
		log.Fatalf(
			"Memory.Write(%d, %d) out of range of memory size %d",
			searchTermPointer,
			searchTermSize,
			mod.Memory().Size(),
		)
	}

	search := mod.ExportedFunction("search")
	pointerSize, err := search.Call(ctx, searchTermPointer, searchTermSize)
	if err != nil {
		log.Fatal(err)
	}

	matchesPointer := uint32(pointerSize[0] >> 32)
	matchesSize := uint32(pointerSize[0])

	if matchesPointer != 0 {
		defer func() {
			if _, err := free.Call(ctx, uint64(matchesPointer)); err != nil {
				log.Fatal(err)
			}
		}()
	}

	bytes, ok := mod.Memory().Read(matchesPointer, matchesSize)
	if !ok {
		log.Fatalf(
			"Memory.Read(%d, %d) out of range of memory size %d",
			matchesPointer,
			matchesSize,
			mod.Memory().Size(),
		)
	}

	var matches []string
	if err := json.Unmarshal(bytes, &matches); err != nil {
		log.Fatal(err)
	}

	return matches
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

	wasmRuntime, mod, ctx := initPluginSystem()
	defer wasmRuntime.Close(ctx)

	input := widget.NewEntry()
	input.OnChanged = func(searchTerm string) {
		if err := dirList.Set([]string{}); err != nil {
			log.Fatal(err.Error())
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
		scrollableList.Show()

		paths := fileScan(mod, ctx, searchTerm)
		dirList.Set(paths)
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
