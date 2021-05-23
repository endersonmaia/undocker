load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "rootfs_lib",
    srcs = ["rootfs.go"],
    importpath = "github.com/motiejus/code/undocker/rootfs",
    visibility = ["//visibility:public"],
)

go_test(
    name = "rootfs_test",
    srcs = ["rootfs_test.go"],
    embed = [":rootfs"],
    deps = [
        "@com_github_stretchr_testify//assert",
        "@com_github_stretchr_testify//require",
    ],
)
