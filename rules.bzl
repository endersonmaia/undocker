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

def lxcbundle(name, src, version, overlay_tars = []):
    if type(version) != "int":
        fail("version must be an int, got {}".format(type(version)))
    nameversion = "{}.{}".format(name, version)
    stage0 = nameversion + "/_stage0/rootfs"
    stage1 = nameversion + "/_stage1/rootfs"
    rootfs(
        name = stage0,
        src = src,
        out = stage0 + ".tar",
    )
    pkg_tar(
        name = stage1,
        deps = [stage0] + overlay_tars,
        extension = "tar.xz",
        out = stage1 + ".tar.xz",
    )
    lxcconfig(name, version, src = src, out = name + "/_stage1/meta.tar.xz")

    pkg_tar(
        name = nameversion,
        srcs = [
            nameversion + "/_stage1/rootfs",
            nameversion + "/_stage1/meta",
        ],
        out = "{}.tar".format(nameversion),
    )

_lxcconfig = rule(
    _lxcconfig_impl,
    doc = "Generate lxc config from a docker container image",
    attrs = {
        "src": _input_container,
        "_undocker": _undocker_cli,
    },
)

def lxcconfig(name, version, src, out = None):
    nameversion = "{}.{}".format(name, version)
    _lxcconfig(name = nameversion + "/_stage0/config", src = src)
    pkg_tar(
        name = nameversion + "/_stage1/meta",
        extension = "tar.xz",
        srcs = [nameversion + "/_stage0/config"],
        remap_paths = {
            name: "",
        },
        out = out or "meta.tar.xz",
    )
