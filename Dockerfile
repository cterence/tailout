FROM golang:1.25.1@sha256:a5e935dbd8bc3a5ea24388e376388c9a69b40628b6788a81658a801abbec8f2e as fetch-stage

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

FROM golang:1.25.1@sha256:a5e935dbd8bc3a5ea24388e376388c9a69b40628b6788a81658a801abbec8f2e AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/app

FROM gcr.io/distroless/base-debian12@sha256:d605e138bb398428779e5ab490a6bbeeabfd2551bd919578b1044718e5c30798 AS deploy-stage
WORKDIR /
COPY --from=build-stage /app/app /app
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app"]
