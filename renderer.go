package galendar

import (
	"fmt"
	"strings"
)

type Renderer interface {
	Name() string
	RenderMonth(cfg Config, month *Calendar) error
	RenderYear(cfg Config, year *Calendar) error
}

var renderers map[string]Renderer

func DefaultRenderer() Renderer {
	return PDFRenderer{}
}

func RegisterRenderer(renderer Renderer) {
	if renderers == nil {
		renderers = map[string]Renderer{}
	}

	renderers[renderer.Name()] = renderer
}

func RendererByName(name string) (Renderer, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	renderer, ok := renderers[name]
	if !ok {
		return nil, fmt.Errorf("invalid name: %q", name)
	}

	return renderer, nil
}
