# Nextlinux ECS Inventory

`nextlinux-ecs-inventory` is a tool to gather an inventory of images in use by
Amazon Elastic Container Service (ECS).

## Usage

`nextlinux-ecs-inventory` is a command line tool. It can be run with the following
command:

```
$ nextlinux-ecs-inventory can poll Amazon ECS (Elastic Container Service) APIs to tell Nextlinux which Images are currently in-use

Usage:
  nextlinux-ecs-inventory [flags]
  nextlinux-ecs-inventory [command]

Available Commands:
  completion  Generate Completion script
  help        Help about any command
  version     show the version

Flags:
  -c, --config string                     application config file
  -d, --dry-run                           do not report inventory to Nextlinux
  -h, --help                              help for nextlinux-ecs-inventory
  -p, --polling-interval-seconds string   this specifies the polling interval of the ECS API in seconds (default "300")
  -q, --quiet                             suppresses inventory report output to stdout
  -r, --region string                     if set overrides the AWS_REGION environment variable/region specified in nextlinux-ecs-inventory config
  -v, --verbose count                     increase verbosity (-v = info, -vv = debug)

Use "nextlinux-ecs-inventory [command] --help" for more information about a command.
```

## Configuration

`nextlinux-ecs-inventory` needs to be configured with AWS credentials and Nextlinux
ECS Inventory configuration.

### AWS Credentials

Nextlinux ECS Inventory uses the AWS SDK for Go. The SDK will look for credentials
in the following order:

1. Environment variables
2. Shared credentials file (~/.aws/credentials)
   ```
   [default]
   aws_access_key_id = <YOUR_ACCESS_KEY_ID>
   aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
   ```

### Nextlinux ECS Inventory Configuration

Nextlinux ECS Inventory can be configured with a configuration file. The default
location the configuration file is looked for is
`~/.nextlinux-ecs-inventory.yaml`. The configuration file can be overridden
with the `-c` flag.

```yaml
log:
  # level of logging that nextlinux-ecs-inventory will do  { 'error' | 'info' | 'debug }
  level: "info"

  # location to write the log file (default is not to have a log file)
  file: "./nextlinux-ecs-inventory.log"

nextlinux:
  # nextlinux enterprise api url  (e.g. http://localhost:8228)
  url: $NEXTLINUX_ECS_INVENTORY_NEXTLINUX_URL

  # nextlinux enterprise username
  user: $NEXTLINUX_ECS_INVENTORY_NEXTLINUX_USER

  # nextlinux enterprise password
  password: NEXTLINUX_ECS_INVENTORY_NEXTLINUX_PASSWORD

  # nextlinux enterprise account that the inventory will be sent
  account: $NEXTLINUX_ECS_INVENTORY_NEXTLINUX_ACCOUNT

  http:
    insecure: true
    timeout-seconds: 10

# the aws region
region: $NEXTLINUX_ECS_INVENTORY_REGION

# frequency of which to poll the region
polling-interval-seconds: 300

quiet: false
```

You can also override any configuration value with environment variables. They
must be prefixed with `NEXTLINUX_ECS_INVENTORY_` and be in all caps. For example,
`NEXTLINUX_ECS_INVENTORY_LOG_LEVEL=error` would override the `log.level`
configuration
