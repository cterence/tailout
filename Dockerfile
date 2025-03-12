FROM golang:1.24.1@sha256:8678013a2add364dc3d5df2acc2b36893fbbd60ebafa5d5149bc22158512f021 as fetch-stage

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

FROM ghcr.io/a-h/templ:latest@sha256:570e4837721ae42dfb77320501fc6e7e4a48a713d09a9575334afcd9f2cca43a AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM cosmtrek/air@sha256:5e8c6edac88b6151ba1d213cb6f181046a2ff77e427ee9f580192a86bf3e94ae as development
COPY --from=generate-stage /ko-app/templ /bin/templ
COPY --chown=65532:65532 . /app
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["air"]

FROM golang:1.24.1@sha256:8678013a2add364dc3d5df2acc2b36893fbbd60ebafa5d5149bc22158512f021 AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:74ddbf52d93fafbdd21b399271b0b4aac1babf8fa98cab59e5692e01169a1348 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
