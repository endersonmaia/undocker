load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("//src/undocker:rules.bzl", "rootfs", "temp")
load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_bundle",
)

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/motiejus/code/undocker",
    visibility = ["//visibility:private"],
    deps = [
        "//src/undocker/internal/cmdlxcconfig:go_default_library",
        "//src/undocker/internal/cmdrootfs:go_default_library",
        "@com_github_jessevdk_go_flags//:go_default_library",
    ],
)

go_binary(
    name = "undocker",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

genrule(
    name = "alpine-rootfs",
    srcs = ["@alpine//image:image"],
    outs = ["alpine-rootfs.tar"],
    cmd = "$(location :undocker) rootfs $< $@",
    tools = [":undocker"],
)

# later debugging
pkg_tar(
    name = "alpine-container-plus",
    deps = ["@alpine//image:image"],
)

rootfs(
    name = "alpine-rootfs2",
    #src = "@alpine//image:image",
    src = ":alpine-container-plus",
)

filegroup(
    name = "maingo",
    srcs = ["main.go"],
    visibility = ["//visibility:public"],
)

temp(
    name = "temp",
    data = ":maingo",
    out = "maingo.txt",
)
