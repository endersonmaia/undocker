load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("//src/undocker:rules.bzl", "rootfs")
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

rootfs(
    name = "alpine-rootfs",
    src = "@alpine//image:image",
)
