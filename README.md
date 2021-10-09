# helm-dependency-fetch

A simple (and very hacky) tool to experiment with fetching helm dependencies quickly

The tool reads dependencies from a `Chart.yaml` (API v2) file, or `requirements.yaml` (API v1)
 file and fetches them into the charts folder. 

## Limitations

The tool always fetches the latest indexes, exactly once per index.  
It resolves dependency versions on each run, regardless of whether there are existing charts.  
Only http(s) and file based URL schemes are supported.  
Lock files are not generated.

## Usage

`helm-dependency-fetch` can be installed from one of the [GitHub releases](https://github.com/shteou/helm-dependency-fetch/releases).  
Ensure it's added somewhere on your `PATH` and simply run `helm-dependency-fetch` from your chart folder.

`helm-dependency-fetch` is used in place of `helm dependency build`. Once it has populated the charts folder, you can use standard
helm tools for the remainder of the workflow.

```
➜  helm-dependency-fetch git:(master) ✗ helm-dependency-fetch -h
helm-dependency-fetch [chartDir]
  helm-dependency-fetch is a drop in replacement for helm dependency build
  It will fetch the chart dependencies for the supplied chart. If no chart is supplied
  the current directory is assumed to be the chart directory
  Note, lock files are not generated, and are ignored.
  The tool will fetch the latest dependencies on each execution.

Flags:
  -h, --help  help for this command
```

In this example we use helm dependency fetch to gather the dependencies of a v2 chart and then use helm template.

```
➜  helm-dependency-fetch git:(master) ✗ ls -al *.yaml
.rw-r--r-- 151 stew 27 Jun  1:09 -N Chart.yaml

➜  helm-dependency-fetch git:(master) ✗ helm-dependency-fetch
Fetching index from https://my-repository.com/
	Fetching chart: https://my-repository.com/charts/my-service-1.0.13.tgz

➜  helm-dependency-fetch git:(master) ✗ helm template my-release .
  ...
```

Or similarly with a v1 chart, with a separate requirements file, in another directory

```
➜  helm-dependency-fetch git:(master) ✗ ls -al my-chart/*.yaml
.rw-r--r--  84 stew 27 Jun  1:01 -N Chart.yaml
.rwxr-xr-x 146 stew 15 Dec  2020 -N requirements.yaml

➜  helm-dependency-fetch git:(master) ✗ helm-dependency-fetch my-chart
Fetching my-service @ >= 1.0.0
Fetching index from https://my-repository.com/
	Fetching chart: https://my-repository.com/charts/my-service-1.0.13.tgz

➜  helm-dependency-fetch git:(master) ✗ helm template my-release my-chart
  ...
```

## Why?

This tool was written because Helm's current behaviour for `helm dependency build` is quite
slow. It's even more slow when using 'unmanaged' repositories.

A short write-up of this behaviour can be found [here](https://stewartplatt.com/blog/speeding-up-helm-dependency-build/).
