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

_lxcconfig = rule(
    _lxcconfig_impl,
    doc = "Generate lxc config from a docker container image",
    attrs = {
        "src": _input_container,
        "_undocker": _undocker_cli,
    },
)

def lxcconfig(name, src):
    _lxcconfig(name = name+"/config", src = src)
    pkg_tar(
        name = name + "txz",
        extension = "tar.xz",
        srcs = [name+"/config"],
        remap_paths = {
            name: "",
        },
    )
