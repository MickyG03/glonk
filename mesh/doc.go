// Package mesh draws 3D models through a perspective Camera onto an
// engine.FrameBuffer, as wireframes or as solid, lit surfaces.
//
// A [Mesh] holds vertices, the edges and triangular faces between them, and where
// it sits in the world. [Cube], [Grid] and [LoadOBJ] make meshes; [Mesh.Draw] draws
// the edges, and [Mesh.DrawShaded] and [Mesh.DrawShadedPixels] draw the faces lit
// by a [Light], into cells or into a half-block pixel Canvas.
//
// Coordinates: x points right, y up and z away from the viewer, so a Camera at the
// origin with Yaw and Pitch 0 looks down +z. Angles are in radians. OBJ files are
// right-handed and look down -z; LoadOBJ converts them.
//
// Drawing does not panic on a bad mesh or camera. An edge or face that names a
// vertex the mesh does not have is skipped, and [Mesh.Validate] reports it. A
// [Camera] with a nonsense FOV or Near sees nothing, and one left at zero uses a
// default.
package mesh
