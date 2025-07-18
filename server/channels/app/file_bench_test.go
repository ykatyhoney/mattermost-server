// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"math/rand"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"
)

var randomJPEG []byte
var randomGIF []byte
var zero10M = make([]byte, 10*1024*1024)
var rgba *image.RGBA

func prepareTestImages(tb testing.TB) {
	if rgba != nil {
		return
	}

	// Create a random image (pre-seeded for predictability)
	rgba = image.NewRGBA(image.Rectangle{
		image.Point{0, 0},
		image.Point{2048, 2048},
	})
	_, err := rand.New(rand.NewSource(1)).Read(rgba.Pix)
	if err != nil {
		tb.Fatal(err)
	}

	// Encode it as JPEG and GIF
	buf := &bytes.Buffer{}
	err = jpeg.Encode(buf, rgba, &jpeg.Options{Quality: 50})
	if err != nil {
		tb.Fatal(err)
	}
	randomJPEG = buf.Bytes()

	buf = &bytes.Buffer{}
	err = gif.Encode(buf, rgba, nil)
	if err != nil {
		tb.Fatal(err)
	}
	randomGIF = buf.Bytes()
}

func BenchmarkUploadFile(b *testing.B) {
	prepareTestImages(b)
	th := Setup(b).InitBasic()
	defer th.TearDown()

	teamID := model.NewId()
	channelID := model.NewId()
	userID := model.NewId()

	mb := func(i int) int {
		return (i + 512*1024) / (1024 * 1024)
	}

	files := []struct {
		title string
		ext   string
		data  []byte
	}{
		{fmt.Sprintf("random-%dMb-gif", mb(len(randomGIF))), ".gif", randomGIF},
		{fmt.Sprintf("random-%dMb-jpg", mb(len(randomJPEG))), ".jpg", randomJPEG},
		{fmt.Sprintf("zero-%dMb", mb(len(zero10M))), ".zero", zero10M},
	}

	fileBenchmarks := []struct {
		title string
		f     func(b *testing.B, n int, data []byte, ext string)
	}{
		{
			title: "raw-ish DoUploadFile",
			f: func(b *testing.B, n int, data []byte, ext string) {
				info1, appErr := th.App.DoUploadFile(th.Context, time.Now(), teamID, channelID,
					userID, fmt.Sprintf("BenchmarkDoUploadFile-%d%s", n, ext), data, true)
				require.Nil(b, appErr)
				err := th.App.Srv().Store().FileInfo().PermanentDelete(th.Context, info1.Id)
				require.NoError(b, err)
				appErr = th.App.RemoveFile(info1.Path)
				require.Nil(b, appErr)
			},
		},
		{
			title: "raw UploadFileX Content-Length",
			f: func(b *testing.B, n int, data []byte, ext string) {
				info, aerr := th.App.UploadFileX(th.Context, channelID,
					fmt.Sprintf("BenchmarkUploadFileTask-%d%s", n, ext),
					bytes.NewReader(data),
					UploadFileSetTeamId(teamID),
					UploadFileSetUserId(userID),
					UploadFileSetTimestamp(time.Now()),
					UploadFileSetContentLength(int64(len(data))),
					UploadFileSetRaw())
				if aerr != nil {
					b.Fatal(aerr)
				}
				err := th.App.Srv().Store().FileInfo().PermanentDelete(th.Context, info.Id)
				require.NoError(b, err)
				appErr := th.App.RemoveFile(info.Path)
				require.Nil(b, appErr)
			},
		},
		{
			title: "raw UploadFileX chunked",
			f: func(b *testing.B, n int, data []byte, ext string) {
				info, aerr := th.App.UploadFileX(th.Context, channelID,
					fmt.Sprintf("BenchmarkUploadFileTask-%d%s", n, ext),
					bytes.NewReader(data),
					UploadFileSetTeamId(teamID),
					UploadFileSetUserId(userID),
					UploadFileSetTimestamp(time.Now()),
					UploadFileSetContentLength(-1),
					UploadFileSetRaw())
				if aerr != nil {
					b.Fatal(aerr)
				}
				err := th.App.Srv().Store().FileInfo().PermanentDelete(th.Context, info.Id)
				require.NoError(b, err)
				appErr := th.App.RemoveFile(info.Path)
				require.Nil(b, appErr)
			},
		},
		{
			title: "image UploadFileX Content-Length",
			f: func(b *testing.B, n int, data []byte, ext string) {
				info, aerr := th.App.UploadFileX(th.Context, channelID,
					fmt.Sprintf("BenchmarkUploadFileTask-%d%s", n, ext),
					bytes.NewReader(data),
					UploadFileSetTeamId(teamID),
					UploadFileSetUserId(userID),
					UploadFileSetTimestamp(time.Now()),
					UploadFileSetContentLength(int64(len(data))))
				if aerr != nil {
					b.Fatal(aerr)
				}
				err := th.App.Srv().Store().FileInfo().PermanentDelete(th.Context, info.Id)
				require.NoError(b, err)
				appErr := th.App.RemoveFile(info.Path)
				require.Nil(b, appErr)
			},
		},
		{
			title: "image UploadFileX chunked",
			f: func(b *testing.B, n int, data []byte, ext string) {
				info, aerr := th.App.UploadFileX(th.Context, channelID,
					fmt.Sprintf("BenchmarkUploadFileTask-%d%s", n, ext),
					bytes.NewReader(data),
					UploadFileSetTeamId(teamID),
					UploadFileSetUserId(userID),
					UploadFileSetTimestamp(time.Now()),
					UploadFileSetContentLength(int64(len(data))))
				if aerr != nil {
					b.Fatal(aerr)
				}
				err := th.App.Srv().Store().FileInfo().PermanentDelete(th.Context, info.Id)
				require.NoError(b, err)
				appErr := th.App.RemoveFile(info.Path)
				require.Nil(b, appErr)
			},
		},
	}

	for _, file := range files {
		for _, fb := range fileBenchmarks {
			b.Run(file.title+"-"+fb.title, func(b *testing.B) {
				for i := 0; b.Loop(); i++ {
					fb.f(b, i, file.data, file.ext)
				}
			})
		}
	}
}
