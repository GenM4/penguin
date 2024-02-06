package files

import (
	"log"
	"os"
	"path/filepath"
)

type FileData struct {
	SrcFilepath  string
	SrcFilename  string
	BaseFilepath string
	BaseFilename string
	AsmFilename  string
	AsmFilepath  string
	ObjFilename  string
	ObjFilepath  string
	ExecFilepath string
}

func GenerateFilepaths(args []string) FileData {

	var result FileData
	result.SrcFilepath = args[len(args)-1]
	result.SrcFilename = filepath.Base(result.SrcFilepath)
	result.BaseFilepath = filepath.Dir(result.SrcFilepath)
	result.BaseFilename = removeFileExtension(result.SrcFilename)
	result.AsmFilename = result.BaseFilename + ".asm"
	result.AsmFilepath = filepath.Join(result.BaseFilepath, result.AsmFilename)
	result.ObjFilename = result.BaseFilename + ".o"
	result.ObjFilepath = filepath.Join(result.BaseFilepath, result.ObjFilename)
	result.ExecFilepath = filepath.Join(result.BaseFilepath, result.BaseFilename)

	return result
}

func OpenTargetFile(filepath string) *os.File {
	asmFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	log.Println("Opened file: " + filepath)

	if err = os.Truncate(filepath, 0); err != nil {
		panic(err)
	}

	return asmFile
}

func removeFileExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}
