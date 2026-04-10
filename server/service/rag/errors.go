package rag

import "errors"

// ErrDocumentNoStoragePath is returned when a document has no file on disk (e.g. non-local storage).
var ErrDocumentNoStoragePath = errors.New("document has no storage path")
