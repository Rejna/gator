# `gator` - RSS feeds manager CLI

`gator` is a Go learning project. It was created during the course of [Boot.dev's](https://boot.dev) "Build a Blog Aggregator" guided project.

## Requirements

* `go >= 1.23.0`
* `postgres >= 15`
* `sqlc` (development only)

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

* `goose`

  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

## Install

```bash
go install github.com/Rejna/gator
```

## Getting started

Before running the CLI, you need to create a config file in your home directory. It should be called `.gatorconfig.json` and it should look something like this:

```json
{
  "db_url":"postgres://<postgres username>:<postgres password>@<postgres address>:<postgres port>/<db name>?sslmode=disable",
  "current_user_name":""
}
```

`db_url` should contain the connection string to your local Postgres database with the empty DB already created. `current_user_name` can be left empty - it will be populated later by `register`/`login` commands.

## Basic usage

```bash
# Run DB migrations from the root of this repo
goose <postgress username> <postgres connection string> up -dir ./sql/schema

# register new user
./gator register <username>

# add the new RSS feed and start following it
./gator addfeed "<feed name>" "<feed url>"

# scrape data (watch out - this command runs until it's manually stopped)
./gator agg 15s

# show 10 latest posts from the followed feeds
./gator browse 10
```

## Available commands

### `register`

```bash
./gator register <username>
```

Adds a user to a local Postgress database, so that they can store data about their RSS feeds subscriptions. This user is automatically set as the current user.

### `login`

```bash
./gator login <username>
```

Sets the current user of `gator`.

### `reset`

```bash
./gator reset
```

Clear database of all data. Mostly for testing purposes.

### `users`

```bash
./gator users
```

List all registered users.

### `addfeed`

```bash
./gator addfeed "<feed name>" "<feed URL>"
```

Add new RSS feed to be scraped. This feed is automatically followed by the current user.

### `feeds`

```bash
./gator feeds
```

List all added RSS feeds.

### `agg`

```bash
./gator agg <time interval>
```

Continously scrape all registered RSS feeds, once every `time interval`. This parameter accepts Go-style durations, eg. `10s`, `1m30s`, `1h`.

### `follow`

```bash
./gator follow "<feed URL>"
```

Make the current user follow RSS feed with the feed's URL as an argument.

### `unfollow`

```bash
./gator unfollow "<feed URL>"
```

Current user stops following feed with feed's URL as an argument.

### `following`

```bash
./gator following
```

List all feeds that are followed by the currently logged in user.

### `browse`

```bash
./gator browse <result limit>
```

Print the latest `result limit` number of posts from RSS feeds that currently logged in user is following. By default, it prints 2 latest results.
