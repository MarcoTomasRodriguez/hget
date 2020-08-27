# hget

![](https://github.com/MarcoTomasRodriguez/hget/workflows/CI/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/b9f13f0d5ce04d629a36f9da50da372d)](https://www.codacy.com/manual/MarcoTomasRodriguez/hget?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=MarcoTomasRodriguez/hget&amp;utm_campaign=Badge_Grade)

This is a mantained version of [huydx/hget](https://github.com/huydx/hget).

hget allows you to download at the maximum speed possible using download threads and to stop and resume tasks.

## Why should I use hget?

hget works especially well when the bottleneck is on the server, rather than the computer.
The parallelism allows the computer to download at his maximum possible speed (but it has its overheads).

If there is no bottleneck in the server, it is probably better to use 1 download thread.
Else you should set the number of download threads as low as possible to achieve the maximum speed.

For example, if the computer (and the internet connection) allows you to download at 5MB/s and the server only gives you 1MB/s,
then it would be a great option to use ~5 download threads.
But if the computer allows you to download at 5MB/s and the server provides you >5MB/s, then there's no need to use more than 1
download thread.

## Install

```bash
go get -u github.com/MarcoTomasRodriguez/hget
cd $GOPATH/src/github.com/MarcoTomasRodriguez/hget
make install
```

This will install the program with the golang default installer.

Alternatively, you can build the binary directly with:

```bash
go get -u github.com/MarcoTomasRodriguez/hget
cd $GOPATH/src/github.com/MarcoTomasRodriguez/hget
make clean build
```

## Usage

### Download

```bash
hget [-n parallelism] [URL]
```

`-n` Download threads (Default: CPUs).

### List

```bash
hget list
```

### Resume

```bash
hget resume [Task | URL]
```

### Clear

```bash
hget clear
```

### Remove

```bash
hget remove [Task | URL]
```

## Configuration

The configuration file is located at $HOME/.hget/config.toml

### Default

```toml
# enables/disables the display of the progress bar.
display_progress_bar = true

# sets the length of the hash used to prevent collisions.
# Constrains: 0 < x <= 32
use_hash_length = 16

# enables/disables the collision protection using a hash while moving the file from inside the program to outside.
save_with_hash = true

# sets the bytes to copy in a row from the response body. 
# Constrains: 0 < x
copy_n_bytes = 250

# restricts the logs to what the user wants to get: 0 means no logs, 1 only important logs and 2 all logs.
# Constrains: 0 <= x <= 2
log_level = 2

# defines in which directory the downloaded file will be moved to.
# if it is empty, then the download folder will be the terminal cwd.
download_folder = ""
```
