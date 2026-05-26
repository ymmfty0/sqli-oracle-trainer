# Roadmap

## v0.1 — Boolean oracle MVP
- Flask lab with vulnerable `/boolean?id=...`
- Go HTTP client
- Observation struct
- Body contains oracle
- Manual true/false payload checks

## v0.2 — Boolean blind extraction
- Payload builder for character checks
- Single-character extraction
- Full string extraction loop
- Charset configuration

## v0.3 — Time-based SQLi
- `/time?id=...` lab endpoint
- Timing oracle
- Threshold configuration
- Baseline timing notes

## v0.4 — Union-based SQLi helper
- Column count checks
- Marker-based response detection
- Basic UNION payload generation

## v0.5 — Error-based SQLi
- SQL error detection
- Status/body based error oracle

## v0.6 — Concurrency
- Goroutines for boolean extraction
- Worker pool
- Context cancellation
- Max concurrency

## v0.7 — Project structure and interfaces
- Oracle interface
- PayloadBuilder interface
- Extractor package
- Client package