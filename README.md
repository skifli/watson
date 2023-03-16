# watson

[![Go Report Card](https://goreportcard.com/badge/github.com/skifli/watson)](https://goreportcard.com/report/github.com/skifli/watson)
![Lines of Code](https://img.shields.io/github/languages/code-size/skifli/watson)
[![Release Downloads](https://img.shields.io/github/downloads/skifli/watson/total.svg)](https://github.com/skifli/watson/releases)

- [watson](#watson)
  - [Installation](#installation)
    - [Using pre-built binaries](#using-pre-built-binaries)
    - [Running from source](#running-from-source)
  - [Usage](#usage)
  - [Stargazers over time](#stargazers-over-time)

watson allows you to easily search for social media accounts across a multitude of platforms.

## Installation

### Using pre-built binaries

Pre-built binaries are made available for every `x.x` release. If you want more frequent updates, then [run from source](#running-from-source). Download the binary for your OS from the [latest release](https://github.com/skifli/watson/releases/latest). There are quick links at the top of every release for popular OSes.

> **Note** If you are on **Linux or macOS**, you may have to execute **`chmod +x path_to_binary`** in a shell to be able to run the binary.

### Running from source

Use this method if none of the pre-built binaries work on your system, or if you want more frequent updates. It is possible that your system's architecture is different to the one that the binaries were compiled for **(AMD)**.

> **Note** You can check your system's architecture by viewing the value of the **`GOHOSTARCH`** environment variable.

* Make sure you have [Go](https://go.dev) installed and is in your system environment variables as **`go`**. If you do not have go installed, you can install it from [here](https://go.dev/dl/).
* Download and extract the repository from [here](https://github.com/skifli/watson/archive/refs/heads/master.zip). Alternatively, you can clone the repository with [Git](https://git-scm.com/) by running `git clone https://github.com/skifli/watson` in a terminal.
* Navigate into the `/src` directory of your clone of this repository.
* Run the command `go build main.go`.
* The compiled binary is in the same folder, named `main.exe` if you are on Windows, else `main`.

## Usage

```
Usage: main.exe [--sites SITES] [--colourless] [--printall] [--readtimeout READTIMEOUT] [--writetimeout WRITETIMEOUT] [--reqsperthread REQSPERTHREAD] [USERNAME]

Positional arguments:
  USERNAME               The username to check for.

Options:
  --sites SITES          The file containing the sites to search. [default: ./sites.json]
  --colourless           Disables coloured output. [default: false]
  --printall             Print all sites, even ones which matches are not found for. [default: false]
  --readtimeout READTIMEOUT
                         Timeout for reading request response (in milliseconds). [default: 500]
  --writetimeout WRITETIMEOUT
                         Timeout for writing request (in milliseconds). [default: 500]
  --reqsperthread REQSPERTHREAD
                         The amount of requests per thread. Can significantly increase or decrease speed. [default: 3]
  --help, -h             display this help and exit
```

## Stargazers over time

[![Stargazers over time](https://starchart.cc/skifli/watson.svg)](https://starchart.cc/skifli/watson)