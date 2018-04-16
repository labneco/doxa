// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package b64image

import (
	"encoding/base64"
	"image"
	"image/color"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestColorize(t *testing.T) {
	Convey("Testing Colorize function", t, func() {
		imgData, _ := ioutil.ReadFile("testdata/avatar.png")
		imgString := base64.StdEncoding.EncodeToString(imgData)
		Convey("Applying a fully opaque color", func() {
			clr := color.RGBA{R: 32, G: 224, B: 224, A: 255}
			dstImageString := Colorize(imgString, clr)
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(dstImageString))
			destImg, _, _ := image.Decode(reader)
			Convey("Result image should not be the original one", func() {
				So(dstImageString, ShouldNotEqual, imgString)
			})
			Convey("The target image should have the same dimensions", func() {
				So(destImg.Bounds().Dx(), ShouldEqual, 180)
				So(destImg.Bounds().Dy(), ShouldEqual, 180)
			})
			Convey("The color at 2,2 should be the given color", func() {
				So(ColorsEqual(destImg.At(2, 2), clr), ShouldBeTrue)
			})
			Convey("The color at 90,90 should be the original color", func() {
				So(ColorsEqual(destImg.At(90, 90), color.RGBA{R: 217, G: 222, B: 226, A: 255}), ShouldBeTrue)
			})
		})
		Convey("Unreadable image should be returned as is", func() {
			clr := color.RGBA{R: 32, G: 224, B: 224, A: 255}
			dstImageString := Colorize("foo bar", clr)
			So(dstImageString, ShouldEqual, "foo bar")
		})
		Convey("Testing random color", func() {
			dstImageString := Colorize(imgString, color.RGBA{})
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(dstImageString))
			destImg, _, _ := image.Decode(reader)
			Convey("Result image should not be the original one", func() {
				So(dstImageString, ShouldNotEqual, imgString)
			})
			Convey("The target image should have the same dimensions", func() {
				So(destImg.Bounds().Dx(), ShouldEqual, 180)
				So(destImg.Bounds().Dy(), ShouldEqual, 180)
			})
			Convey("The color at 2,2 should be the same as 4,4 and not the empty color", func() {
				So(ColorsEqual(destImg.At(2, 2), destImg.At(4, 4)), ShouldBeTrue)
				So(ColorsEqual(destImg.At(2, 2), color.RGBA{}), ShouldBeFalse)
			})
			Convey("The color at 90,90 should be the original color", func() {
				So(ColorsEqual(destImg.At(90, 90), color.RGBA{R: 217, G: 222, B: 226, A: 255}), ShouldBeTrue)
			})

		})
	})
}
