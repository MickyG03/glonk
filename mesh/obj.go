package mesh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/kyrokohan/glonk/geometry"
)

// maxOBJLine is the longest logical line LoadOBJ accepts, in bytes. A polygon with
// tens of thousands of corners still fits; anything longer is more likely a
// corrupt or hostile file than a model.
const maxOBJLine = 1 << 20

// LoadOBJ reads a Wavefront OBJ model, the plain-text format nearly every 3D tool
// can export, and returns it as a Mesh ready to draw.
//
// It understands vertices (v), faces (f, in the v, v/vt, v//vn and v/vt/vn forms)
// and line elements (l). Texture coordinates, normals, materials, groups and
// anything else are skipped, and all objects in the file are merged into one Mesh.
// Indices start at 1, and a negative index counts back from the vertices read so
// far. A face with more than three corners is cut into triangles, which are used
// for shading, while Edges follows the face's outline so a wireframe shows the
// original polygons, with shared edges listed once.
//
// OBJ files are right-handed and look down -Z, while this package looks down +Z.
// LoadOBJ converts: it negates every Z and reverses each face's corner order, so
// the outside of a model stays its outside. A model that looks inside out was
// wound the other way in the file; set Mesh.DoubleSided to draw it anyway.
//
// The input is not trusted. Malformed numbers, bad indices and the like return an
// error naming the line, and a file with no faces and no lines is an error.
func LoadOBJ(r io.Reader) (Mesh, error) {
	var (
		m          Mesh
		seen       = map[[2]int]bool{}
		sc         = bufio.NewScanner(r)
		lineNo     int             // the line being read
		startLine  int             // the line a continued statement began on
		continued  strings.Builder // the text of a statement carried over from lines ending in \
		hasPending bool
	)
	sc.Buffer(make([]byte, 0, 64*1024), maxOBJLine)

	addEdge := func(a, b int) {
		if a == b {
			return
		}
		if a > b {
			a, b = b, a
		}
		if key := [2]int{a, b}; !seen[key] {
			seen[key] = true
			m.Edges = append(m.Edges, key)
		}
	}

	statement := func(line int, text string) error {
		if i := strings.IndexByte(text, '#'); i >= 0 {
			text = text[:i]
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			return nil
		}

		switch fields[0] {
		case "v":
			v, err := parseOBJVertex(fields[1:])
			if err != nil {
				return fmt.Errorf("obj line %d: %w", line, err)
			}
			m.Vertices = append(m.Vertices, v)

		case "f":
			corners, err := parseOBJCorners(fields[1:], len(m.Vertices))
			if err != nil {
				return fmt.Errorf("obj line %d: %w", line, err)
			}
			if len(corners) < 3 {
				return fmt.Errorf("obj line %d: a face needs at least 3 corners, got %d", line, len(corners))
			}
			// Fan from the first corner. Each triangle is listed backwards because
			// the Z flip mirrors the model.
			for i := 1; i+1 < len(corners); i++ {
				m.Faces = append(m.Faces, [3]int{corners[0], corners[i+1], corners[i]})
			}
			for i, c := range corners {
				addEdge(c, corners[(i+1)%len(corners)])
			}

		case "l":
			corners, err := parseOBJCorners(fields[1:], len(m.Vertices))
			if err != nil {
				return fmt.Errorf("obj line %d: %w", line, err)
			}
			if len(corners) < 2 {
				return fmt.Errorf("obj line %d: a line needs at least 2 points, got %d", line, len(corners))
			}
			for i := 0; i+1 < len(corners); i++ {
				addEdge(corners[i], corners[i+1])
			}
		}
		return nil
	}

	for sc.Scan() {
		lineNo++
		text := sc.Text()

		if !hasPending {
			startLine = lineNo
		}
		// Pieces go into a Builder so each is copied once. Gluing them with string
		// concatenation copied everything so far for every piece, which made a long
		// run of continued lines take time proportional to its square.
		if strings.HasSuffix(text, "\\") {
			continued.WriteString(strings.TrimSuffix(text, "\\"))
			continued.WriteByte(' ')
			hasPending = true
			if continued.Len() > maxOBJLine {
				return Mesh{}, fmt.Errorf("obj line %d: %w", startLine, bufio.ErrTooLong)
			}
			continue
		}

		continued.WriteString(text)
		err := statement(startLine, continued.String())
		continued.Reset()
		hasPending = false
		if err != nil {
			return Mesh{}, err
		}
	}
	if err := sc.Err(); err != nil {
		return Mesh{}, fmt.Errorf("obj line %d: %w", lineNo+1, err)
	}
	if hasPending { // the file ended on a backslash
		if err := statement(startLine, continued.String()); err != nil {
			return Mesh{}, err
		}
	}

	if len(m.Faces) == 0 && len(m.Edges) == 0 {
		return Mesh{}, errors.New("obj: no faces or lines")
	}
	return m, nil
}

// parseOBJVertex reads the coordinates of a v statement. Anything after the first
// three numbers (an optional w, or vertex colors some exporters add) is ignored.
// Z is negated to convert from OBJ's right-handed space.
func parseOBJVertex(args []string) (geometry.Vector3, error) {
	if len(args) < 3 {
		return geometry.Vector3{}, fmt.Errorf("a vertex needs 3 numbers, got %d", len(args))
	}

	var c [3]float64
	for i := range c {
		f, err := strconv.ParseFloat(args[i], 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return geometry.Vector3{}, fmt.Errorf("bad number %q", args[i])
		}
		c[i] = f
	}
	return geometry.Vector3{X: c[0], Y: c[1], Z: -c[2]}, nil
}

// parseOBJCorners turns the corner tokens of an f or l statement ("7", "7/2",
// "7//3", "7/2/3") into 0-based vertex indices. have is how many vertices have
// been read so far, which is what negative indices count back from.
func parseOBJCorners(args []string, have int) ([]int, error) {
	corners := make([]int, 0, len(args))

	for _, tok := range args {
		vertex, _, _ := strings.Cut(tok, "/")
		if vertex == "" {
			return nil, fmt.Errorf("missing vertex index in %q", tok)
		}
		n, err := strconv.Atoi(vertex)
		if err != nil {
			return nil, fmt.Errorf("bad vertex index %q", vertex)
		}

		i := n - 1
		if n < 0 {
			i = have + n
		}
		if i < 0 || i >= have { // index 0 lands on i = -1
			return nil, fmt.Errorf("vertex index %d out of range (%d vertices so far)", n, have)
		}
		corners = append(corners, i)
	}
	return corners, nil
}
