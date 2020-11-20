# pup-cli

Manage your pins across services with the pup cli!

# Configuration

Each provider has their own flags, and each flag can be provided via environment variable as well.

Environment variables are of the style `<PROVIDER>_OPTION_NAME_DASHES_UNDERSCORED`.

For example, the `-host` flag for the `pipin` subcommands becomes `PIPIN_HOST`.

# Operations

Each provider has three operations, `ls`, `add` and `rm`

# Providers

Currently, there are providers for [Pinata](https://pinata.cloud/) and the PiPin service, also in this repo.
