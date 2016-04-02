package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

var templateFunctions template.FuncMap

func setupTemplates(router *gin.Engine) {
	templateFunctions = template.FuncMap{
		"timeago": func(t time.Time) template.HTML {
			return timeToHTML(&t)
		},

		"timeagoNil": func(t *time.Time) template.HTML {
			if t == nil {
				return template.HTML("<span class=\"text-muted\"><em>never</em></span>")
			}

			return timeToHTML(t)
		},

		"visicon": func(vis string, addClass string) template.HTML {
			cls := ""

			switch vis {
			case "private":
				cls = "lock"
			case "internal":
				cls = "shield"
			case "public":
				fallthrough
			default:
				cls = "globe"
			}

			return template.HTML("<i class=\"fa fa-" + cls + " visibility-" + vis + " " + addClass + "\"></i>")
		},

		"fileURI": func(pm *postMedata) string {
			return fileURI(pm, false)
		},

		"rawURI": func(pm *postMedata) string {
			return fileURI(pm, true)
		},

		"filesize": func(size int) string {
			return formatFilesize(uint64(size))
		},

		"shorten": func(value string, maxlen int) string {
			length := len(value)

			if length <= maxlen {
				return value
			}

			halfs := maxlen / 2
			runes := []rune(value)

			return fmt.Sprintf("%s…%s", string(runes[:halfs]), string(runes[(length-halfs):]))
		},
	}

	pattern := filepath.Join(config.Directories.Resources, "templates", "*")

	if config.Environment == gin.ReleaseMode {
		router.HTMLRender = render.HTMLProduction{Template: loadTemplate(pattern)}
	} else {
		router.HTMLRender = debugTemplateRenderer{Glob: pattern}
	}
}

type debugTemplateRenderer struct {
	Glob string
}

func (r debugTemplateRenderer) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: r.loadTemplate(),
		Name:     name,
		Data:     data,
	}
}

func (r debugTemplateRenderer) loadTemplate() *template.Template {
	return loadTemplate(r.Glob)
}

func loadTemplate(pattern string) *template.Template {
	t := template.New("")
	t.Funcs(templateFunctions)

	return template.Must(t.ParseGlob(pattern))
}

func timeToHTML(t *time.Time) template.HTML {
	iso := t.Format("2006-01-02T15:04:05-0700")
	pretty := t.Format("Mon, Jan 2 2006 15:04")

	return template.HTML("<time class=\"rel\" datetime=\"" + iso + "\">" + pretty + "</time>")
}
