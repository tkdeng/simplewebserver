package server

import (
	"os"
	"path/filepath"
	"strings"

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

type App struct {
	*fiber.App
}

// New loads a new server
func New(root string, config ...fiber.Config) (App, error) {
	// load config file
	loadConfig(root)

	// compile src
	compile()

	if len(config) == 0 {
		config = append(config, fiber.Config{
			AppName:                 Config.AppTitle,
			ServerHeader:            Config.Title,
			TrustedProxies:          Config.Proxies,
			EnableTrustedProxyCheck: true,
			EnableIPValidation:      true,
		})
	} else {
		config[0].AppName = Config.AppTitle
		config[0].ServerHeader = Config.Title

		if config[0].TrustedProxies == nil {
			config[0].TrustedProxies = Config.Proxies
		} else {
			config[0].TrustedProxies = append(config[0].TrustedProxies, Config.Proxies...)
		}
	}

	app := fiber.New(config[0])

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

	compressAssets := !Config.DebugMode
	app.Get("/theme/*", static.New(Config.Root+"/theme", static.Config{Compress: compressAssets}))
	// app.Get("/assets/wasm/*", static.New(Config.Root+"/wasm", static.Config{Compress: compressAssets}))
	app.Get("/assets/*", static.New(Config.Root+"/assets", static.Config{Compress: compressAssets}))
	if Config.PublicURI != "" {
		app.Get(Config.PublicURI, static.New(Config.Root+"/public", static.Config{Compress: compressAssets, Browse: true}))
	}

	app.Use("/404", func(c fiber.Ctx) error {
		url := goutil.Clean(c.Path())
		if url != "/404" {
			return c.Next()
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		c.SendStatus(404)

		if buf, err := os.ReadFile(Config.Root + "/pages.dist/404/index.html"); err == nil {
			return c.Send(buf)
		}

		return c.Send([]byte("<h1>Error 404</h1><h2>Page Not Found!</h2>"))
	})

	app.Use("/*", static.New(Config.Root+"/pages.dist", static.Config{Compress: compressAssets}))

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
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		c.SendStatus(404)

		if buf, err := os.ReadFile(Config.Root + "/pages.dist/404/index.html"); err == nil {
			return c.Send(buf)
		}

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
