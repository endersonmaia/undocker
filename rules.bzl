load("@rules_pkg//:pkg.bzl", "pkg_tar")

_undocker_cli = attr.label(
    doc = "undocker cli; private and may not be overridden",
    cfg = "host",
    default = Label("//src/undocker:undocker"),
    allow_single_file = True,
    executable = True,
)
_input_container = attr.label(
    doc = "Input container tarball",
    mandatory = True,
    allow_single_file = [".tar"],
)

def _rootfs_impl(ctx):
    out = ctx.outputs.out
    if out == None:
        out = ctx.actions.declare_file(ctx.attr.name + ".tar")
    ctx.actions.run(
        outputs = [out],
        inputs = ctx.files.src,
        executable = ctx.executable._undocker,
        arguments = [
            "rootfs",
            ctx.files.src[0].path,
            out.path,
        ],
        mnemonic = "RootFS",
    )
    return DefaultInfo(
        files = depset([out]),
        runfiles = ctx.runfiles(files = ctx.files.src),
    )

rootfs = rule(
    _rootfs_impl,
    doc = "Generate a rootfs from a docker container image",
    attrs = {
        "src": _input_container,
        "_undocker": _undocker_cli,
        "out": attr.output(),
    },
)

def _lxcconfig_impl(ctx):
    out = ctx.actions.declare_file(ctx.attr.name)
    ctx.actions.run(
        outputs = [out],
        inputs = ctx.files.src,
        executable = ctx.executable._undocker,
        arguments = [
            "lxcconfig",
            ctx.files.src[0].path,
            out.path,
        ],
        mnemonic = "LXCConfig",
    )
    return DefaultInfo(
        files = depset([out]),
        runfiles = ctx.runfiles(files = ctx.files.src),
    )

def lxcbundle(name, src, version):
    if type(version) != "int":
        fail("version must be an int, got {}".format(type(version)))
    rootfsname = name + "/_/rootfs"
    rootfs(name = rootfsname, src = src, out = rootfsname + ".tar")
    lxcconfig(name, src = src, out = name + "/_/meta.tar.xz")
    native.genrule(
        name = name + "-rootfs",
        srcs = [rootfsname],
        outs = [rootfsname + ".tar.xz"],
        cmd = "xz -T0 -f $< > $@",
        message = "XZ",
    )
    outname = "{}.{}.tar".format(name, version)
    pkg_tar(
        name = name,
        srcs = [
            name + "-rootfs",
            name + "-meta",
        ],
        out = outname,
    )

_lxcconfig = rule(
    _lxcconfig_impl,
    doc = "Generate lxc config from a docker container image",
    attrs = {
        "src": _input_container,
        "_undocker": _undocker_cli,
    },
)

def lxcconfig(name, src, out = None):
    _lxcconfig(name = name + "/_/config", src = src)
    pkg_tar(
        name = name + "-meta",
        extension = "tar.xz",
        srcs = [name + "/_/config"],
        remap_paths = {
            name: "",
        },
        out = out or "meta.tar.xz",
    )
