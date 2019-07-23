# Installation guide for 🐳 Docker
Factom Open API packed into single Docker-alpine image, which includes API binary, user management binary & config file.

## Step 1: Prepare a database
Postgres database is required for Factom Open API.
You can run Postgres DB into container on the same machine or connect to any internal/external Postgres DB by filling the config file on the next step.

### Option 1: Run postgres DB into Docker container
Create container `foa-db` (default configuration of Open API) with no external access into it:
```bash
docker run -d --name foa-db postgres
```

### Option 2: Use your own Postgres DB
Nothing to do.
Just prepare DB connection credentials for the next step.

## Step 2: Prepare configuration file

### Download config template
Create any folder (e.g. ~/.foa) for config and download config template into it
```bash
mkdir ~/.foa
curl -o ~/.foa/config.yaml https://raw.githubusercontent.com/DeFacto-Team/Factom-Open-API/master/config.yaml.EXAMPLE
```

### Introduction to config
There are few sections into config file:
* `admin` (Admin params)
* `api` (API params)
* `store` (DB params)
* `factom` (Factom params)

You may use custom config params: uncomment the line and put your value to override the default value.

#### Admin params
Login & password for Admin UI. Leave empty to disable Admin UI.

#### API params
By default Open API uses HTTP port 8081.<br />
Log levels: `3` — Warn, `4` — Info, `5` – Debug, `6` – Debug+DB

#### DB params
If you use Postgres DB into `foa-db` container, then use the default config.
Otherwise, specify connection to your internal/external Postgres DB.

#### Factom params
Entry Credits (EC) purchase fixed amounts of data in the Factom network.<br />
<b>You need EC address filled with Entry Credits to write data on the Factom.</b><br />
- EC address may be imported or generated via Open API Admin UI<br />
- <a href="https://ec.de-facto.pro" target="_blank">Fund your EC address in the EC store</a>
<br /><br />
By default Open API is connected to <a href="https://factomd.net" target="_blank">Factom Open Node</a>, that means you don't need to setup your own node on the Factom blockchain to work with blockchain. But if you want to use your own node, you may specify it into the config.<br />

### Fill the config
```bash
nano ~/.foa/config.yaml
```

### Change owner
```bash
chown 1000:1000 ~/.foa/config.yaml
```

## Step 3: Run Open API container
If you use Postgres DB as Docker container, use the following command with `--link` flag:
```bash
docker run -d -p 8081:8081 --name factom-open-api --link foa-db -v ~/.foa:/home/app/values defactoteam/factom-open-api:1.0.0
```

If you do not use Postgres DB as Docker container, use the following command without `--link` flag:
```bash
docker run -d -p 8081:8081 --name factom-open-api -v ~/.foa:/home/app/values defactoteam/factom-open-api:1.0.0
```

1. If you changed API HTTP port into config, don't forget to edit `-p 8081:8081`.
2. Check the path to config file and edit it, if you use another path, than suggested in this guide (i.e. not `~/.foa`)
3. Don't use `latest` tag for docker image, it's better to install specific releases. Latest release: `1.0.0`.
