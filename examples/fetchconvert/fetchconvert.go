package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

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

func main() {
	fetcher := wpool.NewDispatcher("fetcher", 10)
	extractor := wpool.NewDispatcher("extractor", 10)
	downloader := wpool.NewDispatcher("downloader", 10)

	fetch := func(w wpool.Work, dispatcherName string) {
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
		extractor.WorkQueue <- img

		log.WithFields(log.Fields{
			"dispatcher": dispatcherName,
			"url":        img.Link,
		}).Info("fetched url")
	}

	extract := func(w wpool.Work, dispatcherName string) {
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
			downloader.WorkQueue <- img
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

	download := func(w wpool.Work, dispatcherName string) {
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

	fetcher.Start(fetch)
	extractor.Start(extract)
	downloader.Start(download)

	imgur_gallery_url := "https://api.imgur.com/3/gallery/hot/viral/0.json"
	response, err := goreq.Request{Uri: imgur_gallery_url}.Do()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var collection ImgurResponse
	response.Body.FromJsonTo(&collection)
	response.Body.Close()

	for _, img := range collection.Data {
		fetcher.WorkQueue <- img
	}

	//	fmt.Println(collection)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
