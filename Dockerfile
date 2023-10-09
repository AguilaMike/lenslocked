FROM golang:1.21-alpine AS base
WORKDIR /app

FROM base AS dev
RUN go install github.com/cosmtrek/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest
WORKDIR /app
EXPOSE $PORT_APP $PORT_DELVE
CMD ["air"]

# ### Executable builder
# FROM base AS builder
# WORKDIR /app

# # Application dependencies
# COPY . /app
# RUN go mod download \
#     && go mod verify

# RUN go build -o my-great-program -a .

# ### Production
# FROM alpine:latest

# RUN apk update \
#     && apk add --no-cache \
#     ca-certificates \
#     curl \
#     tzdata \
#     && update-ca-certificates

# # Copy executable
# COPY --from=builder /app/my-great-program /usr/local/bin/my-great-program
# EXPOSE 8080

# ENTRYPOINT ["/usr/local/bin/my-great-program"]



# # # Set the current working directory inside the container
# # WORKDIR /app

# # # Copy go.mod and go.sum files to the workspace
# # COPY go.mod go.sum ./

# # # Download all dependencies
# # RUN go mod download

# # # Copy the source from the current directory to the workspace
# # COPY . .


# # # # Installing Delve
# # # RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
# # # ENV GO111MODULE=off
# # # # Build the Go app
# # # # RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -o main .
# # RUN go build -o main .

# # # Expose ports
# # EXPOSE 8080 4000

# # # Command to run the executable
# # CMD ["air"]



# # # #build stage
# # # FROM golang:alpine AS builder
# # # RUN apk add --no-cache git
# # # WORKDIR /go/src/app
# # # COPY . .
# # # RUN go get -d -v ./...
# # # RUN go build -o /go/bin/app -v ./...

# # # #final stage
# # # FROM alpine:latest
# # # RUN apk --no-cache add ca-certificates
# # # COPY --from=builder /go/bin/app /app
# # # ENTRYPOINT /app
# # # LABEL Name=lenslocked Version=0.0.1
# # # EXPOSE 4000
