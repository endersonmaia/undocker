load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "rootfs_lib",
    srcs = [
        "rootfs.go",
        "rootfs_tests.go",
    ],
    importpath = "github.com/motiejus/code/undocker/rootfs",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_stretchr_testify//assert",
        "@com_github_stretchr_testify//require",
    ],
)
