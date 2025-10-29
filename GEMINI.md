# Gemini Interaction Guide

This document provides guidance for interacting with the JellyWolProxy project using Gemini.

## Testing

To run the tests, use the following command:

```bash
make test
```

## Linting

To run the linter, use the following command:

```bash
make lint
```

## Building

To build the project, use the following command:

```bash
make build
```

## Running

To run the proxy, use the following command:

```bash
./jellywolproxy [flags]
```

Available flags:

- `--config`: Path to the `config.json` file. (Default: `config.json`)
- `--port`: Port for the proxy to run on. (Default: `3881`)
- `--log-level`: Set the logging level (`Debug`, `Info`, `Warn`, `Error`). This overrides the `logLevel` in the config file.

## History

A history of actions taken by the Gemini assistant is stored in the `.gemini_history` file.
