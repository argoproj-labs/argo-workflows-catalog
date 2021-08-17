# Contributing

## Adding a New Item to the Catalog

Install Golang.

Fork the repository, then get the code:

```bash
go get github.com/argoproj-labs/argo-workflows-catalog
```

Copy the "hello world" template:

```bash
cd ~/go/src/github.com/argoproj-labs/argo-workflows-catalog
cp -R templates/hello-world templates/my-template
```

Edit `templates/my-template/manifests.yaml` to add your template. Write a breif description of your contribution in the `metadata.annotations` section of your template.


Run `make`:

```bash
make
```

Commit your changes and make a pull request.