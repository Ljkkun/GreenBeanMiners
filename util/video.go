package util

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"log"
	"os"
	"strings"
)

func GetFrame(videoPath, snapshotPath string, frameNum int) (snapshotName string, err error) {
	buf := bytes.NewBuffer(nil)
	err = ffmpeg.Input(videoPath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()

	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return "", err
	}

	img, err := imaging.Decode(buf)
	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return "", err
	}

	err = imaging.Save(img, snapshotPath)
	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return "", err
	}

	fmt.Println("--snapshotPath--", snapshotPath)

	names := strings.Split(snapshotPath, "\"")
	fmt.Println("----names----", names)

	// 这里把 snapshotPath 的 string 类型转换成 []string
	snapshotName = names[len(names)-1]
	fmt.Println("----snapshotName----", snapshotName)
	// ----snapshotName---- ./assets/testImage.png

	return snapshotName, nil
}
