@echo on

cd /d C:\Cabinet

tar --exclude=CabinetREST/.git --exclude=CabinetREST/storage -czf - CabinetREST | ssh cabinet-vps "mkdir -p /root/Programs && rm -rf /root/Programs/CabinetREST && cd /root/Programs && tar -xzf -"

ssh cabinet-vps "mkdir -p /root/Programs/CabinetREST/storage"

scp C:\Cabinet\CabinetREST\storage\cabinet.db cabinet-vps:/root/Programs/CabinetREST/storage/cabinet.db

pause