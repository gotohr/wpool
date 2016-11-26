package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/franela/goreq"
	"github.com/gotohr/wpool"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

type ImgurImage struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`

	Html string
	Src  string

	Treshold int
}

type ImgurResponse struct {
	Data []ImgurImage `json:"data"`
}

func FetchUriContent(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	uri := w.(string)
	response, err := goreq.Request{
		Uri:     uri,
		Timeout: 5000 * time.Millisecond,
	}.Do()
	if err != nil {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"uri":        uri,
			"err":        err.Error(),
		}).Errorln("error fetching uri")
		return
	}
	//	defer response.Body.Close()
	destination.WorkQueue <- response.Body
}

func ExtractImgUri(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	body := w.(*goreq.Body)
	html, _ := body.ToString()
	body.Close()
	//	fmt.Println(html)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(html))
	if err != nil {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"err":        err.Error(),
		}).Errorln("error extracting image url")
		return
	}
	selector := "img.post-image-placeholder"
	src, found := doc.Find(selector).Attr("src")
	if found {
		destination.WorkQueue <- "http:" + src
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"src":        src,
		}).Info("extracted url")
	} else {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
		}).Errorln("src not found")
	}
}

func DownloadImage(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	u4, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	content, _ := ioutil.ReadAll(w.(io.Reader))
	w.(*goreq.Body).Close()
	fileName := "images/" + u4.String() + ".jpg"
	fileWriteError := ioutil.WriteFile(fileName, content, 0644)
	if fileWriteError != nil {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"fileName":   fileName,
			"err":        fileWriteError.Error(),
		}).Error("error storing image file")
		return
	}

	log.WithFields(log.Fields{
		"dispatcher": dispatcherName,
		"fileName":   fileName,
	}).Info("downloaded")

}
