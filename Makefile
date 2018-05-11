all: test

test:
	goapp test

runserver:
	dev_appserver.py app.yaml

deploy: test
	gcloud app deploy app.yaml index.yaml --project ctlstatus