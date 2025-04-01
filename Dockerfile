FROM golang:1.23 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies to ensure go.sum is complete
RUN go mod tidy
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application with CGO disabled for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server

# Stage 2: Create a minimal runtime image using Google's distroless
FROM gcr.io/distroless/base-debian11

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /server

# Expose the port the application will run on
EXPOSE 8080

# 运行应用
CMD ["/server"]