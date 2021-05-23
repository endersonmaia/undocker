def _rootfs_impl(ctx):
    outf = ctx.attr.name + ".tar"
    out = ctx.actions.declare_file(outf)
    ctx.actions.run(
        outputs = [out],
        executable = ctx.files._undocker[0],
        arguments = [
            "rootfs",
            ctx.attr.src.files.to_list()[0].path,
            outf,
        ],
    )
    return [DefaultInfo(
            files = depset([out]),
            runfiles = ctx.runfiles(files = ctx.files.src),
    )]

rootfs = rule(
    doc = "Generate a rootfs from a docker container image",
    implementation = _rootfs_impl,
    attrs = {
        "src": attr.label(
            doc = "Input container tarball",
            mandatory = True,
            allow_single_file = [".tar"],
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

def _temp_impl(ctx):
    for f in ctx.files.data:
        print(f.path)
    ctx.actions.run_shell(
        outputs = [ctx.outputs.out],
        command = "find $@",
    )

temp = rule(
    implementation = _temp_impl,
    attrs = {
        "data": attr.label(mandatory = True, allow_single_file = True),
        "out": attr.output(mandatory = True)
    },
)
