# Beta Subscribers CLI Tool

The Beta Subscribers CLI tool is now included directly in the Docker container, allowing you to manage beta program subscribers without needing to build the CLI tool separately.

## Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all subscribers or filter by status | `./cli list --status=invited` |
| `stats` | Show subscriber statistics | `./cli stats` |
| `remove` | Remove a subscriber by email | `./cli remove user@example.com` |

## Using the CLI on the Server

When you SSH into your server, you can use the CLI tool to manage subscribers:

```bash
# SSH into your server
ssh user@your-server

# Run the CLI commands
docker exec -it form-submission-service ./cli list
docker exec -it form-submission-service ./cli stats
docker exec -it form-submission-service ./cli remove user@example.com
```

## CLI Options

- `--db-path=<path>`: Specify a custom database path (default: uses the path from environment variables)
- `--status=<status>`: Filter subscribers by status when using the `list` command

## Examples

### List all subscribers
```bash
docker exec -it form-submission-service ./cli list
```

### List only invited subscribers
```bash
docker exec -it form-submission-service ./cli list --status=invited
```

### Show subscriber statistics
```bash
docker exec -it form-submission-service ./cli stats
```

### Remove a subscriber
```bash
docker exec -it form-submission-service ./cli remove user@example.com
``` 