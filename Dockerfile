FROM golang:1.15-alpine as builder
ARG BUNDLE_GITHUB__COM
RUN apk update && apk add --no-cache git make gcc alpine-sdk
RUN git config --global url."https://$BUNDLE_GITHUB__COM:x-oauth-basic@github.com/".insteadOf "https://github.com/"
COPY . /service_source
WORKDIR /service_source
RUN make app.dependencies.download
RUN make build.master

FROM alpine:3.11.6
ARG APP_NAME
ENV APP_NAME $APP_NAME
ENV SERVICE_NAME $APP_NAME
RUN apk update && apk add --no-cache tzdata ca-certificates bash
WORKDIR /service
COPY --from=builder /service_source/bin ./bin
CMD ["./bin/app"]
