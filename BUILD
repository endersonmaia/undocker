load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "undocker_lib",
    srcs = ["main.go"],
    importpath = "github.com/motiejus/code/undocker",
    visibility = ["//visibility:private"],
    deps = ["@com_github_jessevdk_go_flags//:go-flags"],
)

go_binary(
    name = "undocker",
    embed = [":undocker_lib"],
    visibility = ["//visibility:public"],
)
