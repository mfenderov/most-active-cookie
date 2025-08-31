# Most Active Cookie

[![CI](https://github.com/mfenderov/most-active-cookie/workflows/CI/badge.svg)](https://github.com/mfenderov/most-active-cookie/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mfenderov/most-active-cookie)](https://goreportcard.com/report/github.com/mfenderov/most-active-cookie)

Find the most active cookie(s) from CSV log files for a specific date.

## Install

**CLI (Go users):**
```bash
go install github.com/mfenderov/most-active-cookie/cmd/most-active-cookie@latest
```

**CLI (Binary):**
```bash
# Linux
curl -L -O https://github.com/mfenderov/most-active-cookie/releases/latest/download/most-active-cookie-linux-amd64
# macOS  
curl -L -O https://github.com/mfenderov/most-active-cookie/releases/latest/download/most-active-cookie-darwin-amd64
# Windows + other platforms: https://github.com/mfenderov/most-active-cookie/releases/latest
```

**Library:**
```bash
go get github.com/mfenderov/most-active-cookie
```

## Usage

**CLI:**
```bash
most-active-cookie -f cookie_log.csv -d 2018-12-09
```

**Library:**
```go
import "github.com/mfenderov/most-active-cookie"

cookies, err := cookie.FindMostActiveCookies("cookie_log.csv", "2018-12-09")
```

## Input Format

```csv
cookie,timestamp
AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00
SAZuXPGUrfbcn5UA,2018-12-09T10:13:00+00:00
```

Date format: `YYYY-MM-DD` (UTC timezone). Returns all cookies with maximum count, sorted alphabetically.