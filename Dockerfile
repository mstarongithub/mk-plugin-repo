FROM golang:1.22.1 as buildstage-go

WORKDIR /app

# Add go.sum here once non-standard-lib packages are included
COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app

CMD ["app"]
