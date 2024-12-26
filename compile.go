package server

import (
	"fmt"
	"os"
	"time"

	"github.com/tkdeng/goutil"
	"github.com/tkdeng/staticweb"
)

func compile() {
	os.MkdirAll(Config.Root, 0755)
	os.MkdirAll(Config.Root+"/pages", 0755)
	os.MkdirAll(Config.Root+"/theme", 0755)
	os.MkdirAll(Config.Root+"/assets", 0755)
	os.MkdirAll(Config.Root+"/db", 0755)

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

	PrintMsg("confirm", "Compiled Server!", 50, true)
}

func compTemplates() {
	if err := staticweb.Compile(Config.Root+"/pages", Config.Root+"/pages.dist"); err != nil {
		fmt.Println(err)
	}

	lastUpdate := time.Now().UnixMilli()

	fw := goutil.FileWatcher()
	fw.OnAny = func(path, op string) {
		if now := time.Now().UnixMilli(); now-lastUpdate > 1000 {
			lastUpdate = now
			if err := staticweb.Compile(Config.Root+"/pages", Config.Root+"/pages.dist"); err != nil {
				fmt.Println(err)
			}
		}
	}
	fw.WatchDir(Config.Root + "/pages")
}
