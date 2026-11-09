// Package mesh draws 3D models through a perspective Camera onto an
// engine.FrameBuffer, as wireframes or as solid, lit surfaces.
//
// A [Mesh] holds vertices, the edges and triangular faces between them, where it
// sits in the world (its [Transform]) and how it looks (its [Material]). [Cube],
// [Sphere], [Cylinder], [Cone], [Plane], [Grid] and [LoadOBJ] make meshes;
// [Mesh.Draw] draws the edges, and
// [Mesh.DrawShaded] and [Mesh.DrawShadedPixels] draw the faces lit by any number of
// lights, each a [Light] such as [Directional] or [Ambient] or one of your own, into
// cells or into a half-block pixel Canvas.
//
// Yaw and Pitch turn a model the way they turn a [Camera]: positive Yaw swings the
// front toward +x and positive Pitch tips it up.
//
// Coordinates: x points right, y up and z away from the viewer, so a Camera at the
// origin with Yaw and Pitch 0 looks down +z. Angles are in radians. OBJ files are
// right-handed and look down -z; LoadOBJ converts them.
//
// Drawing does not panic on a bad mesh, transform or camera. An edge or face that
// names a vertex the mesh does not have is skipped, and [Mesh.Validate] reports it, as
// it does a [Transform] that cannot be applied, which draws nothing. A [Camera] with
// a nonsense FOV, Near or CellAspect sees nothing, and one left at zero uses a
// default.
package mesh
