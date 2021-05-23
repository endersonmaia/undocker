load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["cmdrootfs.go"],
    importpath = "github.com/motiejus/code/undocker/internal/cmdrootfs",
    visibility = ["//src/undocker:__subpackages__"],
    deps = [
        "//src/undocker/rootfs:go_default_library",
        "@com_github_jessevdk_go_flags//:go_default_library",
        "@org_uber_go_multierr//:go_default_library",
    ],
)
