# Temp Dir Creation
New-Item C:\Temp -ItemType Directory

# Download SteamCMD
Invoke-WebRequest -Uri https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip -UseBasicParsing -OutFile C:\Temp\steamcmd.zip

# Extract SteamCMD
Expand-Archive -Path C:\Temp\steamcmd.zip -DestinationPath C:\steamcmd -Force

# Install Valheim
Set-Location C:\steamcmd
Start-Process -FilePath .\steamcmd.exe -ArgumentList "+login anonymous +force_install_dir .\Valheim +app_update 896660 validate +exit" -Wait

# Copy Start Batch
Copy-Item .\valheim\start_headless_server.bat -Destination .\valheim\start.bat

# Update Parameters
$startBatch = Get-Content .\valheim\start.bat 
($startBatch -replace 'My server','techdecline') -replace 'secret','Passw0rd!' | Set-Content -Path .\valheim\start.bat


# Update Windows Firewall
New-NetFirewallRule -Name Valheim_tcp -DisplayName Valheim -Enabled True -Action Allow -Direction Inbound -LocalPort 2456-2458 -Protocol TCP
New-NetFirewallRule -Name Valheim_upd -DisplayName Valheim -Enabled True -Action Allow -Direction Inbound -LocalPort 2456-2458 -Protocol UDP

# Start dedicated Server
. ".\valheim\start.bat" 