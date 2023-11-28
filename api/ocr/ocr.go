package ocr

import (
	"net/http"
	"os"
	"strings"

	helper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/gin-gonic/gin"
	"github.com/otiai10/gosseract/v2"
)

func OcrHandler(c *gin.Context) {
	var (
		url  = c.Query("url")
		lang = c.DefaultQuery("lang", "eng")
		text string
	)

	fileName := helper.DownloadFile(url)
	client := gosseract.NewClient()
	defer client.Close()

	if url != "" {
		client.SetLanguage(lang)

		client.SetImage(fileName)
		text, _ = client.Text()
	} else {
		langs, _ := gosseract.GetAvailableLanguages()
		text = strings.Join(langs, "\n")
	}

	os.Remove(fileName)
	c.String(http.StatusOK, text)
}
