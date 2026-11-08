// Package sprite draws single-cell sprites onto an engine.FrameBuffer in layers.
//
// A sprite is one character in the default colors. Positions are in cells, with
// the origin at the top left, x to the right and y down; a fraction is floored.
//
// A sprite at a NaN or infinite position, or off screen, is not drawn, and a nil
// sprite added to a Layer is ignored.
package sprite
