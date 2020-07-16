package main

import (
	"image/png"
	"io/ioutil"
	"os"
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
		maintainer := annotations["workflows.argoproj.io/maintainer"].(string)
		if !strings.HasPrefix(maintainer, "@") {
			panic("invalid maintainer, must be Github username starting with \"@\": " + maintainer)
		}
		url := "https://raw.githubusercontent.com/argoproj-labs/argo-workflows-catalog/master/" + filename
		cards = append(cards, `<div class="card box-shadow">
    <div class="card-header">
        <div class="media-body">
            <h3 class="card-title">`+name+`</h3>
            <p class="card-subtitle text-muted">By <a href="https://github.com/`+strings.TrimSuffix(maintainer, "@")+`">`+maintainer+`</a> <span class="badge badge-light">`+version+`</span></p>
        </div>
    </div>
  <img class="card-img-top" src="`+name+`/icon.png" alt="Icon">
  <div class="card-body">
    <p class="card-text">`+description+`</p>
    <p class="card-text"><a href="`+url+`" class='btn btn-light'>Get</a></p>
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
		if img.Bounds().Max.X > 320 || img.Bounds().Max.Y > 180 {
			tall := img.Bounds().Max.Y > img.Bounds().Max.X
			if tall {
				img = resize.Resize(320, 0, img, resize.Lanczos3)
			} else {
				img = resize.Resize(0, 180, img, resize.Lanczos3)
			}
			img, err = cutter.Crop(img, cutter.Config{Width: 320, Height: 180, Mode: cutter.Centered})
			if err != nil {
				panic("failed to crop icon: " + err.Error())
			}
			err := os.Mkdir("docs/"+name, 0777)
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
	}

	err = ioutil.WriteFile("docs/index.html", []byte(`<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" integrity="sha384-wvfXpqpZZVQGK6TAh5PVlGOfQNHSoD2xbE+QkPxCAFlNEevoEH3Sl0sibVcOQVnN" crossorigin="anonymous">

    <title>Argo Workflows Catalog</title>
  </head>
  <body>
    <div class="pricing-header px-3 py-3 pt-md-5 pb-md-4 mx-auto text-center">
      <h1 class="display-4">
        Template Catalog
      </h1>
      <p class="lead">Free reusable templates for Argo Workflows.</p>
    </div>
    <div class="container">
      <div class="card-deck text-center">
        `+strings.Join(cards, "")+`
      </div>
    </div>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
  </body>
</html>`), 0644)
	if err != nil {
		panic("failed to save index: " + err.Error())
	}
}
