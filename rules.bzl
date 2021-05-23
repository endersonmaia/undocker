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
        executable = ctx.files._undocker[0],
        arguments = [
            "rootfs",
            ctx.files.src[0].path,
            out.path,
        ],
        mnemonic = "RootFS",
    )
    return [DefaultInfo(
        files = depset([out]),
        runfiles = ctx.runfiles(files = ctx.files.src),
    )]

rootfs = rule(
    doc = "Generate a rootfs from a docker container image",
    implementation = _rootfs_impl,
    attrs = {
        "src": _input_container,
        "_undocker": _undocker,
    },
)

def _lxcconfig_impl(ctx):
    out = ctx.actions.declare_file(ctx.attr.name + ".conf")
    ctx.actions.run(
        outputs = [out],
        inputs = ctx.files.src,
        executable = ctx.files._undocker[0],
        arguments = [
            "lxcconfig",
            ctx.files.src[0].path,
            out.path,
        ],
        mnemonic = "LXCConfig",
    )
    return [DefaultInfo(
        files = depset([out]),
        runfiles = ctx.runfiles(files = ctx.files.src),
    )]

lxcconfig = rule(
    doc = "Generate lxc config from a docker container image",
    implementation = _lxcconfig_impl,
    attrs = {
        "src": _input_container,
        "_undocker": _undocker,
    },
)
