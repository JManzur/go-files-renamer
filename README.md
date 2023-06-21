# Go Files Renamer

A Go script to rename files to lowercase

## Building the binary

```bash
go build -o go-files-renamer main.go
```

## Usage:

**Basic usage:**

```bash
./go-files-renamer -path /path/to/files
```

**Options:**

Verbose mode:
```bash
./go-files-renamer -path /path/to/files -v
```

Limit the number of simultaneous goroutines:
```bash
./go-files-renamer -path /path/to/files -maxGoroutines 100
```

Send logs to a different path:
```bash
./go-files-renamer -path /path/to/files -logFilePath /path/to/logs
```

