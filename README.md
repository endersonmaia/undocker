[![godocs.io](http://godocs.io/git.sr.ht/~motiejus/undocker?status.svg)](http://godocs.io/git.sr.ht/~motiejus/undocker)
[![builds.sr.ht status](https://builds.sr.ht/~motiejus/undocker.svg)](https://builds.sr.ht/~motiejus/undocker?)

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

Undocker has no dependencies outside Golang stdlib.

Usage: convert docker image to rootfs
-------------------------------------

Download `busybox` docker image from docker hub and convert it to a rootfs:

```
$ skopeo copy docker://docker.io/busybox:latest docker-archive:busybox.tar
$ undocker busybox.tar - | tar -tv | head -10
drwxr-xr-x 0/0               0 2021-05-17 22:07 bin/
-rwxr-xr-x 0/0         1149184 2021-05-17 22:07 bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/[[ link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/acpid link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/add-shell link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/addgroup link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/adduser link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/adjtimex link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/ar link to bin/[
hrwxr-xr-x 0/0               0 2021-05-17 22:07 bin/arch link to bin/[
```

You can also refer [here][2] for other ways to download Docker images. There
are many.

Converting a [1.1GB Docker image with 77
layers](https://hub.docker.com/r/homeassistant/home-assistant) takes around 4
seconds and on a reasonably powerful Intel laptop.

Usage example: systemd-nspawn
-----------------------------

Start with systemd-nspawn:

```
systemd-nspawn -D $PWD busybox httpd -vfp 8080
```

Usage example: plain old systemd
--------------------------------

```
systemd-run \
  --wait --pty --collect --service-type=exec \
  -p PrivateUsers=true \
  -p DynamicUser=yes \
  -p ProtectProc=invisible \
  -p RootDirectory=$PWD \
  -- busybox httpd -vfp 8080
```

Good things like `PrivateUsers`, `DynamicUser`, `ProtectProc` and other
[systemd protections][1] are available, just like to any systemd unit.

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

- Pull requests (patchsets) with accompanying tests.
- Regression reports.

If you found a container that undocker cannot extract, or extracts incorrectly
and you need this that work with undocker, do not submit an issue: submit a
patchset.

Reports of regression reports must provide examples of "works before" and "does
not work after". Issues without an accompanying patch will most likely be
rejected.

LICENSE
-------

MIT

[1]: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
[2]: https://fly.io/blog/docker-without-docker/
