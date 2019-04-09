# LinkChain Install Document

## Install LinkChain

This guide will explain how to install the `lcd` and `lccli` entrypoints onto your system.

### Install Go

Install `go` by following the [official docs](https://golang.org/doc/install). Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=\$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=\$PATH:\$GOBIN" >> ~/.bash_profile
echo "export GO111MODULE=on" >> ~/.bash_profile
source ~/.bash_profile
```

::: tip
**Go 1.10.1+** is minimum version required for the linkchain.
:::

### Install golang-dep

```bash
go get -u github.com/golang/dep/cmd/dep

```

### Install the binaries

Next, let's download the latest version of linkchain. Here we'll use the `master` branch, which contains the latest stable release.

```bash
mkdir -p $GOPATH/src/github.com/
cd $GOPATH/src/github.com/
git clone https://github.com/mihongtech/linkchain.git
cd linkchain && git checkout master
dep ensure -v

```
### Install the binaries in unix

Then,install `lcd` and `lccli` in unix

```bash
cd build/unix
./build.sh

```
### Install the binaries in windows

Then,install `lcd` and `lccli` in windows

```bash
cd build/win
build.bat

```

### Verify install 
> *NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `lcd` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ which lcd
$ which lccli
```

### Next

Now you can [join the mainnet](./join-mainnet.md), [the public testnet](./join-testnet.md) or [create you own  testnet](./deploy-testnet.md)