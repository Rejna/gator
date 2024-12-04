# `gator` - RSS feeds manager CLI

`gator` is a Go learning project. It was created during the course of [Boot.dev's](https://boot.dev) "Build a Blog Aggregator" guided project.

## Requirements

* `go >= 1.23.0`
* `postgres >= 15`
* `sqlc`

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

* `goose`

  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

## Available commands

### `register`

```bash
./gator register <username>
```

Adds a user to a local Postgress database, so that they can store data about their RSS feeds subscriptions.

### `login`

```bash
./gator login <username>
```

Sets the current user of `gator`.

### `reset`

```bash
./gator reset
```

### `users`

```bash
./gator users
```

### `agg`

```bash
./gator agg <time interval>
```

### `feeds`

```bash
./gator feeds
```


### `addfeed`

```bash
./gator addfeed "<feed name>" "<feed URL>"
```


### `follow`

```bash
./gator follow <>
```


### `unfollow`

```bash
./gator unfollow "<feed URL>"
```


### `following`

```bash
./gator following
```


### `browse`

```bash
./gator browse <result limit>
```