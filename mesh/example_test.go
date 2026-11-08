package mesh_test

import (
	"fmt"
	"strings"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
	"github.com/kyrokohan/glonk/mesh"
)

// printFrame prints the characters of fb, one line per row.
func printFrame(fb *engine.FrameBuffer) {
	for y := range fb.Height() {
		row := make([]rune, fb.Width())
		for x := range row {
			row[x] = fb.At(x, y).Character
		}
		fmt.Println(strings.TrimRight(string(row), " "))
	}
}

func ExampleLoadOBJ() {
	const square = `# a flat rectangle, 4 wide and 2 tall
v 0 0 0
v 4 0 0
v 4 2 0
v 0 2 0
f 1 2 3 4
`
	m, err := mesh.LoadOBJ(strings.NewReader(square))
	if err != nil {
		panic(err)
	}
	m = m.Fit(2) // centered on the origin, longest side 2

	lo, hi := m.Bounds()
	size := hi.Sub(lo)
	fmt.Printf("%d vertices, %d edges, %d triangles\n", len(m.Vertices), len(m.Edges), len(m.Faces))
	fmt.Printf("%.1f wide, %.1f tall\n", size.X, size.Y)
	fmt.Println(m.Validate())
	// Output:
	// 4 vertices, 4 edges, 2 triangles
	// 2.0 wide, 1.0 tall
	// <nil>
}

// A Camera left at zero sees 90 degrees from the origin, looking down +z.
func ExampleMesh_Draw() {
	fb, err := engine.NewFrameBuffer(30, 11)
	if err != nil {
		panic(err)
	}
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)

	cube := mesh.Cube(2)
	cube.Position = geometry.Vector3{Z: 3} // three units in front of the camera
	cube.Draw(fb, mesh.Camera{}, engine.Cell{Character: '#'})

	printFrame(fb)
	// Output:
	// ┌────────────────────────────┐
	// │                            │
	// │        ############        │
	// │        ###      ###        │
	// │        #  ######  #        │
	// │        #  #    #  #        │
	// │        #  ######  #        │
	// │        ###      ###        │
	// │        ############        │
	// │                            │
	// └────────────────────────────┘
}

func ExampleMesh_DrawShaded() {
	fb, err := engine.NewFrameBuffer(30, 11)
	if err != nil {
		panic(err)
	}
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)

	cube := mesh.Cube(2)
	cube.Position = geometry.Vector3{Z: 3}
	cube.Yaw = 0.6 // turn it so that two sides show
	light := mesh.Light{Direction: geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, Ambient: 0.2}
	cube.DrawShaded(fb, mesh.Camera{}, light, 208)

	printFrame(fb)
	// Output:
	// ┌────────────────────────────┐
	// │                            │
	// │              **            │
	// │        ********--          │
	// │        ********---         │
	// │        ********---         │
	// │        ********---         │
	// │        ********--          │
	// │              **            │
	// │                            │
	// └────────────────────────────┘
}

func ExampleMesh_Validate() {
	m := mesh.Mesh{
		Vertices: []geometry.Vector3{{}, {X: 1}, {Y: 1}},
		Faces:    [][3]int{{0, 1, 5}}, // there is no vertex 5
	}

	fmt.Println(m.Validate())
	// Output:
	// mesh: face 0 uses vertex 5, but only vertices 0 to 2 exist
}

func ExampleMesh_Clone() {
	a := mesh.Cube(2)
	b := a.Clone()
	b.Vertices[0].X = 99

	fmt.Println(a.Vertices[0].X, b.Vertices[0].X)
	// Output:
	// -1 99
}

func ExampleCamera_Project() {
	cam := mesh.Camera{Position: geometry.Vector3{Z: -5}}

	x, y, ok := cam.Project(geometry.Vector3{X: 1}, 80, 24)
	fmt.Printf("%.1f %.1f %v\n", x, y, ok)

	_, _, ok = cam.Project(geometry.Vector3{Z: -10}, 80, 24) // behind the camera
	fmt.Println(ok)
	// Output:
	// 44.8 12.0 true
	// false
}
