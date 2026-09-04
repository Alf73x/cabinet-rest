$ErrorActionPreference = "Stop"
try {

$LocalDb   = "C:\Cabinet\CabinetREST_DeployDB\DB\cabinet.db"
$Vps       = "cabinet-vps"
$RemoteDir = "/root/Programs/CabinetREST"

Write-Host "Uploading database..."

scp $LocalDb "${Vps}:${RemoteDir}/storage/cabinet.db.new"

if ($LASTEXITCODE -ne 0) {
    throw "Database upload failed"
}

Write-Host "Database uploaded."
Write-Host "Installing database on VPS..."

$RemoteScript = @'
set -e

cd /root/Programs/CabinetREST

if [ ! -s storage/cabinet.db.new ]; then
    echo "ERROR: cabinet.db.new does not exist or is empty"
    exit 1
fi

mkdir -p storage/backups

BACKUP="storage/backups/cabinet-$(date +%Y%m%d-%H%M%S).db"

echo "Stopping REST..."

docker compose --env-file .deploy.env stop cabinet-rest

echo "Creating backup..."

cp storage/cabinet.db "$BACKUP"

# Старые WAL/SHM нельзя оставлять рядом с новой SQLite DB.
rm -f storage/cabinet.db-wal
rm -f storage/cabinet.db-shm

echo "Installing new database..."

mv storage/cabinet.db.new storage/cabinet.db

echo "Starting REST..."

docker compose --env-file .deploy.env up -d

echo "Waiting for REST..."

for i in $(seq 1 15); do
    if curl -fsS http://127.0.0.1:8082/health > /dev/null; then
        echo "REST health check OK"
        echo "Database deployment successful"
        echo "Backup: $BACKUP"
        exit 0
    fi

    echo "REST is not ready yet... ($i/15)"
    sleep 2
done

echo "Health check FAILED"
echo "Rolling database back..."

docker compose --env-file .deploy.env stop cabinet-rest

cp "$BACKUP" storage/cabinet.db

rm -f storage/cabinet.db-wal
rm -f storage/cabinet.db-shm

docker compose --env-file .deploy.env up -d

echo "Previous database restored."

docker logs --tail 100 cabinet-rest || true

exit 1
'@

# Преобразуем Windows CRLF в Unix LF перед отправкой в bash.
$RemoteScript = $RemoteScript -replace "`r", ""

$RemoteScript | ssh $Vps "bash -s"

if ($LASTEXITCODE -ne 0) {
    throw "Database deployment failed"
}

Write-Host ""
Write-Host "================================="
Write-Host "Database deployed successfully."
Write-Host "================================="

}
catch {
    Write-Host ""
    Write-Host "=================================" -ForegroundColor Red
    Write-Host "DEPLOY FAILED" -ForegroundColor Red
    Write-Host "=================================" -ForegroundColor Red

    Write-Host ""
    Write-Host "Error:" -ForegroundColor Yellow
    Write-Host $_.Exception.Message -ForegroundColor Red

    Write-Host ""
    Write-Host "Details:" -ForegroundColor Yellow
    Write-Host $_
}
finally {
    Write-Host ""
    Read-Host "Press Enter to close"
}