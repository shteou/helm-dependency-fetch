# helm-dependency-fetch

A simple (and very hacky) tool to experiment with fetching helm dependencies quickly

The tool reads dependencies from a `Chart.yaml` (API v2) file, or `requirements.yaml` (API v1)
 file and fetches them into the charts folder. 

## Limitations

The tool always fetches the latest indexes, exactly once per index.  
It resolves dependency versions on each run, regardless of whether there are existing charts.  
Only http(s) chart URLs are supported.  
Lock files are not supported/generated.

## Usage

`helm-dependency-fetch` can be installed from one of the [GitHub releases](https://github.com/shteou/helm-dependency-fetch/releases).  
Ensure it's added somewhere on your `PATH` and simply run `helm-dependency-fetch` from your chart folder.

`helm-dependency-fetch` can be used in place of `helm dependency build`. Once it has populated the charts folder, you can use standard
helm tools for the remainder of the workflow.

Note: the chart manifest must exist in the current working directory, it offers no flags for overriding this.

For example, here we have an v1 chart manifest (with separate requirements.yaml file).

```
➜  helm-dependency-fetch git:(master) ✗ ls -al *.yaml
.rw-r--r-- 8.9k stew 27 Jun  1:09 -N Chart.yaml
.rwxr-xr-x 1.6k stew 15 Dec  2020 -N requirements.yaml

➜  helm-dependency-fetch git:(master) ✗ helm-dependency-fetch
Fetching my-service @ >= 1.0.0
Fetching index from https://my-repository.com/
	Fetching chart: https://my-repository.com/charts/my-service-1.0.13.tgz

➜  helm-dependency-fetch git:(master) ✗ helm template my-release .
  ...
```

## Why?

This tool was written because Helm's current behaviour for `helm dependency build` is quite
slow. It's even more slow when using 'unmanaged' repositories.

A short write-up of this behaviour can be found [here](https://stewartplatt.com/blog/speeding-up-helm-dependency-build/).
