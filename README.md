# Clang Tidy Cache

A fairly simple wrapper application around the clang-tidy executable. It will attempt to fingerprint each source invocation and store the results in a local user cache. This can be useful when building software projects of a reasonable scale.

## Configuration

In order to keep the wrapper reasonably clean, the user will have to write a configuration file at the following location:

`~/.ctcache/config.json`

The configuration file contains the information about where `clang-tidy-cache` can find the real `clang-tidy` executable. Here is an example below:

```json
{
  "clang_tidy_path": "/usr/local/Cellar/llvm@6/6.0.1_1/bin/clang-tidy"
}
```

## Installing

To get the latest version checkout the releases page on github:

https://github.com/ejfitzgerald/clang-tidy-cache/releases
