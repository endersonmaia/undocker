def _rootfs_impl(ctx):
    output = ctx.outputs.output
    ctx.actions.run(
        outputs = [output],
        executable = ctx.files._undocker[0],
        arguments = [
            ctx.attr.src.path,
            output.path,
        ],
    )

rootfs = rule(
    doc = "Generate a rootfs from a docker container image",
    implementation = _rootfs_impl,
    outputs = {
        "out": "%{name}.tar",
    },
    attrs = {
        "src": attr.label(
            doc = "Input container tarball",
            mandatory = True,
            allow_single_file = [".tar"],
        ),
        "output": attr.output(
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
