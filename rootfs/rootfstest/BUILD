load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["rootfstest.go"],
    importpath = "github.com/motiejus/code/undocker/rootfs/rootfstest",
    visibility = ["//visibility:public"],
    deps = ["@com_github_stretchr_testify//require:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["rootfstest_test.go"],
    embed = [":go_default_library"],
    deps = ["@com_github_stretchr_testify//assert:go_default_library"],
)
