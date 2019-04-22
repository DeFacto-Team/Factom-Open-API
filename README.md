# Factom Open API
Factom Open API is a lightweight REST API for Factom blockchain. It connects to existing factomd node and has a built-in Factom wallet, so you don't need to run separate instance for signing data before writing it on the Factom blockchain.

## Main features
* **Instant start:** use Open API immediately after installation without long syncing data from the blockchain
* **Write data** on the blockchain at **fixed predictable cost** (**$1 for 1,000 entries** of 1 KB)
* **BaaS-ready:** user-based API access, counting usage, limits
* **Read chains & entries** from the blockchain easily
* **Search chains & entries** by tags (external IDs)
* **Pagination, sorting, filtering** results with query params
* **Generic factomd interface:** send direct requests to factomd API via special REST path

## Installation guides (developer release)
* 🐳 <a href="https://github.com/DeFacto-Team/Factom-Open-API/blob/master/guides/INSTALL_DOCKER.md">Run in Docker</a>
* 🛠 Use binaries (guide is not ready yet)

## User management app

For access & work with Factom Open API you need to create user(s).
In the next version the user management will be possible via admin endpoint and Web UI, but for current release we developed the admin binary.

### You run Factom Open API as 🐳 Docker container
The binary is embedded into Open API container, so you can run it via terminal:
```bash
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml create anton
```
You will see access key into terminal.
By default, new users **are enabled** and **have no writes limit**.

You can manage users with additional binary commands:
```bash
# create user `anton` and generate API access key
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml create anton

# disable access to API for user `anton`
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml disable anton

# enable access to API for user `anton`
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml enable anton

# delete user `anton`
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml delete anton

# rotate API access key for user `anton`
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml rotate-key anton

# set writes limit for user `anton` to `1000` // 0 for unlimited
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml set-limit anton 1000

# show users, API keys & params
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml ls

# show help
docker exec -ti factom-open-api ./user -c=/home/app/values/config.yaml help
```
