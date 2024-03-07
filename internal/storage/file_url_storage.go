package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type FileURLStorage struct {
	*MemoryURLStorage
	filepath string
}

var _ URLStorage = (*FileURLStorage)(nil)

func (f *FileURLStorage) AddURL(ctx context.Context, userID int, id string, url string) error {
	err := f.MemoryURLStorage.AddURL(ctx, userID, id, url)
	if err != nil {
		return err
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	event := Event{
		UUID:        newUUID,
		UserID:      userID,
		ShortURL:    id,
		OriginalURL: url,
	}

	file, err := os.OpenFile(f.filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	if err2 := json.NewEncoder(file).Encode(&event); err2 != nil {
		err3 := f.MemoryURLStorage.DeleteURL(ctx, userID, id)
		if err3 != nil {
			return err3
		}
		return err2
	}

	return nil
}
