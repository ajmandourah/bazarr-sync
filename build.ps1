$Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build -o .\build\ .\cmd\bazarr-sync\
$Env:GOOS = "windows"; $Env:GOARCH = "amd64"; go build -o .\build\ .\cmd\bazarr-sync\
