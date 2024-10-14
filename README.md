# Pun: Package UNikernels

This project aims to simplify the process of packaging unikernels for
[urunc](https://github.com/nubificus/urunc). Combining `pun` and `urunc` a user
can package and execute unikernels as easy as containers.

## Execution modes

`pun` supports two modes of execution: either a) in a local execution printing a
LLB, or b) as a frontend for Docker's buildkit. The easiest way is to use it as
a frontend.

### As buildkit's frontend

In order to use `pun` with as Buildkit's frontend, we just need to add a new
line in the top of the Dockerfile. In pparticular, every file that we want to
use for pun needs to start with the following line.


```
#syntax=harbor.nbfc.io/nubificus/urunc/pun/llb:latest
```

Then, we can just execute docker build as usual. Buildkit will fetch an image
containing `pun` and it will use it as a frontend. Therefore, no building is
required.

### Using buildctl

In order to use `pun` with buildctl, we have to build it locally and then feed
its output to buildctl.

#### How to build

We can build `pun` by simply running make:
```
make
```

> **_NOTE:_**  `pun` was created with Golang version 1.22.0

#### How to use

`pun` makes use of Buildkit and LLB. As a result, the application itself does
not produce any artifacts, but just a LLB graph which we can later feed to
buildkit. Therefore, in order to use `pun`, we need to firstly install buildkit.
For more information regarding building and installing buildkit, please refer
to buildkit
[instructions](https://github.com/moby/buildkit?tab=readme-ov-file#quick-start).

As long as buildkit is installed in our system, we can package any unikernel
with the following command:
```
./pun -f Containerfile | sudo buildctl build ... --local context=<path_to_local_context> --output type=<type>,<type-specific-args>
```

`pun` takes a single argument and that is the `Containerfile` a
Dockerfile-syntax file with the packaging instructions.

Regarding the buildctl arguments:
- `--local context` specifies the directory where the user wants to set the
  local context. It is similar to the build context in the docker build command.
THerefore, if we specify any `COPY` instructions in the `Containerfile`, the
paths will be relative to this argument.
- `--output type=<type>` specifies the output format. Buildkit supports various
  [outputs](https://github.com/moby/buildkit/tree/master?tab=readme-ov-file#output).
  Just for convenience we mention the `docker` output, which produces an output
  that we cna pass to `docker load` in order to place our image in the local
  docker registry. We can also specify the name of the image, using the
  `name=<name` in the ``<type-specific-option>`.

For instance:
```
./pun -f Containerfile | sudo buildctl build ... --local context=/home/ubuntu/unikernels/ --output type=docker,name=harbor.nbfc.io/nubificus/urunc/pun:latest | sudo docker load
```

#### The Containerfile format

`pun` supports Dockerfile-style files as input. Therefore, any such file can be
given as input. However, it is important to note, that `pun` has been built in
order to produce images for `urunc`. Therefore, currently only the following
instructions are supported:
- `FROM`: Specifies the base image. It can be any image or just `scratch`
- `COPY`: Copies local files inside the image as a new layer.
- 'LABEL': Specifies annotations for the image.

All the other instructions will get ignored.

## Examples

### Packaging a rumprun unikernel with `pun` as buildkit's frontend
In case we have already built a rumprun unikernel, we can easily package it with
a normal docker command that will use `pun`. To do that we need the following
`Containerfile`:
```
#syntax=harbor.nbfc.io/nubificus/urunc/pun/llb:latest
FROM scratch

COPY test-redis.hvt /unikernel/test-redis.hvt
COPY redis.conf /conf/redis.conf

LABEL com.urunc.unikernel.binary=/unikernel/test-redis.hvt
LABEL "com.urunc.unikernel.cmdline"='redis-server /data/conf/redis.conf'
LABEL "com.urunc.unikernel.unikernelType"="rumprun"
LABEL "com.urunc.unikernel.hypervisor"="hvt"
```

We can then build the image with the following command:
```
docker build -f Containerfile -t harbor.nbfc.io/nubificus/urunc/redis-rumprun-hvt:test .
```

The image will get loaded in the local docker registry.

### Packaging a rumprun unikernel with `pun` as buildkit's frontend

In case we want to use an existing [unikraft](unikraft.org) unikernel image from
[unikraft's catalog](https://github.com/unikraft/catalog), we can transform it
to an image that `urunc` can execute with `pun`. In that case the
`Containerfile` should look like:
```
#syntax=harbor.nbfc.io/nubificus/urunc/pun/llb:latest
FROM unikraft.org/nginx:1.15

LABEL com.urunc.unikernel.binary="/unikraft/bin/kernel"
LABEL "com.urunc.unikernel.cmdline"="nginx -c /nginx/conf/nginx.conf"
LABEL "com.urunc.unikernel.unikernelType"="unikraft"
LABEL "com.urunc.unikernel.hypervisor"="qemu"
```

We can then build the image with the following command:
```
docker build -f Containerfile -t harbor.nbfc.io/nubificus/urunc/redis-rumprun-hvt:test .
```

The image will get loaded in the local docker registry.

### Packaging a rumprun unikernel with `pun` with buildctl

In case we have already built a rumprun unikernel, we can easily package it with
`pun` using the following `Containerfile`:
```
FROM scratch

COPY test-redis.hvt /unikernel/test-redis.hvt
COPY redis.conf /conf/redis.conf

LABEL com.urunc.unikernel.binary=/unikernel/test-redis.hvt
LABEL "com.urunc.unikernel.cmdline"='redis-server /data/conf/redis.conf'
LABEL "com.urunc.unikernel.unikernelType"="rumprun"
LABEL "com.urunc.unikernel.hypervisor"="hvt"
```

We can then build the image with the following command:
```
./pun -f Containerfile | sudo buildctl build ... --local context=${PWD} --output type=docker,name=harbor.nbfc.io/nubificus/urunc/redis-rumprun-hvt:test | sudo docker load
```

The image will get loaded in the local docker registry.

### Packaging a unikraft unikernel with `pun` with buildctl

In case we want to use an existing [unikraft](unikraft.org) unikernel image from
[unikraft's catalog](https://github.com/unikraft/catalog), we can transform it
to an image that `urunc` can execute with `pun`. In that case the
`Containerfile` should look like:
```
FROM unikraft.org/nginx:1.15

LABEL com.urunc.unikernel.binary="/unikraft/bin/kernel"
LABEL "com.urunc.unikernel.cmdline"="nginx -c /nginx/conf/nginx.conf"
LABEL "com.urunc.unikernel.unikernelType"="unikraft"
LABEL "com.urunc.unikernel.hypervisor"="qemu"
```

We can then build the image with the following command:
```
./pun -f Containerfile | sudo buildctl build ... --local context=${PWD} --output type=docker,name=harbor.nbfc.io/nubificus/urunc/nginx-unikraft-qemu:test | sudo docker load
```

The image will get loaded in the local docker registry.
