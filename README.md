Undocker
--------

Converts a Docker image (a bunch of layers) to a flattened "rootfs" tarball.

Why?
----

Docker images seems to be the lingua franca of distributing application
containers. These are very wide-spread. However, is Docker the best runtime
environment? Not for everyone.

Undocker bridges the gap between application images (in docker image format)
and container runtimes: now you can run a Docker image with systemd-nspawn
and/or lxc, without doing the `docker pull; docker start; docker export` dance.

Usage -- extract docker image
-----------------------------

Download `nginx` docker image from docker hub and convert it to a rootfs:

```
skopeo copy docker://docker.io/nginx:latest docker-archive:nginx.tar
undocker rootfs nginx.tar - | tar -xv
```

(the same can be done with `docker pull` and `docker save`)

Usage -- systemd-nspawn example
-------------------------------

Once the image is converted to a root file-system, it can be started using
classic utilities which expect a rootfs:

```
systemd-nspawn -D $PWD nginx -g 'daemon off;'
```

Usage -- lxc example
--------------------

Preparing the image for use with lxc:

```
undocker rootfs nginx.tar - | xz -T0 > nginx.tar.xz
undocker lxcconfig nginx.tar config
tar -cJf meta.tar.xz config
```

Import it to lxc and run it:

```
lxc-create -n bb -t local -- -m meta.tar.xz -f nginx.tar.xz
lxc-start -F -n bb -s lxc.net.0.type=none
lxc-start -F -n bb -s lxc.net.0.type=none -- /docker-entrypoint.sh nginx -g "daemon off;"
```

Note: automatic entrypoint does not work well with parameters with spaces; not
sure what lxc expects here to make it work.

About the implementation
------------------------

Extracting docker image layers may be harder than you have thought. See
`rootfs/doc.go` for more details.

The rootfs code is dependency-free (it uses Go's stdlib alone). The existing
project dependencies are convenience-only.

Contributions
-------------

I will accept pull request for code (including tests) and documentation. I am
unlikely to react to bug reports without a patch.
