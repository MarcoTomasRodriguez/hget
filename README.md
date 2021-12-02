<div id="top"></div>

<div align="center">
  <a href="https://github.com/marcotomasrodriguez/hget">
    <img src="https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/svg/logo.svg" alt="Logo" width="80" height="80">
  </a>
  <h2 align="center">hget</h2>
  <p align="center">
    Interruptible and resumable download accelerator.
    <br />
    <a href="https://github.com/marcotomasrodriguez/hget"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/marcotomasrodriguez/hget">View Demo</a>
    ·
    <a href="https://github.com/marcotomasrodriguez/hget/issues">Report Bug</a>
    ·
    <a href="https://github.com/othneildrew/Best-README-Template/issues">Request Feature</a>
    <br />
    <br />
    <a href="https://opensource.org/licenses/MIT">
      <img src="https://img.shields.io/badge/License-MIT-blue.svg">
    </a>
    <a href="https://github.com/MarcoTomasRodriguez/hget/workflows/CI">
      <img src="https://github.com/MarcoTomasRodriguez/hget/workflows/CI/badge.svg">
    </a>
  </p>
</div>

## Description

hget allows you to download files at maximum speed using workers (goroutines), and shines when the bottleneck is on the server, rather than the client.

It takes advantage of the `Range` header, allowing many workers to download in parallel a file. That said, it is not a Swiss Army Knife: if there is no bottleneck in the server, it is probably better to use just 1 download worker; else, you should set the number of workers to the minimum value that allows you to download at the maximum of your internet speed.

For example, if the client can download 100MB/s and the server only provides you 10MB/s per worker, then it would be wise to use ~10 download workers. Nevertheless, if the client has 10MB/s and the server provides you >10MB/s, there is no need to use more than 1.

Other features of this software are,

1. Interruptible downloads: press <kbd>Ctrl</kbd> + <kbd>C</kbd> or <kbd>⌘</kbd> + <kbd>C</kbd> and the download will stop gracefully.
2. Resumable downloads: use `hget resume ID` to resume an interrupted download.
3. Prevent file collision: enable the collision protection in the configuration and downloads with the same name will not collide (a random string is included before the filename).

<p align="right">(<a href="#top">back to top</a>)</p>

## Installation

### Using Go

```bash
go install github.com/MarcoTomasRodriguez/hget
```

### Binary

Precompiled binaries for Linux and MacOS are available at https://github.com/MarcoTomasRodriguez/hget/releases.

## Usage

### Download

```bash
hget [-n workers] URL
```

`-n` Download workers (Default: CPUs).

![Download demo](https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/gif/root.gif)

### List

```bash
hget list
```

![List demo](https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/gif/list.gif)

### Resume

```bash
hget resume ID
```

![Resume demo](https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/gif/resume.gif)

### Remove

```bash
hget remove ID
```

![Remove demo](https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/gif/remove.gif)

### Clear

```bash
hget clear
```

![Clear demo](https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/gif/clear.gif)

<p align="right">(<a href="#top">back to top</a>)</p>

## Configuration

The configuration file is located by default at `$HOME/.hget/config.toml`.

```toml
# Folder used by the program to save temporal files, such as ongoing and paused downloads.
program_folder = "$HOME/.hget" # This will not work. Write an absolute path instead.

# Restricts the logs to what the user wants to get. 0 means no logs, 1 only important logs and 2 all logs.
log_level = 2

# Defines the directory in which the downloaded file will be moved.
download.folder = "$PWD" # This will not work. Write an absolute path instead.

# Sets the bytes to copy in a row from the response body.
download.copy_n_bytes = 300

# Enables/disables the collision protection using a random string when saving the file to the final destination.
download.collision_protection = false
```

<p align="right">(<a href="#top">back to top</a>)</p>

## Acknowledgments

- [huydx](https://github.com/huydx) for creating the initial version of this project.

<p align="right">(<a href="#top">back to top</a>)</p>

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>
