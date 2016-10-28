package main

import (
	"bytes"
	"io/ioutil"

	"github.com/PuerkitoBio/goquery"
	"github.com/franela/goreq"
	"github.com/gotohr/wpool"

	log "github.com/Sirupsen/logrus"
)

type ImgurImage struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`

	Html string
	Src  string
}

type ImgurResponse struct {
	Data []ImgurImage `json:"data"`
}

func Fetch(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	img := w.(ImgurImage)

	response, err := goreq.Request{Uri: img.Link}.Do()
	if err != nil {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"url":        img.Link,
			"err":        err.Error(),
		}).Errorln("error fetching url")
		return
	}
	defer response.Body.Close()
	html, _ := response.Body.ToString()
	img.Html = html
	destination.WorkQueue <- img

	log.WithFields(log.Fields{
		"dispatcher": dispatcherName,
		"url":        img.Link,
	}).Info("fetched url")
}

func Extract(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	img := w.(ImgurImage)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(img.Html))
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
		img.Src = src
		destination.WorkQueue <- img
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

func Download(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
	img := w.(ImgurImage)
	response, err := goreq.Request{Uri: "http:" + img.Src}.Do()
	if err != nil {
		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"url":        img.Src,
			"err":        err.Error(),
		}).Errorln("error downloding url")
		return
	}

	defer response.Body.Close()

	content, _ := ioutil.ReadAll(response.Body)
	fileName := "images/" + img.Id + ".jpg"
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
