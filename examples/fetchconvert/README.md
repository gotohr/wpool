We have 3 Dispatchers, with 10 workers each.
```go
	fetcher := wpool.NewDispatcher("fetcher", 10)
	extractor := wpool.NewDispatcher("extractor", 10)
	downloader := wpool.NewDispatcher("downloader", 10)
```
