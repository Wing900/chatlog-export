package model

import "path/filepath"

type MediaV4 struct {
	Type       string `json:"type"`
	Key        string `json:"key"`
	Dir1       string `json:"dir1"`
	Dir2       string `json:"dir2"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifyTime int64  `json:"modifyTime"`
}

func (m *MediaV4) Wrap() *Media {

	var path string
	switch m.Type {
	case "image":
		path = filepath.Join("msg", "attach", m.Dir1, m.Dir2, "Img", m.Name)
	case "video":
		path = filepath.Join("msg", "video", m.Dir1, m.Name)
	case "file":
		path = filepath.Join("msg", "file", m.Dir1, m.Name)
	}

	return &Media{
		Type:       m.Type,
		Key:        m.Key,
		Path:       path,
		Name:       m.Name,
		Size:       m.Size,
		ModifyTime: m.ModifyTime,
	}
}
