# Chatter

Chatter is a command-line peer-to-peer text messaging tool written in Go. With Chatter, users can connect with others and engage in real-time text-based conversations. This tool provides a convenient and lightweight way to communicate with other users directly from the command line.

![Demo](./demo.png)

## Features

-   **Peer-to-peer messaging**: Connect with other users directly without any intermediary servers.
-   **Real-time communication**: Engage in text-based conversations in real-time.
-   **Command-line interface**: Conveniently chat with others directly from the command line interface.
-   **Lightweight**: Chatter is designed to be lightweight and efficient.

## Getting Started

### Prerequisites

Before using Chatter, ensure you have Go installed on your system. You can download and install Go from the [official Go website](https://golang.org/dl/).

### Installation

To install Chatter, simply run the following command:

```bash
go install github.com/chiragsoni81245/chatter
```

**Note-** after this install you should get a executable file at your GOPATH, so make sure if you have your GOPATH configured into your PATH environment variable so that you can access that executable from anywhere, otherwise you can also run it by exclusively going to your GOPATH directoy

To check your go environment variables like GOPATH, you can use this command

```bash
go env
```

### Usage

Once Chatter is installed, you can start using it from the command line. Here are some basic usage examples:

1. **Start Chatter** To start Chatter, simply run the following command:

```bash
chatter start
```

This will give you can interactive interface where you can do two operations, **Connect** and **Chat**, you can see the usage of that in the demo video
