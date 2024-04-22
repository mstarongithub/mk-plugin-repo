FROM golang:1.22.1 as buildstage-go

WORKDIR /app

# Add go.sum here once non-standard-lib packages are included
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/go-cache GOCACHE=/go-cache CGO_ENABLED=1 GOOS=linux go build -o /server && touch db.sqlite

FROM gcr.io/distroless/base-debian12 AS release-stage

WORKDIR /

COPY --from=buildstage-go /server /server
COPY --from=buildstage-go /app/db.sqlite /db.sqlite

EXPOSE 8080

CMD [ "/server" ]
