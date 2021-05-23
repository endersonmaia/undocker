Undocker
--------

Convert a Docker image to a root file system.

Why?
---

Docker images seems to be the lingua franca of distributing application
containers. It's hell convenient. But is it the best runtime environment? Not
for everyone.

Undocker bridges the gap between image configurations and container runtimes:
now you can run a Docker image with systemd-nspawn and/or lxc.

Usage -- extract docker image
-----------------------------

Download `nginx` docker image from docker hub and convert it to a rootfs:
```
skopeo copy docker://docker.io/nginx:latest docker-archive:nginx.tar
undocker rootfs nginx.tar - | tar -xv
```

Usage -- systemd-nspawn example
-------------------------------

```
systemd-nspawn -D $PWD nginx -g 'daemon off;'
```

Usage -- lxc example
--------------------

Converting and creating the archive:

```
undocker rootfs nginx.tar - | xz -T0 > nginx.tar.xz
undocker lxcconfig nginx.tar config
tar -cJf meta.tar.xz config
```

Importing it to lxc and running it:

```
lxc-create -n bb -t local -- -m meta.tar.xz -f nginx.tar.xz
lxc-start -F -n bb -s lxc.net.0.type=none
lxc-start -F -n bb -s lxc.net.0.type=none -- /docker-entrypoint.sh nginx -g "daemon off;"
```

Note: automatic entrypoint does not work well with parameters with spaces; not
sure what lxc expects here to make it work.

Contributions
-------------

These are the contributions I will accept:

- pull requests for code.
- documentation updates.

I am very unlikely to react to bug reports (even if they are legit) without
accopmanying pull requests.
