# builder stage
FROM golang:1.17.5 as builder
ARG DATE
ARG COMMIT
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN echo "building k8s lab prometheus server on commit: ${COMMIT} date: ${DATE}"
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/marcosQuesada/prometheus-operator/pkg/config.Commit=${COMMIT} -X github.com/marcosQuesada/prometheus-operator/pkg/config.Date=${DATE}" .

# final stage
FROM alpine:3.11.5
COPY --from=builder /app/* /app/
