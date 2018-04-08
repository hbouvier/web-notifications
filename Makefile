GOCC=go
PROJECTNAME=web-push-notification
VERSION=$(shell cat VERSION.txt)

all: deps gen run

clean:
	rm storage/fungen_auto.go
run:
	${GOCC} run ${PROJECTNAME}.go

deps:
	${GOCC} get github.com/sherclockholmes/webpush-go github.com/kulshekhar/fungen 

gen:
	${GOCC} generate storage/registration.go