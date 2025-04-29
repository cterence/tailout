FROM golang:1.24.2@sha256:f52b85c1b58271ba213ac59878c4fe2fdbb41ac842c5c69f9e7a9d16d53c4ad8 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:e603e0e1a6caa989fbb60c829d84ab9e357c805334cba130e0abe323628fc53b AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:5e8c6edac88b6151ba1d213cb6f181046a2ff77e427ee9f580192a86bf3e94ae as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.24.2@sha256:f52b85c1b58271ba213ac59878c4fe2fdbb41ac842c5c69f9e7a9d16d53c4ad8 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:27769871031f67460f1545a52dfacead6d18a9f197db77110cfc649ca2a91f44 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
