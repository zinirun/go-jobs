package main

import (
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/zinirun/go-jobs/scrapper"
)

const fileName string = "jobs.csv"
const downloadFileName string = "go-jobs.csv"

func handleHome(c echo.Context) error {
	return c.File("public/index.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, downloadFileName)
}
func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
