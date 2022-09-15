package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type DeComp interface {
	Tra_Dir(pat string) ([]string, error)
	Sele_Path(pat []string) error
}

type Job_Dir struct {
	Path       string
	ChoosePath string
}

func (Dir Job_Dir) Tra_Dir(pat string) ([]string, error) { //递归目录
	files, err := ioutil.ReadDir(pat)
	if err != nil {
		log.Fatalln(err)
	}
	file_list := make([]string, 10)
	for _, f := range files {
		if f.IsDir() {
			fpath := pat + "\\" + f.Name()
			L, _ := Dir.Tra_Dir(fpath)
			file_list = append(file_list, L...)
		} else {
			fpath := pat + "\\" + f.Name()
			file_list = append(file_list, fpath)
		}

	}
	return file_list, nil
}

func (Dir Job_Dir) Sele_Path(pat []string) error {
	for _, element := range pat {
		fp := path.Base(filepath.ToSlash(element)) //文件名
		ext := path.Ext(fp)                        //文件后缀
		//name := strings.TrimSuffix(fp, ext)        //解压的目录名
		workdir := strings.TrimSuffix(element, fp)
		if ext == ".rar" {
			er := os.Chdir(workdir) //更改工作目录
			if er != nil {
				return er
			}
			cmd := exec.Command("unrar", "x", "-ad", "-o+", element) //windows命令解压rar
			err := cmd.Run()
			if err != nil {
				return err
			}

		}
		if ext == ".zip" {
			err := Un_zip(element, workdir)
			if err != nil {
				fmt.Println(workdir, "下的zip文件已被损坏")
				log.Fatalln(err)
			}
		}
	}

	return nil

}

func Un_zip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	var decodeName string
	for _, f := range zipReader.File {
		if f.Flags == 0 {
			//如果标致位是0  则是默认的本地编码   默认为gbk
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := ioutil.ReadAll(decoder)
			decodeName = string(content)
		} else {
			//如果标志为是 1 << 11也就是 2048  则是utf-8编码
			decodeName = f.Name
		}
		fpath := filepath.Join(destDir, decodeName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func mvfile(filelist []string, destdir string) error {
	err := os.Mkdir(destdir, os.ModePerm)
	if err != nil {
		return err
	}
	for _, ele := range filelist {
		if strings.Contains(ele, "网络设备") && strings.Contains(ele, ".xlsx") {
			ind := strings.LastIndex(ele, "\\")
			fn := ele[ind:]
			err = os.Rename(ele, destdir+"\\"+fn)
		}
	}

	return nil
}

func main() {
	work_dir, _ := os.Getwd()

	F := Job_Dir{
		Path:       work_dir,
		ChoosePath: "网络设备",
	}
	var dec DeComp
	dec = F
	fl, err := dec.Tra_Dir(F.Path)
	if err != nil {
		log.Fatalln(err)
	}
	err = dec.Sele_Path(fl)
	if err != nil {
		log.Fatalln()
	}
	fl2, err := dec.Tra_Dir(work_dir)
	err = mvfile(fl2, work_dir+"\\"+"抽表")
	if err != nil {
		log.Fatalln(err)
	}
}
