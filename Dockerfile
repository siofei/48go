#build stage
FROM golang

WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN mkdir /go/bin/app && go build -o /go/bin/app/48go 48go.go
RUN mkdir /48LiveGo
COPY config.yaml /48LiveGo
RUN apt update && apt-get -y install ffmpeg
CMD ["/go/bin/app/48go"]

#final stage
# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
# COPY --from=builder /go/bin/app /app
# ENTRYPOINT /app
# LABEL Name=48go Version=0.0.1