# go-fbvideo
A small tool to upload video to a facebook resource (user, page, group, event)

### Example:
```
package main

import (
	"log"

	"github.com/canhlinh/go-fbvideo"
)

func main() {
	uploadSession := fbvideo.NewUploadSession("../testdata/video.mp4", 1849649771774166, "EAADGAjjQ7sgBANGPH6KB2fsEZAf7EjCXJR0uoaVxQr25ODc4ZAeyrAoehV3ZAEOz9mrOKD5SxyJbVK7OwgsLeFhoM24c638f91zgNEnQ3ApGnN3OPe3t6S0ZAX3NwlTd3OC0sTZBKbUiw0TOz76l2jbEXwrZArb5qbDwUDfP0ZAsH1fxSv4BE4BRWZCPEzYgvzv8DRfGJgRts2tievuZATy6J")
	if err := uploadSession.Upload(); err != nil {
		log.Fatal(err)
	}
}

```