### docker build -t dialogflow-mattermost:latest .
### docker run --rm -it dialogflow-mattermost:latest

FROM golang:latest as builder

COPY . /go/src/dialogflow-mattermost/
WORKDIR /go/src/dialogflow-mattermost/
RUN go get
RUN CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s"

FROM alpine:3.5
RUN apk update && apk add ca-certificates
COPY --from=builder /go/src/dialogflow-mattermost/dialogflow-mattermost /dialogflow-mattermost
WORKDIR /
ENTRYPOINT ["/dialogflow-mattermost"]
CMD ["--help"]