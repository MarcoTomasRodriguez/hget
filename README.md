# hget

This is a mantained version of [huydx/hget](https://github.com/huydx/hget).

Unfortunately, the original author stopped giving support to this project; since the idea is cool, I decided to continue it.

In the future, this could be merged back.

## Install

```
$ go get -d github.com/MarcoTomasRodriguez/hget
$ cd $GOPATH/src/github.com/MarcoTomasRodriguez/hget
$ make install
```

This will install the program with the golang default installer.

Alternatively, you can build the binary directly with:

```
$ go get -d github.com/MarcoTomasRodriguez/hget
$ cd $GOPATH/src/github.com/MarcoTomasRodriguez/hget
$ make clean build
```

## Usage

```
hget [-n parallel] [Url] // Downloads a file using n threads. The default is the number of cores.
hget tasks // Gets all the interrupted tasks.
hget resume [TaskName | URL] // Resumes a task given a TaskName or URL.
```

To interrupt any on-downloading process, just ctrl-c or ctrl-d at the middle of the download, hget will safely save your data and you will be able to resume later.

### Download

![](https://i.gyazo.com/89009c7f02fea8cb4cbf07ee5b75da0a.gif)

### Resume

![](https://i.gyazo.com/caa69808f6377421cb2976f323768dc4.gif)


