@echo off
cd /d C:\Cabinet

tar --exclude=CabinetREST/.git --exclude=CabinetREST/storage -czf - CabinetREST | ssh user@192.168.7.233 "rm -rf /home/user/Programs/CabinetREST && cd /home/user/Programs && tar -xzf -"

ssh user@192.168.7.233 "mkdir -p /home/user/Programs/CabinetREST/storage"
scp C:\Cabinet\CabinetREST\storage\ecabinet.db user@192.168.7.233:/home/user/Programs/CabinetREST/storage/ecabinet.db

pause