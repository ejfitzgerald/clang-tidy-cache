# Clang Tidy Cache

A fairly simple wrapper application around the clang-tidy executable. It will attempt to fingerprint each source invocation and store the results in a local user cache. This can be useful when building software projects of a reasonable scale.

## Configuration

By default, the wrapper will look for the `clang-tidy` executable on the path. This can be changed by setting the `CLANG_TIDY_CACHE_BINARY` environment variable, or by writing a configuration file at the following location:

`~/.ctcache/config.json`

The configuration file contains the information about where `clang-tidy-cache` can find the real `clang-tidy` executable. Here is an example below:

```json
{
  "clang_tidy_path": "/usr/local/Cellar/llvm@6/6.0.1_1/bin/clang-tidy"
}
```

By default, the cache is stored in a filesystem under `~/.ctcache/cache`. This can be changed by setting `CLANG_TIDY_CACHE_DIR` environment variable.

For easy integration in your CI system, set `CLANG_TIDY_CACHE_DIR` to a directory that you can share across your pipelines.

## Installing

To get the latest version checkout the releases page on github:

https://github.com/ejfitzgerald/clang-tidy-cache/releases
