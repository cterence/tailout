FROM golang:1.25.5@sha256:6396b3d8039d2050ab7a3c5c6e1cbeed8bf6d2ddc0403e1ab39d78749227ca19 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:ff67369f452b6f5f65ed4ab67b36ca39457c7c5214be2876c4302e38d9ec1ace AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:72a5cd981f3f5dfcba7264b28c0e963362d3d542e3462fea77ca5a5e98436697 as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.25.5@sha256:6396b3d8039d2050ab7a3c5c6e1cbeed8bf6d2ddc0403e1ab39d78749227ca19 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:f5a3067027c2b322cd71b844f3d84ad3deada45ceb8a30f301260a602455070e AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
