$ErrorActionPreference = "Stop"

$version = "0.17.2"
$dockerUser = "kozerenko"

$vueImage = "${dockerUser}/message-ui"
$goImage = "${dockerUser}/message-api"

$env:HUSKY = "0"
node ver.js

# 1. Vue.js: Контекст — корінь фронтенду, Dockerfile — у папці docker/
$vueContext    = ".\message-ui"
$vueDockerfile = ".\message-ui\docker\Dockerfile"

Write-Host "--- Building & Pushing Vue.js Image ---" -ForegroundColor Cyan
docker build -f $vueDockerfile -t "${vueImage}:${version}" -t "${vueImage}:latest" $vueContext

docker push "${vueImage}:${version}"
docker push "${vueImage}:latest"

# 2. Go: Контекст — корінь бекенду, Dockerfile — у папці docker/
$goContext    = ".\message-backend"
$goDockerfile = ".\message-backend\docker\Dockerfile"

Write-Host "--- Building & Pushing Go Image ---" -ForegroundColor Cyan
docker build -f $goDockerfile -t "${goImage}:${version}" -t "${goImage}:latest" $goContext

docker push "${goImage}:${version}"
docker push "${goImage}:latest"

Write-Host "--- Successfully pushed images to Docker Hub! ---" -ForegroundColor Green