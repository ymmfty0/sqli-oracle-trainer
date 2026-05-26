# Roadmap

## Status legend

- ✅ Done
- 🟡 In progress
- ⬜ Planned
- 🔁 Refactor later

---

## v0.1 — Boolean Oracle MVP

**Status:** ✅ Done

Implemented:

- ✅ Flask lab with vulnerable `/boolean?id=...`
- ✅ Go HTTP client
- ✅ URL validation
- ✅ HTTP timeout configuration
- ✅ `Observation` struct
- ✅ Body contains oracle
- ✅ Manual true/false payload checks
- ✅ Loop over true/false payloads
- ✅ Query parameters via `url.Values`

Current oracle marker:

```text
Product found
```

Current true/false examples:

```sql
1' AND 1=1-- -
1' AND 1=2-- -
```

---

## v0.2 — Boolean Blind Extraction

**Status:** 🟡 In progress

Implemented:

- ✅ Charset constants:
    - lowercase
    - uppercase
    - digits
    - symbols
    - default charset
- ✅ ASCII range constants:
    - `ASCIIMin = 32`
    - `ASCIIMax = 126`
- ✅ Character equality payload builder
- ✅ ASCII equality payload builder
- ✅ Charset-based extraction loop
- ✅ ASCII equality extraction loop
- ✅ Extraction result building with `strings.Builder`
- ✅ Stop extraction when no character is found
- ✅ Extraction functions return `(string, error)`

Current extraction methods:

```go
ExtractByCharset(maxLen, charset)
ExtractByASCII(maxLen, minCode, maxCode)
```

Needs cleanup:

- 🟡 Improve error messages in `SendPayload`
- 🟡 Rename unclear error contexts:
    - `send request` during request creation
    - `create new request` during `client.Do`
    - `read resp erro`
- 🟡 Move oracle marker into a constant
- 🟡 Remove hardcoded `maxLen = 21` from `main`
- 🟡 Add comments/docs for charset vs ASCII extraction

Planned:

- ⬜ CLI flags:
    - target URL
    - max length
    - charset mode
    - extraction mode
- ⬜ JSON output for extracted result

---

## v0.3 — Blind ANDing / ASCII Comparison

**Status:** ⬜ Planned

Goal:

Use comparison-based blind checks instead of direct equality checks.

Planned:

- ⬜ Payload builder for ASCII greater-than checks
- ⬜ Manual test for `unicode(substr(...)) > N`
- ⬜ Binary search for one character
- ⬜ Binary search for full string extraction
- ⬜ Compare request count between:
    - charset brute-force
    - ASCII equality
    - ASCII binary search

Expected benefit:

```text
Fewer requests during blind extraction.
```

---

## v0.4 — Time-based SQLi

**Status:** ⬜ Planned

Goal:

Implement time-based blind SQLi extraction using response delay as oracle.

Planned:

- ⬜ Add `/time?id=...` lab endpoint
- ⬜ Add time-based true/false payloads
- ⬜ Add timing oracle using `Observation.Elapsed`
- ⬜ Add threshold configuration
- ⬜ Add baseline timing measurement
- ⬜ Add time-based extraction loop
- ⬜ Add notes about jitter, false positives and retries

Oracle model:

```text
true  -> response time > threshold
false -> response time <= threshold
```

---

## v0.5 — Union-based SQLi Helper

**Status:** ⬜ Planned

Goal:

Implement helpers for visible-output SQLi exploitation.

Planned:

- ⬜ Add `/union?id=...` lab endpoint
- ⬜ Column count checks
- ⬜ `ORDER BY` helper
- ⬜ `UNION SELECT NULL,...` helper
- ⬜ Marker-based response detection
- ⬜ Basic data extraction via UNION

Focus:

```text
Not blind extraction, but direct visible output extraction.
```

---

## v0.6 — Error-based SQLi

**Status:** ⬜ Planned

Goal:

Detect SQL injection behavior using SQL errors and HTTP status codes.

Planned:

- ⬜ Add `/error?id=...` lab endpoint
- ⬜ SQL error detection
- ⬜ Status-code oracle
- ⬜ Body-based error oracle
- ⬜ Error marker detection:
    - `SQL error`
    - `syntax error`
    - `database error`

---

## v0.7 — Concurrency

**Status:** ⬜ Planned

Goal:

Speed up boolean extraction using goroutines.

Planned:

- ⬜ Goroutines for boolean charset extraction
- ⬜ Worker pool
- ⬜ Max concurrency setting
- ⬜ Context cancellation
- ⬜ Stop workers after first true match
- ⬜ Rate limiting
- ⬜ Compare sequential vs parallel extraction

Notes:

Concurrency should be implemented after the sequential extraction logic is stable.

---

## v0.8 — Project Structure and Interfaces

**Status:** ⬜ Planned

Goal:

Refactor the project into reusable packages and introduce interfaces naturally.

Planned:

- ⬜ Split project into packages
- ⬜ Move `Observation` into model package
- ⬜ Move HTTP client into client package
- ⬜ Add `Oracle` interface
- ⬜ Add `PayloadBuilder` interface
- ⬜ Add extractor package
- ⬜ Add output package
- ⬜ Add JSON report output

Potential structure:

```text
sqli-oracle-trainer/
├── cmd/
│   └── sqli-trainer/
├── internal/
│   ├── client/
│   ├── model/
│   ├── oracle/
│   ├── payload/
│   ├── extractor/
│   └── output/
├── lab/
├── docs/
└── README.md
```

---

## v0.9 — dbexec Transition

**Status:** ⬜ Planned

Goal:

Use the patterns learned in this project to prepare for `dbexec`.

Concepts to transfer:

- HTTP/client abstraction
- result models
- oracle interfaces
- payload builders
- extraction engine
- concurrency control
- clean project structure
- error handling
- reporting

Related future `dbexec` concept:

```go
type Provider interface {
	Name() string
	Check(ctx context.Context, target Target, cred Credential) Result
}
```

---

## Current Focus

Current version:

```text
v0.2 — Boolean Blind Extraction
```

Immediate tasks:

- 🟡 Clean up `SendPayload` error messages
- 🟡 Run charset extraction
- 🟡 Run ASCII extraction
- 🟡 Commit current progress
- ⬜ Start v0.3: ASCII greater-than payload builder