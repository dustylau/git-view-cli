$projectRoot = Join-Path $PSScriptRoot '../..'
$compose = @('compose', '--project-directory', $projectRoot, '-f', (Join-Path $projectRoot 'compose.yaml'))
docker @compose up -d --no-recreate runner
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
docker @compose run --no-deps --build --rm bootstrap @args
exit $LASTEXITCODE
