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
If you use Postgres DB into `foa-db` container, then use the default config.
Otherwise, specify connection to your internal/external Postgres DB.

#### Factom params
❗️ You need to fill `factom`.`esaddress` in order to use Factom Open API.<br />
By default Open API is connected to <a href="https://factomd.net" target="_blank">Factom Open Node</a>, that means you don't need to setup your own node on the Factom blockchain to work with blockchain. But if you want to use your own node, you may specify it into the config.<br />

### Fill the config
```bash
nano ~/.foa/config.yaml
```

## Step 3: Download and install Factom Open API binaries
…

*By default all binaries uses the config located at `<USER_FOLDER>/.foa/config.yaml`. If you use custom location for config file, please don't forget to provide it with `-c` flag while running binaries.*
```bash
./foa -c=/somewhere/placed/config.yaml
```