# Antigravity Quota Tool

> [!IMPORTANT]
> This tool is 100% AI built.

`antigravity-quota` is a lightweight Go command line interface (CLI) tool designed to query the running Antigravity language server for your model usage quota and print a clean, human-readable summary.

## Features

- **Automatic environment detection**: Automatically reads the address and CSRF token when run inside active Antigravity workspace terminal sessions.
- **Explicit override support**: Override connections using command line flags (`--address` and `--token`).
- **Raw JSON output**: Output the raw JSON response using the `--json` flag.

---

## Installation

To build and install the tool to your `GOBIN` directory (e.g., `~/go/bin`), run the following command from the project root:

```bash
make install
```

Make sure your `GOBIN` directory is added to your shell's `PATH` environment variable.

---

## Usage

Simply run:
```bash
antigravity-quota
```

### Options

- `-address string`: Specify a custom Antigravity language server address (e.g., `localhost:41475`).
- `-token string`: Specify a custom Antigravity CSRF token.
- `-json`: Outputs the raw JSON response instead of the formatted summary.

Example with overrides:
```bash
antigravity-quota -address localhost:41475 -token <csrf-token>
```

---

## Uninstalling

To uninstall the binary from your local bin directory, run:

```bash
make uninstall
```
