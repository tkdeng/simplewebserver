package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tkdeng/simplewebserver/render"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/tkdeng/goutil"
)

type ConfigData struct {
	Title    string
	AppTitle string
	Desc     string

	PublicURI string

	Origins []string
	Proxies []string

	OriginErrHandler func(c fiber.Ctx, err error) error

	PortHTTP uint16
	PortSSL  uint16

	DebugMode bool

	Root string
}

var Config = ConfigData{
	Title:    "Web Server",
	AppTitle: "WebServer",
	Desc:     "A Web Server.",

	PortHTTP: 8080,
	PortSSL:  8443,
}

var Engine *render.Engine

type App struct {
	*fiber.App
}

// New loads a new server
func New(root string) (App, error) {
	// load config file
	loadConfig(root)

	// compile src
	compile()

	var err error
	render.SetVarChar("{{", "}}")
	Engine, err = render.New(Config.Root + "/pages")
	if err != nil {
		return App{}, err
	}

	app := fiber.New(fiber.Config{
		Views:                   Engine,
		PassLocalsToViews:       true,
		AppName:                 Config.AppTitle,
		ServerHeader:            Config.Title,
		TrustedProxies:          Config.Proxies,
		EnableTrustedProxyCheck: true,
		EnableIPValidation:      true,
	})

	compressAssets := !Config.DebugMode
	app.Get("/theme/*", static.New(Config.Root+"/theme", static.Config{Compress: compressAssets}))
	// app.Get("/assets/wasm/*", static.New(Config.Root+"/wasm", static.Config{Compress: compressAssets}))
	app.Get("/assets/*", static.New(Config.Root+"/assets", static.Config{Compress: compressAssets}))
	if Config.PublicURI != "" {
		app.Get(Config.PublicURI, static.New(Config.Root+"/public", static.Config{Compress: compressAssets, Browse: true}))
	}

	if Config.OriginErrHandler == nil {
		Config.OriginErrHandler = func(c fiber.Ctx, err error) error {
			c.SendStatus(403)
			return c.SendString(err.Error())
		}
	}

	// enforce specific domain and ip origins
	app.Use(VerifyOrigin(Config.Origins, Config.Proxies, Config.OriginErrHandler))

	// auto redirect http to https
	if Config.PortSSL != 0 {
		app.Use(RedirectSSL(Config.PortHTTP, Config.PortSSL))
	}

	return App{app}, nil
}

// Listen to both http and https ports and
// auto generate a self signed ssl certificate
// (will also auto renew every year)
//
// by using self signed certs, you can use a proxy like cloudflare and
// not have to worry about verifying a certificate athority like lets encrypt
func (app *App) Listen() error {
	app.Use(func(c fiber.Ctx) error {
		url := goutil.Clean(c.Path())

		if url == "/" || url == "" {
			url = "index"
		}
		url = strings.Trim(url, "/")

		if path, err := goutil.JoinPath(Config.Root, "pages.dist", url+".html"); err == nil {
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)

				if i, err := strconv.Atoi(url); err == nil && i >= 100 && i <= 599 {
					c.SendStatus(i)
				} else {
					c.SendStatus(200)
				}

				return c.Render(url, fiber.Map{})
			}
		}

		if path, err := goutil.JoinPath(Config.Root, "pages.dist", "404.html"); err == nil {
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
				c.SendStatus(404)
				return c.Render("404", fiber.Map{})
			}
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		c.SendStatus(404)
		return c.Send([]byte("<h1>Error 404</h1><h2>Page Not Found!</h2>"))
	})

	return ListenAutoTLS(app.App, Config.PortHTTP, Config.PortSSL, Config.Root+"/db/ssl/auto_ssl")
}

func loadConfig(root string) {
	// load config file
	if path, err := filepath.Abs(root); err == nil {
		root = path
	}
	root = strings.TrimSuffix(root, "/")

	goutil.ReadConfig(root+"/config.yml", &Config)
	Config.Root = root
}
