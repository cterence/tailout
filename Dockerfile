FROM golang:1.25.6@sha256:4c973c7cf9e94ad40236df4a9a762b44f6680560a7fb8a4e69513b5957df7217 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:fbc9554184984615e935febfc6e4fdf5e1e892f5ab0aa5587b96385aa1c0e779 AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:942190e51856de6eb2cb62660cb9a1d6a65625fd5d03dfaef7ee5f526d5f481f as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.25.6@sha256:4c973c7cf9e94ad40236df4a9a762b44f6680560a7fb8a4e69513b5957df7217 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:347a41e7f263ea7f7aba1735e5e5b1439d9e41a9f09179229f8c13ea98ae94cf AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
