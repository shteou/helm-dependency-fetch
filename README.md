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
slow. It's even more slow when using 'unmanaged' repositories.

A short write-up of this behaviour can be found [here](https://stewartplatt.com/blog/speeding-up-helm-dependency-build/).
