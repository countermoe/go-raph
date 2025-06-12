# go-raph

minimalist go dependency graph visualization

## screenshot

<img width="924" alt="Screenshot 2025-06-12 at 1 50 40 AM" src="https://github.com/user-attachments/assets/4c9f4119-4462-407f-b09e-946a733552aa" />

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
