def _rootfs_impl(ctx):
    ctx.actions.run(
        executable =  ctx.files._undocker,
        arguments = [
            ctx.attr.src,
            ctx.output,
        ],
    )

rootfs = rule(
    doc = "Generate a rootfs from a docker container image",
    implementation = _rootfs_impl,
    attrs = {
        "src": attr.label(
            doc = "Input container tarball",
            mandatory = True,
            allow_single_file = [".tar"],
        ),
        "out": attr.output(
            doc = "Output rootfs tarball",
            mandatory = True,
        ),
        "_undocker": attr.label(
            doc = "undocker cli; private and may not be overridden",
            cfg = "host",
            default = Label("//src/undocker:undocker"),
            allow_single_file = True,
            executable = True,
        ),
    },
)
