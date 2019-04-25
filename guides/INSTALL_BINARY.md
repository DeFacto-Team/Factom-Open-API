# Installation guide for 🛠 binaries
Factom Open API consists of API binary, user management binary & config file.

## Step 1: Prepare a database
Postgres database is required for Factom Open API.
You can connect to any internal/external Postgres DB by filling the config file.
Please prepare DB connection credentials for the next step.

## Step 2: Prepare configuration file

### Download config template
Create any folder (e.g. ~/.foa) for config and download config template into it
```bash
mkdir ~/.foa
curl -o ~/.foa/config.yaml https://raw.githubusercontent.com/DeFacto-Team/Factom-Open-API/master/config.yaml.EXAMPLE
```

### Introduction to config
There are few sections into config file:
* `api` (API params)
* `store` (DB params)
* `factom` (Factom params)

You may use custom config params: uncomment the line and put your value to override the default value.

#### API params
By default Open API uses HTTP port 8081.<br />
Log levels: `3` — Warn, `4` — Info, `5` – Debug, `6` – Debug+DB

#### DB params
Specify connection to your internal/external Postgres DB.

#### Factom params
❗️ You need to fill `factom`.`esaddress` in order to use Factom Open API.<br />
By default Open API is connected to <a href="https://factomd.net" target="_blank">Factom Open Node</a>, that means you don't need to setup your own node on the Factom blockchain to work with blockchain. But if you want to use your own node, you may specify it into the config.<br />

### Fill the config
```bash
nano ~/.foa/config.yaml
```

## Step 3: Download Factom Open API binaries and run them
Download with cURL:
```bash
curl -o foa https://github.com/DeFacto-Team/Factom-Open-API/releases/download/1.0.0/foa
curl -o user https://github.com/DeFacto-Team/Factom-Open-API/releases/download/1.0.0/user
```

Run Factom Open API with default config location:
```bash
./foa
```

Run user management app with default config location:
```bash
./user
```

*By default all binaries uses the config located at `<USER_FOLDER>/.foa/config.yaml`. If you use custom location for config file, please don't forget to provide it with `-c` flag while running binaries.*
```bash
./foa -c=/somewhere/placed/config.yaml
```

## (Optional) Step 4: Run Factom Open API binary as a daemon
*This instruction is for Linux.*
Setup running the Factom Open API as a daemon is needed to automatically start Factom Open API when your server starts after reboot.

### Move binary
Move `foa` binary to `/usr/bin/`.

### Create new user
It’s better to run Factom Open API as non-root user, so you need new user:
```bash
adduser foa
```

### Create service file to run Factom Open API as daemon
```bash
nano /etc/systemd/system/foa.service
```

Fill this file with the following daemon config:
```bash
[Unit]
Description=Run the Factom Open API service
Documentation=https://github.com/DeFacto-Team/Factom-Open-API
After=network-online.target
[Service]
User=foa
Group=foa
EnvironmentFile=-/etc/default/foa
ExecStart=/usr/bin/foa $FOA_OPTS
KillMode=control-group
Restart=on-failure
[Install]
WantedBy=multi-user.target
```

### Create environment file for Factom Open API service

```bash
nano /etc/default/foa
```

Fill this file with `FOA_OPTS` param:
```bash
FOA_OPTS = ''
```

### Enable Factom Open API daemon
```bash
systemctl daemon-reload
systemctl start foa
systemctl enable foa
```
