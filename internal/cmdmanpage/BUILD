load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["cmdmanpage.go"],
    importpath = "git.sr.ht/~motiejus/code/undocker/internal/cmdmanpage",
    visibility = ["//src/undocker:__subpackages__"],
    deps = ["@com_github_jessevdk_go_flags//:go_default_library"],
)
