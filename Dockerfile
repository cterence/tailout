FROM golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:ff3af18c3c8254fde008b26aeb890efb00f90842b3e7723f2fd2791f9f7e5ecb AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:44cb0266577d20eb3003812c0d89bf9ad056f6dd48dbadb961e8da51f99bd99c as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:937c7eaaf6f3f2d38a1f8c4aeff326f0c56e4593ea152e9e8f74d976dde52f56 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
