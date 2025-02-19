# syntax=docker/dockerfile:1

FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY versioning.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /blogserver-version

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /blogserver-version /blogserver-version

COPY versions ./versions

EXPOSE 8080

ENTRYPOINT ["/blogserver-version"]