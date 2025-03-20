os_family := if os_family() == "unix" { "linux" } else { os_family() }

build OS = os_family:
    PRODUCTION=true wails3 task build:{{OS}}

dev:
    wails3 dev

lint:
    cd frontend && watchexec -e svelte,js,ts,css,json biome lint

format:
    go fmt
    cd frontend && biome format --fix
