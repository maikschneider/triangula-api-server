package logic

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/RH12503/Triangula-CLI/polygons"
	"github.com/RH12503/Triangula-CLI/util"

	"github.com/RH12503/Triangula-CLI/triangles"
	"github.com/RH12503/Triangula/normgeom"
	"github.com/fatih/color"
)

// decodePoints reads and decodes an JSON file containing the data of points.
func decodePoints(inputFile string) (normgeom.NormPointGroup, error) {
	jsonPoints, err := ioutil.ReadFile(inputFile)
	if err != nil {
		color.Red("error reading input file")
		return normgeom.NormPointGroup{}, err
	}
	var points normgeom.NormPointGroup
	err = json.Unmarshal(jsonPoints, &points)
	if err != nil {
		color.Red("error decoding input file")
		return normgeom.NormPointGroup{}, err
	}
	return points, nil
}

// RenderSVG renders a triangulation to a SVG.
func RenderSVG(inputFile, outputFile, imageFile, shape string) {
	color.Yellow("Reading image file...")

	img, err := util.DecodeImage(imageFile)

	if err != nil {
		return
	}

	color.Yellow("Reading input file...")
	points, err := decodePoints(inputFile)

	if err != nil {
		return
	}

	color.Yellow("Generating SVG...")
	filename := outputFile

	if !strings.HasSuffix(filename, ".svg") {
		filename += ".svg"
	}

	switch shape {
	case "triangles":
		err = triangles.WriteSVG(filename, points, img)
		break
	case "polygons":
		err = polygons.WriteSVG(filename, points, img)
		break
	default:
		err = errors.New("invalid shape type")
	}

	if err != nil {
		color.Red("error generating SVG")
		return
	}

	color.Green("Successfully generated SVG at %s", filename)
}
