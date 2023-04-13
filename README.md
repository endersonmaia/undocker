[![godocs.io](http://godocs.io/git.jakstys.lt/motiejus/undocker?status.svg)](http://godocs.io/git.jakstys.lt/motiejus/undocker)

Undocker
--------

Converts a Docker image (a bunch of layers) to a flattened "rootfs" tarball.

Why?
----

Docker images became a popular way to distribute applications with their
dependencies; however, Docker is not the best runtime environment. At least not
for everyone. May boring technology run our software.

Undocker bridges the gap between application images (in docker image format)
and application isolation ("container") runtimes: once the docker image is
extracted, it can be run with old-fashioned tools: lxc, systemd-nspawn,
systemd, FreeBSD Jails, and many others.


Installation
------------

Build it like this for the "current" platform:

```
$ make undocker
```

`make -B` will print the extra flags (`-X <...>`) for cross-compiling with
other archs. It's all `go build <...>` in the back, and depends only on Go's
compiler and stdlib.

Usage: convert docker image to rootfs
-------------------------------------

Download `busybox` docker image from docker hub and convert it to a rootfs:

```
$ skopeo copy docker://docker.io/busybox:latest docker-archive:busybox.tar
$ undocker busybox.tar - | tar -xv | sponge | head -10; echo '<...>'
bin/
bin/[
bin/[[
bin/acpid
bin/add-shell
bin/addgroup
bin/adduser
bin/adjtimex
bin/ar
bin/arch
<...>
```

Refer [here][2] for other ways to download Docker images. There are many.

On author's laptop converting a [1.1GB Docker image with 77
layers](https://hub.docker.com/r/homeassistant/home-assistant) takes around 3
seconds and uses ~65MB of residential memory.

Usage example: systemd
----------------------

```
systemd-run \
  --wait --pty --collect --service-type=exec \
  -p RootDirectory=$PWD \
  -p ProtectProc=invisible \
  -p PrivateUsers=true \
  -p DynamicUser=yes \
  -- busybox httpd -vfp 8080
```

[Systemd protections][1] like `PrivateUsers`, `DynamicUser`, `ProtectProc` and
others are available, just like to any systemd unit.

Similar Projects
----------------

* [rootfs_builder](https://github.com/ForAllSecure/rootfs_builder)

Changelog
---------

**v1.0**

* initial release: `rootfs.Flatten` and a simple command-line application.

Contributions
-------------

The following contributions may be accepted:

- Patchsets, with accompanying tests.
- Regression reports.

If you found a container that undocker cannot extract, or extracts incorrectly
and you need this that work with undocker, do not submit an issue: submit a
patchset.

Reports of regression reports must provide examples of "works before" and "does
not work after". Issues without an accompanying patch will most likely be
rejected.

Communication
-------------

Feel free to ping me [directly][motiejus-comms].

LICENSE
-------

MIT

[1]: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
[2]: https://fly.io/blog/docker-without-docker/

[motiejus-comms]: https://jakstys.lt/contact/
