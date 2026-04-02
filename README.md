# Server for exporting HTML to PDF

## Quick start

1. Clone this repository:

    ```bash
    git clone https://github.com/TOsmanov/go-pdf.git
    ```

2. Edit `docker-debug.yaml` configuration for your site:

    - Set the DOM selector. For example, if the main content on your site is located in `<div id="content">main content</div>`, set `#content`:

        ```yaml
        core:
            pdf:
                selector: "#content"
        ```

    - Set the logo URL (used on the title page):

        ```yaml
        core:
            pdf:
                service-pages:
                    logo-path: "https://example.com/logo.jpg"
        ```

3. Make the `start.sh` script executable and run it:

    ```bash
    chmod +rx start.sh
    ./start.sh
    ```

4. Open in a browser: http://127.0.0.1:10001/face

| URI           | Action                                             |
| ------------- | -------------------------------------------------- |
| `/face`       | Returns the face page (a form for sending requests) |
| `/pdf`        | Exports to PDF                                    |
| `/docx`       | Exports to DOCX                                   |
| `/reload`     | Restarts or stops the server                      |

## CLI

### Build

```bash
go clean --modcache && \
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -mod=readonly -o gopdf_cli ./cmd/main.go
```

## Server

### Build

```bash
go clean --modcache && \
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -mod=readonly -o gopdf_server ./server/main.go
```

Running the server without Docker:

```bash
./gopdf_server
```

### Docker

```bash
docker compose up -d --build server
```

#### Debug mode

To enable debug message output in Docker, use the `docker-debug.yaml` configuration file. Edit `docker-compose.yml`:

```diff
        environment:
-            CONFIG_PATH: "/app/config/prod.yaml"          # uncomment for production
+            CONFIG_PATH: "/app/config/docker-debug.yaml"  # and comment this
```

## Linting

### Run locally

```bash
golangci-lint run
```

### Run with Docker

```bash
docker run --rm -v $(pwd):/app/ -w /app/ golangci/golangci-lint:v1.58 golangci-lint --timeout=3m -v run
```

## Tests

### Run locally with file output

Use the `SAVE_FILES` environment variable to save test results as files:

```bash
SAVE_FILES=true go test ./...
```

### Run with Docker

Running tests:

```bash
docker compose run --rm --build tests go test ./...
```

Running tests with saved results:

```bash
docker compose run --rm --build -e SAVE_FILES=true tests go test -cover ./...
docker compose run --rm tests chmod -R 777 tests
```
