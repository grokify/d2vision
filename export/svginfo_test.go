package export

import "testing"

func TestInspectSVG(t *testing.T) {
	tests := []struct {
		name string
		svg  string
		want SVGInfo
		err  bool
	}{
		{"width height and SMIL", `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="1000"><circle><animate dur="6.5s"/><animateMotion dur="3s"/></circle></svg>`,
			SVGInfo{800, 1000, 6.5}, false},
		{"px units", `<svg width="640px" height="480px"></svg>`, SVGInfo{640, 480, 0}, false},
		{"viewBox fallback", `<svg width="100%" viewBox="0 0 400.4 300"></svg>`, SVGInfo{400, 300, 0}, false},
		{"milliseconds", `<svg width="10" height="10"><animate dur="750ms"/></svg>`, SVGInfo{10, 10, 0.75}, false},
		{"indefinite ignored", `<svg width="10" height="10"><animate dur="indefinite"/></svg>`, SVGInfo{10, 10, 0}, false},
		{"css animation", `<svg width="10" height="10"><style>.e{animation-duration: 1.2s}</style></svg>`, SVGInfo{10, 10, 1.2}, false},
		{"css shorthand as emitted by D2", `<svg width="10" height="10"><style>.e{animation: dashdraw 4.932820s linear infinite;}</style></svg>`, SVGInfo{10, 10, 4.93282}, false},
		{"no size", `<svg></svg>`, SVGInfo{}, true},
		{"not svg", `<html></html>`, SVGInfo{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InspectSVG([]byte(tt.svg))
			if (err != nil) != tt.err {
				t.Fatalf("err = %v, wantErr %v", err, tt.err)
			}
			if !tt.err && got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
