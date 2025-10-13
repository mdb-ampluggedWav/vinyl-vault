package services

type FileUploadResult struct {
	Path, Filename string
	Size           uint64
}

type FileValidationError struct {
	Field, Message string
}

type FileService struct {
	uploadDir string
}

func NewFileService(uploadDir string) *FileService {
	return &FileService{uploadDir: uploadDir}
}
