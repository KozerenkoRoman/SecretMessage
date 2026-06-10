# Вхідні параметри
$ProjectRoot = Get-Location
$SourceDir = Join-Path -Path $ProjectRoot -ChildPath "src"
$OutputFile = "LoveLetter-ui.txt"

# Видаляємо попередній файл
if (Test-Path $OutputFile) {
    Remove-Item $OutputFile
}

$files =
    @(
        Get-ChildItem -Path $ProjectRoot -File
        Get-ChildItem -Path $SourceDir -Recurse -File
    ) |
    Where-Object {
        ($_.Extension -in ".js", ".css", ".scss", ".html", ".ts", ".go", ".mod", ".sum", ".sql", ".vue") -and
        ($_.Name -notlike "*.spec.ts") -and
        ($_.Name -notlike "*_test.go") -and
	    ($_.Name -notlike "package-lock.json")
		
    } |
    Sort-Object FullName

$files | ForEach-Object {

    $filePath = $_.FullName

    if ($filePath.StartsWith($SourceDir)) {
        $relativePath = [System.IO.Path]::GetRelativePath($SourceDir, $filePath)
        $relativePath = "src/$relativePath"
    }
    else {
        $relativePath = [System.IO.Path]::GetRelativePath($ProjectRoot, $filePath)
    }

    Add-Content -Path $OutputFile -Value ("`n/* ===== FILE: $relativePath ===== */")

    $content = Get-Content -Path $filePath -Raw

    $content = $content -replace "/\*[\s\S]*?\*/", ""
    $content = $content -replace "(?m)//.*$", ""
    $content = $content -replace "`t", " "
    $content = $content -replace "`r?`n", " "
    $content = $content -replace "\s{2,}", " "
    $content = $content -replace "\s*([{};:,])\s*", '$1'
    $content = $content.Trim()

    Add-Content -Path $OutputFile -Value $content
}

Write-Host "Мініфіковано в: $OutputFile"