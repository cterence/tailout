FROM golang:1.25.1@sha256:ab1f5c47de0f2693ed97c46a646bde2e4f380e40c173454d00352940a379af60 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:70a8ea3b71b2bc4522fe3eb67f7dcaa0f2bb95b7f0bd087917bbb0410694ca7a AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:3def16c884b2b9409a4c7bb987cd76d656a12d0e0436fe8c583df93014e0356c as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.25.1@sha256:ab1f5c47de0f2693ed97c46a646bde2e4f380e40c173454d00352940a379af60 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
