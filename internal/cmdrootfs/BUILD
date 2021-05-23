load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["cmdrootfs.go"],
    importpath = "github.com/motiejus/code/undocker/internal/cmdrootfs",
    visibility = ["//src/undocker:__subpackages__"],
    deps = [
        "//src/undocker/rootfs:go_default_library",
        "@com_github_jessevdk_go_flags//:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["cmdrootfs_test.go"],
    embed = [":go_default_library"],
    deps = [
        "@com_github_jessevdk_go_flags//:go_default_library",
        "@com_github_stretchr_testify//assert:go_default_library",
        "@com_github_stretchr_testify//require:go_default_library",
    ],
)
