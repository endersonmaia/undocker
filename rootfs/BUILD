load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "rootfs.go",
        "tree.go",
    ],
    importpath = "github.com/motiejus/code/undocker/rootfs",
    visibility = ["//visibility:public"],
    deps = ["@org_uber_go_multierr//:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "rootfs_test.go",
        "tree_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//src/undocker/rootfs/rootfstest:go_default_library",
        "@com_github_stretchr_testify//assert:go_default_library",
        "@com_github_stretchr_testify//require:go_default_library",
    ],
)
