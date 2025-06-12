# go-raph

minimalist go dependency graph visualization

## usage

```bash
# analyze current directory
go run main.go

# analyze specific directory
go run main.go /path/to/project

# or with flag
go run main.go -path /path/to/project

# custom port
go run main.go -port 3000
```

## node types

- red: main module
- blue: packages 
- yellow: external imports
