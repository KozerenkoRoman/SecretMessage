$ErrorActionPreference = "Stop"

$dockerUser = "kozerenko"
$vueImage   = "${dockerUser}/message-ui"
$goImage    = "${dockerUser}/message-api"

$env:HUSKY = "0"
node ver.js

$version = (Get-Content -Raw -Path ".\message-ui\package.json" | ConvertFrom-Json).version

$vueContext    = ".\message-ui"
$vueDockerfile = ".\message-ui\docker\Dockerfile"

Write-Host "--- Building & Pushing Vue.js Image (v$version) ---" -ForegroundColor Cyan
docker build -f $vueDockerfile -t "${vueImage}:${version}" -t "${vueImage}:latest" $vueContext

docker push "${vueImage}:${version}"
docker push "${vueImage}:latest"

$goContext    = ".\message-backend"
$goDockerfile = ".\message-backend\docker\Dockerfile"

Write-Host "--- Building & Pushing Go Image (v$version) ---" -ForegroundColor Cyan
docker build -f $goDockerfile -t "${goImage}:${version}" -t "${goImage}:latest" $goContext

docker push "${goImage}:${version}"
docker push "${goImage}:latest"

Write-Host "--- Successfully pushed images to Docker Hub! ---" -ForegroundColor Green