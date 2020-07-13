package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Masterminds/semver"
	"sigs.k8s.io/yaml"
)

//https://getbootstrap.com/docs/4.0/examples/pricing/
type obj = map[string]interface{}

func main() {
	dir, err := ioutil.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	var cards []string
	for _, f := range dir {
		name := f.Name()
		filename := "templates/" + name + "/manifests.yaml"
		println(filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			panic("failed to read manifests:" + err.Error())
		}
		manifest := make(obj)
		err = yaml.Unmarshal(data, &manifest)
		if err != nil {
			panic("invalid YAML:" + err.Error())
		}
		// format and normalise the manifest
		data, err = yaml.Marshal(manifest)
		if err != nil {
			panic("failed to marshall YAML:" + err.Error())
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic("failed to same manifest:" + err.Error())
		}
		kind := manifest["kind"].(string)
		if kind != "WorkflowTemplate" {
			panic(name + " kind must be WorkflowTemplate")
		}
		annotations := manifest["metadata"].(obj)["annotations"].(obj)
		version := annotations["workflows.argoproj.io/version"].(string)
		_, err = semver.NewConstraint(version)
		if err != nil {
			panic("invalid version (must be semver constraint): " + err.Error())
		}
		description := annotations["workflows.argoproj.io/description"].(string)
		if !strings.HasSuffix(description, ".") {
			panic("description must end with period")
		}
		tags := strings.Split(annotations["workflows.argoproj.io/tags"].(string), ",")
		if len(tags) == 1 && tags[0] == "" {
			panic("must have at least one tag")
		}
		icon := annotations["workflows.argoproj.io/icon"].(string)
		maintainer := annotations["workflows.argoproj.io/maintainer"].(string)
		if !strings.Contains(maintainer, "@") {
			panic("invalid maintainer, must be email address: " + maintainer)
		}
		cards = append(cards, fmt.Sprintf(`<div class="card box-shadow">
    <div class="card-header d-flex flex-row">
        <div>
        	<h1><i class='fa'>%s</i></h1>
        </div>
        <div class="media-body">
            <h3 class="card-title">%s</h3>
            <p class="card-subtitle text-muted">By %s <span class="badge badge-light">%s</span></p>
        </div>
    </div>
  <div class="card-body">
    <p class="card-text">%s</p>
    <a href="%s" class="btn btn-light">Download</a>
  </div>
</div>`, icon, name, maintainer, version, description,  "https://raw.githubusercontent.com/alexec/argo-workflows-catalog/master/"+filename))

	}

	err = ioutil.WriteFile("docs/index.html", []byte(fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" integrity="sha384-wvfXpqpZZVQGK6TAh5PVlGOfQNHSoD2xbE+QkPxCAFlNEevoEH3Sl0sibVcOQVnN" crossorigin="anonymous">

    <title>%s</title>
  </head>
  <body>
    <div class="pricing-header px-3 py-3 pt-md-5 pb-md-4 mx-auto text-center">
      <h1 class="display-4">Template Catalog</h1>
      <p class="lead">Free reusable templates for Argo Workflows.</p>
    </div>
    <div class="container">
      <div class="card-deck text-center">
        %s
      </div>
    </div>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
  </body>
</html>`, "Argo Workflow Catalog", strings.Join(cards, ""))), 0644)
	if err != nil {
		panic("failed to save index: " + err.Error())
	}
}
