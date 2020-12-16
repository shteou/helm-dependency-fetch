# helm-dependency-fetch

A simple (and very hacky) tool to experiment with fetching helm dependencies quickly

The tool reads dependencies from a helm (v2) `requirements.yaml` file and fetches
them into the charts folder. It re-resolved versions each time.
