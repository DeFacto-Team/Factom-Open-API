# Factom Open API
Factom Open API is a lightweight REST API for Factom blockchain. It connects to existing factomd node and has a built-in Factom wallet, so you don't need to run separate instance for signing data before writing it on the Factom blockchain.

## Main features
* **Instant start:** use Open API immediately after installation without long syncing data from the blockchain
* **Write data** on the blockchain at **fixed predictable cost** (**$1 for 1,000 entries** of 1 KB)
* **BaaS-ready:** user-based API access, counting usage, limits
* **Read all chain entries at once**: using single request (no need to read all entry blocks one by one)
* **Search chains & entries** by tags (external IDs)
* **Pagination, sorting, filtering** results with query params
* **Generic factomd interface:** all factomd API requests are supported via special REST path

## API Reference

### Methods
* **Chains**
  * <a href="https://docs.openapi.de-facto.pro/chains/create-chain" target="_blank">POST /chains</a> – *Create chain*
  * <a href="https://docs.openapi.de-facto.pro/chains/get-chains" target="_blank">GET /chains</a> – *Get user's chains*
  * <a href="https://docs.openapi.de-facto.pro/chains/search-chains" target="_blank">POST /chains/search</a> – *Search user's chains by ExtIDs*
  * <a href="https://docs.openapi.de-facto.pro/chains/get-chain" target="_blank">GET /chains/:chainId</a> – *Get chains by ChainID*
  * <a href="https://docs.openapi.de-facto.pro/chains/get-chain-entries" target="_blank">GET /chains/:chainId/entries</a> – *Get chain entries*
  * <a href="https://docs.openapi.de-facto.pro/chains/get-chain-first-entry" target="_blank">GET /chains/:chainId/entries/first</a> – *Get first entry of chain*
  * <a href="https://docs.openapi.de-facto.pro/chains/get-chain-last-entry" target="_blank">GET /chains/:chainId/entries/last</a> – *Get last entry of chain*
  * <a href="https://docs.openapi.de-facto.pro/chains/search-chain-entries" target="_blank">POST /chains/:chainId/entries/search</a> – *Search entries in chain by ExtIDs*
* **Entries**
  * <a href="https://docs.openapi.de-facto.pro/entries/create-entry" target="_blank">POST /entries</a> – *Create entry in chain*
  * <a href="https://docs.openapi.de-facto.pro/entries/get-entry" target="_blank">GET /entries/:entryHash</a> – *Get entry by EntryHash*
* **Generic**
  * <a href="https://docs.openapi.de-facto.pro/factomd/factomd-method" target="_blank">POST /factomd/:method</a> – *Generic factomd interface*
* **Info**
  * <a href="https://docs.openapi.de-facto.pro/user/get-user" target="_blank">GET /user</a> – *Get user info*
  * <a href="https://docs.openapi.de-facto.pro/api/api-info" target="_blank">GET /</a> – *Get API info*

### Documentation
* Documentation on Gitbook: https://docs.openapi.de-facto.pro
* Built-in Swagger specification: `http://<factom_open_api_server_ip_and_port>/docs/index.html`

## Installation guides
* 🐳 <a href="https://github.com/DeFacto-Team/Factom-Open-API/blob/master/guides/INSTALL_DOCKER.md">Install with Docker</a>
* 🛠 Install with binaries (guide under development)

## Clients
* <a href="https://github.com/DeFacto-Team/Factom-Open-API-PHP" target="_blank">PHP</a>
* Golang (under development)
* JS (under development)

## User management app (temporarily)

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
