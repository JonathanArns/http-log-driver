# http-log-driver

This is a simple docker log driver that sends logs as json objects via http.

## Background

This was created to provide access to container logs to ditm.

## Install

```
$ docker plugin install jonathanarns/http-log-driver
```

## Check

```
$ docker plugin ls
ID                  NAME                                  DESCRIPTION         ENABLED
8c7587db6fdc        jonathanarns/http-log-driver:latest   File log driver     true

```

## Usage

Run a container using this plugin:

```
$ docker run --log-driver jonathanarns/http-log-driver --log-opt endpoint=http://localhost:8080 alpine date
```
Note that `endpoint` is a required argument. It is also the only argument.

## Uninstall

To uninstall, please make sure that no containers are still using this plugin. After that, disable and remove the plugin like this:

```
$ docker plugin disable jonathanarns/http-log-driver
$ docker plugin rm jonathanarns/http-log-driver
```
