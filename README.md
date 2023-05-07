# BashMon 
A proof of concept project that monitors Bash command inputs inside and outside a container using uprobe from eBPF.

## Introduction
BashMon has two subdirectories inside the project.
- `agent`: An agent source code for actually monitoring all Bash inputs from inside and outside a container. This will send data to `master` using REST API.
- `master`: A REST API server for collecting data from agent and storing it to a sqlite3 DB.

## How to Use
### 1. Agent
In order to self compile `agent`, you need `clang` for compiling eBPF code into object files.

```bash
make build
```
Will generate an executable file named `monitor_bash`. `monitor_bash` offers following command-line arguments from user:

- `master`: The master server's IP and port in string format.
- `verbose`: A boolean format that decides whether or not to print out the whole detection process when a Bash event happens.

An example execution command will be:
```bash
sudo ./monitor_bash --master=http://172.23.14.15:9000 --verbose=true
```
> Since `monitor_bash` requires to perform memlock, you need `sudo` permission to run the program. Otherwise, it will fail due to `operation not permitted`.

Once `monitor_bash` has started, it will automatically look for all Overlay FS inside the current host and look for `/bin/bash` files to attach. Since Overlay FS will perform CoW, those `/bin/bash` files might have same inodes. Program automatically detects this and attaches uprobe to the only ones that have different inodes. uprobe will be attached to symbol `readline` to monitor user's inputs
> Be aware that this is just a proof of concept, rather a implementation of monitoring Bash programs inside each containers. This will not detect shell executions nor detect signals comming from the Bash.

### 2. Master
In order to self compile `master`, you can use following command.
```bash
go build
```
Will generate an executable file named `master`. `master` offers following command-line arguments from user:
- `k8s`: Whether or not to enable Kubernetes detection. If enabled, once agents send container events, master will look for pods that actually runs the containers using Kubernetes API.
- `port`: The integer value to run the service on specific port. Defaults to `9000`
- `verbose`: A boolean format that decides whether or not to print out the API receiving process when an agent sends data.

An example execution command will be:
```bash
./master --k8s=true --verbose=true --port=9000
```
> Please be aware that using `--k8s true --verbose true` might set verbose to `false` due to malfunctions inside commandline argument parsing from Golang.

Once `master` has started, it will start an API server on the port designated by the user. Then will listen for agent's `POST` requests and store those requests inside the sqlite databse.

## Contribution
Since this is rather a proof of concept, only critical issues will be taken into the mainstream. 
