package animation

import (
	"fmt"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"
)

// Arrangement is a sprite Arranger that uses high-level concepts to
// transform a sprite Node.
type Arrangement struct {
	Offset     geom.Point     // distance between parent and pivot
	Pivot      geom.Point     // point on sized, unrotated node
	Size       *geom.Point    // optional bounding rectangle for scaling
	Rotation   float32        // radians counter-clockwise
	Texture    sprite.Texture // optional Node Texture
	Transforms []Transform    // active transformations to apply on Arrange

	// TODO: Physics *physics.Physics
}

func (ar *Arrangement) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	ar2 := *ar
	for _, a := range ar.Transforms {
		tween := a.Tween(a.T0, a.T1, t)
		a.Transformer.Transform(&ar2, tween)
	}
	e.SetTexture(n, t, ar2.Texture)
	e.SetTransform(n, t, ar2.Affine())

	//ar.squash(t)
}

// Squash plays through transformas and physics, updating the Arrangement
// and removing any outdated transforms.
//
// TODO: automatically do this? export? if automatic, merge into Arrange.
func (ar *Arrangement) Squash(t clock.Time) {
	remove := 0
	for _, a := range ar.Transforms {
		if t < a.T1 {
			// stop squashing at the first animation that cannot be squashed.
			// animations are not commutative.
			break
		}
		a.Transformer.Transform(ar, 1)
		fmt.Printf("squash: %+v\n", ar)
		remove++
	}
	ar.Transforms = ar.Transforms[remove:]
}

func (ar *Arrangement) Affine() f32.Affine {
	var a f32.Affine
	a.Identity()
	if ar == nil {
		return a
	}
	a.Translate(&a, float32(ar.Offset.X), float32(ar.Offset.Y))
	if ar.Rotation != 0 {
		a.Rotate(&a, ar.Rotation)
	}
	x, y := float32(ar.Pivot.X), float32(ar.Pivot.Y)
	a.Translate(&a, -x, -y)

	if ar.Size != nil {
		a.Scale(&a, float32(ar.Size.X), float32(ar.Size.Y))
	}
	return a
}

type Transform struct {
	T0, T1      clock.Time
	Tween       func(t0, t1, t clock.Time) float32
	Transformer Transformer
}

type Transformer interface {
	Transform(ar *Arrangement, tween float32)
}

// Rotate rotates counter-clockwise, measured in radians.
type Rotate float32

func (r Rotate) Transform(ar *Arrangement, tween float32) {
	ar.Rotation += tween * float32(r)
}

// Move moves the Arrangement offset.
type Move geom.Point

func (m Move) Transform(ar *Arrangement, tween float32) {
	ar.Offset.X += m.X * geom.Pt(tween)
	ar.Offset.Y += m.Y * geom.Pt(tween)
}

func (m Move) String() string {
	return fmt.Sprintf("Move(%s,%s)", m.X, m.Y)
}
