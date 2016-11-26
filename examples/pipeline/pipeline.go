package main

import (
	"fmt"

	"github.com/gotohr/beanrpc"

	"github.com/franela/goreq"
	"github.com/gotohr/wpool"

	log "github.com/Sirupsen/logrus"
)

func main() {

	pl := wpool.NewPipeline([]wpool.PElement{
		wpool.PElement{"htmlFetcher", 1, FetchUriContent},
		wpool.PElement{"imgUriExtractor", 20, ExtractImgUri},
		wpool.PElement{"imgFetcher", 30, FetchUriContent},
		wpool.PElement{"imgDowloader", 10, DownloadImage},
	})

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
			fmt.Println(img.Link)
			pl.Dispatchers[0].WorkQueue <- img.Link
		}
		log.Println("process Params->", imgur_gallery_url)
	})

	go PushJobs(r)

	//blocking method!
	r.Run()
}

func PushJobs(r *beanrpc.BeanWorker) {
	//	time.Sleep(1 * time.Second)

	r.Put("process", "https://api.imgur.com/3/gallery/hot/viral/0.json", 1)
}
