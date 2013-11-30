package projects

import "errors"

var (
	ErrNoCachedFileList = errors.New("The cached file list did not exist")
)
