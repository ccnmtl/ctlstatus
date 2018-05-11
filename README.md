# ctlstatus

Google App Engine hosted status page.

## Installation

* Install Go
* Install the Go [Google App Engine](https://cloud.google.com/appengine/docs/standard/go/)
  * Add the `google-cloud-sdk/bin` to your PATH
  * Add the `/google-cloud-sdk/platform/google_appengine/goroot-1.9/bin/` to your PATH
* Checkout the project: https://github.com/ccnmtl/ctlstatus
* Download and install the dependencies
  * go get github.com/russross/blackfriday-tool
  * go get google.golang.org/appengine/aetest

## Run
`make runserver`

## Deploy
`make deploy`

Notes: gCloud will ask you to authenticate before the deploy can take place.
