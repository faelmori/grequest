package grequest

import (
	"io"
	"os"
	"path/filepath"
)

func getFileExtensionByContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "application/zip":
		return ".zip"
	case "application/x-rar-compressed":
		return ".rar"
	case "application/json":
		return ".json"
	case "application/xml":
		return ".xml"
	case "text/html":
		return ".html"
	case "text/plain":
		return ".txt"
	case "text/css":
		return ".css"
	case "text/javascript":
		return ".js"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/x-msvideo":
		return ".avi"
	case "video/mpeg":
		return ".mpeg"
	default:
		return ""
	}
}

func readFileByPath(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, err
}

func getFileNameByPath(path string) string {
	_, fileName := filepath.Split(path)
	return fileName
}

func saveToFile(fileName string, src io.Reader) error {
	_ = os.MkdirAll(filepath.Dir(fileName), 0777)
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return err
	}
	return nil
}
