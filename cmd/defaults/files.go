package defaults

var Files map[string][]byte

func addFile(path, content string) {
	if Files == nil {
		Files = make(map[string][]byte)
	}
	Files[path] = []byte(content)
}
