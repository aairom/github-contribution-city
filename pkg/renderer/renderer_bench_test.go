package renderer

import (
	"testing"
)

func BenchmarkRenderSkylineSVG(b *testing.B) {
	scene := createTestScene()
	r := NewSVGRenderer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RenderSkyline(scene, "github-dark")
	}
}

func BenchmarkRenderIsometric3DSVG(b *testing.B) {
	scene := createTestScene()
	r := NewSVGRenderer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RenderIsometric3D(scene, "cyberpunk")
	}
}
