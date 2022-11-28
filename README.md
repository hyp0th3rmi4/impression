# PubSubber

Impression latch onto a PubSub topic via a temporary subscription and copies messages either to console or file systems up until a termination condition is reached (CTRL+C or maximum number of messages).

## Synopsis

```
$> go mod download
$> make clean build
$> ./bin/impression 
```

The commands above will build the source code and run impression. The command line help will provide additional details about
the usage of the command.

