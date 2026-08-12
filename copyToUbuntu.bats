@echo off
cd /d C:\Cabinet

tar --exclude=CabinetREST/.git -czf - CabinetREST | ssh user@192.168.7.233 "cd /home/user/Programs && tar -xzf -"

pause