FROM golang:1.22
WORKDIR /app

# Pre-copy go.mod to cache dependencies
COPY go.mod ./
RUN go mod download

# Copy the rest
COPY . .

# Build
RUN go build -o stackguard

EXPOSE 8080
CMD ["./stackguard"]
