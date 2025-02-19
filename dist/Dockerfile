# syntax=docker/dockerfile:1

FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY dist.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /blogserver-dist

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /blogserver-dist /blogserver-dist

EXPOSE 8080

ENTRYPOINT ["/blogserver-dist"]