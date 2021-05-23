load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["cmdlxcconfig.go"],
    importpath = "github.com/motiejus/code/undocker/internal/cmdlxcconfig",
    visibility = ["//src/undocker:__subpackages__"],
    deps = [
        "//src/undocker/lxcconfig:go_default_library",
        "@com_github_jessevdk_go_flags//:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["cmdlxcconfig_test.go"],
    embed = [":go_default_library"],
    deps = [
        "@com_github_jessevdk_go_flags//:go_default_library",
        "@com_github_stretchr_testify//assert:go_default_library",
        "@com_github_stretchr_testify//require:go_default_library",
    ],
)
