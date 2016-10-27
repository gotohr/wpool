package main

import (
	"time"

	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/gotohr/beanrpc"

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

	fetch := func(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
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

	extract := func(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
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

	download := func(w wpool.Work, dispatcherName string, destination *wpool.Dispatcher) {
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

	fetcher.Start(fetch, &extractor)
	extractor.Start(extract, &downloader)
	downloader.Start(download, nil)

	r := beanrpc.New("localhost:11300")

	// opens tube for procesing
	if err := r.Open("mytube"); err != nil {
		log.Println(err)
		return
	}

	// register method
	r.On("process", func(c *beanrpc.Context) {

		log.Println("Buffered output->", string(c.Buff()))

		log.Println("Job id->", c.Id())

		//bind your type
		var imgur_gallery_url string

		if err := c.Bind(&imgur_gallery_url); err != nil {
			log.Println(err)
		}
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
		log.Println("process Params->", imgur_gallery_url)
	})

	go PushJobs(r)

	//blocking method!
	r.Run()
}

func PushJobs(r *beanrpc.BeanWorker) {
	time.Sleep(1 * time.Second)

	r.Put("process", "https://api.imgur.com/3/gallery/hot/viral/0.json", 1)
}
