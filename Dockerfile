# ---- Svelte builder
# First build the frontend into static files
# The Go app requires those as it embeds them into the final executable
FROM node:21 as buildstage-svelte

WORKDIR /app

COPY frontend/package*.json .
RUN npm install

COPY frontend .
RUN npm run build
RUN ls -a

# ---- Go builder
# Then build the Go app
FROM golang:1.22.1 as buildstage-go

WORKDIR /app

# Add go.sum here once non-standard-lib packages are included
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
COPY --from=buildstage-svelte /app/build ./frontend/build
RUN --mount=type=cache,target=/go-cache GOCACHE=/go-cache CGO_ENABLED=1 GOOS=linux go build -o /server

# ---- Final slim container
FROM gcr.io/distroless/base-debian12 AS release-stage
WORKDIR /
ARG log_level
ENV env_log_level $log_level
COPY --from=buildstage-go /server /server
EXPOSE 8080
CMD [ "/server" ]
