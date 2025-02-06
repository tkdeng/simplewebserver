package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	bash "github.com/tkdeng/gobash"
	regex "github.com/tkdeng/goregex"
	"github.com/tkdeng/goutil"
	"github.com/tkdeng/staticweb"
)

func compile() {
	os.MkdirAll(Config.Root, 0755)
	os.MkdirAll(Config.Root+"/pages", 0755)
	os.MkdirAll(Config.Root+"/theme", 0755)
	os.MkdirAll(Config.Root+"/assets", 0755)
	os.MkdirAll(Config.Root+"/db", 0755)
	os.MkdirAll(Config.Root+"/wasm", 0755)

	if Config.PublicURI != "" {
		os.MkdirAll(Config.Root+"/public", 0755)
	}

	//todo: sandbox downloads directory
	// os.MkdirAll(Config.Root+"/downloads", 2600)

	PrintMsg("warn", "Compiling Server Pages...", 50, false)

	os.RemoveAll(Config.Root + "/pages.dist")
	if err := os.Mkdir(Config.Root+"/pages.dist", 0755); err != nil {
		panic(err)
	}

	compTemplates()

	PrintMsg("warn", "Compiling WASM Scripts...", 50, false)

	os.RemoveAll(Config.Root + "/wasm.dist")
	if err := os.Mkdir(Config.Root+"/wasm.dist", 0755); err != nil {
		panic(err)
	}

	compWasm()

	PrintMsg("confirm", "Compiled Server!", 50, true)
}

func compTemplates() {
	staticweb.Live(Config.Root+"/pages", Config.Root+"/pages.dist")
}

func compWasm() {
	dirList, err := os.ReadDir(Config.Root + "/wasm")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, dir := range dirList {
		if dir.IsDir() {
			if path, err := goutil.JoinPath(Config.Root+"/wasm", dir.Name()); err == nil {
				compileWasmFile(path, dir.Name())
			}
		}
	}

	lastUpdate := time.Now().UnixMilli()

	fw := goutil.FileWatcher()
	fw.OnAny = func(path, op string) {
		if now := time.Now().UnixMilli(); now-lastUpdate > 1000 {
			lastUpdate = now

			path = string(regex.Comp(`^(%1[\\/][\w_-]+).*$`, Config.Root+"/wasm").RepStr([]byte(path), []byte("$1")))
			compileWasmFile(path, filepath.Base(path))
		}
	}
	fw.WatchDir(Config.Root + "/wasm")
}

func compileWasmFile(path string, fileName string) {
	if buf, err := os.ReadFile(path + "/go.mod"); err == nil {
		//* Compile Go WASM
		regex.Comp(`(?m)^module\s+(.+)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			fileName = string(data(1))
			return nil
		})

		out, err := goutil.JoinPath(Config.Root+"/wasm.dist", fileName+".wasm")
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = bash.Run([]string{`go`, `build`, `-o`, out}, path, []string{"GOOS=js", "GOARCH=wasm"})
		if err != nil {
			fmt.Println(err)
		}
	}
}
