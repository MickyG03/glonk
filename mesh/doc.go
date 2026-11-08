// Package mesh draws wireframe 3D meshes through a perspective Camera onto an
// engine.FrameBuffer.
//
// Drawing does not panic on a bad mesh or camera. An edge or face that names a
// vertex the mesh does not have is skipped, and Mesh.Validate reports it. A Camera
// with a nonsense FOV or Near sees nothing, and one left at zero uses a default (see
// Camera).
package mesh
