// Package sprite draws sprites onto an engine.FrameBuffer in layers.
//
// A [Sprite] is one [engine.Cell] at a position: a character with its colors and
// attributes. Positions are in cells, with the origin at the top left, x to the right
// and y down; a fraction is floored. A sprite replaces the whole cell under it,
// background included, as [engine.FrameBuffer.Set] does.
//
// A [Layer] draws its sprites in the order they were added, so later ones cover
// earlier ones. Add, Remove and Clear change what is in it, and Len counts.
//
// A sprite at a NaN or infinite position, or off screen, is not drawn, and a nil
// sprite added to a Layer is ignored.
package sprite
