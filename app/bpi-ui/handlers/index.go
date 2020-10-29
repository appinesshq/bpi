package handlers

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type indexGroup struct {
	tmpl    *template.Template
	auth    *Auth
	apiHost string
}

func newIndex(tmpl string, apiHost string, auth *Auth) (indexGroup, error) {
	rawTmpl, err := ioutil.ReadFile(tmpl)
	if err != nil {
		return indexGroup{}, errors.Wrap(err, "reading index page")
	}

	t := template.New("index")
	if _, err := t.Parse(string(rawTmpl)); err != nil {
		return indexGroup{}, errors.Wrap(err, "creating template")
	}

	ig := indexGroup{
		tmpl:    t,
		auth:    auth,
		apiHost: apiHost,
	}

	return ig, nil
}

func (ig *indexGroup) handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var markup bytes.Buffer
	vars := map[string]interface{}{
		"AuthToken": ig.auth.Token,
		"ApiHost":   ig.apiHost,
	}
	if err := ig.tmpl.Execute(&markup, vars); err != nil {
		return errors.Wrapf(err, "executing template")
	}

	io.Copy(w, &markup)
	return nil
}
