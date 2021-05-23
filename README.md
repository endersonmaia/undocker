Undocker
--------

Converts a Docker image (a bunch of layers) to a flattened "rootfs" tarball.

Why?
----

Docker images became a popular way to distribute applications with their
dependencies. However, Docker itself is not the best runtime environment. At
least not for everyone.

Undocker bridges the gap between application images (in docker image format)
and container runtimes: now you can run a Docker image with old-fashioned
tools: lxc, systemd-nspawn or systemd itself.

Usage -- extract docker image
-----------------------------

Download `busybox` docker image from docker hub and convert it to a rootfs:

```
skopeo copy docker://docker.io/busybox:latest docker-archive:busybox.tar
undocker busybox.tar - | tar -xv
```

Almost the same can be done with a combination of `docker pull` and `docker
save`.

Usage -- systemd-nspawn example
-------------------------------

Start with systemd-nspawn:

```
systemd-nspawn -D $PWD busybox httpd -vfp 8080
```

Usage -- plain old systemd
--------------------------

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

Notes & gotchas
---------------

`unocker` does not magically enable you to run containers from the internet.
Many will need significant tuning or not work at all; one will still need to
understand [what's inside](https://xkcd.com/1988/).

Contributions
-------------

I want this project to be useful for others, but not become a burden for me. If
undocker fails for you (for example, you found a container that undocker cannot
extract, or extracts incorrectly), **you** are on the hook to triage and fix
it.

Therefore, the following contributions are welcome:

- Pull rquests (diffs) with accompanying tests.
- Documentation.

Issues without accompanying patches will most likely be rejected, with one
exception: reports about regressions do not have to contain patches, but a
failing commit is mandatory, and a failing test case is highly appreciated.

[1]: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
