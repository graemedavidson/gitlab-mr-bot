# Build Image
FROM golang:alpine3.15 AS build-env
COPY *.go go.mod go.sum /src/
RUN cd /src && go build

# Run Image
FROM alpine
WORKDIR /app
COPY --from=build-env /src/gitlab-mr-webhook /app/
COPY ./templates/ /app/templates/
COPY ./static/ /app/static/
ENTRYPOINT ["./gitlab-mr-webhook"]
