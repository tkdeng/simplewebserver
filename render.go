package server

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/tkdeng/goutil"
)

var dynamicRouteReg = regexp.MustCompile(`\[([\w_-]+)(\?|)\]`)

func getRoute(url string) (string, error) {
	if url == "/" || url == "" {
		url = "index"
	}

	dir := Config.Root + "/pages.dist"
	path, err := goutil.JoinPath(dir, url+".html")
	if err != nil {
		return "", err
	}

	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return strings.TrimSuffix(strings.TrimPrefix(path, Config.Root+"/pages.dist/"), ".html"), nil
	}
	return "", io.EOF
}
