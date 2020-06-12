package http

import (
	"github.com/disintegration/imaging"
	"github.com/filebrowser/filebrowser/v2/files"
	"net/http"
	"net/url"
)

var thumbnailHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:      d.user.Fs,
		Path:    r.URL.Path,
		Modify:  d.user.Perm.Modify,
		Expand:  true,
		Checker: d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir || file.Type != "image" {
		return http.StatusNotFound, nil
	}

	return thumbnailFileHandler(w, r, file)
})

func thumbnailFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer fd.Close()

	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		// As per RFC6266 section 4.3
		w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(file.Name))
	}

	srcImg, err := imaging.Decode(fd, imaging.AutoOrientation(true))
	if err != nil {
		return errToStatus(err), err
	}
	format, err := imaging.FormatFromExtension(file.Extension)
	if err != nil {
		return http.StatusNotFound, err
	}
	dstImage := imaging.Resize(srcImg, srcImg.Bounds().Dx(), 0, imaging.NearestNeighbor)
	imaging.Encode(w, dstImage, format)
	return 0, nil
}
