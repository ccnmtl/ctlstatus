all: test

test: build
	goapp test

runserver: build
	dev_appserver.py app.yaml

