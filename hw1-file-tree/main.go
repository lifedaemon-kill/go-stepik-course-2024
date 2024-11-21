package main

import (
	"io"
	"os"
	"sort"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := getTree(out, path, printFiles, "")
	return err
}

func getTree(out io.Writer, path string, printFiles bool, prefix string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if !printFiles {
		n := 0
		for _, entry := range entries {
			if entry.IsDir() {
				entries[n] = entry
				n++
			}
		}
		entries = entries[:n]
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for i := 0; i < len(entries)-1; i++ {
		if entries[i].IsDir() {
			if _, err := out.Write([]byte(prefix + "├───" + entries[i].Name() + "\n")); err != nil {
				return err
			}

			if err = getTree(out, path+string(os.PathSeparator)+entries[i].Name(), printFiles, prefix+"│\t"); err != nil {
				return err
			}

		} else if printFiles {
			fileInfo, err := os.Stat(path + string(os.PathSeparator) + entries[i].Name())
			if err != nil {
				return err
			}

			fileSize := fileInfo.Size()
			var strFileSize string
			if fileSize == 0 {
				strFileSize = "empty"
			} else {
				strFileSize = strconv.FormatInt(fileSize, 10) + "b"
			}

			_, err = out.Write([]byte(prefix + "├───" + entries[i].Name() + " (" + strFileSize + ")" + "\n"))
			if err != nil {
				return err
			}
		}
	}
	last := len(entries) - 1
	if last < 0 {
		return nil
	}
	if entries[last].IsDir() {
		_, err := out.Write([]byte(prefix + "└───" + entries[last].Name() + "\n"))
		if err != nil {
			return err
		}
		if err = getTree(out, path+string(os.PathSeparator)+entries[last].Name(), printFiles, prefix+"\t"); err != nil {
			return err
		}
	} else if printFiles {
		fileInfo, err := os.Stat(path + string(os.PathSeparator) + entries[last].Name())
		if err != nil {
			return err
		}

		fileSize := fileInfo.Size()
		var strFileSize string
		if fileSize == 0 {
			strFileSize = "empty"
		} else {
			strFileSize = strconv.FormatInt(fileSize, 10) + "b"
		}

		_, err = out.Write([]byte(prefix + "└───" + entries[last].Name() + " (" + strFileSize + ")" + "\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

/*
├───
└───
│
*/
