os_family := if os_family() == "unix" { "linux" } else { os_family() }
dev := env("DEV", "false")

build OS = os_family:
    DEV={{dev}} GOOS={{OS}} wails3 build{{ if OS == "linux" { " -tags gtk3" } else { "" } }}

dev:
    DEV=true wails3 dev

lint:
    cd frontend && watchexec -e svelte,js,ts,css,json biome lint

format:
    go fmt
    cd frontend && pnpm biome check --write --linter-enabled=false
