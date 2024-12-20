package utils

import (
	"image"
	"image/draw"

	"github.com/nfnt/resize"
)

func AvatarCompose(bg image.Image, avatars []image.Image) image.Image {
	dest := image.NewRGBA(bg.Bounds())
	draw.Draw(dest, image.Rect(0, 0,
		bg.Bounds().Max.X, bg.Bounds().Max.Y),
		bg, image.Pt(0, 0), draw.Over)
	//如果是两个头像到4个头像那么分行
	lsize := 1
	if len(avatars) >= 2 && len(avatars) <= 4 {
		lsize = 2
	} else if len(avatars) > 4 {
		//如果是两个头像到大于4个的分三行
		lsize = 3
	}
	//如果只有一个头像，则直接居中显示
	if len(avatars) == 1 {
		splitSize := bg.Bounds().Max.X / 2
		img := resize.Resize(uint(splitSize), 0, avatars[0], resize.Lanczos3)
		lx := splitSize / 2
		ly := splitSize / 2
		rx := splitSize / 2 * 3
		ry := splitSize / 2 * 3

		draw.Draw(dest, image.Rect(lx, ly,
			rx, ry),
			img, image.Pt(0, 0), draw.Over)
		return dest
	}
	//计算一下单个头像的大小
	splitSize := bg.Bounds().Max.X / lsize

	var lx, ly int
	var rx, ry int
	for i, img := range avatars {
		img = resize.Resize(uint(splitSize), 0, img, resize.Lanczos3)
		col := i / lsize
		row := i % lsize

		lx = row * splitSize
		ly = col * splitSize
		rx = (row + 1) * splitSize
		ry = (col + 1) * splitSize
		draw.Draw(dest, image.Rect(lx, ly,
			rx, ry),
			img, image.Pt(0, 0), draw.Over)
	}
	return dest
}
