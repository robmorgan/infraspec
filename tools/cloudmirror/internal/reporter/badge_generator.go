package reporter

import (
	"fmt"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// BadgeGenerator generates SVG badges for coverage
type BadgeGenerator struct{}

// NewBadgeGenerator creates a new badge generator
func NewBadgeGenerator() *BadgeGenerator {
	return &BadgeGenerator{}
}

// GenerateBadge generates an SVG badge for a coverage report
func (g *BadgeGenerator) GenerateBadge(report *models.CoverageReport) ([]byte, error) {
	coverage := report.CoveragePercent
	color := g.getColor(coverage)
	label := fmt.Sprintf("%s coverage", report.ServiceName)
	value := fmt.Sprintf("%.1f%%", coverage)

	svg := g.generateSVG(label, value, color)
	return []byte(svg), nil
}

// GenerateSimpleBadge generates a simple coverage badge
func (g *BadgeGenerator) GenerateSimpleBadge(serviceName string, coverage float64) ([]byte, error) {
	color := g.getColor(coverage)
	label := serviceName
	value := fmt.Sprintf("%.1f%%", coverage)

	svg := g.generateSVG(label, value, color)
	return []byte(svg), nil
}

func (g *BadgeGenerator) getColor(coverage float64) string {
	switch {
	case coverage >= 80:
		return "#4c1" // bright green
	case coverage >= 60:
		return "#a3c51c" // yellow-green
	case coverage >= 40:
		return "#dfb317" // yellow
	case coverage >= 20:
		return "#fe7d37" // orange
	default:
		return "#e05d44" // red
	}
}

func (g *BadgeGenerator) generateSVG(label, value, color string) string {
	// Calculate widths based on text length
	labelWidth := len(label)*6 + 10
	valueWidth := len(value)*6 + 10
	totalWidth := labelWidth + valueWidth

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
  <title>%s: %s</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r">
    <rect width="%d" height="20" rx="3" fill="#fff"/>
  </clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)">%s</text>
    <text x="%d" y="140" transform="scale(.1)">%s</text>
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)">%s</text>
    <text x="%d" y="140" transform="scale(.1)">%s</text>
  </g>
</svg>`,
		totalWidth,
		label, value,
		label, value,
		totalWidth,
		labelWidth,
		labelWidth, valueWidth, color,
		totalWidth,
		labelWidth*5, label,
		labelWidth*5, label,
		labelWidth*10+valueWidth*5, value,
		labelWidth*10+valueWidth*5, value,
	)
}

// GenerateOverallBadge generates a badge for overall coverage across services
func (g *BadgeGenerator) GenerateOverallBadge(multi *models.MultiServiceReport) ([]byte, error) {
	color := g.getColor(multi.OverallCoverage)
	label := "AWS parity"
	value := fmt.Sprintf("%.1f%%", multi.OverallCoverage)

	svg := g.generateSVG(label, value, color)
	return []byte(svg), nil
}
