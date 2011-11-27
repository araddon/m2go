include $(GOROOT)/src/Make.inc

TARG=m2go
GOFMT=gofmt -spaces=true -tabindent=false -tabwidth=2

GOFILES=\
	m2go.go\
	tnet.go\
	signals.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w m2go_test.go
	${GOFMT} -w tnet_test.go
	${GOFMT} -w example.go
	${GOFMT} -w ${GOFILES}

