FROM golang:1.25.4@sha256:3976069ea4797af8ebc1b78d56e0c6b42e85ec31703c83e0a1ffa34cb127a248 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:ff67369f452b6f5f65ed4ab67b36ca39457c7c5214be2876c4302e38d9ec1ace AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:55a323d0f627b69307347f578b856ab30c812cc40603a4288688852964594c27 as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.25.4@sha256:3976069ea4797af8ebc1b78d56e0c6b42e85ec31703c83e0a1ffa34cb127a248 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
