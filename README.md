# Housekeeper v2

New and improved housekeeper, now in Go

## Usage
The program relies on having a list of account to actually check. This list can either be provided manually, or the python script in `accounts_retriever.py` can be used. This script will get an up-to-date mapping from the Metavisor repository on Gerrit. 

The recommended way of using Housekeeper is through Docker. For the most common use cases, there are make targets (take a look in the `Makefile`).
