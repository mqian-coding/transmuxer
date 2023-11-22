package host

import (
	"concurrency-practice/internal/store"
	"path"
)

func GetManifestPath(filename string) string {
	if store.IsServerInitialized() {
		return path.Join(store.TheServer.StaticDir, "host", filename)
	}
	return ""
}

func GetPlaylistPath(filename string) string {
	if store.IsServerInitialized() {
		return path.Join(GetManifestPath(filename), "playlist.m3u8")
	}
	return ""
}

func GetChunklistPath(filename string) string {
	if store.IsServerInitialized() {
		return path.Join(GetManifestPath(filename), "chunklist.m3u8")
	}
	return ""
}

func GetSegmentsDir(filename string) string {
	if store.IsServerInitialized() {
		return path.Join(store.TheServer.StaticDir, "host", filename, "segments")
	}
	return ""
}

func GetSegmentsPath(filename, segmentName string) string {
	if store.IsServerInitialized() {
		return path.Join(GetSegmentsDir(filename), segmentName+".ts")
	}
	return ""
}
