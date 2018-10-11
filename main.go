package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/skratchdot/open-golang/open"

	"github.com/mholt/archiver"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		panic("no file provided")
	}

	epubFile := args[0]
	port := "10086"
	if len(args) > 1 {
		port = args[1]
	}
	file, err := os.Open(epubFile)
	checkError(err)
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	checkError(err)

	sum := fmt.Sprintf("%x", sha256.Sum256(bytes))

	_, err = os.Stat(sum)
	if os.IsNotExist(err) {
		err = archiver.Zip.Open(epubFile, "bookshelf/"+sum)
		checkError(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("bookshelf", "bookshelf")

	box := packr.NewBox("./bib")
	box.Walk(func(path string, f packr.File) error {
		extName := filepath.Ext(path)
		mt := mime.TypeByExtension(extName)

		e.GET("/"+path, func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "max-age=3600")
			r := strings.NewReader(box.String(path))
			return c.Stream(http.StatusOK, mt, r)
		})
		return nil
	})
	e.GET("/i/", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "max-age=3600")
		r := box.String("i/index.html")
		return c.HTML(http.StatusOK, r)
	})

	// Start the Server
	addr := fmt.Sprintf("%s:%s", "localhost", port)
	e.HideBanner = true

	time.AfterFunc(2*time.Second, func() {
		open.Start(fmt.Sprintf("http://localhost:%s/i/?book=%s", port, sum))
	})

	e.Start(addr)

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
