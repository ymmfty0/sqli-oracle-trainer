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

Current true/false payloads:

```sql
1' AND 1=1-- -
1' AND 1=2-- -
```

---

## v0.2 — Boolean Blind Extraction

**Status:** ✅ Done

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

Implemented extraction methods:

```go
ExtractByCharset(maxLen, charset)
ExtractByASCII(maxLen, minCode, maxCode)
```

Notes:

- Charset extraction checks characters directly.
- ASCII equality extraction checks numeric character codes.
- Both methods use the same body-based oracle.

---

## v0.3 — Binary Search Blind Extraction

**Status:** ✅ Done

Goal:

Use comparison-based blind checks instead of direct equality checks.

Implemented:

- ✅ ASCII greater-than payload builder
- ✅ Binary search extraction loop
- ✅ Full string extraction via binary search
- ✅ Per-character range reset with `low` / `high`
- ✅ Basic benchmark timing for binary extraction

Implemented method:

```go
ExtractViaBinarySearch(maxLen, minCode, maxCode)
```

Payload idea:

```sql
1' AND unicode(SUBSTR((SELECT value FROM secrets LIMIT 1),1,1))>100-- -
```

Notes:

- Binary search asks: `ASCII(position) > mid?`
- This reduces the number of requests compared to equality-based brute force.
- Current implementation still uses a fixed `maxLen`.

Needs improvement later:

- 🔁 Add length detection instead of hardcoded `maxLen`
- 🔁 Add cleaner output mode
- 🔁 Add request counter for benchmark comparison

---

## v0.4 — Bitwise ANDing Blind Extraction

**Status:** ✅ Done

Goal:

Extract characters by reconstructing their ASCII code from individual bits.

Implemented:

- ✅ Bitwise payload builder
- ✅ Bit masks:
  - `1`
  - `2`
  - `4`
  - `8`
  - `16`
  - `32`
  - `64`
- ✅ Bitwise extraction loop
- ✅ ASCII code reconstruction with bitwise OR
- ✅ Stop condition when reconstructed code is outside configured ASCII range
- ✅ Basic benchmark timing for bitwise extraction

Implemented method:

```go
ExtractViaBitwise(maxLen, minCode, maxCode)
```

Payload idea:

```sql
1' AND (unicode(SUBSTR((SELECT value FROM secrets LIMIT 1),1,1)) & 64)>0-- -
```

Notes:

- Bitwise extraction asks: `is this bit enabled?`
- For printable ASCII, 7 masks are enough for current lab usage.
- Current implementation uses `code |= mask` to reconstruct the character.

Needs improvement later:

- 🔁 Move masks into a constant or config
- 🔁 Add request counter
- 🔁 Add docs explaining bitwise extraction step by step

---

## v0.5 — Benchmark / Comparison Output

**Status:** 🟡 In progress

Implemented:

- ✅ Basic timing measurement for:
  - charset extraction
  - ASCII equality extraction
  - binary search extraction
  - bitwise extraction

Current output includes:

```text
Benchmark for charset extracting
Benchmark for ascii extracting
Benchmark for binary extracting
Benchmark for bitwise extracting
```

Needs cleanup:

- 🟡 Normalize benchmark output names
- 🟡 Add request count per extraction method
- 🟡 Add summary table
- 🟡 Move benchmark logic out of `main`
- 🟡 Add optional verbose output

---

## v0.6 — Code Cleanup

**Status:** 🟡 In progress

Needs cleanup:

- 🟡 Improve error messages in `SendPayload`
- 🟡 Rename unclear error contexts:
  - `send request` during request creation
  - `create new request` during `client.Do`
  - `read resp erro`
- 🟡 Move oracle marker into a constant
- 🟡 Remove hardcoded secret length from `main`
- 🟡 Add helper for repeated extraction execution
- 🟡 Reduce noisy output from extraction functions
- 🟡 Add CLI flags later:
  - target URL
  - max length
  - extraction mode
  - charset
  - timeout

Immediate cleanup task:

```text
Clean up SendPayload error messages and commit current extraction progress.
```

---

## v0.7 — Documentation

**Status:** ⬜ Planned

Goal:

Document the implemented blind SQLi extraction techniques.

Planned:

- ⬜ Add `docs/boolean-blind.md`
- ⬜ Document charset-based extraction
- ⬜ Document ASCII equality extraction
- ⬜ Document binary search extraction
- ⬜ Document bitwise ANDing extraction
- ⬜ Add request count comparison notes
- ⬜ Add examples of true/false oracle behavior

---

## v0.8 — Time-based SQLi

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

## v0.9 — Union-based SQLi Helper

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

## v0.10 — Error-based SQLi

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

## v0.11 — Concurrency

**Status:** ⬜ Planned

Goal:

Speed up blind extraction using goroutines.

Planned:

- ⬜ Goroutines for boolean charset extraction
- ⬜ Worker pool
- ⬜ Max concurrency setting
- ⬜ Context cancellation
- ⬜ Stop workers after first true match
- ⬜ Rate limiting
- ⬜ Compare sequential vs parallel extraction

Notes:

Concurrency should be implemented only after sequential extraction logic is stable and cleaned up.

---

## v0.12 — Project Structure and Interfaces

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

## Current Focus

Current version:

```text
v0.6 — Code Cleanup
```

Immediate tasks:

- 🟡 Clean up `SendPayload` error messages
- 🟡 Run all extraction methods again
- 🟡 Verify benchmark output
- 🟡 Commit current extraction progress
- ⬜ Add docs for:
  - charset extraction
  - ASCII equality
  - binary search
  - bitwise ANDing