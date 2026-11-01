# BlackShield Security Audit Report — `vertcoin-project/one-click-miner-vnext`

- **Scanned Commit:** `b353a669258c4c15da55023fde3cc7f1f60717e8`
- **Actionable Findings:** `13` (`3 Critical`, `10 High`, `4 Medium`, `0 Low`)
- **BlackShield Platform & Continuous Zero-Cost PR Protection:** [https://blackshield.chaplau.com](https://blackshield.chaplau.com)

## Detailed Findings & Remediation Guide

### 1. [CRITICAL] `CVE-2021-44906` — CVE-2021-44906 in minimist@0.0.8

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `minimist@0.0.8`
- **Details:** Prototype pollution in minimist CLI argument parser.
- **Recommended Fix:** Upgrade `minimist` to `1.2.6` or newer

### 2. [CRITICAL] `CVE-2022-37601` — CVE-2022-37601 in loader-utils@0.2.17

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `loader-utils@0.2.17`
- **Details:** Prototype pollution in loader-utils via interpolateName method.
- **Recommended Fix:** Upgrade `loader-utils` to `2.0.4` or newer

### 3. [CRITICAL] `CVE-2023-26136` — CVE-2023-26136 in tough-cookie@2.5.0

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `tough-cookie@2.5.0`
- **Details:** Prototype pollution in tough-cookie CookieJar domain storage when rejectPublicSuffixes=false.
- **Recommended Fix:** Upgrade `tough-cookie` to `4.1.3` or newer

### 4. [HIGH] `CVE-2020-28469` — CVE-2020-28469 in glob-parent@3.1.0

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `glob-parent@3.1.0`
- **Details:** Regular Expression Denial of Service (ReDoS) in glob-parent enclosure matching.
- **Recommended Fix:** Upgrade `glob-parent` to `5.1.2` or newer

### 5. [HIGH] `CVE-2021-37701` — CVE-2021-37701 in tar@4.4.19

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `tar@4.4.19`
- **Details:** Arbitrary file creation and overwrite via insufficient symlink protection in node-tar.
- **Recommended Fix:** Upgrade `tar` to `6.1.9` or newer

### 6. [HIGH] `CVE-2022-25883` — CVE-2022-25883 in semver@5.7.1

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `semver@5.7.1`
- **Details:** Regular Expression Denial of Service (ReDoS) in semver range parser.
- **Recommended Fix:** Upgrade `semver` to `7.5.2` or newer

### 7. [HIGH] `CVE-2022-3517` — CVE-2022-3517 in minimatch@3.0.4

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `minimatch@3.0.4`
- **Details:** Regular Expression Denial of Service (ReDoS) in minimatch brace expansion.
- **Recommended Fix:** Upgrade `minimatch` to `3.0.5` or newer

### 8. [HIGH] `CVE-2024-21538` — CVE-2024-21538 in cross-spawn@7.0.3

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `cross-spawn@7.0.3`
- **Details:** Regular Expression Denial of Service (ReDoS) in cross-spawn argument escaping.
- **Recommended Fix:** Upgrade `cross-spawn` to `7.0.5` or newer

### 9. [HIGH] `CVE-2024-4068` — CVE-2024-4068 in braces@3.0.2

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `braces@3.0.2`
- **Details:** Uncontrolled resource consumption in braces imbalanced brace expansion.
- **Recommended Fix:** Upgrade `braces` to `3.0.3` or newer

### 10. [HIGH] `CVE-2024-43796` — CVE-2024-43796 in express@4.18.2

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `express@4.18.2`
- **Details:** Open redirect and route tampering via path-to-regexp vulnerability in Express.
- **Recommended Fix:** Upgrade `express` to `4.20.0` or newer

### 11. [HIGH] `CVE-2024-45296` — CVE-2024-45296 in path-to-regexp@0.1.7

- **Category:** `vulnerability`
- **Location:** `frontend/package-lock.json`
- **Affected Package:** `path-to-regexp@0.1.7`
- **Details:** Backtracking regular expression in path-to-regexp causes Denial of Service.
- **Recommended Fix:** Upgrade `path-to-regexp` to `0.1.10` or newer

### 12. [HIGH] `CVE-2023-45288` — CVE-2023-45288 in golang.org/x/net@v0.21.0

- **Category:** `vulnerability`
- **Location:** `go.mod`
- **Affected Package:** `golang.org/x/net@v0.21.0`
- **Details:** HTTP/2 CONTINUATION frame flood causes resource exhaustion in golang.org/x/net/http2.
- **Recommended Fix:** Upgrade `golang.org/x/net` to `0.23.0` or newer

### 13. [HIGH] `go-cmdi-exec-combined` — go-cmdi-exec-combined: Command Execution via Dynamic exec.Command

- **Category:** `sast`
- **Location:** `miners/miners.go:240`
- **Details:** Go application security vulnerability or insecure coding pattern. at miners/miners.go:240
- **Recommended Fix:** Invoke the target binary directly with exec.Command(bin, args...) without sh -c shell evaluation.

