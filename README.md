# helm-dependency-fetch

A simple (and very hacky) tool to experiment with fetching helm dependencies quickly

The tool reads dependencies from a `Chart.yaml` (API v2) file, or `requirements.yaml` (API v1)
 file and fetches them into the charts folder. 

## Limitations

The tool always fetches the latest indexes, exactly once. It resolves dependency versions on
each run, regardless of whether there are existing charts. It does not support chart
conditions to optimise downloading.

## Why?

This tool was written because Helm's current behaviour for `helm dependency build` is quite
slow. From a rough analysis, it appears to:

1) Pre-fetch all required index files
2) Iterate over each dependency, resolving the latest compatible version, and downloading it

Both of these stages appear to be sub-optimal. The initial pre-fetch of index files appears
to try to parallelise fetching indexes, but don't de-duplicate them. Then, after resolving
the target dependency versions, the Download Manager appears to re-fetch indexes for _each_
dependency, ignoring the previously fetched index cache.

For charts with more than a few dependencies, this can easily lead to execution times of
upwards of a minute, whereas this tool is able to build the charts folder in a fraction of
that. In an example with 14 dependencies, from 2 different repositories, execution time
went from ~70s down to about 10s.
