// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"runtime"
	"time"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"

	"github.com/crawshaw/balloon/animation"
	"github.com/crawshaw/balloon/text"
)

var sheet struct {
	sheet  sprite.Texture
	skull1 sprite.SubTex
	skull2 sprite.SubTex
	skull3 sprite.SubTex

	swing1 sprite.SubTex
	swing2 sprite.SubTex
	swing3 sprite.SubTex
}

var (
	start time.Time
	eng   sprite.Engine
	font  *truetype.Font
)

var (
	scene     *sprite.Node
	gameScene *sprite.Node
	menuScene *sprite.Node
	overScene *sprite.Node
)

func timerInit() {
	start = time.Now()
	if err := loadSheet(); err != nil {
		log.Fatal(err)
	}

	var err error
	font, err = loadFont()
	if err != nil {
		panic(err)
	}

	menuSceneInit()
	gameSceneInit()
	overSceneInit()
	scene = menuScene
}

func menuSceneInit() {
	menuScene = new(sprite.Node)
	eng.Register(menuScene)

	addSkull := func(offsetX, size geom.Pt, subTex sprite.SubTex, duration int) {
		skull := &sprite.Node{
			Arranger: &animation.Arrangement{
				Offset: geom.Point{X: offsetX, Y: -size},
				Size:   &geom.Point{size, size},
				Pivot:  geom.Point{size / 2, size / 2},
				SubTex: subTex,
			},
		}
		eng.Register(skull)
		menuScene.AppendChild(skull)

		skullAnim := new(sprite.Node)
		eng.Register(skullAnim)
		menuScene.AppendChild(skullAnim)
		skullAnim.Arranger = &animation.Animation{
			Current: "init",
			States: map[string]animation.State{
				"init": animation.State{
					Duration: duration / 4,
					Next:     "falling",
				},
				"falling": animation.State{
					Duration: duration,
					Next:     "reset",
					Transforms: map[*sprite.Node]animation.Transform{
						skull: animation.Transform{
							Transformer: animation.Move{Y: geom.Height + size*2},
						},
					},
				},
				"reset": animation.State{
					Duration: 0,
					Next:     "falling",
					Transforms: map[*sprite.Node]animation.Transform{
						skull: animation.Transform{
							Transformer: animation.Move{Y: -geom.Height - size*2},
						},
					},
				},
			},
		}
	}

	addSkull(24, 36, sheet.skull1, 240)
	addSkull(48, 18, sheet.skull2, 100)
	addSkull(96, 36, sheet.skull3, 160)

	addText(menuScene, "Gopher Fall!", 20, geom.Point{24, 24})
	addText(menuScene, "Tap to start", 14, geom.Point{48, 48})
}

func overSceneInit() {
	overScene = new(sprite.Node)
	eng.Register(overScene)

	addText(overScene, "GAME OVER", 20, geom.Point{28, 28})
	addText(overScene, "Tap to play again", 14, geom.Point{32, 48})
}

func addText(parent *sprite.Node, str string, size geom.Pt, pos geom.Point) {
	p := &sprite.Node{
		Arranger: &animation.Arrangement{
			Offset: pos,
		},
	}
	eng.Register(p)
	parent.AppendChild(p)
	pText := &sprite.Node{
		Arranger: &text.String{
			Size:  size,
			Color: color.Black,
			Font:  font,
			Text:  str,
		},
	}
	eng.Register(pText)
	p.AppendChild(pText)
}

func gameSceneInit() {
	gameScene = new(sprite.Node)
	eng.Register(gameScene)

	game.scissor = newScissorArm2(eng)
	game.scissor.arrangement.Offset.Y = 2 * 72
	gameScene.AppendChild(game.scissor.node)

	n1 := new(sprite.Node)
	eng.Register(n1)
	n1.Arranger = &animation.Arrangement{
		Offset: geom.Point{X: 0, Y: geom.Height - 12 - 2},
	}
	gameScene.AppendChild(n1)

	t := new(sprite.Node)
	eng.Register(t)
	n1.AppendChild(t)
	game.scoreText = &text.String{
		Size:  12,
		Color: color.Black,
		Font:  font,
	}
	t.Arranger = game.scoreText

	updateGame(0)

	//Fprint(os.Stdout, gameScene, NotNilFilter)
}

func loadFont() (*truetype.Font, error) {
	font := ""
	switch runtime.GOOS {
	case "android":
		font = "/system/fonts/DroidSansMono.ttf"
	case "darwin":
		//font = "/Library/Fonts/Andale Mono.ttf"
		font = "/Library/Fonts/Arial.ttf"
		//font = "/Library/Fonts/儷宋 Pro.ttf"
	case "linux":
		font = "/usr/share/fonts/truetype/droid/DroidSansMono.ttf"
	default:
		return nil, fmt.Errorf("go.mobile/app/debug: unsupported runtime.GOOS %q", runtime.GOOS)
	}
	b, err := ioutil.ReadFile(font)
	if err != nil {
		return nil, err
	}
	return freetype.ParseFont(b)
}

func loadSheet() error {
	t, err := loadTexture("skull.png")
	if err != nil {
		return err
	}
	sheet.sheet = t

	sheet.skull1 = sprite.SubTex{t, image.Rect(0, 0, 24, 31)}
	sheet.skull2 = sprite.SubTex{t, image.Rect(24, 0, 48, 31)}
	sheet.skull3 = sprite.SubTex{t, image.Rect(48, 0, 72, 31)}

	t, err = loadTexture("swing_full.png")
	if err != nil {
		return err
	}
	sheet.swing1 = sprite.SubTex{t, image.Rect(0, 0, 24, 32)}
	sheet.swing2 = sprite.SubTex{t, image.Rect(0, 0, 48, 32)}
	sheet.swing3 = sprite.SubTex{t, image.Rect(0, 0, 72, 32)}
	return nil
}

func loadTexture(path string) (sprite.Texture, error) {
	f, err := app.Open(path)
	if err != nil {
		return nil, err
	}
	mb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	m, err := png.Decode(bytes.NewReader(mb))
	// any reason we can't skip the read all business??
	// m, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	t, err := eng.LoadTexture(m)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func touch(e event.Touch) {
	if e.Type != event.TouchStart {
		return
	}

	switch scene {
	case overScene:
		scene = menuScene
	case menuScene:
		startGame()
		scene = gameScene
	case gameScene:
		game.nextTouch = &e
	default:
		log.Printf("touch in unknown state %v", e)
	}
}

func now() clock.Time {
	d := time.Since(start)
	return clock.Time(60 * d / time.Second)
}
