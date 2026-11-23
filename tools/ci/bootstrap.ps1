$projectRoot = Join-Path $PSScriptRoot '../..'
docker compose --project-directory $projectRoot -f (Join-Path $projectRoot 'compose.yaml') run --build --rm bootstrap @args
exit $LASTEXITCODE
