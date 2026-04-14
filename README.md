# driftwatch

A CLI tool that detects configuration drift between deployed services and their declared state in version control.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Point `driftwatch` at your services config and a running environment to detect drift:

```bash
# Check a single service against its declared state
driftwatch check --service api-gateway --source ./config/api-gateway.yaml

# Scan all services defined in a manifest
driftwatch scan --manifest ./services.yaml --env production

# Output results as JSON
driftwatch scan --manifest ./services.yaml --format json
```

Example output:

```
[DRIFT]  api-gateway     replicas: declared=3, actual=2
[DRIFT]  auth-service    image tag: declared=v1.4.2, actual=v1.4.1
[OK]     payment-service no drift detected
```

---

## Configuration

`driftwatch` reads from a `driftwatch.yaml` file in the working directory. You can override the config path with `--config`:

```bash
driftwatch scan --config /etc/driftwatch/config.yaml
```

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

This project is licensed under the [MIT License](LICENSE).