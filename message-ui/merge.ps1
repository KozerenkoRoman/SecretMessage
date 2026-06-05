# Вхідні параметри
$SourceDir = Join-Path -Path (Get-Location) -ChildPath "src"
$OutputFile = "LoveLetter-ui.txt"

# Видаляємо попередній файл
if (Test-Path $OutputFile) {
    Remove-Item $OutputFile
}

Get-ChildItem -Path $SourceDir -Recurse -File |
    Where-Object {
        ($_.Extension -in ".js", ".css", ".scss", ".html", ".ts", ".go", ".mod", ".sum", ".sql", ".vue") -and
        ($_.Name -notlike "*.spec.ts") -and
		($_.Name -notlike "*_test.go")
    } |
    Sort-Object FullName |
    ForEach-Object {

        $filePath = $_.FullName
        $fileName = $_.Name
        $relativePath = [System.IO.Path]::GetRelativePath($SourceDir, $filePath)

        # Заголовок файлу
        Add-Content -Path $OutputFile -Value ("`n/* ===== FILE: $relativePath ===== */")

        # Читаємо файл як один рядок
        $content = Get-Content -Path $filePath -Raw

        # Видаляємо багаторядкові коментарі /* ... */
        $content = $content -replace "/\*[\s\S]*?\*/", ""

        # Видаляємо однорядкові коментарі // ...
        $content = $content -replace "(?m)//.*$", ""

        # Замінюємо таби на пробіли
        $content = $content -replace "`t", " "

        # Прибираємо переноси рядків
        $content = $content -replace "`r?`n", " "

        # Стискаємо всі пробіли підряд в один
        $content = $content -replace "\s{2,}", " "

        # Прибираємо пробіли навколо службових символів
        $content = $content -replace "\s*([{};:,])\s*", '$1'
        $content = $content.Trim()
        Add-Content -Path $OutputFile -Value $content
    }

Write-Host "Мініфіковано в: $OutputFile"