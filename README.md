<div id="top"></div>

<div align="center">
  <a href="https://github.com/marcotomasrodriguez/hget">
    <img src="https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/assets/svg/logo.svg" alt="Logo" width="80" height="80">
  </a>
  <h2 align="center">hget</h2>
  <p align="center">
    Interruptible and resumable download accelerator.
    <br />
    <a href="https://github.com/MarcoTomasRodriguez/hget#description"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/marcotomasrodriguez/hget#usage">View Demo</a>
    ·
    <a href="https://github.com/marcotomasrodriguez/hget/issues">Report Bug</a>
    ·
    <a href="https://github.com/othneildrew/Best-README-Template/issues">Request Feature</a>
    <br />
    <br />
    <img src="https://img.shields.io/github/license/marcotomasrodriguez/hget?label=Licence&style=flat-square&color=blue">
    <img src="https://img.shields.io/github/release/marcotomasrodriguez/hget?display_name=tag&sort=semver&label=Release&style=flat-square&color=blue">
    <img src="https://img.shields.io/github/workflow/status/marcotomasrodriguez/hget/Test?label=Tests&style=flat-square">
  </p>
</div>

## Description

hget is a command-line tool written in Go that downloads files at maximum speed by using range downloads with multiple workers and shines when the bottleneck is on the server.

It takes advantage of the `Range` header, allowing many workers to download in parallel a file. That said, it is not a Swiss Army Knife: if there is no bottleneck in the server, it is probably better to use just 1 download worker; else, you should set the number of workers to the minimum value that allows you to download at the maximum of your internet speed.

For example, if the client can download 100MB/s and the server only provides you 10MB/s per worker, then it would be wise to use ~10 download workers (beware of context switching). Nevertheless, if the client has 10MB/s and the server provides you >10MB/s, there is no need to use more than 1.

### Additional features

- Interruptible downloads: press <kbd>Ctrl</kbd> + <kbd>C</kbd> or <kbd>⌘</kbd> + <kbd>C</kbd> and the download will stop gracefully.
- Resumable downloads: use `hget resume ID` to resume an interrupted download.

<p align="right">(<a href="#top">back to top</a>)</p>

## Installation

### Using Go

```bash
go install github.com/MarcoTomasRodriguez/hget@latest
```

### Binary

Precompiled binaries for Linux and MacOS are available at https://github.com/MarcoTomasRodriguez/hget/releases.

<p align="right">(<a href="#top">back to top</a>)</p>

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

## Acknowledgments

- [huydx](https://github.com/huydx) for creating the initial version of this project.

<p align="right">(<a href="#top">back to top</a>)</p>

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>
