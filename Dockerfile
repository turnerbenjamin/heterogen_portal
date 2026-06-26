# Stage 1: build web resources
FROM node:24 AS web-build
WORKDIR /src/webresources

COPY webresources/package.json webresources/package-lock.json* webresources/tsconfig.json webresources/rollup.config.js ./ 
COPY webresources ./
RUN npm ci && npm run build

# Stage 2: build Go binary
FROM golang:1.25 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
COPY . .

# Bring built webresources from web-build into the builder stage
COPY --from=web-build /src /src

# Copy built static files into cmd/static
RUN mkdir -p cmd/static/js && cp webresources/node_modules/htmx.org/dist/htmx.min.js cmd/static/js/

# Build the Go binary
WORKDIR /src/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/heterogen

# Stage 3: final image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /usr/local/bin/heterogen /usr/local/bin/heterogen

ENV SQLSERVER_DSN=""
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/heterogen"]