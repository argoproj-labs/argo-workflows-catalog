package main

import (
	"image/png"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"sigs.k8s.io/yaml"
)

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
		kind := manifest["kind"].(string)
		if kind != "WorkflowTemplate" {
			panic(name + " kind must be WorkflowTemplate")
		}
		annotations := manifest["metadata"].(obj)["annotations"].(obj)
		version := strings.TrimSpace(annotations["workflows.argoproj.io/version"].(string))
		_, err = semver.NewConstraint(version)
		if err != nil {
			panic("invalid version (must be semver constraint): " + err.Error())
		}
		description := strings.TrimSpace(annotations["workflows.argoproj.io/description"].(string))
		if !strings.HasSuffix(description, ".") {
			panic("description must end with period: " + description)
		}
		tags := strings.Split(annotations["workflows.argoproj.io/tags"].(string), ",")
		sort.Strings(tags)
		if len(tags) == 1 && tags[0] == "" {
			panic("must have at least one tag")
		}
		maintainer := annotations["workflows.argoproj.io/maintainer"].(string)
		if !strings.HasPrefix(maintainer, "@") {
			panic("invalid maintainer, must be Github username starting with \"@\": " + maintainer)
		}
		url := "https://raw.githubusercontent.com/argoproj-labs/argo-workflows-catalog/master/" + filename
		cards = append(cards, `<div class="col-sm-4">
<div class="shadow p-3 mb-5 bg-white rounded">
    <h4><i class='fa fa-sitemap'></i> `+name+`</h4>
    <p class="text-muted">By <a href="https://github.com/`+strings.TrimPrefix(maintainer, "@")+`">`+maintainer+`</a> <span class="badge badge-light">`+version+`</span></p>
    <p style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">`+description+`</p>
	<p>`+formatTags(tags)+`</p>
    <div><button type="button" class="btn btn-light" data-toggle="modal" data-target="#`+name+`Modal">Get <i class="fa fa-angle-right"></i></button></div>
<div class="modal" id="`+name+`Modal" tabindex="-1" aria-labelledby="`+name+`Label" aria-hidden="true">
  <div class="modal-dialog modal-dialog-scrollable modal-lg">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title" id="`+name+`Label"><i class='fa fa-sitemap'></i> `+name+`</h5>
        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
          <span aria-hidden="true">&times;</span>
        </button>
      </div>
      <div class="modal-body">`+description+`</div>
      <div class="modal-body"><pre><code>kubectl apply -f `+url+`</code></pre></div>
      <div class="modal-body"><pre><code>`+string(data)+`</code></pre></div>
    </div>
  </div>
</div>
</div>
</div>`)

		icon, err := os.Open("templates/" + name + "/icon.png")
		if err != nil {
			panic("failed to open icon:" + err.Error())
		}
		img, err := png.Decode(icon)
		if err != nil {
			panic("failed to decode icon:" + err.Error())
		}
		err = icon.Close()
		if err != nil {
			panic("failed to close icon: " + err.Error())
		}
		max := img.Bounds().Max
		tall := max.X/max.Y > 2
		if tall {
			img = resize.Resize(0, 160, img, resize.Lanczos3)
		} else {
			img = resize.Resize(320, 0, img, resize.Lanczos3)
		}
		img, err = cutter.Crop(img, cutter.Config{Width: 320, Height: 160, Mode: cutter.Centered})
		if err != nil {
			panic("failed to crop icon: " + err.Error())
		}
		err = os.Mkdir("docs/"+name, 0777)
		if err != nil && !os.IsExist(err) {
			panic("failed to create directory: " + err.Error())
		}
		out, err := os.Create("docs/" + name + "/icon.png")
		if err != nil {
			panic("failed to create icon: " + err.Error())
		}
		err = png.Encode(out, img)
		if err != nil {
			panic("failed to encode icon: " + err.Error())
		}
		err = out.Close()
		if err != nil {
			panic("failed to close icon: " + err.Error())
		}
	}

	err = ioutil.WriteFile("docs/index.html", []byte(`<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css" integrity="sha384-B0vP5xmATw1+K9KRQjQERJvTumQW0nPEzvF6L/Z6nronJ3oUOFUFpCjEUQouq2+l" crossorigin="anonymous">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" integrity="sha384-wvfXpqpZZVQGK6TAh5PVlGOfQNHSoD2xbE+QkPxCAFlNEevoEH3Sl0sibVcOQVnN" crossorigin="anonymous">

    <title>Argo Workflows Catalog</title>
  </head>
  <body>
	<nav class="navbar bg-dark">
	  <div class="container-fluid">
		<a class="navbar-brand" style="color:white">
			<i class="fa fa-bookmark"></i>
			<b>Template Catalog</b>: Free reusable templates for Argo Workflows.
		</a>
		<div class="d-flex">
			<a href="https://github.com/argoproj-labs/argo-workflows-catalog/blob/master/README.md" class='btn btn-light'>Contribute <i class="fa fa-angle-right"></i></a>
		</div>
	  </div>
	</nav>
    <div class="container p-5">
      <div class="row">
        `+strings.Join(cards, "")+`
      </div>
    </div>
<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js" integrity="sha384-DfXdz2htPH0lsSSs5nCTpuj/zy4C+OGpamoFVy38MVBnE+IbbVYUew+OrCXaRkfj" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/js/bootstrap.bundle.min.js" integrity="sha384-Piv4xVNRyMGpqkS2by6br4gNJ7DXjqk09RmUpJ8jgGtD7zP9yug3goQfGII0yAns" crossorigin="anonymous"></script>
  </body>
</html>`), 0644)
	if err != nil {
		panic("failed to save index: " + err.Error())
	}
}

func formatTags(tags []string) string {
	html := ""
	for _, tag := range tags {
		html += `<span class="badge badge-secondary m-1">` + tag + `</span>`
	}
	return html
}
