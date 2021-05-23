load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_docker//container:container.bzl", "container_bundle")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "git.sr.ht/~motiejus/code/undocker",
    visibility = ["//visibility:private"],
    deps = [
        "//src/undocker/rootfs:go_default_library",
        "@com_github_jessevdk_go_flags//:go_default_library",
    ],
)

go_binary(
    name = "undocker",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["main_test.go"],
    embed = [":go_default_library"],
    deps = [
        "@com_github_jessevdk_go_flags//:go_default_library",
        "@com_github_stretchr_testify//assert:go_default_library",
        "@com_github_stretchr_testify//require:go_default_library",
    ],
)
