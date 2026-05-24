# Stage 1: build web resources
FROM node:20 as web-build
WORKDIR /src/webresources
COPY webresources/package.json webresources/package-lock.json* webresources/tsconfig.json webresources/rollup.config.js ./ 
COPY webresources ./webresources
RUN cd webresources && npm ci && npm run build

# Stage 2: build Go binary
FROM golang:1.25 as builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
COPY . .

# bring built webresources from web-build into the builder stage
COPY --from=web-build /src/webresources /src/webresources

# copy built static files into cmd/static (rollup output goes to cmd/static/js)
RUN cp -r webresources/node_modules/@azure/msal-browser/lib/msal-browser.min.js cmd/static/js/ 2>/dev/null || true
RUN cp -r webresources/node_modules/htmx.org/dist/htmx.min.js cmd/static/js/ 2>/dev/null || true

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